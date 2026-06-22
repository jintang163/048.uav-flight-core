#ifndef __MAVLINK_HANDLER_H__
#define __MAVLINK_HANDLER_H__

#include "types.h"
#include "flight_config.h"
#include "link_manager.h"

#define MAVLINK_COMM_RADIO 0
#define MAVLINK_COMM_LTE   1

typedef enum {
    MAV_STATE_UNINIT = 0,
    MAV_STATE_BOOT = 1,
    MAV_STATE_CALIBRATING = 2,
    MAV_STATE_STANDBY = 3,
    MAV_STATE_ACTIVE = 4,
    MAV_STATE_CRITICAL = 5,
    MAV_STATE_EMERGENCY = 6,
    MAV_STATE_POWEROFF = 7
} MAVState;

typedef enum {
    MAV_MODE_FLAG_SAFETY_ARMED = 128,
    MAV_MODE_FLAG_MANUAL_INPUT_ENABLED = 64,
    MAV_MODE_FLAG_HIL_ENABLED = 32,
    MAV_MODE_FLAG_STABILIZE_ENABLED = 16,
    MAV_MODE_FLAG_GUIDED_ENABLED = 8,
    MAV_MODE_FLAG_AUTO_ENABLED = 4,
    MAV_MODE_FLAG_TEST_ENABLED = 2,
    MAV_MODE_FLAG_CUSTOM_MODE_ENABLED = 1
} MAVModeFlag;

typedef enum {
    MAV_CMD_NAV_WAYPOINT = 16,
    MAV_CMD_NAV_TAKEOFF = 22,
    MAV_CMD_NAV_LAND = 21,
    MAV_CMD_NAV_LOITER_UNLIM = 17,
    MAV_CMD_NAV_LOITER_TIME = 19,
    MAV_CMD_NAV_RETURN_TO_LAUNCH = 20,
    MAV_CMD_DO_SET_MODE = 176,
    MAV_CMD_COMPONENT_ARM_DISARM = 400,
    MAV_CMD_DO_CHANGE_SPEED = 178,
    MAV_CMD_DO_SET_HOME = 179,
    MAV_CMD_DO_SET_PARAM = 223,
    CMD_DO_THRUST_LEARNING_CONFIG = 4201,
    CMD_DO_SET_PID_GAINS = 4202
} MAVCommand;

void mavlink_handler_init(void);
void mavlink_handler_update(void);
void mavlink_send_heartbeat(void);
void mavlink_send_attitude(void);
void mavlink_send_global_position_int(void);
void mavlink_send_battery_status(void);
void mavlink_send_rc_channels_raw(void);
void mavlink_send_sys_status(void);
void mavlink_send_mission_count(uint16_t count);
void mavlink_send_mission_item(uint16_t index);
void mavlink_send_mission_ack(uint8_t type);
void mavlink_send_statustext(uint8_t severity, const char *text);
void mavlink_send_command_ack(uint16_t cmd, uint8_t result);
void mavlink_send_local_position_ned(void);
void mavlink_send_altitude(void);
void mavlink_send_vfr_hud(void);
void mavlink_receive_byte(uint8_t byte);
void mavlink_set_target_attitude(float roll, float pitch, float yaw, float thrust);
void mavlink_set_target_position(int32_t lat, int32_t lon, int32_t alt);
bool mavlink_get_command(ControlCommand *cmd);

void mavlink_send_geofence_violation(uint16_t fence_id, uint8_t violation_type,
                                     uint8_t severity, float distance,
                                     float lat, float lon, float alt);

void mavlink_send_log_entry(uint16_t id, uint32_t num_logs, uint32_t latest_log_num,
                            uint32_t size, uint32_t time_utc);
void mavlink_send_log_data(uint16_t id, uint32_t offset, uint8_t count, const uint8_t *data);

void mavlink_send_link_status(LinkType active_link, LinkStatus *radio_status, LinkStatus *lte_status);

void mavlink_send_obstacle_avoidance_event(void *event);
void mavlink_send_obstacle_avoidance_complete(void *event);
void mavlink_send_obstacle_avoidance_failed(void *event, const char *reason);

void mavlink_set_primary_link(uint8_t link);
void mavlink_receive_byte_from_link(uint8_t link, uint8_t byte);

void mavlink_send_thrust_learning_status(void);
void mavlink_send_thrust_curve(uint8_t start_index, uint8_t count);
void mavlink_send_pid_gains(void);

#endif
