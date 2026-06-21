#include "motor_control.h"
#include "pwm_output.h"
#include "flight_config.h"
#include <string.h>

static MotorState motor_state;

void motor_control_init(void)
{
    memset(&motor_state, 0, sizeof(MotorState));
    for (int i = 0; i < 4; i++) {
        motor_state.pwm[i] = MOTOR_PWM_MIN;
        motor_state.normalized[i] = 0.0f;
    }
    motor_state.armed = false;
    motor_state.safety_off = false;
}

void motor_control_update(void)
{
    if (!motor_state.armed) {
        for (int i = 0; i < 4; i++) {
            motor_state.normalized[i] = 0.0f;
            motor_state.pwm[i] = MOTOR_PWM_MIN;
        }
    }
    pwm_output_update(motor_state.normalized);
    motor_state.last_update = HAL_GetTick();
}

void motor_control_set_mix(MotorMixInput *input)
{
    if (!input) return;

    float throttle = CONSTRAIN(input->throttle, 0.0f, 1.0f);
    float roll = CONSTRAIN(input->roll, -1.0f, 1.0f);
    float pitch = CONSTRAIN(input->pitch, -1.0f, 1.0f);
    float yaw = CONSTRAIN(input->yaw, -1.0f, 1.0f);

    float motors[4];
    motor_control_mix_quad_x(throttle, roll, pitch, yaw, motors);

    for (int i = 0; i < 4; i++) {
        motor_state.normalized[i] = CONSTRAIN(motors[i], 0.0f, 1.0f);
        motor_state.pwm[i] = (uint32_t)(MOTOR_PWM_MIN + 
            motor_state.normalized[i] * (MOTOR_PWM_MAX - MOTOR_PWM_MIN));
    }
}

void motor_control_arm(void)
{
    motor_state.armed = true;
    pwm_output_arm();
}

void motor_control_disarm(void)
{
    motor_state.armed = false;
    for (int i = 0; i < 4; i++) {
        motor_state.normalized[i] = 0.0f;
        motor_state.pwm[i] = MOTOR_PWM_MIN;
    }
    pwm_output_update(motor_state.normalized);
    pwm_output_disarm();
}

bool motor_control_is_armed(void)
{
    return motor_state.armed;
}

void motor_control_set_pwm(uint8_t motor, uint32_t pwm)
{
    if (motor >= 4) return;
    if (pwm < MOTOR_PWM_MIN) pwm = MOTOR_PWM_MIN;
    if (pwm > MOTOR_PWM_MAX) pwm = MOTOR_PWM_MAX;
    motor_state.pwm[motor] = pwm;
    motor_state.normalized[motor] = 
        (float)(pwm - MOTOR_PWM_MIN) / (float)(MOTOR_PWM_MAX - MOTOR_PWM_MIN);
}

void motor_control_set_all_pwm(uint32_t pwm)
{
    for (int i = 0; i < 4; i++) {
        motor_control_set_pwm(i, pwm);
    }
}

uint32_t motor_control_get_pwm(uint8_t motor)
{
    if (motor >= 4) return 0;
    return motor_state.pwm[motor];
}

void motor_control_get_all_pwm(uint32_t *pwm)
{
    if (!pwm) return;
    for (int i = 0; i < 4; i++) {
        pwm[i] = motor_state.pwm[i];
    }
}

void motor_control_set_idle(void)
{
    for (int i = 0; i < 4; i++) {
        motor_state.normalized[i] = MOTOR_IDLE_THROTTLE;
        motor_state.pwm[i] = (uint32_t)(MOTOR_PWM_MIN + 
            MOTOR_IDLE_THROTTLE * (MOTOR_PWM_MAX - MOTOR_PWM_MIN));
    }
}

void motor_control_stop(void)
{
    for (int i = 0; i < 4; i++) {
        motor_state.normalized[i] = 0.0f;
        motor_state.pwm[i] = MOTOR_PWM_MIN;
    }
    pwm_output_update(motor_state.normalized);
}

void motor_control_mix_quad_x(float throttle, float roll, float pitch, float yaw, float *motor_out)
{
    if (!motor_out) return;

    float roll_coeff = 0.5f;
    float pitch_coeff = 0.5f;
    float yaw_coeff = 0.5f;

    motor_out[MOTOR_1] = throttle - roll_coeff * roll + pitch_coeff * pitch + yaw_coeff * yaw;
    motor_out[MOTOR_2] = throttle + roll_coeff * roll + pitch_coeff * pitch - yaw_coeff * yaw;
    motor_out[MOTOR_3] = throttle + roll_coeff * roll - pitch_coeff * pitch + yaw_coeff * yaw;
    motor_out[MOTOR_4] = throttle - roll_coeff * roll - pitch_coeff * pitch - yaw_coeff * yaw;

    float max_motor = motor_out[0];
    for (int i = 1; i < 4; i++) {
        if (motor_out[i] > max_motor) {
            max_motor = motor_out[i];
        }
    }

    if (max_motor > 1.0f && throttle > MOTOR_MIN_THROTTLE) {
        float scale = (1.0f - throttle) / (max_motor - throttle);
        for (int i = 0; i < 4; i++) {
            motor_out[i] = throttle + (motor_out[i] - throttle) * scale;
        }
    }

    for (int i = 0; i < 4; i++) {
        motor_out[i] = CONSTRAIN(motor_out[i], 0.0f, 1.0f);
    }
}

float motor_control_get_throttle_curve(float throttle)
{
    throttle = CONSTRAIN(throttle, 0.0f, 1.0f);
    return throttle * throttle;
}

void motor_control_enable_safety(void)
{
    motor_state.safety_off = false;
}

void motor_control_disable_safety(void)
{
    motor_state.safety_off = true;
}

void motor_control_get_state(MotorState *state)
{
    if (!state) return;
    *state = motor_state;
}
