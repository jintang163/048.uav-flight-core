#ifndef MAVLINK_HEARTBEAT_H
#define MAVLINK_HEARTBEAT_H

#include "mavlink_v2.h"
#include "messages/common.h"
#include <stdint.h>
#include <stdbool.h>

#define HEARTBEAT_INTERVAL_MS        1000
#define HEARTBEAT_TIMEOUT_MS         3000
#define HEARTBEAT_LOST_MS           10000

typedef enum {
    LINK_STATUS_CONNECTED = 0,
    LINK_STATUS_WARNING = 1,
    LINK_STATUS_LOST = 2
} link_status_t;

typedef struct {
    uint8_t target_sysid;
    uint8_t target_compid;
    uint64_t last_heartbeat_time;
    uint64_t last_send_time;
    uint32_t heartbeat_interval_ms;
    uint32_t timeout_ms;
    uint32_t lost_ms;
    link_status_t status;
    uint32_t received_count;
    uint32_t lost_count;
    mavlink_heartbeat_t last_heartbeat;
    bool has_valid_heartbeat;
} heartbeat_monitor_t;

typedef void (*heartbeat_status_cb_t)(link_status_t status, void* user_data);

void heartbeat_monitor_init(heartbeat_monitor_t* monitor, uint8_t target_sysid, uint8_t target_compid);
void heartbeat_monitor_set_timing(heartbeat_monitor_t* monitor, uint32_t interval_ms,
                                   uint32_t timeout_ms, uint32_t lost_ms);
void heartbeat_monitor_set_callback(heartbeat_monitor_t* monitor, heartbeat_status_cb_t cb, void* user_data);

bool heartbeat_monitor_update(heartbeat_monitor_t* monitor, const mavlink_message_t* msg, uint64_t now_ms);
link_status_t heartbeat_monitor_get_status(const heartbeat_monitor_t* monitor);
bool heartbeat_monitor_should_send(const heartbeat_monitor_t* monitor, uint64_t now_ms);
uint16_t heartbeat_create_msg(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                              uint8_t type, uint8_t autopilot, uint8_t base_mode,
                              uint32_t custom_mode, uint8_t system_status);

uint32_t heartbeat_get_received_count(const heartbeat_monitor_t* monitor);
uint32_t heartbeat_get_lost_count(const heartbeat_monitor_t* monitor);
bool heartbeat_get_last(const heartbeat_monitor_t* monitor, mavlink_heartbeat_t* out);

#endif
