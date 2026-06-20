#include "mavlink_handler.h"
#include "flight_controller.h"
#include "sensor_manager.h"
#include "mission_manager.h"
#include "task_attitude_estimation.h"
#include "motor_control.h"
#include "types.h"

static MAVLinkHandlerData mavlink_data;

void mavlink_handler_init(void)
{
    memset(&mavlink_data, 0, sizeof(MAVLinkHandlerData));
    mavlink_data.heartbeat_interval = HEARTBEAT_INTERVAL;
    mavlink_data.last_heartbeat_sent = 0;
    mavlink_data.last_heartbeat_received = 0;
    mavlink_data.gcs_connected = false;
    mavlink_data.system_id = MAV_SYS_ID;
    mavlink_data.component_id = MAV_COMP_ID;
}

void mavlink_handler_update(void)
{
    uint32_t now = HAL_GetTick();

    if (now - mavlink_data.last_heartbeat_sent >= mavlink_data.heartbeat_interval) {
        mavlink_send_heartbeat();
        mavlink_data.last_heartbeat_sent = now;
    }

    if (now - mavlink_data.last_heartbeat_received > GCS_HEARTBEAT_TIMEOUT) {
        mavlink_data.gcs_connected = false;
    }

    if (mavlink_data.stream_attitude && (now - mavlink_data.last_attitude_sent >= 200)) {
        mavlink_send_attitude();
        mavlink_data.last_attitude_sent = now;
    }

    if (mavlink_data.stream_position && (now - mavlink_data.last_position_sent >= 200)) {
        mavlink_send_global_position_int();
        mavlink_data.last_position_sent = now;
    }

    if (mavlink_data.stream_battery && (now - mavlink_data.last_battery_sent >= 1000)) {
        mavlink_send_battery_status();
        mavlink_data.last_battery_sent = now;
    }

    if (mavlink_data.stream_sys_status && (now - mavlink_data.last_sys_status_sent >= 1000)) {
        mavlink_send_sys_status();
        mavlink_data.last_sys_status_sent = now;
    }
}

void mavlink_receive_byte(uint8_t byte)
{
    mavlink_message_t msg;
    mavlink_status_t status;

    if (mavlink_parse_char(MAVLINK_COMM_0, byte, &msg, &status)) {
        mavlink_handler_process_message(&msg);
    }
}

void mavlink_handler_process_message(mavlink_message_t *msg)
{
    switch (msg->msgid) {
        case MAVLINK_MSG_ID_HEARTBEAT:
            mavlink_data.last_heartbeat_received = HAL_GetTick();
            mavlink_data.gcs_connected = true;
            break;

        case MAVLINK_MSG_ID_COMMAND_LONG:
            mavlink_handler_handle_command_long(msg);
            break;

        case MAVLINK_MSG_ID_MANUAL_CONTROL:
            mavlink_handler_handle_manual_control(msg);
            break;

        case MAVLINK_MSG_ID_SET_MODE:
            mavlink_handler_handle_set_mode(msg);
            break;

        case MAVLINK_MSG_ID_MISSION_COUNT:
            mavlink_handler_handle_mission_count(msg);
            break;

        case MAVLINK_MSG_ID_MISSION_ITEM_INT:
            mavlink_handler_handle_mission_item_int(msg);
            break;

        case MAVLINK_MSG_ID_MISSION_CLEAR_ALL:
            mission_manager_clear_mission();
            mavlink_send_mission_ack(msg->sysid, msg->compid, MAV_MISSION_ACCEPTED);
            break;

        case MAVLINK_MSG_ID_MISSION_SET_CURRENT:
            {
                mavlink_mission_set_current_t current;
                mavlink_msg_mission_set_current_decode(msg, &current);
                mission_manager_goto_waypoint(current.seq);
            }
            break;

        default:
            break;
    }
}

