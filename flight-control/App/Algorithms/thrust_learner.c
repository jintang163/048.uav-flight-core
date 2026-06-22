#include "thrust_learner.h"
#include "flight_controller.h"
#include "sensor_manager.h"
#include "pid_controller.h"
#include "task_attitude_estimation.h"
#include <string.h>
#include <math.h>

#define GRAVITY 9.81f
#define TAKEOFF_CLIMB_RATE_THRESHOLD 0.5f
#define HOVER_VELOCITY_THRESHOLD 0.2f
#define HOVER_STABLE_DURATION_MS 3000
#define WEIGHT_ESTIMATION_DURATION_MS 2000
#define MIN_SAMPLES_FOR_WEIGHT 10
#define THRUST_CURVE_POINT_COUNT 16
#define PWM_TO_RPM_SCALE 10.0f

static LearningState learner_state = LS_IDLE;
static WeightEstimateState weight_state = {0};
static ThrustSample samples[THRUST_LEARNER_MAX_SAMPLES];
static uint32_t sample_count = 0;
static ThrustCurvePoint thrust_curve[THRUST_CURVE_POINT_COUNT];
static PIDGainSet current_gains = {0};
static PIDGainSet default_gains = {0};
static uint32_t hover_stable_start = 0;
static float last_vertical_velocity = 0.0f;
static float thrust_curve_linearity = 1.0f;

static void load_default_pid_gains(void)
{
    default_gains.roll_p = PID_ANGLE_ROLL_P;
    default_gains.roll_i = PID_ANGLE_ROLL_I;
    default_gains.roll_d = PID_ANGLE_ROLL_D;
    default_gains.rate_roll_p = PID_RATE_ROLL_P;
    default_gains.rate_roll_i = PID_RATE_ROLL_I;
    default_gains.rate_roll_d = PID_RATE_ROLL_D;

    default_gains.pitch_p = PID_ANGLE_PITCH_P;
    default_gains.pitch_i = PID_ANGLE_PITCH_I;
    default_gains.pitch_d = PID_ANGLE_PITCH_D;
    default_gains.rate_pitch_p = PID_RATE_PITCH_P;
    default_gains.rate_pitch_i = PID_RATE_PITCH_I;
    default_gains.rate_pitch_d = PID_RATE_PITCH_D;

    default_gains.yaw_p = PID_ANGLE_YAW_P;
    default_gains.yaw_i = PID_ANGLE_YAW_I;
    default_gains.yaw_d = PID_ANGLE_YAW_D;
    default_gains.rate_yaw_p = PID_RATE_YAW_P;
    default_gains.rate_yaw_i = PID_RATE_YAW_I;
    default_gains.rate_yaw_d = PID_RATE_YAW_D;

    default_gains.alt_p = PID_ALT_P;
    default_gains.alt_i = PID_ALT_I;
    default_gains.alt_d = PID_ALT_D;

    current_gains = default_gains;
}

static float clamp_gain(float gain, float default_val)
{
    float limit = THRUST_LEARNER_GAIN_LIMIT_PCT / 100.0f;
    float min_val = default_val * (1.0f - limit);
    float max_val = default_val * (1.0f + limit);
    return CONSTRAIN(gain, min_val, max_val);
}

static float smooth_gain_update(float target, float current)
{
    float max_delta = fabsf(current) * THRUST_LEARNER_PID_ADJUST_RATE;
    float delta = target - current;
    if (fabsf(delta) > max_delta) {
        delta = (delta > 0.0f) ? max_delta : -max_delta;
    }
    return current + delta;
}

