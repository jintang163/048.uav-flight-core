#include "task_formation.h"
#include "formation_manager.h"
#include "flight_controller.h"
#include "sensor_manager.h"
#include "mission_manager.h"
#include "coordinate.h"
#include "pid_controller.h"
#include "uwb_driver.h"
#include "rgb_led.h"
#include <math.h>
#include <string.h>

static TaskHandle_t task_handle = NULL;
static PIDController pos_xy_pid;
static PIDController pos_z_pid;

#define FORMATION_POS_KP 1.5f
#define FORMATION_POS_KI 0.08f
#define FORMATION_POS_KD 0.4f
#define FORMATION_POS_I_MAX 3.0f
#define FORMATION_POS_OUT_MAX 5.0f

#define FORMATION_ALT_KP 1.2f
#define FORMATION_ALT_KI 0.12f
#define FORMATION_ALT_KD 0.6f
#define FORMATION_ALT_I_MAX 100.0f
#define FORMATION_ALT_OUT_MAX 5.0f

#define FORMATION_ERROR_ENHANCE_GAIN 2.0f

static float position_error_magnitude = 0.0f;
static bool error_exceeds_threshold = false;

void task_formation_init(void)
{
    formation_manager_init();
    rgb_led_init();
    uwb_driver_init();

    pid_init(&pos_xy_pid,
             FORMATION_POS_KP,
             FORMATION_POS_KI,
             FORMATION_POS_KD,
             FORMATION_POS_I_MAX,
             FORMATION_POS_OUT_MAX);

    pid_init(&pos_z_pid,
             FORMATION_ALT_KP,
             FORMATION_ALT_KI,
             FORMATION_ALT_KD,
             FORMATION_ALT_I_MAX,
             FORMATION_ALT_OUT_MAX);

    xTaskCreate(task_formation_main,
                "Formation",
                TASK_FORMATION_STACK_SIZE,
                NULL,
                TASK_FORMATION_PRIORITY,
                &task_handle);
}

static void calculate_formation_control(float dt,
                                        float *target_vel_n,
                                        float *target_vel_e,
                                        float *target_vel_d,
                                        float *target_yaw)
{
    Vector3f target_offset;
    formation_manager_get_target_offset(&target_offset);

    Vector3f actual_relative = {0.0f, 0.0f, 0.0f};
    float leader_vel_n = 0.0f;
    float leader_vel_e = 0.0f;
    float leader_yaw = 0.0f;

    uint8_t leader_id = formation_manager_get_leader();
    UWBNeighborInfo leader_info;
    if (formation_manager_get_neighbor(leader_id, &leader_info)) {
        actual_relative.x = -leader_info.relative_pos.x;
        actual_relative.y = -leader_info.relative_pos.y;
        actual_relative.z = -leader_info.relative_pos.z;
        leader_vel_n = leader_info.velocity.x;
        leader_vel_e = leader_info.velocity.y;
        leader_yaw = leader_info.yaw;
    }

    float measurement_n = actual_relative.x - target_offset.x;
    float measurement_e = actual_relative.y - target_offset.y;
    float measurement_d = actual_relative.z - target_offset.z;

    if (fabsf(leader_yaw) > 0.01f) {
        float cos_yaw = cosf(leader_yaw);
        float sin_yaw = sinf(leader_yaw);
        float m_n = measurement_n * cos_yaw + measurement_e * sin_yaw;
        float m_e = -measurement_n * sin_yaw + measurement_e * cos_yaw;
        measurement_n = m_n;
        measurement_e = m_e;
    }

    position_error_magnitude = sqrtf(measurement_n * measurement_n +
                                     measurement_e * measurement_e +
                                     measurement_d * measurement_d);

    error_exceeds_threshold = (position_error_magnitude > FORMATION_POSITION_ERROR_MAX);

    if (error_exceeds_threshold) {
        pid_set_gains(&pos_xy_pid,
                      FORMATION_POS_KP * FORMATION_ERROR_ENHANCE_GAIN,
                      FORMATION_POS_KI * FORMATION_ERROR_ENHANCE_GAIN,
                      FORMATION_POS_KD * FORMATION_ERROR_ENHANCE_GAIN);
        pid_set_gains(&pos_z_pid,
                      FORMATION_ALT_KP * FORMATION_ERROR_ENHANCE_GAIN,
                      FORMATION_ALT_KI,
                      FORMATION_ALT_KD * FORMATION_ERROR_ENHANCE_GAIN);
    } else {
        pid_set_gains(&pos_xy_pid, FORMATION_POS_KP, FORMATION_POS_KI, FORMATION_POS_KD);
        pid_set_gains(&pos_z_pid, FORMATION_ALT_KP, FORMATION_ALT_KI, FORMATION_ALT_KD);
    }

    float vel_correction_n = pid_compute(&pos_xy_pid, 0.0f, measurement_n, dt);
    float vel_correction_e = pid_compute(&pos_xy_pid, 0.0f, measurement_e, dt);
    float vel_correction_d = pid_compute(&pos_z_pid, 0.0f, measurement_d, dt);

    float decel_factor = formation_manager_get_collision_deceleration();

    *target_vel_n = (leader_vel_n + vel_correction_n) * decel_factor;
    *target_vel_e = (leader_vel_e + vel_correction_e) * decel_factor;
    *target_vel_d = vel_correction_d * decel_factor;
    *target_yaw = leader_yaw;
}

