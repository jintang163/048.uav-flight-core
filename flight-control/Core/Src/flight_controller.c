#include "flight_controller.h"
#include "task_attitude_estimation.h"
#include "sensor_manager.h"
#include "motor_control.h"
#include "pwm_output.h"

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