void mavlink_handler_handle_command_long(mavlink_message_t *msg)
{
    mavlink_command_long_t cmd;
    mavlink_msg_command_long_decode(msg, &cmd);

    switch (cmd.command) {
        case MAV_CMD_COMPONENT_ARM_DISARM:
            if (cmd.param1 == 1.0f) {
                flight_controller_arm();
                mavlink_send_command_ack(msg->sysid, msg->compid, cmd.command, MAV_RESULT_ACCEPTED);
            } else {
                flight_controller_disarm();
                mavlink_send_command_ack(msg->sysid, msg->compid, cmd.command, MAV_RESULT_ACCEPTED);
            }
            break;

        case MAV_CMD_NAV_TAKEOFF:
            mission_manager_start_takeoff(cmd.param7);
            mavlink_send_command_ack(msg->sysid, msg->compid, cmd.command, MAV_RESULT_ACCEPTED);
            break;

        case MAV_CMD_NAV_LAND:
            mission_manager_start_land();
            mavlink_send_command_ack(msg->sysid, msg->compid, cmd.command, MAV_RESULT_ACCEPTED);
            break;

        case MAV_CMD_NAV_RETURN_TO_LAUNCH:
            mission_manager_start_rtl();
            mavlink_send_command_ack(msg->sysid, msg->compid, cmd.command, MAV_RESULT_ACCEPTED);
            break;

        case MAV_CMD_MISSION_START:
            mission_manager_start();
            mavlink_send_command_ack(msg->sysid, msg->compid, cmd.command, MAV_RESULT_ACCEPTED);
            break;

        case MAV_CMD_MISSION_STOP:
            mission_manager_stop();
            mavlink_send_command_ack(msg->sysid, msg->compid, cmd.command, MAV_RESULT_ACCEPTED);
            break;

        case MAV_CMD_DO_SET_HOME:
            {
                GPSPosition current_pos;
                GPSVelocity current_vel;
                sensor_manager_get_gps(&current_pos, &current_vel);

                if (cmd.param1 == 1.0f) {
                    flight_controller_set_home_position(&current_pos);
                    mission_manager_set_home(&current_pos);
                    mavlink_send_command_ack(msg->sysid, msg->compid, cmd.command, MAV_RESULT_ACCEPTED);
                } else {
                    GPSPosition home_pos;
                    home_pos.lat = (int32_t)(cmd.param5 * 1e7);
                    home_pos.lon = (int32_t)(cmd.param6 * 1e7);
                    home_pos.alt = (int32_t)(cmd.param7 * 1000);
                    flight_controller_set_home_position(&home_pos);
                    mission_manager_set_home(&home_pos);
                    mavlink_send_command_ack(msg->sysid, msg->compid, cmd.command, MAV_RESULT_ACCEPTED);
                }
            }
            break;

        default:
            mavlink_send_command_ack(msg->sysid, msg->compid, cmd.command, MAV_RESULT_UNSUPPORTED);
            break;
    }
}

void mavlink_handler_handle_manual_control(mavlink_message_t *msg)
{
    mavlink_manual_control_t manual;
    mavlink_msg_manual_control_decode(msg, &manual);

    ControlCommand cmd;
    cmd.roll = (float)manual.x / 1000.0f * MAX_TILT_ANGLE;
    cmd.pitch = (float)manual.y / 1000.0f * MAX_TILT_ANGLE;
    cmd.yaw = (float)manual.r / 1000.0f * MAX_YAW_RATE;
    cmd.throttle = (float)manual.z / 1000.0f;

    flight_controller_set_mavlink_command(&cmd);

    if (manual.buttons & 1) {
        flight_controller_arm();
    }
    if (manual.buttons & 2) {
        flight_controller_disarm();
    }
}

void mavlink_handler_handle_set_mode(mavlink_message_t *msg)
{
    mavlink_set_mode_t mode;
    mavlink_msg_set_mode_decode(msg, &mode);

    FlightMode flight_mode;
    switch (mode.custom_mode) {
        case 0:
            flight_mode = FLIGHT_MODE_MANUAL;
            break;
        case 1:
            flight_mode = FLIGHT_MODE_ALT_HOLD;
            break;
        case 2:
            flight_mode = FLIGHT_MODE_POS_HOLD;
            break;
        case 3:
            flight_mode = FLIGHT_MODE_AUTO;
            break;
        case 4:
            flight_mode = FLIGHT_MODE_RTL;
            break;
        case 5:
            flight_mode = FLIGHT_MODE_LAND;
            break;
        default:
            flight_mode = FLIGHT_MODE_MANUAL;
            break;
    }

    flight_controller_set_mode(flight_mode);
    mavlink_send_command_ack(msg->sysid, msg->compid, MAV_CMD_DO_SET_MODE, MAV_RESULT_ACCEPTED);
}