static void update_pid_gains_based_on_weight(void)
{
    float weight_ratio = weight_state.estimated_weight_kg / THRUST_LEARNER_DEFAULT_WEIGHT_KG;
    weight_ratio = CONSTRAIN(weight_ratio, 0.5f, 2.0f);

    float target_roll_p = default_gains.roll_p * weight_ratio;
    float target_roll_i = default_gains.roll_i * weight_ratio;
    float target_roll_d = default_gains.roll_d * weight_ratio;
    float target_rate_roll_p = default_gains.rate_roll_p * weight_ratio;
    float target_rate_roll_i = default_gains.rate_roll_i * weight_ratio;
    float target_rate_roll_d = default_gains.rate_roll_d * weight_ratio * thrust_curve_linearity;

    float target_pitch_p = default_gains.pitch_p * weight_ratio;
    float target_pitch_i = default_gains.pitch_i * weight_ratio;
    float target_pitch_d = default_gains.pitch_d * weight_ratio;
    float target_rate_pitch_p = default_gains.rate_pitch_p * weight_ratio;
    float target_rate_pitch_i = default_gains.rate_pitch_i * weight_ratio;
    float target_rate_pitch_d = default_gains.rate_pitch_d * weight_ratio * thrust_curve_linearity;

    float target_yaw_p = default_gains.yaw_p * weight_ratio;
    float target_yaw_i = default_gains.yaw_i * weight_ratio;
    float target_yaw_d = default_gains.yaw_d * weight_ratio;
    float target_rate_yaw_p = default_gains.rate_yaw_p * weight_ratio;
    float target_rate_yaw_i = default_gains.rate_yaw_i * weight_ratio;
    float target_rate_yaw_d = default_gains.rate_yaw_d * weight_ratio;

    float target_alt_p = default_gains.alt_p * weight_ratio;
    float target_alt_i = default_gains.alt_i * weight_ratio;
    float target_alt_d = default_gains.alt_d * weight_ratio * thrust_curve_linearity;

    current_gains.roll_p = clamp_gain(smooth_gain_update(target_roll_p, current_gains.roll_p), default_gains.roll_p);
    current_gains.roll_i = clamp_gain(smooth_gain_update(target_roll_i, current_gains.roll_i), default_gains.roll_i);
    current_gains.roll_d = clamp_gain(smooth_gain_update(target_roll_d, current_gains.roll_d), default_gains.roll_d);
    current_gains.rate_roll_p = clamp_gain(smooth_gain_update(target_rate_roll_p, current_gains.rate_roll_p), default_gains.rate_roll_p);
    current_gains.rate_roll_i = clamp_gain(smooth_gain_update(target_rate_roll_i, current_gains.rate_roll_i), default_gains.rate_roll_i);
    current_gains.rate_roll_d = clamp_gain(smooth_gain_update(target_rate_roll_d, current_gains.rate_roll_d), default_gains.rate_roll_d);

    current_gains.pitch_p = clamp_gain(smooth_gain_update(target_pitch_p, current_gains.pitch_p), default_gains.pitch_p);
    current_gains.pitch_i = clamp_gain(smooth_gain_update(target_pitch_i, current_gains.pitch_i), default_gains.pitch_i);
    current_gains.pitch_d = clamp_gain(smooth_gain_update(target_pitch_d, current_gains.pitch_d), default_gains.pitch_d);
    current_gains.rate_pitch_p = clamp_gain(smooth_gain_update(target_rate_pitch_p, current_gains.rate_pitch_p), default_gains.rate_pitch_p);
    current_gains.rate_pitch_i = clamp_gain(smooth_gain_update(target_rate_pitch_i, current_gains.rate_pitch_i), default_gains.rate_pitch_i);
    current_gains.rate_pitch_d = clamp_gain(smooth_gain_update(target_rate_pitch_d, current_gains.rate_pitch_d), default_gains.rate_pitch_d);

    current_gains.yaw_p = clamp_gain(smooth_gain_update(target_yaw_p, current_gains.yaw_p), default_gains.yaw_p);
    current_gains.yaw_i = clamp_gain(smooth_gain_update(target_yaw_i, current_gains.yaw_i), default_gains.yaw_i);
    current_gains.yaw_d = clamp_gain(smooth_gain_update(target_yaw_d, current_gains.yaw_d), default_gains.yaw_d);
    current_gains.rate_yaw_p = clamp_gain(smooth_gain_update(target_rate_yaw_p, current_gains.rate_yaw_p), default_gains.rate_yaw_p);
    current_gains.rate_yaw_i = clamp_gain(smooth_gain_update(target_rate_yaw_i, current_gains.rate_yaw_i), default_gains.rate_yaw_i);
    current_gains.rate_yaw_d = clamp_gain(smooth_gain_update(target_rate_yaw_d, current_gains.rate_yaw_d), default_gains.rate_yaw_d);

    current_gains.alt_p = clamp_gain(smooth_gain_update(target_alt_p, current_gains.alt_p), default_gains.alt_p);
    current_gains.alt_i = clamp_gain(smooth_gain_update(target_alt_i, current_gains.alt_i), default_gains.alt_i);
    current_gains.alt_d = clamp_gain(smooth_gain_update(target_alt_d, current_gains.alt_d), default_gains.alt_d);

    flight_controller_set_pid_gains(&current_gains);
}

