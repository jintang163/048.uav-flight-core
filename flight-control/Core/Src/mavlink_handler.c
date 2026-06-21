#include "mavlink_handler.h"
#include "flight_controller.h"
#include "sensor_manager.h"
#include "mission_manager.h"
#include "task_attitude_estimation.h"
#include "motor_control.h"
#include "blackbox_logger.h"
#include "types.h"
#include "sbus_rc.h"
#include "link_manager.h"
#include "4g_driver.h"
#include "main.h"

#define HEARTBEAT_INTERVAL      1000
#define GCS_HEARTBEAT_TIMEOUT   5000

typedef struct {
    uint32_t heartbeat_interval;
    uint32_t last_heartbeat_sent;
    uint32_t last_heartbeat_received;
    bool gcs_connected;
    uint8_t system_id;
    uint8_t component_id;
    uint16_t mission_count;
    uint16_t mission_item_index;
    bool stream_attitude;
    bool stream_position;
    bool stream_battery;
    bool stream_sys_status;
    uint32_t last_attitude_sent;
    uint32_t last_position_sent;
    uint32_t last_battery_sent;
    uint32_t last_sys_status_sent;
    bool dual_link_enabled;
    uint8_t primary_link;
} MAVLinkHandlerData;

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
    mavlink_data.dual_link_enabled = true;
    mavlink_data.primary_link = MAVLINK_COMM_RADIO;
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
            if (current_rx_link == MAVLINK_COMM_RADIO) {
                link_manager_notify_heartbeat(LINK_TYPE_RADIO);
            } else {
                link_manager_notify_heartbeat(LINK_TYPE_4G);
            }
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

        case MAVLINK_MSG_ID_LOG_REQUEST_LIST:
            mavlink_handler_handle_log_request_list(msg);
            break;

        case MAVLINK_MSG_ID_LOG_REQUEST_DATA:
            mavlink_handler_handle_log_request_data(msg);
            break;

        case MAVLINK_MSG_ID_LOG_REQUEST_END:
            mavlink_handler_handle_log_request_end(msg);
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
    LinkType active_link;
    LinkStatus radio_status, lte_status;

    uint8_t base_mode = 0;
    if (flight_controller_is_armed()) {
        base_mode |= MAV_MODE_FLAG_SAFETY_ARMED;
    }
    base_mode |= MAV_MODE_FLAG_CUSTOM_MODE_ENABLED;

    uint32_t custom_mode = (uint32_t)mode;
    active_link = link_manager_get_active_link();
    link_manager_get_link_status(LINK_TYPE_RADIO, &radio_status);
    link_manager_get_link_status(LINK_TYPE_4G, &lte_status);

    custom_mode |= ((uint32_t)active_link << 8);

    mavlink_msg_heartbeat_pack(mavlink_data.system_id, mavlink_data.component_id, &msg,
                               MAV_TYPE_QUADROTOR, MAV_AUTOPILOT_GENERIC, base_mode, custom_mode,
                               flight_controller_get_state() == FCS_EMERGENCY ? MAV_STATE_CRITICAL : MAV_STATE_ACTIVE);

    mavlink_send_message(&msg);

    mavlink_send_link_status(active_link, &radio_status, &lte_status);
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

static uint8_t current_rx_link = MAVLINK_COMM_RADIO;