void mavlink_handler_handle_mission_count(mavlink_message_t *msg)
{
    mavlink_mission_count_t count_msg;
    mavlink_msg_mission_count_decode(msg, &count_msg);

    mission_manager_clear_mission();
    mavlink_data.mission_count = count_msg.count;
    mavlink_data.mission_item_index = 0;
    mavlink_send_mission_request_int(msg->sysid, msg->compid, 0);
}

void mavlink_handler_handle_mission_item_int(mavlink_message_t *msg)
{
    mavlink_mission_item_int_t item;
    mavlink_msg_mission_item_int_decode(msg, &item);

    if (item.seq != mavlink_data.mission_item_index) {
        return;
    }

    MissionItem mission_item;
    mission_item.type = item.command;
    mission_item.lat = item.x;
    mission_item.lon = item.y;
    mission_item.alt = item.z;
    mission_item.heading = item.param4;
    mission_item.hold_time = (uint16_t)item.param1;
    mission_item.radius = item.param2;

    mission_manager_add_waypoint(&mission_item);
    mavlink_data.mission_item_index++;

    if (mavlink_data.mission_item_index < mavlink_data.mission_count) {
        mavlink_send_mission_request_int(msg->sysid, msg->compid, mavlink_data.mission_item_index);
    } else {
        mavlink_send_mission_ack(msg->sysid, msg->compid, MAV_MISSION_ACCEPTED);
    }
}

void mavlink_send_heartbeat(void)
{
    mavlink_message_t msg;
    FlightMode mode = flight_controller_get_mode();

    uint8_t base_mode = 0;
    if (flight_controller_is_armed()) {
        base_mode |= MAV_MODE_FLAG_SAFETY_ARMED;
    }
    base_mode |= MAV_MODE_FLAG_CUSTOM_MODE_ENABLED;

    uint32_t custom_mode = (uint32_t)mode;

    mavlink_msg_heartbeat_pack(mavlink_data.system_id, mavlink_data.component_id, &msg,
                               MAV_TYPE_QUADROTOR, MAV_AUTOPILOT_GENERIC, base_mode, custom_mode,
                               flight_controller_get_state() == FCS_EMERGENCY ? MAV_STATE_CRITICAL : MAV_STATE_ACTIVE);

    mavlink_send_message(&msg);
}

void mavlink_send_attitude(void)
{
    mavlink_message_t msg;
    AttitudeState attitude;
    task_attitude_estimation_get_state(&attitude);

    mavlink_msg_attitude_pack(mavlink_data.system_id, mavlink_data.component_id, &msg,
                              HAL_GetTick() / 1000,
                              attitude.euler.roll,
                              attitude.euler.pitch,
                              attitude.euler.yaw,
                              attitude.angular_velocity.x,
                              attitude.angular_velocity.y,
                              attitude.angular_velocity.z);

    mavlink_send_message(&msg);
}

void mavlink_send_global_position_int(void)
{
    mavlink_message_t msg;
    GPSPosition pos;
    GPSVelocity vel;
    sensor_manager_get_gps(&pos, &vel);

    mavlink_msg_global_position_int_pack(mavlink_data.system_id, mavlink_data.component_id, &msg,
                                         HAL_GetTick() / 1000,
                                         pos.lat,
                                         pos.lon,
                                         pos.alt,
                                         pos.alt,
                                         (int16_t)(vel.vn * 100),
                                         (int16_t)(vel.ve * 100),
                                         (int16_t)(vel.vd * 100),
                                         (uint16_t)(atan2f(vel.ve, vel.vn) * 100));

    mavlink_send_message(&msg);
}

void mavlink_send_battery_status(void)
{
    mavlink_message_t msg;
    BatteryState battery;
    sensor_manager_get_battery(&battery);

    int16_t voltages[10];
    voltages[0] = (int16_t)(battery.voltage * 1000);

    mavlink_msg_battery_status_pack(mavlink_data.system_id, mavlink_data.component_id, &msg,
                                    0,
                                    MAV_BATTERY_FUNCTION_ALL,
                                    MAV_BATTERY_TYPE_LIPO,
                                    1,
                                    voltages,
                                    (int16_t)(battery.current * 100),
                                    (int32_t)battery.capacity_used,
                                    -1,
                                    (int8_t)battery.battery_percent,
                                    0, 0, 0);

    mavlink_send_message(&msg);
}

