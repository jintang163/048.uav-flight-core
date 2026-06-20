#include "pwm_output.h"
#include "main.h"

extern TIM_HandleTypeDef htim1;

static PWM_Output pwm_output;

bool pwm_output_init(void)
{
    pwm_output.min_pwm = PWM_MIN;
    pwm_output.max_pwm = PWM_MAX;
    pwm_output.idle_pwm = PWM_IDLE;
    pwm_output.armed = false;
    pwm_output.initialized = true;

    for (int i = 0; i < 4; i++) {
        pwm_output.pwm_value[i] = PWM_MIN;
        pwm_output.normalized[i] = 0.0f;
    }

    pwm_output_update();

    HAL_TIM_PWM_Start(&htim1, TIM_CHANNEL_1);
    HAL_TIM_PWM_Start(&htim1, TIM_CHANNEL_2);
    HAL_TIM_PWM_Start(&htim1, TIM_CHANNEL_3);
    HAL_TIM_PWM_Start(&htim1, TIM_CHANNEL_4);

    return true;
}

void pwm_output_set_motor(uint8_t motor, float value)
{
    if (motor >= 4) {
        return;
    }

    pwm_output.normalized[motor] = CONSTRAIN(value, 0.0f, 1.0f);
    pwm_output.pwm_value[motor] = (uint32_t)(pwm_output.min_pwm +
                                            pwm_output.normalized[motor] *
                                            (pwm_output.max_pwm - pwm_output.min_pwm));
}

void pwm_output_set_all(float *values)
{
    for (int i = 0; i < 4; i++) {
        pwm_output_set_motor(i, values[i]);
    }
}

void pwm_output_set_pwm(uint8_t motor, uint32_t pwm)
{
    if (motor >= 4) {
        return;
    }

    pwm_output.pwm_value[motor] = CONSTRAIN(pwm, pwm_output.min_pwm, pwm_output.max_pwm);
    pwm_output.normalized[motor] = (float)(pwm_output.pwm_value[motor] - pwm_output.min_pwm) /
                                  (float)(pwm_output.max_pwm - pwm_output.min_pwm);
}

void pwm_output_set_all_pwm(uint32_t pwm)
{
    for (int i = 0; i < 4; i++) {
        pwm_output_set_pwm(i, pwm);
    }
}

void pwm_output_update(void)
{
    uint32_t pwm_output_value[4];

    for (int i = 0; i < 4; i++) {
        if (pwm_output.armed) {
            pwm_output_value[i] = pwm_output.pwm_value[i];
        } else {
            pwm_output_value[i] = pwm_output.min_pwm;
        }
        pwm_output_value[i] = CONSTRAIN(pwm_output_value[i], PWM_MIN, PWM_MAX_OUT);
    }

    __HAL_TIM_SET_COMPARE(&htim1, TIM_CHANNEL_1, pwm_output_value[0]);
    __HAL_TIM_SET_COMPARE(&htim1, TIM_CHANNEL_2, pwm_output_value[1]);
    __HAL_TIM_SET_COMPARE(&htim1, TIM_CHANNEL_3, pwm_output_value[2]);
    __HAL_TIM_SET_COMPARE(&htim1, TIM_CHANNEL_4, pwm_output_value[3]);
}

void pwm_output_arm(void)
{
    pwm_output.armed = true;
    pwm_output_set_all_pwm(PWM_IDLE);
    pwm_output_update();
}

void pwm_output_disarm(void)
{
    pwm_output.armed = false;
    pwm_output_set_all_pwm(PWM_MIN);
    pwm_output_update();
}

bool pwm_output_is_armed(void)
{
    return pwm_output.armed;
}

void pwm_output_set_idle(void)
{
    for (int i = 0; i < 4; i++) {
        pwm_output.pwm_value[i] = PWM_IDLE;
        pwm_output.normalized[i] = 0.05f;
    }
    pwm_output_update();
}

void pwm_output_stop(void)
{
    for (int i = 0; i < 4; i++) {
        pwm_output.pwm_value[i] = PWM_MIN;
        pwm_output.normalized[i] = 0.0f;
    }
    pwm_output_update();
}

uint32_t pwm_output_get_pwm(uint8_t motor)
{
    if (motor >= 4) {
        return 0;
    }
    return pwm_output.pwm_value[motor];
}
