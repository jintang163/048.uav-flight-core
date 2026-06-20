#ifndef MAVLINK_MESSAGES_COMMON_H
#define MAVLINK_MESSAGES_COMMON_H

#include <stdint.h>

#define MAVLINK_MSG_ID_HEARTBEAT          0
#define MAVLINK_MSG_ID_SYS_STATUS         1
#define MAVLINK_MSG_ID_GPS_RAW_INT        24
#define MAVLINK_MSG_ID_ATTITUDE           30
#define MAVLINK_MSG_ID_GLOBAL_POSITION_INT 33
#define MAVLINK_MSG_ID_RC_CHANNELS_RAW    35
#define MAVLINK_MSG_ID_MISSION_ITEM       39
#define MAVLINK_MSG_ID_MISSION_CURRENT    42
#define MAVLINK_MSG_ID_MISSION_COUNT      44
#define MAVLINK_MSG_ID_COMMAND_LONG       76
#define MAVLINK_MSG_ID_COMMAND_ACK        77
#define MAVLINK_MSG_ID_BATTERY_STATUS     147

#define MAV_TYPE_GENERIC                  0
#define MAV_TYPE_FIXED_WING               1
#define MAV_TYPE_QUADROTOR                2
#define MAV_TYPE_GROUND_ROVER             10
#define MAV_TYPE_SURFACE_BOAT             11
#define MAV_TYPE_SUBMARINE                12
#define MAV_TYPE_HEXAROTOR                13
#define MAV_TYPE_OCTOROTOR                14

#define MAV_AUTOPILOT_GENERIC             0
#define MAV_AUTOPILOT_PX4                 12
#define MAV_AUTOPILOT_ARMAZILA            15

#define MAV_STATE_UNINIT                  0
#define MAV_STATE_BOOT                    1
#define MAV_STATE_CALIBRATING             2
#define MAV_STATE_STANDBY                 3
#define MAV_STATE_ACTIVE                  4
#define MAV_STATE_CRITICAL                5
#define MAV_STATE_EMERGENCY               6
#define MAV_STATE_POWEROFF                7

#define MAV_CMD_NAV_WAYPOINT              16
#define MAV_CMD_NAV_TAKEOFF               22
#define MAV_CMD_NAV_LAND                  21
#define MAV_CMD_NAV_LOITER_UNLIM          17
#define MAV_CMD_NAV_RETURN_TO_LAUNCH      20
#define MAV_CMD_DO_SET_MODE               176
#define MAV_CMD_COMPONENT_ARM_DISARM      400
#define MAV_CMD_REQUEST_AUTOPILOT_CAPABILITIES 520

#define MAV_RESULT_ACCEPTED               0
#define MAV_RESULT_TEMPORARILY_REJECTED   1
#define MAV_RESULT_DENIED                 2
#define MAV_RESULT_UNSUPPORTED            3
#define MAV_RESULT_FAILED                 4

#define MAV_MODE_FLAG_CUSTOM_MODE_ENABLED 1
#define MAV_MODE_FLAG_TEST_ENABLED        2
#define MAV_MODE_FLAG_AUTO_ENABLED        4
#define MAV_MODE_FLAG_GUIDED_ENABLED      8
#define MAV_MODE_FLAG_STABILIZE_ENABLED   16
#define MAV_MODE_FLAG_HIL_ENABLED         32
#define MAV_MODE_FLAG_MANUAL_INPUT_ENABLED 64
#define MAV_MODE_FLAG_SAFETY_ARMED        128

typedef struct __attribute__((packed)) {
    uint32_t custom_mode;
    uint8_t type;
    uint8_t autopilot;
    uint8_t base_mode;
    uint8_t system_status;
    uint8_t mavlink_version;
} mavlink_heartbeat_t;