void mavlink_send_sys_status(void)
{
    mavlink_message_t msg;

    uint32_t onboard_control_sensors_present = 0;
    onboard_control_sensors_present |= (1 << MAV_SYS_STATUS_SENSOR_3D_GYRO);
    onboard_control_sensors_present |= (1 << MAV_SYS_STATUS_SENSOR_3D_ACCEL);
    onboard_control_sensors_present |= (1 << MAV_SYS_STATUS_SENSOR_3D_MAG);
    onboard_control_sensors_present |= (1 << MAV_SYS_STATUS_SENSOR_GPS);
    onboard_control_sensors_present |= (1 << MAV_SYS_STATUS_SENSOR_ABSOLUTE_PRESSURE);

    uint32_t onboard_control_sensors_enabled = onboard_control_sensors_present;

    uint32_t onboard_control_sensors_health = 0;
    if (!sensor_manager_check_imu_timeout()) {
        onboard_control_sensors_health |= (1 << MAV_SYS_STATUS_SENSOR_3D_GYRO);
        onboard_control_sensors_health |= (1 << MAV_SYS_STATUS_SENSOR_3D_ACCEL);
    }
    if (!sensor_manager_check_mag_timeout()) {
        onboard_control_sensors_health |= (1 << MAV_SYS_STATUS_SENSOR_3D_MAG);
    }
    if (!sensor_manager_check_gps_timeout()) {
        onboard_control_sensors_health |= (1 << MAV_SYS_STATUS_SENSOR_GPS);
    }
    if (!sensor_manager_check_baro_timeout()) {
        onboard_control_sensors_health |= (1 << MAV_SYS_STATUS_SENSOR_ABSOLUTE_PRESSURE);
    }

    mavlink_msg_sys_status_pack(mavlink_data.system_id, mavlink_data.component_id, &msg,
                                onboard_control_sensors_present,
                                onboard_control_sensors_enabled,
                                onboard_control_sensors_health,
                                0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0);

    mavlink_send_message(&msg);
}

void mavlink_send_command_ack(uint8_t sysid, uint8_t compid, uint16_t command, uint8_t result)
{
    mavlink_message_t msg;
    mavlink_msg_command_ack_pack(mavlink_data.system_id, mavlink_data.component_id, &msg,
                                 command, result, 0, 0, sysid, compid);
    mavlink_send_message(&msg);
}

void mavlink_send_mission_request_int(uint8_t sysid, uint8_t compid, uint16_t seq)
{
    mavlink_message_t msg;
    mavlink_msg_mission_request_int_pack(mavlink_data.system_id, mavlink_data.component_id, &msg,
                                         sysid, compid, seq, 0);
    mavlink_send_message(&msg);
}

void mavlink_send_mission_ack(uint8_t sysid, uint8_t compid, uint8_t type)
{
    mavlink_message_t msg;
    mavlink_msg_mission_ack_pack(mavlink_data.system_id, mavlink_data.component_id, &msg,
                                 sysid, compid, type, 0);
    mavlink_send_message(&msg);
}

void mavlink_send_message(mavlink_message_t *msg)
{
    uint8_t buf[MAVLINK_MAX_PACKET_LEN];
    uint16_t len = mavlink_msg_to_send_buffer(buf, msg);

    HAL_UART_Transmit(&huart3, buf, len, 10);
}

bool mavlink_handler_is_gcs_connected(void)
{
    return mavlink_data.gcs_connected;
}

void mavlink_send_geofence_violation(uint16_t fence_id, uint8_t violation_type,
                                     uint8_t severity, float distance,
                                     float lat, float lon, float alt)
{
    char text[128];
    snprintf(text, sizeof(text),
             "GEOFENCE VIOLATION: fence=%d type=%d sev=%d dist=%.1fm pos=%.6f,%.6f,%.1fm",
             fence_id, violation_type, severity, distance, lat, lon, alt);
    
    mavlink_send_statustext(severity == 0 ? MAV_SEVERITY_WARNING : 
                            severity == 1 ? MAV_SEVERITY_CRITICAL : MAV_SEVERITY_EMERGENCY,
                            text);
}