static void update_formation_lights(void)
{
    FormationLightCommand cmd;
    formation_manager_get_light_command(&cmd);

    if (cmd.effect != LIGHT_EFFECT_STATIC || cmd.r != 0 || cmd.g != 0 || cmd.b != 0) {
        rgb_led_set_effect((LedEffectType)cmd.effect, cmd.r, cmd.g, cmd.b);
    }
}

void task_formation_main(void *argument)
{
    UNUSED(argument);

    TickType_t last_wake_time = xTaskGetTickCount();
    const TickType_t period = pdMS_TO_TICKS(1000 / TASK_FORMATION_FREQ);
    float dt = 1.0f / TASK_FORMATION_FREQ;

    while (1) {
        GPSPosition gps_pos;
        GPSVelocity gps_vel;
        sensor_manager_get_gps(&gps_pos, &gps_vel);

        FlightMode mode = flight_controller_get_mode();

        if (mode == FLIGHT_MODE_FORMATION) {
            uwb_driver_update();
            formation_manager_update(dt);

            FormationState formation_state = formation_manager_get_state();

            if (formation_state == FORMATION_STATE_EXECUTING) {
                ControlCommand formation_cmd;
                float target_vel_n, target_vel_e, target_vel_d, target_yaw;

                calculate_formation_control(dt, &target_vel_n, &target_vel_e, &target_vel_d, &target_yaw);

                float vel_error_n = target_vel_n - gps_vel.vn;
                float vel_error_e = target_vel_e - gps_vel.ve;

                formation_cmd.roll = CONSTRAIN(vel_error_e * 0.5f, -MAX_TILT_ANGLE, MAX_TILT_ANGLE);
                formation_cmd.pitch = CONSTRAIN(-vel_error_n * 0.5f, -MAX_TILT_ANGLE, MAX_TILT_ANGLE);
                formation_cmd.yaw = target_yaw;

                float baro_alt;
                sensor_manager_get_baro(&baro_alt, NULL, NULL);

                float alt_error = -target_vel_d * 0.5f;
                formation_cmd.throttle = CONSTRAIN(0.5f + alt_error, 0.3f, 0.8f);

                flight_controller_set_mavlink_command(&formation_cmd);
            }

            update_formation_lights();
        } else {
            pid_reset(&pos_xy_pid);
            pid_reset(&pos_z_pid);
            position_error_magnitude = 0.0f;
            error_exceeds_threshold = false;
            rgb_led_off();
        }

        vTaskDelayUntil(&last_wake_time, period);
    }
}