typedef struct __attribute__((packed)) {
    uint32_t onboard_control_sensors_present;
    uint32_t onboard_control_sensors_enabled;
    uint32_t onboard_control_sensors_health;
    uint16_t load;
    uint16_t voltage_battery;
    int16_t current_battery;
    int8_t battery_remaining;
    uint16_t drop_rate_comm;
    uint16_t errors_comm;
    uint16_t errors_count1;
    uint16_t errors_count2;
    uint16_t errors_count3;
    uint16_t errors_count4;
} mavlink_sys_status_t;

typedef struct __attribute__((packed)) {
    uint64_t time_usec;
    int32_t lat;
    int32_t lon;
    int32_t alt;
    int32_t eph;
    int32_t epv;
    uint16_t vel;
    int16_t vn;
    int16_t ve;
    int16_t vd;
    uint16_t cog;
    uint8_t fix_type;
    uint8_t satellites_visible;
    int32_t alt_ellipsoid;
    uint32_t h_acc;
    uint32_t v_acc;
    uint32_t vel_acc;
    uint32_t hdg_acc;
} mavlink_gps_raw_int_t;

typedef struct __attribute__((packed)) {
    uint64_t time_boot_ms;
    float roll;
    float pitch;
    float yaw;
    float rollspeed;
    float pitchspeed;
    float yawspeed;
} mavlink_attitude_t;

typedef struct __attribute__((packed)) {
    uint32_t time_boot_ms;
    int32_t lat;
    int32_t lon;
    int32_t alt;
    int32_t relative_alt;
    int16_t vx;
    int16_t vy;
    int16_t vz;
    uint16_t hdg;
} mavlink_global_position_int_t;

typedef struct __attribute__((packed)) {
    uint32_t time_boot_ms;
    uint16_t chan1_raw;
    uint16_t chan2_raw;
    uint16_t chan3_raw;
    uint16_t chan4_raw;
    uint16_t chan5_raw;
    uint16_t chan6_raw;
    uint16_t chan7_raw;
    uint16_t chan8_raw;
    uint8_t port;
    uint8_t rssi;
} mavlink_rc_channels_raw_t;

typedef struct __attribute__((packed)) {
    float param1;
    float param2;
    float param3;
    float param4;
    double x;
    double y;
    double z;
    uint16_t seq;
    uint8_t frame;
    uint16_t command;
    uint8_t current;
    uint8_t autocontinue;
    uint8_t mission_type;
    uint8_t target_system;
    uint8_t target_component;
} mavlink_mission_item_t;

typedef struct __attribute__((packed)) {
    uint16_t seq;
    uint8_t target_system;
    uint8_t target_component;
} mavlink_mission_current_t;

typedef struct __attribute__((packed)) {
    uint16_t count;
    uint8_t target_system;
    uint8_t target_component;
    uint8_t mission_type;
} mavlink_mission_count_t;

typedef struct __attribute__((packed)) {
    float param1;
    float param2;
    float param3;
    float param4;
    float param5;
    float param6;
    float param7;
    uint16_t command;
    uint8_t target_system;
    uint8_t target_component;
    uint8_t confirmation;
} mavlink_command_long_t;

typedef struct __attribute__((packed)) {
    uint16_t command;
    uint8_t result;
    uint8_t progress;
    uint8_t result_param2;
    uint8_t target_system;
    uint8_t target_component;
    int32_t result_param1;
} mavlink_command_ack_t;

typedef struct __attribute__((packed)) {
    uint32_t time_boot_ms;
    uint8_t id;
    uint16_t battery_function;
    uint16_t type;
    int16_t temperature;
    uint16_t voltages[10];
    int16_t current_battery;
    int32_t current_consumed;
    int32_t energy_consumed;
    int8_t battery_remaining;
    int32_t time_remaining;
    uint8_t charge_state;
    uint32_t voltages_ext[4];
    uint8_t mode;
    uint8_t fault_bitmask;
} mavlink_battery_status_t;

