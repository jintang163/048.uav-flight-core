#ifndef __LINK_MANAGER_H__
#define __LINK_MANAGER_H__

#include "types.h"
#include "flight_config.h"

typedef enum {
    LINK_TYPE_RADIO = 0,
    LINK_TYPE_4G = 1,
    LINK_TYPE_COUNT = 2
} LinkType;

typedef enum {
    LINK_STATE_DISCONNECTED = 0,
    LINK_STATE_CONNECTING = 1,
    LINK_STATE_CONNECTED = 2,
    LINK_STATE_DEGRADED = 3
} LinkState;

typedef struct __attribute__((packed)) {
    int8_t rssi;
    float snr;
    float packet_loss;
    uint32_t latency_ms;
} LinkQuality;

typedef struct __attribute__((packed)) {
    LinkType type;
    LinkState state;
    LinkQuality quality;
    bool connected;
    uint32_t last_heartbeat;
    uint32_t bytes_sent;
    uint32_t bytes_received;
} LinkStatus;

typedef struct __attribute__((packed)) {
    bool auto_switch_enabled;
    int8_t radio_rssi_threshold;
    int8_t lte_rssi_threshold;
    uint32_t hysteresis_time_ms;
} LinkManagerConfig;

void link_manager_init(void);
void link_manager_update(void);
void link_manager_set_preferred_link(LinkType type);
LinkType link_manager_get_active_link(void);
bool link_manager_get_link_status(LinkType type, LinkStatus *status);
void link_manager_update_link_quality(LinkType type, LinkQuality *quality);
void link_manager_notify_heartbeat(LinkType type);
const char* link_type_to_string(LinkType type);

#endif
