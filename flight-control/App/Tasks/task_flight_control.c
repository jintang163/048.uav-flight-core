#include "task_flight_control.h"
#include "flight_controller.h"
#include "task_attitude_estimation.h"
#include "sensor_manager.h"
#include "coordinate.h"

static TaskHandle_t task_handle = NULL;
static CascadePID roll_pid, pitch_pid, yaw_pid;
static PIDController alt_pid, pos_x_pid, pos_y_pid;

void task_flight_control_get_pid_gains(PIDGainSet *gains)
{
    if (gains == NULL) {
        return;
    }
    gains->roll_p = roll_pid.angle_pid.kp;
    gains->roll_i = roll_pid.angle_pid.ki;
    gains->roll_d = roll_pid.angle_pid.kd;
    gains->rate_roll_p = roll_pid.rate_pid.kp;
    gains->rate_roll_i = roll_pid.rate_pid.ki;
    gains->rate_roll_d = roll_pid.rate_pid.kd;

    gains->pitch_p = pitch_pid.angle_pid.kp;
    gains->pitch_i = pitch_pid.angle_pid.ki;
    gains->pitch_d = pitch_pid.angle_pid.kd;
    gains->rate_pitch_p = pitch_pid.rate_pid.kp;
    gains->rate_pitch_i = pitch_pid.rate_pid.ki;
    gains->rate_pitch_d = pitch_pid.rate_pid.kd;

    gains->yaw_p = yaw_pid.angle_pid.kp;
    gains->yaw_i = yaw_pid.angle_pid.ki;
    gains->yaw_d = yaw_pid.angle_pid.kd;
    gains->rate_yaw_p = yaw_pid.rate_pid.kp;
    gains->rate_yaw_i = yaw_pid.rate_pid.ki;
    gains->rate_yaw_d = yaw_pid.rate_pid.kd;

    gains->alt_p = alt_pid.kp;
    gains->alt_i = alt_pid.ki;
    gains->alt_d = alt_pid.kd;
}

void task_flight_control_set_pid_gains(PIDGainSet *gains)
{
    if (gains == NULL) {
        return;
    }
    pid_set_gains(&roll_pid.angle_pid, gains->roll_p, gains->roll_i, gains->roll_d);
    pid_set_gains(&roll_pid.rate_pid, gains->rate_roll_p, gains->rate_roll_i, gains->rate_roll_d);

    pid_set_gains(&pitch_pid.angle_pid, gains->pitch_p, gains->pitch_i, gains->pitch_d);
    pid_set_gains(&pitch_pid.rate_pid, gains->rate_pitch_p, gains->rate_pitch_i, gains->rate_pitch_d);

    pid_set_gains(&yaw_pid.angle_pid, gains->yaw_p, gains->yaw_i, gains->yaw_d);
    pid_set_gains(&yaw_pid.rate_pid, gains->rate_yaw_p, gains->rate_yaw_i, gains->rate_yaw_d);

    pid_set_gains(&alt_pid, gains->alt_p, gains->alt_i, gains->alt_d);
}

void task_flight_control_init(void)
{
    cascade_pid_init(&roll_pid,
                     PID_ANGLE_ROLL_P, PID_ANGLE_ROLL_I, PID_ANGLE_ROLL_D,
                     PID_ANGLE_ROLL_I_MAX, PID_ANGLE_ROLL_OUT_MAX,
                     PID_RATE_ROLL_P, PID_RATE_ROLL_I, PID_RATE_ROLL_D,
                     PID_RATE_ROLL_I_MAX, PID_RATE_ROLL_OUT_MAX);

    cascade_pid_init(&pitch_pid,
                     PID_ANGLE_PITCH_P, PID_ANGLE_PITCH_I, PID_ANGLE_PITCH_D,
                     PID_ANGLE_PITCH_I_MAX, PID_ANGLE_PITCH_OUT_MAX,
                     PID_RATE_PITCH_P, PID_RATE_PITCH_I, PID_RATE_PITCH_D,
                     PID_RATE_PITCH_I_MAX, PID_RATE_PITCH_OUT_MAX);

    cascade_pid_init(&yaw_pid,
                     PID_ANGLE_YAW_P, PID_ANGLE_YAW_I, PID_ANGLE_YAW_D,
                     PID_ANGLE_YAW_I_MAX, PID_ANGLE_YAW_OUT_MAX,
                     PID_RATE_YAW_P, PID_RATE_YAW_I, PID_RATE_YAW_D,
                     PID_RATE_YAW_I_MAX, PID_RATE_YAW_OUT_MAX);

    pid_init(&alt_pid, PID_ALT_P, PID_ALT_I, PID_ALT_D, PID_ALT_I_MAX, PID_ALT_OUT_MAX);
    pid_init(&pos_x_pid, PID_POS_X_P, PID_POS_X_I, PID_POS_X_D, PID_POS_X_I_MAX, PID_POS_X_OUT_MAX);
    pid_init(&pos_y_pid, PID_POS_Y_P, PID_POS_Y_I, PID_POS_Y_D, PID_POS_Y_I_MAX, PID_POS_Y_OUT_MAX);

    flight_controller_init();

    xTaskCreate(task_flight_control_main,
                "FlightCtrl",
                TASK_FLIGHT_CONTROL_STACK_SIZE,
                NULL,
                TASK_FLIGHT_CONTROL_PRIORITY,
                &task_handle);
}

