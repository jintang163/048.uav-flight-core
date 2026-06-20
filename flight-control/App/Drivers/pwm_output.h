#ifndef __PWM_OUTPUT_H__
#define __PWM_OUTPUT_H__

#include "types.h"
#include "flight_config.h"

typedef struct {
    uint32_t pwm_value[4];
    float normalized[4];
    uint32_t min_pwm;
    uint32_t max_pwm;
    uint32_t idle_pwm;
    bool armed;
    bool initialized;
} PWM_Output;

bool pwm_output_init(void);
void pwm_output_set_motor(uint8_t motor, float value);
void pwm_output_set_all(float *values);
void pwm_output_set_pwm(uint8_t motor, uint32_t pwm);
void pwm_output_set_all_pwm(uint32_t pwm);
void pwm_output_update(void);
void pwm_output_arm(void);
void pwm_output_disarm(void);
bool pwm_output_is_armed(void);
void pwm_output_set_idle(void);
void pwm_output_stop(void);
uint32_t pwm_output_get_pwm(uint8_t motor);

#endif
