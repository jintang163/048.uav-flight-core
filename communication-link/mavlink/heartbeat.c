#include "heartbeat.h"
#include "encoder.h"
#include <string.h>

void heartbeat_monitor_init(heartbeat_monitor_t* monitor, uint8_t target_sysid, uint8_t target_compid)
{
    memset(monitor, 0, sizeof(heartbeat_monitor_t));
    monitor->target_sysid = target_sysid;
    monitor->target_compid = target_compid;
    monitor->heartbeat_interval_ms = HEARTBEAT_INTERVAL_MS;
    monitor->timeout_ms = HEARTBEAT_TIMEOUT_MS;
    monitor->lost_ms = HEARTBEAT_LOST_MS;
    monitor->status = LINK_STATUS_LOST;
    monitor->has_valid_heartbeat = false;
}

void heartbeat_monitor_set_timing(heartbeat_monitor_t* monitor, uint32_t interval_ms,
                                   uint32_t timeout_ms, uint32_t lost_ms)
{
    monitor->heartbeat_interval_ms = interval_ms;
    monitor->timeout_ms = timeout_ms;
    monitor->lost_ms = lost_ms;
}

void heartbeat_monitor_set_callback(heartbeat_monitor_t* monitor, heartbeat_status_cb_t cb, void* user_data)
{
    (void)monitor;
    (void)cb;
    (void)user_data;
}

bool heartbeat_monitor_update(heartbeat_monitor_t* monitor, const mavlink_message_t* msg, uint64_t now_ms)
{
    if (mavlink_get_msg_id(msg) != MAVLINK_MSG_ID_HEARTBEAT) {
        return false;
    }
    if (monitor->target_sysid != MAVLINK_SYS_ID_ALL &&
        mavlink_get_system_id(msg) != monitor->target_sysid) {
        return false;
    }
    if (monitor->target_compid != MAVLINK_COMP_ID_ALL &&
        mavlink_get_component_id(msg) != monitor->target_compid) {
        return false;
    }

    mavlink_heartbeat_t hb;
    mavlink_msg_heartbeat_decode(msg, &hb);

    monitor->last_heartbeat_time = now_ms;
    monitor->last_heartbeat = hb;
    monitor->has_valid_heartbeat = true;
    monitor->received_count++;

    link_status_t new_status = LINK_STATUS_CONNECTED;
    if (monitor->status != new_status) {
        monitor->status = new_status;
    }

    return true;
}

link_status_t heartbeat_monitor_get_status(const heartbeat_monitor_t* monitor)
{
    return monitor->status;
}

bool heartbeat_monitor_should_send(const heartbeat_monitor_t* monitor, uint64_t now_ms)
{
    return (now_ms - monitor->last_send_time) >= monitor->heartbeat_interval_ms;
}

uint16_t heartbeat_create_msg(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                              uint8_t type, uint8_t autopilot, uint8_t base_mode,
                              uint32_t custom_mode, uint8_t system_status)
{
    mavlink_heartbeat_t hb;
    memset(&hb, 0, sizeof(hb));
    hb.type = type;
    hb.autopilot = autopilot;
    hb.base_mode = base_mode;
    hb.custom_mode = custom_mode;
    hb.system_status = system_status;
    hb.mavlink_version = 3;
    return mavlink_encode_heartbeat(encoder, msg, &hb);
}

uint32_t heartbeat_get_received_count(const heartbeat_monitor_t* monitor)
{
    return monitor->received_count;
}

uint32_t heartbeat_get_lost_count(const heartbeat_monitor_t* monitor)
{
    return monitor->lost_count;
}

bool heartbeat_get_last(const heartbeat_monitor_t* monitor, mavlink_heartbeat_t* out)
{
    if (!monitor->has_valid_heartbeat) {
        return false;
    }
    memcpy(out, &monitor->last_heartbeat, sizeof(mavlink_heartbeat_t));
    return true;
}