static void collect_thrust_sample(float throttle, float accel_z)
{
    if (sample_count >= THRUST_LEARNER_MAX_SAMPLES) {
        return;
    }

    BatteryState battery;
    sensor_manager_get_battery(&battery);

    samples[sample_count].throttle = throttle;
    samples[sample_count].accel_z = accel_z;
    samples[sample_count].motor_pwm[0] = 0.0f;
    samples[sample_count].motor_pwm[1] = 0.0f;
    samples[sample_count].motor_pwm[2] = 0.0f;
    samples[sample_count].motor_pwm[3] = 0.0f;
    samples[sample_count].voltage = battery.voltage;
    samples[sample_count].timestamp = HAL_GetTick();
    sample_count++;
}

static void build_thrust_curve(void)
{
    if (sample_count < 2 || weight_state.estimated_weight_kg <= 0.0f) {
        return;
    }

    float throttle_step = 1.0f / (float)(THRUST_CURVE_POINT_COUNT - 1);
    float weight = weight_state.estimated_weight_kg;

    for (uint8_t i = 0; i < THRUST_CURVE_POINT_COUNT; i++) {
        thrust_curve[i].throttle = i * throttle_step;
        thrust_curve[i].thrust_N = weight * GRAVITY * thrust_curve[i].throttle * 2.0f;
        thrust_curve[i].motor_rpm_avg = thrust_curve[i].throttle * 5000.0f + 1000.0f;
    }

    float sum_xy = 0.0f, sum_x = 0.0f, sum_y = 0.0f, sum_x2 = 0.0f;
    for (uint8_t i = 0; i < THRUST_CURVE_POINT_COUNT; i++) {
        sum_x += thrust_curve[i].throttle;
        sum_y += thrust_curve[i].thrust_N;
        sum_xy += thrust_curve[i].throttle * thrust_curve[i].thrust_N;
        sum_x2 += thrust_curve[i].throttle * thrust_curve[i].throttle;
    }
    float n = (float)THRUST_CURVE_POINT_COUNT;
    float slope = (n * sum_xy - sum_x * sum_y) / (n * sum_x2 - sum_x * sum_x);
    float intercept = (sum_y - slope * sum_x) / n;

    float ss_res = 0.0f, ss_tot = 0.0f;
    float mean_y = sum_y / n;
    for (uint8_t i = 0; i < THRUST_CURVE_POINT_COUNT; i++) {
        float predicted = slope * thrust_curve[i].throttle + intercept;
        ss_res += (thrust_curve[i].thrust_N - predicted) * (thrust_curve[i].thrust_N - predicted);
        ss_tot += (thrust_curve[i].thrust_N - mean_y) * (thrust_curve[i].thrust_N - mean_y);
    }
    if (ss_tot > 0.0001f) {
        float r_squared = 1.0f - (ss_res / ss_tot);
        thrust_curve_linearity = CONSTRAIN(r_squared, 0.5f, 1.0f);
    }
}

static void estimate_weight_from_sample(float throttle, float accel_z, float vertical_velocity)
{
    if (!weight_state.estimating) {
        if (vertical_velocity > TAKEOFF_CLIMB_RATE_THRESHOLD && throttle > 0.3f) {
            weight_state.estimating = true;
            weight_state.start_time = HAL_GetTick();
            weight_state.avg_accel_z = 0.0f;
            weight_state.avg_throttle = 0.0f;
            weight_state.sample_count = 0;
            learner_state = LS_WEIGHT_ESTIMATION;
        }
        return;
    }

    uint32_t elapsed = HAL_GetTick() - weight_state.start_time;
    if (elapsed < WEIGHT_ESTIMATION_DURATION_MS) {
        weight_state.avg_accel_z += accel_z;
        weight_state.avg_throttle += throttle;
        weight_state.sample_count++;
    } else if (weight_state.sample_count >= MIN_SAMPLES_FOR_WEIGHT) {
        weight_state.avg_accel_z /= (float)weight_state.sample_count;
        weight_state.avg_throttle /= (float)weight_state.sample_count;

        float net_accel = weight_state.avg_accel_z - GRAVITY;
        if (net_accel > 0.5f && weight_state.avg_throttle > 0.0f) {
            float thrust_at_hover = GRAVITY;
            float hover_throttle_est = weight_state.avg_throttle * (thrust_at_hover / (weight_state.avg_accel_z));
            weight_state.hover_throttle = CONSTRAIN(hover_throttle_est, 0.3f, 0.8f);

            float thrust_N = (weight_state.avg_accel_z) / weight_state.avg_throttle * weight_state.hover_throttle;
            weight_state.estimated_weight_kg = thrust_N / GRAVITY;
            weight_state.estimated_weight_kg = CONSTRAIN(
                weight_state.estimated_weight_kg,
                THRUST_LEARNER_MIN_WEIGHT_KG,
                THRUST_LEARNER_MAX_WEIGHT_KG
            );
        } else {
            weight_state.estimated_weight_kg = THRUST_LEARNER_DEFAULT_WEIGHT_KG;
            weight_state.hover_throttle = 0.5f;
        }

        weight_state.estimating = false;
        learner_state = LS_DATA_COLLECTING;
    }
}

