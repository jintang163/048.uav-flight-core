#ifndef __MOTOR_CONTROL_H__
#define __MOTOR_CONTROL_H__

#include "types.h"
#include "flight_config.h"

typedef enum {
    MOTOR_1 = 0,
    MOTOR_2 = 1,
    MOTOR_3 = 2,
    MOTOR_4 = 3
} MotorID;

typedef struct {
    float roll;
    float pitch;
    float yaw;
    float throttle;
} MotorMixInput;

typedef struct {
    uint32_t pwm[4];
    float normalized[4];
    bool armed;
    bool safety_off;
    uint32_t last_update;
} MotorState;

void motor_control_init(void);
void motor_control_update(void);
void motor_control_set_mix(MotorMixInput *input);
void motor_control_arm(void);
void motor_control_disarm(void);
bool motor_control_is_armed(void);
void motor_control_set_pwm(uint8_t motor, uint32_t pwm);
void motor_control_set_all_pwm(uint32_t pwm);
uint32_t motor_control_get_pwm(uint8_t motor);
void motor_control_get_all_pwm(uint32_t *pwm);
void motor_control_set_idle(void);
void motor_control_stop(void);
void motor_control_mix_quad_x(float throttle, float roll, float pitch, float yaw, float *motor_out);
float motor_control_get_throttle_curve(float throttle);
void motor_control_enable_safety(void);
void motor_control_disable_safety(void);

#endif