#define MAVLINK_MSG_HEARTBEAT_LEN          9
#define MAVLINK_MSG_SYS_STATUS_LEN         31
#define MAVLINK_MSG_GPS_RAW_INT_LEN        52
#define MAVLINK_MSG_ATTITUDE_LEN           28
#define MAVLINK_MSG_GLOBAL_POSITION_INT_LEN 28
#define MAVLINK_MSG_RC_CHANNELS_RAW_LEN    22
#define MAVLINK_MSG_MISSION_ITEM_LEN       38
#define MAVLINK_MSG_MISSION_CURRENT_LEN    4
#define MAVLINK_MSG_MISSION_COUNT_LEN      5
#define MAVLINK_MSG_COMMAND_LONG_LEN       33
#define MAVLINK_MSG_COMMAND_ACK_LEN        10
#define MAVLINK_MSG_BATTERY_STATUS_LEN     54

void mavlink_msg_heartbeat_encode(uint8_t system_id, uint8_t component_id,
                                  mavlink_message_t* msg, const mavlink_heartbeat_t* data);
void mavlink_msg_heartbeat_decode(const mavlink_message_t* msg, mavlink_heartbeat_t* data);

void mavlink_msg_sys_status_encode(uint8_t system_id, uint8_t component_id,
                                   mavlink_message_t* msg, const mavlink_sys_status_t* data);
void mavlink_msg_sys_status_decode(const mavlink_message_t* msg, mavlink_sys_status_t* data);

void mavlink_msg_gps_raw_int_encode(uint8_t system_id, uint8_t component_id,
                                    mavlink_message_t* msg, const mavlink_gps_raw_int_t* data);
void mavlink_msg_gps_raw_int_decode(const mavlink_message_t* msg, mavlink_gps_raw_int_t* data);

void mavlink_msg_attitude_encode(uint8_t system_id, uint8_t component_id,
                                 mavlink_message_t* msg, const mavlink_attitude_t* data);
void mavlink_msg_attitude_decode(const mavlink_message_t* msg, mavlink_attitude_t* data);

void mavlink_msg_global_position_int_encode(uint8_t system_id, uint8_t component_id,
                                            mavlink_message_t* msg, const mavlink_global_position_int_t* data);
void mavlink_msg_global_position_int_decode(const mavlink_message_t* msg, mavlink_global_position_int_t* data);

void mavlink_msg_rc_channels_raw_encode(uint8_t system_id, uint8_t component_id,
                                        mavlink_message_t* msg, const mavlink_rc_channels_raw_t* data);
void mavlink_msg_rc_channels_raw_decode(const mavlink_message_t* msg, mavlink_rc_channels_raw_t* data);

void mavlink_msg_mission_item_encode(uint8_t system_id, uint8_t component_id,
                                     mavlink_message_t* msg, const mavlink_mission_item_t* data);
void mavlink_msg_mission_item_decode(const mavlink_message_t* msg, mavlink_mission_item_t* data);

void mavlink_msg_mission_current_encode(uint8_t system_id, uint8_t component_id,
                                        mavlink_message_t* msg, const mavlink_mission_current_t* data);
void mavlink_msg_mission_current_decode(const mavlink_message_t* msg, mavlink_mission_current_t* data);

void mavlink_msg_mission_count_encode(uint8_t system_id, uint8_t component_id,
                                      mavlink_message_t* msg, const mavlink_mission_count_t* data);
void mavlink_msg_mission_count_decode(const mavlink_message_t* msg, mavlink_mission_count_t* data);

void mavlink_msg_command_long_encode(uint8_t system_id, uint8_t component_id,
                                     mavlink_message_t* msg, const mavlink_command_long_t* data);
void mavlink_msg_command_long_decode(const mavlink_message_t* msg, mavlink_command_long_t* data);

void mavlink_msg_command_ack_encode(uint8_t system_id, uint8_t component_id,
                                    mavlink_message_t* msg, const mavlink_command_ack_t* data);
void mavlink_msg_command_ack_decode(const mavlink_message_t* msg, mavlink_command_ack_t* data);

void mavlink_msg_battery_status_encode(uint8_t system_id, uint8_t component_id,
                                       mavlink_message_t* msg, const mavlink_battery_status_t* data);
void mavlink_msg_battery_status_decode(const mavlink_message_t* msg, mavlink_battery_status_t* data);

#endif
