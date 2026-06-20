#ifndef __PID_CONTROLLER_H__
#define __PID_CONTROLLER_H__

#include "types.h"
#include "flight_config.h"

typedef struct {
    float kp;
    float ki;
    float kd;
    float i_max;
    float out_max;
    float i_term;
    float last_error;
    float last_measurement;
    float setpoint;
    float measurement;
    float output;
    uint32_t last_update;
} PIDController;

void pid_init(PIDController *pid, float kp, float ki, float kd, float i_max, float out_max);
void pid_reset(PIDController *pid);
float pid_compute(PIDController *pid, float setpoint, float measurement, float dt);
float pid_compute_rate(PIDController *pid, float setpoint, float measurement, float dt);
void pid_set_gains(PIDController *pid, float kp, float ki, float kd);
void pid_set_output_limit(PIDController *pid, float out_max);
void pid_set_integral_limit(PIDController *pid, float i_max);

typedef struct {
    PIDController angle_pid;
    PIDController rate_pid;
} CascadePID;

void cascade_pid_init(CascadePID *pid,
                      float angle_kp, float angle_ki, float angle_kd, float angle_i_max, float angle_out_max,
                      float rate_kp, float rate_ki, float rate_kd, float rate_i_max, float rate_out_max);
void cascade_pid_reset(CascadePID *pid);
float cascade_pid_compute(CascadePID *pid, float angle_setpoint, float angle_measurement,
                          float rate_measurement, float dt);

#endif
