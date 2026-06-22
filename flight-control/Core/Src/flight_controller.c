#include "flight_controller.h"
#include "task_attitude_estimation.h"
#include "sensor_manager.h"
#include "motor_control.h"
#include "pwm_output.h"
#include "task_flight_control.h"

static FlightControllerData fc_data;

void flight_controller_init(void)
{
    memset(&fc_data, 0, sizeof(FlightControllerData));

    fc_data.state = FCS_DISARMED;
    fc_data.mode = FLIGHT_MODE_MANUAL;
    fc_data.requested_mode = FLIGHT_MODE_MANUAL;
    fc_data.control_source = CONTROL_SOURCE_RC;
    fc_data.armed = false;
    fc_data.failsafe_active = false;
    fc_data.home_set = false;

    fc_data.rc_command.roll = 0;
    fc_data.rc_command.pitch = 0;
    fc_data.rc_command.yaw = 0;
    fc_data.rc_command.throttle = 0;

    fc_data.mavlink_command.roll = 0;
    fc_data.mavlink_command.pitch = 0;
    fc_data.mavlink_command.yaw = 0;
    fc_data.mavlink_command.throttle = 0;

    fc_data.final_command.roll = 0;
    fc_data.final_command.pitch = 0;
    fc_data.final_command.yaw = 0;
    fc_data.final_command.throttle = 0;

    cascade_pid_init(&fc_data.roll_cascade,
                     PID_ANGLE_ROLL_P, PID_ANGLE_ROLL_I, PID_ANGLE_ROLL_D,
                     PID_ANGLE_ROLL_I_MAX, PID_ANGLE_ROLL_OUT_MAX,
                     PID_RATE_ROLL_P, PID_RATE_ROLL_I, PID_RATE_ROLL_D,
                     PID_RATE_ROLL_I_MAX, PID_RATE_ROLL_OUT_MAX);

    cascade_pid_init(&fc_data.pitch_cascade,
                     PID_ANGLE_PITCH_P, PID_ANGLE_PITCH_I, PID_ANGLE_PITCH_D,
                     PID_ANGLE_PITCH_I_MAX, PID_ANGLE_PITCH_OUT_MAX,
                     PID_RATE_PITCH_P, PID_RATE_PITCH_I, PID_RATE_PITCH_D,
                     PID_RATE_PITCH_I_MAX, PID_RATE_PITCH_OUT_MAX);

    cascade_pid_init(&fc_data.yaw_cascade,
                     PID_ANGLE_YAW_P, PID_ANGLE_YAW_I, PID_ANGLE_YAW_D,
                     PID_ANGLE_YAW_I_MAX, PID_ANGLE_YAW_OUT_MAX,
                     PID_RATE_YAW_P, PID_RATE_YAW_I, PID_RATE_YAW_D,
                     PID_RATE_YAW_I_MAX, PID_RATE_YAW_OUT_MAX);

    pid_init(&fc_data.alt_pid,
             PID_ALT_P, PID_ALT_I, PID_ALT_D,
             PID_ALT_I_MAX, PID_ALT_OUT_MAX);
}

void flight_controller_update(float dt)
{
    task_attitude_estimation_get_state(&fc_data.attitude);

    GPSPosition gps_pos;
    GPSVelocity gps_vel;
    sensor_manager_get_gps(&gps_pos, &gps_vel);

    fc_data.position.position = gps_pos;
    fc_data.position.velocity = gps_vel;
    fc_data.position.altitude = (float)gps_pos.alt / 1000.0f;
    fc_data.position.ground_speed = sqrtf(gps_vel.vn * gps_vel.vn + gps_vel.ve * gps_vel.ve);
    fc_data.position.heading = atan2f(gps_vel.ve, gps_vel.vn);

    float baro_alt;
    sensor_manager_get_baro(&baro_alt, NULL, NULL);
    if (baro_alt > 0) {
        fc_data.position.altitude = baro_alt;
    }

    RCInput rc;
    sensor_manager_get_rc(&rc);

    if (rc.connected) {
        fc_data.control_source = CONTROL_SOURCE_RC;
        fc_data.final_command = fc_data.rc_command;
    } else {
        fc_data.control_source = CONTROL_SOURCE_MAVLINK;
        fc_data.final_command = fc_data.mavlink_command;
    }

    if (fc_data.failsafe_active) {
        fc_data.final_command.roll = 0;
        fc_data.final_command.pitch = 0;
        fc_data.final_command.yaw = 0;

        if (fc_data.mode == FLIGHT_MODE_RTL || fc_data.mode == FLIGHT_MODE_LAND) {
        } else {
            fc_data.final_command.throttle = 0.3f;
        }
    }

    switch (fc_data.state) {
        case FCS_DISARMED:
            if (fc_data.armed) {
                fc_data.state = FCS_ARMED;
                fc_data.arm_time = HAL_GetTick();
                motor_control_arm();
                pwm_output_arm();
            }
            break;

        case FCS_ARMED:
            if (!fc_data.armed) {
                fc_data.state = FCS_DISARMED;
                motor_control_disarm();
                pwm_output_disarm();
            } else if (fc_data.failsafe_active) {
                fc_data.state = FCS_EMERGENCY;
            }
            break;

        case FCS_TAKEOFF:
            if (fc_data.position.altitude >= TAKEOFF_HEIGHT) {
                fc_data.state = FCS_ARMED;
            } else if (!fc_data.armed) {
                fc_data.state = FCS_DISARMED;
                motor_control_disarm();
                pwm_output_disarm();
            }
            break;

        case FCS_LANDING:
            if (fc_data.position.altitude < 0.5f) {
                fc_data.state = FCS_DISARMED;
                fc_data.armed = false;
                motor_control_disarm();
                pwm_output_disarm();
            } else if (!fc_data.armed) {
                fc_data.state = FCS_DISARMED;
                motor_control_disarm();
                pwm_output_disarm();
            }
            break;

        case FCS_EMERGENCY:
            if (!fc_data.failsafe_active && fc_data.armed) {
                fc_data.state = FCS_ARMED;
            } else if (!fc_data.armed) {
                fc_data.state = FCS_DISARMED;
                motor_control_disarm();
                pwm_output_disarm();
            }
            break;

        default:
            break;
    }

    if (fc_data.armed) {
        fc_data.flight_time = (HAL_GetTick() - fc_data.arm_time) / 1000;
    }

    fc_data.final_command.roll = CONSTRAIN(fc_data.final_command.roll, -MAX_TILT_ANGLE, MAX_TILT_ANGLE);
    fc_data.final_command.pitch = CONSTRAIN(fc_data.final_command.pitch, -MAX_TILT_ANGLE, MAX_TILT_ANGLE);
    fc_data.final_command.yaw = CONSTRAIN(fc_data.final_command.yaw, -MAX_YAW_RATE, MAX_YAW_RATE);
    fc_data.final_command.throttle = CONSTRAIN(fc_data.final_command.throttle, 0.0f, 1.0f);

    if (fc_data.armed && dt > 0.0f) {
        fc_data.attitude_target.roll = cascade_pid_compute(&fc_data.roll_cascade,
                                                            fc_data.final_command.roll,
                                                            fc_data.attitude.euler.roll,
                                                            fc_data.attitude.angular_velocity.x,
                                                            dt);
        fc_data.attitude_target.pitch = cascade_pid_compute(&fc_data.pitch_cascade,
                                                             fc_data.final_command.pitch,
                                                             fc_data.attitude.euler.pitch,
                                                             fc_data.attitude.angular_velocity.y,
                                                             dt);
        fc_data.attitude_target.yaw = cascade_pid_compute(&fc_data.yaw_cascade,
                                                           fc_data.final_command.yaw,
                                                           fc_data.attitude.euler.yaw,
                                                           fc_data.attitude.angular_velocity.z,
                                                           dt);
        fc_data.attitude_target.throttle = fc_data.final_command.throttle;

        fc_data.rate_target.roll = fc_data.roll_cascade.rate_pid.output;
        fc_data.rate_target.pitch = fc_data.pitch_cascade.rate_pid.output;
        fc_data.rate_target.yaw = fc_data.yaw_cascade.rate_pid.output;
        fc_data.rate_target.throttle = fc_data.final_command.throttle;
    }
}