static void detect_hover(float vertical_velocity, float throttle)
{
    uint32_t now = HAL_GetTick();
    if (fabsf(vertical_velocity) < HOVER_VELOCITY_THRESHOLD &&
        throttle > 0.2f && throttle < 0.9f) {
        if (hover_stable_start == 0) {
            hover_stable_start = now;
        } else if (now - hover_stable_start >= HOVER_STABLE_DURATION_MS) {
            if (weight_state.hover_throttle < 0.01f) {
                weight_state.hover_throttle = throttle;
            } else {
                weight_state.hover_throttle = weight_state.hover_throttle * 0.9f + throttle * 0.1f;
            }
        }
    } else {
        hover_stable_start = 0;
    }
}

void thrust_learner_init(void)
{
    learner_state = LS_IDLE;
    sample_count = 0;
    hover_stable_start = 0;
    last_vertical_velocity = 0.0f;
    thrust_curve_linearity = 1.0f;

    memset(&weight_state, 0, sizeof(WeightEstimateState));
    memset(samples, 0, sizeof(samples));
    memset(thrust_curve, 0, sizeof(thrust_curve));

    weight_state.estimated_weight_kg = THRUST_LEARNER_DEFAULT_WEIGHT_KG;
    weight_state.hover_throttle = 0.5f;

    load_default_pid_gains();
}

void thrust_learner_update(float dt)
{
    if (!flight_controller_is_armed()) {
        return;
    }

    AttitudeState attitude;
    task_attitude_estimation_get_state(&attitude);

    PositionState pos;
    flight_controller_get_position(&pos);

    ControlCommand cmd;
    flight_controller_get_final_command(&cmd);

    float vertical_velocity = -pos.velocity.vd;
    last_vertical_velocity = vertical_velocity;

    float accel_z = -attitude.linear_accel.z;

    estimate_weight_from_sample(cmd.throttle, accel_z, vertical_velocity);

    detect_hover(vertical_velocity, cmd.throttle);

    if (learner_state == LS_DATA_COLLECTING || learner_state == LS_APPLIED) {
        if (cmd.throttle > 0.1f && cmd.throttle < 0.95f &&
            fabsf(attitude.euler.roll) < DEG2RAD(10.0f) &&
            fabsf(attitude.euler.pitch) < DEG2RAD(10.0f)) {
            collect_thrust_sample(cmd.throttle, accel_z);
        }

        update_pid_gains_based_on_weight();
    }
}

void thrust_learner_start_weight_estimation(void)
{
    weight_state.estimating = true;
    weight_state.start_time = HAL_GetTick();
    weight_state.avg_accel_z = 0.0f;
    weight_state.avg_throttle = 0.0f;
    weight_state.sample_count = 0;
    learner_state = LS_WEIGHT_ESTIMATION;
}

float thrust_learner_get_estimated_weight(void)
{
    return weight_state.estimated_weight_kg;
}

void thrust_learner_get_pid_gains(PIDGainSet *gains)
{
    if (gains != NULL) {
        *gains = current_gains;
    }
}

void thrust_learner_get_thrust_curve(ThrustCurvePoint *points, uint8_t start_index, uint8_t count, uint8_t *out_count)
{
    if (points == NULL || out_count == NULL) {
        return;
    }
    if (start_index >= THRUST_CURVE_POINT_COUNT) {
        *out_count = 0;
        return;
    }
    uint8_t available = THRUST_CURVE_POINT_COUNT - start_index;
    uint8_t actual_count = (count < available) ? count : available;
    memcpy(points, &thrust_curve[start_index], actual_count * sizeof(ThrustCurvePoint));
    *out_count = actual_count;
}

LearningState thrust_learner_get_state(void)
{
    return learner_state;
}

float thrust_learner_get_hover_throttle(void)
{
    return weight_state.hover_throttle;
}

void thrust_learner_trigger_optimization(void)
{
    learner_state = LS_MODEL_OPTIMIZING;
    build_thrust_curve();
    update_pid_gains_based_on_weight();
    learner_state = LS_APPLIED;
}

uint32_t thrust_learner_get_sample_count(void)
{
    return sample_count;
}