void task_flight_control_main(void *argument)
{
    UNUSED(argument);

    TickType_t last_wake_time = xTaskGetTickCount();
    const TickType_t period = pdMS_TO_TICKS(1000 / TASK_FLIGHT_CONTROL_FREQ);
    float dt = 1.0f / TASK_FLIGHT_CONTROL_FREQ;

    while (1) {
        AttitudeState attitude;
        task_attitude_estimation_get_state(&attitude);

        RCInput rc;
        sensor_manager_get_rc(&rc);

        ControlCommand rc_cmd;
        if (rc.connected) {
            float roll_norm = (float)(rc.channels[SBUS_CHANNEL_ROLL] - SBUS_CHANNEL_MID) /
                              (float)(SBUS_CHANNEL_MAX - SBUS_CHANNEL_MID);
            float pitch_norm = (float)(rc.channels[SBUS_CHANNEL_PITCH] - SBUS_CHANNEL_MID) /
                               (float)(SBUS_CHANNEL_MAX - SBUS_CHANNEL_MID);
            float yaw_norm = (float)(rc.channels[SBUS_CHANNEL_YAW] - SBUS_CHANNEL_MID) /
                             (float)(SBUS_CHANNEL_MAX - SBUS_CHANNEL_MID);
            float throttle_norm = (float)(rc.channels[SBUS_CHANNEL_THROTTLE] - SBUS_CHANNEL_MIN) /
                                  (float)(SBUS_CHANNEL_MAX - SBUS_CHANNEL_MIN);

            rc_cmd.roll = roll_norm * MAX_TILT_ANGLE;
            rc_cmd.pitch = pitch_norm * MAX_TILT_ANGLE;
            rc_cmd.yaw = yaw_norm * MAX_YAW_RATE;
            rc_cmd.throttle = throttle_norm;

            flight_controller_set_rc_command(&rc_cmd);

            if (rc.channels[SBUS_CHANNEL_ARM] > SBUS_CHANNEL_MID) {
                flight_controller_arm();
            } else {
                flight_controller_disarm();
            }

            uint16_t mode_channel = rc.channels[SBUS_CHANNEL_MODE];
            if (mode_channel < 500) {
                flight_controller_set_mode(FLIGHT_MODE_MANUAL);
            } else if (mode_channel < 1000) {
                flight_controller_set_mode(FLIGHT_MODE_ALT_HOLD);
            } else if (mode_channel < 1500) {
                flight_controller_set_mode(FLIGHT_MODE_POS_HOLD);
            } else {
                flight_controller_set_mode(FLIGHT_MODE_AUTO);
            }
        }

        flight_controller_update(dt);

        FlightMode mode = flight_controller_get_mode();
        ControlCommand final_cmd;
        flight_controller_get_final_command(&final_cmd);

        float roll_output = 0, pitch_output = 0, yaw_output = 0, throttle_output = 0;

        if (flight_controller_is_armed()) {
            throttle_output = final_cmd.throttle;

            switch (mode) {
                case FLIGHT_MODE_MANUAL:
                    roll_output = final_cmd.roll * PID_RATE_ROLL_OUT_MAX;
                    pitch_output = final_cmd.pitch * PID_RATE_PITCH_OUT_MAX;
                    yaw_output = final_cmd.yaw * PID_RATE_YAW_OUT_MAX;
                    break;

                case FLIGHT_MODE_ALT_HOLD:
                case FLIGHT_MODE_POS_HOLD:
                case FLIGHT_MODE_AUTO:
                case FLIGHT_MODE_RTL:
                case FLIGHT_MODE_LAND:
                    roll_output = cascade_pid_compute(&roll_pid, final_cmd.roll,
                                                      attitude.euler.roll,
                                                      attitude.angular_velocity.x, dt);
                    pitch_output = cascade_pid_compute(&pitch_pid, final_cmd.pitch,
                                                       attitude.euler.pitch,
                                                       attitude.angular_velocity.y, dt);
                    yaw_output = cascade_pid_compute(&yaw_pid, final_cmd.yaw,
                                                     attitude.euler.yaw,
                                                     attitude.angular_velocity.z, dt);
                    break;

                default:
                    break;
            }
        }

        MotorMixInput mix_input;
        mix_input.roll = roll_output;
        mix_input.pitch = pitch_output;
        mix_input.yaw = yaw_output;
        mix_input.throttle = throttle_output;

        motor_control_set_mix(&mix_input);

        vTaskDelayUntil(&last_wake_time, period);
    }
}