void flight_controller_set_mode(FlightMode mode)
{
    fc_data.requested_mode = mode;
    fc_data.mode = mode;
}

FlightMode flight_controller_get_mode(void)
{
    return fc_data.mode;
}

void flight_controller_arm(void)
{
    fc_data.armed = true;
}

void flight_controller_disarm(void)
{
    fc_data.armed = false;
}

bool flight_controller_is_armed(void)
{
    return fc_data.armed;
}

void flight_controller_set_rc_command(ControlCommand *cmd)
{
    fc_data.rc_command = *cmd;
}

void flight_controller_set_mavlink_command(ControlCommand *cmd)
{
    fc_data.mavlink_command = *cmd;
}

void flight_controller_get_attitude_target(ControlCommand *target)
{
    *target = fc_data.attitude_target;
}

void flight_controller_get_rate_target(ControlCommand *target)
{
    *target = fc_data.rate_target;
}

void flight_controller_get_final_command(ControlCommand *cmd)
{
    *cmd = fc_data.final_command;
}

void flight_controller_set_home_position(GPSPosition *pos)
{
    fc_data.home_position = *pos;
    fc_data.home_set = true;
}

void flight_controller_get_home_position(GPSPosition *pos)
{
    *pos = fc_data.home_position;
}

bool flight_controller_is_home_set(void)
{
    return fc_data.home_set;
}

void flight_controller_trigger_failsafe(void)
{
    fc_data.failsafe_active = true;
}

void flight_controller_clear_failsafe(void)
{
    fc_data.failsafe_active = false;
}

bool flight_controller_is_failsafe_active(void)
{
    return fc_data.failsafe_active;
}

FlightControlState flight_controller_get_state(void)
{
    return fc_data.state;
}

void flight_controller_get_position(PositionState *pos)
{
    if (pos != NULL) {
        *pos = fc_data.position;
    }
}

void flight_controller_set_target_altitude(float altitude)
{
    fc_data.position.altitude = altitude;
    fc_data.mode = FLIGHT_MODE_ALT_HOLD;
    fc_data.mavlink_command.throttle = CONSTRAIN(
        (altitude - fc_data.position.altitude) * 0.5f + 0.5f,
        0.0f, 1.0f
    );
}

void flight_controller_goto_position(float lat, float lng, float alt)
{
    fc_data.mode = FLIGHT_MODE_AUTO;
    fc_data.mavlink_command.roll = 0;
    fc_data.mavlink_command.pitch = 0;
    fc_data.mavlink_command.yaw = 0;
    fc_data.mavlink_command.throttle = 0.5f;
}

float flight_controller_get_heading(void)
{
    return fc_data.position.heading;
}

float flight_controller_get_throttle(void)
{
    return fc_data.final_command.throttle;
}

void flight_controller_get_pid_gains(PIDGainSet *gains)
{
    task_flight_control_get_pid_gains(gains);
}

void flight_controller_set_pid_gains(PIDGainSet *gains)
{
    task_flight_control_set_pid_gains(gains);
}

bool flight_controller_get_mavlink_command(ControlCommand *cmd)
{
    if (cmd == NULL) {
        return false;
    }
    *cmd = fc_data.mavlink_command;
    return true;
}
