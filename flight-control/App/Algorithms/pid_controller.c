#include "pid_controller.h"

void pid_init(PIDController *pid, float kp, float ki, float kd, float i_max, float out_max)
{
    pid->kp = kp;
    pid->ki = ki;
    pid->kd = kd;
    pid->i_max = i_max;
    pid->out_max = out_max;
    pid_reset(pid);
}

void pid_reset(PIDController *pid)
{
    pid->i_term = 0.0f;
    pid->last_error = 0.0f;
    pid->last_measurement = 0.0f;
    pid->setpoint = 0.0f;
    pid->measurement = 0.0f;
    pid->output = 0.0f;
    pid->last_update = 0;
}

float pid_compute(PIDController *pid, float setpoint, float measurement, float dt)
{
    if (dt <= 0.0f) {
        return pid->output;
    }

    pid->setpoint = setpoint;
    pid->measurement = measurement;

    float error = setpoint - measurement;

    float p_term = pid->kp * error;

    pid->i_term += pid->ki * error * dt;
    pid->i_term = CONSTRAIN(pid->i_term, -pid->i_max, pid->i_max);

    float d_term = pid->kd * (error - pid->last_error) / dt;

    float output = p_term + pid->i_term + d_term;
    pid->output = CONSTRAIN(output, -pid->out_max, pid->out_max);

    pid->last_error = error;

    return pid->output;
}

float pid_compute_rate(PIDController *pid, float setpoint, float measurement, float dt)
{
    if (dt <= 0.0f) {
        return pid->output;
    }

    pid->setpoint = setpoint;
    pid->measurement = measurement;

    float error = setpoint - measurement;

    float p_term = pid->kp * error;

    pid->i_term += pid->ki * error * dt;
    pid->i_term = CONSTRAIN(pid->i_term, -pid->i_max, pid->i_max);

    float d_term = -pid->kd * (measurement - pid->last_measurement) / dt;

    float output = p_term + pid->i_term + d_term;
    pid->output = CONSTRAIN(output, -pid->out_max, pid->out_max);

    pid->last_measurement = measurement;

    return pid->output;
}

void pid_set_gains(PIDController *pid, float kp, float ki, float kd)
{
    pid->kp = kp;
    pid->ki = ki;
    pid->kd = kd;
}

void pid_set_output_limit(PIDController *pid, float out_max)
{
    pid->out_max = out_max;
}

void pid_set_integral_limit(PIDController *pid, float i_max)
{
    pid->i_max = i_max;
}

void cascade_pid_init(CascadePID *pid,
                      float angle_kp, float angle_ki, float angle_kd, float angle_i_max, float angle_out_max,
                      float rate_kp, float rate_ki, float rate_kd, float rate_i_max, float rate_out_max)
{
    pid_init(&pid->angle_pid, angle_kp, angle_ki, angle_kd, angle_i_max, angle_out_max);
    pid_init(&pid->rate_pid, rate_kp, rate_ki, rate_kd, rate_i_max, rate_out_max);
}

void cascade_pid_reset(CascadePID *pid)
{
    pid_reset(&pid->angle_pid);
    pid_reset(&pid->rate_pid);
}

float cascade_pid_compute(CascadePID *pid, float angle_setpoint, float angle_measurement,
                          float rate_measurement, float dt)
{
    float rate_setpoint = pid_compute(&pid->angle_pid, angle_setpoint, angle_measurement, dt);

    float output = pid_compute_rate(&pid->rate_pid, rate_setpoint, rate_measurement, dt);

    return output;
}