void mavlink_send_message(mavlink_message_t *msg)
{
    uint8_t buf[MAVLINK_MAX_PACKET_LEN];
    uint16_t len = mavlink_msg_to_send_buffer(buf, msg);

    if (msg->msgid == MAVLINK_MSG_ID_HEARTBEAT) {
        HAL_UART_Transmit(&huart3, buf, len, 10);
        HAL_UART_Transmit(&huart4, buf, len, 10);
    } else {
        LinkType active_link = link_manager_get_active_link();

        if (mavlink_data.dual_link_enabled) {
            HAL_UART_Transmit(&huart3, buf, len, 10);
            HAL_UART_Transmit(&huart4, buf, len, 10);
        } else {
            if (active_link == LINK_TYPE_RADIO) {
                HAL_UART_Transmit(&huart3, buf, len, 10);
            } else {
                HAL_UART_Transmit(&huart4, buf, len, 10);
            }
        }
    }
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

static uint16_t log_transfer_id = 0;
static uint32_t log_transfer_offset = 0;
static uint32_t log_transfer_size = 0;
static bool log_transfer_active = false;

void mavlink_handler_handle_log_request_list(mavlink_message_t *msg)
{
    mavlink_log_request_list_t req;
    mavlink_msg_log_request_list_decode(msg, &req);

    BlackboxInfo info;
    blackbox_get_info(&info);

    mavlink_send_log_entry(
        0,
        1,
        0,
        info.total_entries * sizeof(BlackboxLogEntry),
        info.start_time
    );
}

void mavlink_handler_handle_log_request_data(mavlink_message_t *msg)
{
    mavlink_log_request_data_t req;
    mavlink_msg_log_request_data_decode(msg, &req);

    BlackboxInfo info;
    blackbox_get_info(&info);

    uint32_t total_size = info.total_entries * sizeof(BlackboxLogEntry);

    if (req.id != 0) {
        return;
    }

    log_transfer_id = req.id;
    log_transfer_offset = req.offset;
    log_transfer_size = total_size;
    log_transfer_active = true;

    uint32_t bytes_remaining = total_size - req.offset;
    uint32_t bytes_to_send = (bytes_remaining < req.count) ? bytes_remaining : req.count;
    uint8_t chunk_size = 90;

    for (uint32_t sent = 0; sent < bytes_to_send; sent += chunk_size) {
        uint32_t chunk_len = bytes_to_send - sent;
        if (chunk_len > chunk_size) {
            chunk_len = chunk_size;
        }

        uint8_t data[90];
        blackbox_read_data(req.offset + sent, data, chunk_len);

        mavlink_send_log_data(
            req.id,
            req.offset + sent,
            (uint8_t)chunk_len,
            data
        );
    }
}

void mavlink_handler_handle_log_request_end(mavlink_message_t *msg)
{
    log_transfer_active = false;
    log_transfer_id = 0;
    log_transfer_offset = 0;
    log_transfer_size = 0;
}

void mavlink_send_log_entry(uint16_t id, uint32_t num_logs, uint32_t latest_log_num,
                            uint32_t size, uint32_t time_utc)
{
    mavlink_message_t msg;
    mavlink_msg_log_entry_pack(
        mavlink_data.system_id,
        mavlink_data.component_id,
        &msg,
        id,
        num_logs,
        latest_log_num,
        size,
        time_utc
    );
    mavlink_send_message(&msg);
}

void mavlink_send_log_data(uint16_t id, uint32_t offset, uint8_t count, const uint8_t *data)
{
    mavlink_message_t msg;
    uint8_t log_data[90];
    
    memset(log_data, 0, sizeof(log_data));
    if (count > 0 && data != NULL) {
        memcpy(log_data, data, count > 90 ? 90 : count);
    }
    
    mavlink_msg_log_data_pack(
        mavlink_data.system_id,
        mavlink_data.component_id,
        &msg,
        id,
        offset,
        count,
        log_data
    );
    mavlink_send_message(&msg);
}

void mavlink_send_statustext(uint8_t severity, const char *text)
{
    mavlink_message_t msg;
    char status_text[50];

    memset(status_text, 0, sizeof(status_text));
    if (text != NULL) {
        strncpy(status_text, text, sizeof(status_text) - 1);
    }

    mavlink_msg_statustext_pack(
        mavlink_data.system_id,
        mavlink_data.component_id,
        &msg,
        severity,
        status_text,
        0, 0
    );

    mavlink_send_message(&msg);
}

void mavlink_send_rc_channels_raw(void)
{
    mavlink_message_t msg;
    SBUS_Data sbus_data;

    sbus_rc_get_data(&sbus_data);

    uint16_t chan[16];
    for (int i = 0; i < 16; i++) {
        chan[i] = sbus_data.channels[i];
    }

    mavlink_msg_rc_channels_raw_pack(
        mavlink_data.system_id,
        mavlink_data.component_id,
        &msg,
        HAL_GetTick() / 1000,
        0,
        chan[0], chan[1], chan[2], chan[3],
        chan[4], chan[5], chan[6], chan[7],
        chan[8], chan[9], chan[10], chan[11],
        chan[12], chan[13], chan[14], chan[15],
        sbus_data.signal_loss ? 0 : 1,
        0
    );

    mavlink_send_message(&msg);
}

void mavlink_send_local_position_ned(void)
{
    mavlink_message_t msg;
    PositionState pos;

    sensor_manager_get_position(&pos);

    mavlink_msg_local_position_ned_pack(
        mavlink_data.system_id,
        mavlink_data.component_id,
        &msg,
        HAL_GetTick() / 1000,
        0.0f, 0.0f, 0.0f,
        pos.velocity.vn,
        pos.velocity.ve,
        pos.velocity.vd
    );

    mavlink_send_message(&msg);
}

void mavlink_send_altitude(void)
{
    mavlink_message_t msg;
    PositionState pos;

    sensor_manager_get_position(&pos);

    mavlink_msg_altitude_pack(
        mavlink_data.system_id,
        mavlink_data.component_id,
        &msg,
        HAL_GetTick() / 1000,
        pos.altitude / 1000.0f,
        pos.altitude / 1000.0f,
        0.0f, 0.0f, 0.0f, 0.0f
    );

    mavlink_send_message(&msg);
}

void mavlink_send_vfr_hud(void)
{
    mavlink_message_t msg;
    PositionState pos;
    BatteryState battery;

    sensor_manager_get_position(&pos);
    sensor_manager_get_battery(&battery);

    mavlink_msg_vfr_hud_pack(
        mavlink_data.system_id,
        mavlink_data.component_id,
        &msg,
        pos.ground_speed,
        pos.ground_speed,
        pos.heading,
        pos.velocity.vd * -1.0f,
        battery.voltage,
        0.0f
    );

    mavlink_send_message(&msg);
}

#define MAV_COMP_ID_TELEMETRY_RADIO  68
#define MAV_COMP_ID_UDP_BRIDGE       240

static void mavlink_send_radio_status(uint8_t comp_id, int8_t rssi, uint8_t rx_errors,
                                      uint16_t latency_ms, uint8_t packet_loss_pct,
                                      uint32_t bytes_sent, uint32_t bytes_received)
{
    mavlink_message_t msg;
    mavlink_msg_radio_status_pack(
        mavlink_data.system_id,
        comp_id,
        &msg,
        (uint8_t)(rssi + 128),
        (uint8_t)(rssi + 128),
        100,
        (int8_t)(rssi / 2),
        (int8_t)(rssi / 2),
        rx_errors,
        (uint16_t)(packet_loss_pct * 10)
    );
    mavlink_send_message(&msg);
}

static void mavlink_send_named_value_int(const char *name, int32_t value)
{
    mavlink_message_t msg;
    char name_buf[16];

    memset(name_buf, 0, sizeof(name_buf));
    if (name != NULL) {
        strncpy(name_buf, name, sizeof(name_buf) - 1);
    }

    mavlink_msg_named_value_int_pack(
        mavlink_data.system_id,
        mavlink_data.component_id,
        &msg,
        HAL_GetTick(),
        name_buf,
        value
    );
    mavlink_send_message(&msg);
}

void mavlink_send_link_status(LinkType active_link, LinkStatus *radio_status, LinkStatus *lte_status)
{
    LTEStatus lte_drv_status;

    if (radio_status != NULL) {
        mavlink_send_radio_status(
            MAV_COMP_ID_TELEMETRY_RADIO,
            radio_status->quality.rssi,
            0,
            (uint16_t)radio_status->quality.latency_ms,
            (uint8_t)radio_status->quality.packet_loss,
            radio_status->bytes_sent,
            radio_status->bytes_received
        );
    }

    if (lte_status != NULL) {
        mavlink_send_radio_status(
            MAV_COMP_ID_UDP_BRIDGE,
            lte_status->quality.rssi,
            0,
            (uint16_t)lte_status->quality.latency_ms,
            (uint8_t)lte_status->quality.packet_loss,
            lte_status->bytes_sent,
            lte_status->bytes_received
        );

        _4g_driver_get_status(&lte_drv_status);
        mavlink_send_named_value_int("lte_nettype", (int32_t)lte_drv_status.network_type);
    }

    mavlink_send_named_value_int("link_active", (int32_t)active_link);

    if (radio_status != NULL && lte_status != NULL) {
        char text[120];
        snprintf(text, sizeof(text),
                 "LINK: active=%s | RADIO: rssi=%d | 4G: rssi=%d net=%d",
                 link_type_to_string(active_link),
                 radio_status->quality.rssi,
                 lte_status->quality.rssi,
                 lte_drv_status.network_type);
        mavlink_send_statustext(MAV_SEVERITY_INFO, text);
    }
}

void mavlink_send_mission_count(uint16_t count)
{
    mavlink_message_t msg;
    mavlink_msg_mission_count_pack(
        mavlink_data.system_id,
        mavlink_data.component_id,
        &msg,
        0, 0, count, 0
    );
    mavlink_send_message(&msg);
}

void mavlink_send_mission_item(uint16_t index)
{
    MissionItem item;
    mavlink_message_t msg;

    mission_manager_get_waypoint(index, &item);

    mavlink_msg_mission_item_int_pack(
        mavlink_data.system_id,
        mavlink_data.component_id,
        &msg,
        0, 0, index, 0, item.type,
        0, 0, 0, 0, item.hold_time, item.radius, item.heading,
        item.lat, item.lon, item.alt, 0
    );
    mavlink_send_message(&msg);
}

void mavlink_set_target_attitude(float roll, float pitch, float yaw, float thrust)
{
    ControlCommand cmd;
    cmd.roll = roll;
    cmd.pitch = pitch;
    cmd.yaw = yaw;
    cmd.throttle = thrust;
    flight_controller_set_mavlink_command(&cmd);
}

void mavlink_set_target_position(int32_t lat, int32_t lon, int32_t alt)
{
    PositionState pos;
    sensor_manager_get_position(&pos);
    pos.position.lat = lat;
    pos.position.lon = lon;
    pos.position.alt = alt;
}

bool mavlink_get_command(ControlCommand *cmd)
{
    if (cmd == NULL) {
        return false;
    }
    return flight_controller_get_mavlink_command(cmd);
}

void mavlink_send_command_ack(uint16_t cmd, uint8_t result)
{
    mavlink_message_t msg;
    mavlink_msg_command_ack_pack(
        mavlink_data.system_id,
        mavlink_data.component_id,
        &msg,
        cmd, result, 0, 0, 0, 0
    );
    mavlink_send_message(&msg);
}

void mavlink_send_mission_ack(uint8_t type)
{
    mavlink_message_t msg;
    mavlink_msg_mission_ack_pack(
        mavlink_data.system_id,
        mavlink_data.component_id,
        &msg,
        0, 0, type, 0
    );
    mavlink_send_message(&msg);
}

void mavlink_set_primary_link(uint8_t link)
{
    if (link == MAVLINK_COMM_RADIO || link == MAVLINK_COMM_LTE) {
        mavlink_data.primary_link = link;
    }
}

void mavlink_receive_byte_from_link(uint8_t link, uint8_t byte)
{
    mavlink_message_t msg;
    mavlink_status_t status;

    current_rx_link = link;

    if (mavlink_parse_char(MAVLINK_COMM_0, byte, &msg, &status)) {
        mavlink_handler_process_message(&msg);
    }
}
