#include "link_manager.h"
#include <string.h>

#define LINK_HEARTBEAT_TIMEOUT 5000

static LinkStatus link_status[LINK_TYPE_COUNT];
static LinkManagerConfig config;
static LinkType active_link;
static LinkType preferred_link;
static bool initialized = false;

static uint32_t radio_below_threshold_time = 0;
static uint32_t lte_below_threshold_time = 0;
static uint32_t radio_above_threshold_time = 0;
static uint32_t lte_above_threshold_time = 0;

void link_manager_init(void)
{
    memset(&link_status, 0, sizeof(link_status));

    for (int i = 0; i < LINK_TYPE_COUNT; i++) {
        link_status[i].type = (LinkType)i;
        link_status[i].state = LINK_STATE_DISCONNECTED;
        link_status[i].connected = false;
        link_status[i].quality.rssi = -120;
        link_status[i].quality.snr = 0.0f;
        link_status[i].quality.packet_loss = 100.0f;
        link_status[i].quality.latency_ms = 0;
        link_status[i].last_heartbeat = 0;
        link_status[i].bytes_sent = 0;
        link_status[i].bytes_received = 0;
    }

    config.auto_switch_enabled = true;
    config.radio_rssi_threshold = -90;
    config.lte_rssi_threshold = -100;
    config.hysteresis_time_ms = 3000;

    active_link = LINK_TYPE_RADIO;
    preferred_link = LINK_TYPE_RADIO;

    radio_below_threshold_time = 0;
    lte_below_threshold_time = 0;
    radio_above_threshold_time = 0;
    lte_above_threshold_time = 0;

    initialized = true;
}

static void link_manager_update_heartbeat_timeout(void)
{
    uint32_t now = HAL_GetTick();

    for (int i = 0; i < LINK_TYPE_COUNT; i++) {
        if (link_status[i].connected) {
            if (now - link_status[i].last_heartbeat > LINK_HEARTBEAT_TIMEOUT) {
                link_status[i].connected = false;
                link_status[i].state = LINK_STATE_DISCONNECTED;
            }
        }
    }
}

static void link_manager_update_link_state(LinkType type)
{
    LinkStatus *status = &link_status[type];
    int8_t threshold = (type == LINK_TYPE_RADIO) ? config.radio_rssi_threshold : config.lte_rssi_threshold;

    if (!status->connected) {
        if (status->quality.rssi > threshold) {
            status->state = LINK_STATE_CONNECTING;
        } else {
            status->state = LINK_STATE_DISCONNECTED;
        }
    } else {
        if (status->quality.rssi > threshold) {
            if (status->quality.packet_loss < 30.0f) {
                status->state = LINK_STATE_CONNECTED;
            } else {
                status->state = LINK_STATE_DEGRADED;
            }
        } else {
            status->state = LINK_STATE_DEGRADED;
        }
    }
}

static void link_manager_update_threshold_timers(void)
{
    uint32_t now = HAL_GetTick();
    LinkStatus *radio = &link_status[LINK_TYPE_RADIO];
    LinkStatus *lte = &link_status[LINK_TYPE_4G];

    if (radio->quality.rssi < config.radio_rssi_threshold) {
        if (radio_below_threshold_time == 0) {
            radio_below_threshold_time = now;
        }
        radio_above_threshold_time = 0;
    } else {
        if (radio_above_threshold_time == 0) {
            radio_above_threshold_time = now;
        }
        radio_below_threshold_time = 0;
    }

    if (lte->quality.rssi < config.lte_rssi_threshold) {
        if (lte_below_threshold_time == 0) {
            lte_below_threshold_time = now;
        }
        lte_above_threshold_time = 0;
    } else {
        if (lte_above_threshold_time == 0) {
            lte_above_threshold_time = now;
        }
        lte_below_threshold_time = 0;
    }
}

static void link_manager_auto_switch(void)
{
    if (!config.auto_switch_enabled) {
        return;
    }

    uint32_t now = HAL_GetTick();
    LinkStatus *radio = &link_status[LINK_TYPE_RADIO];
    LinkStatus *lte = &link_status[LINK_TYPE_4G];

    if (active_link == LINK_TYPE_RADIO) {
        if (radio_below_threshold_time > 0 &&
            (now - radio_below_threshold_time) > config.hysteresis_time_ms &&
            lte->connected &&
            lte->quality.rssi > config.lte_rssi_threshold) {
            active_link = LINK_TYPE_4G;
        }
    } else if (active_link == LINK_TYPE_4G) {
        if (radio_above_threshold_time > 0 &&
            (now - radio_above_threshold_time) > config.hysteresis_time_ms &&
            radio->connected &&
            radio->quality.rssi > config.radio_rssi_threshold) {
            active_link = LINK_TYPE_RADIO;
        }
    }

    if (!link_status[active_link].connected) {
        for (int i = 0; i < LINK_TYPE_COUNT; i++) {
            if (link_status[i].connected) {
                active_link = (LinkType)i;
                break;
            }
        }
    }
}

void link_manager_update(void)
{
    if (!initialized) {
        return;
    }

    link_manager_update_heartbeat_timeout();

    for (int i = 0; i < LINK_TYPE_COUNT; i++) {
        link_manager_update_link_state((LinkType)i);
    }

    link_manager_update_threshold_timers();
    link_manager_auto_switch();
}

void link_manager_set_preferred_link(LinkType type)
{
    if (type >= LINK_TYPE_COUNT) {
        return;
    }
    preferred_link = type;

    if (link_status[type].connected) {
        active_link = type;
    }
}

LinkType link_manager_get_active_link(void)
{
    return active_link;
}

bool link_manager_get_link_status(LinkType type, LinkStatus *status)
{
    if (type >= LINK_TYPE_COUNT || status == NULL) {
        return false;
    }
    *status = link_status[type];
    return true;
}

void link_manager_update_link_quality(LinkType type, LinkQuality *quality)
{
    if (type >= LINK_TYPE_COUNT || quality == NULL) {
        return;
    }
    link_status[type].quality = *quality;
}

void link_manager_notify_heartbeat(LinkType type)
{
    if (type >= LINK_TYPE_COUNT) {
        return;
    }
    link_status[type].last_heartbeat = HAL_GetTick();
    link_status[type].connected = true;
}

const char* link_type_to_string(LinkType type)
{
    switch (type) {
        case LINK_TYPE_RADIO:
            return "RADIO";
        case LINK_TYPE_4G:
            return "4G";
        default:
            return "UNKNOWN";
    }
}
