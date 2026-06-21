#include "task_link_manager.h"
#include "link_manager.h"
#include "4g_driver.h"
#include "sbus_rc.h"
#include "mavlink_handler.h"
#include <string.h>
#include <stdio.h>

static TaskHandle_t task_handle = NULL;
static LinkType last_active_link = LINK_TYPE_RADIO;
static LinkState last_radio_state = LINK_STATE_DISCONNECTED;
static LinkState last_lte_state = LINK_STATE_DISCONNECTED;

static uint32_t last_link_manager_update = 0;
static uint32_t last_lte_quality_update = 0;
static uint32_t last_radio_quality_update = 0;

static int8_t simulated_radio_rssi = -60;

static void task_link_manager_update_radio_quality(void)
{
    LinkQuality quality;
    SBUS_Data sbus_data;

    memset(&quality, 0, sizeof(quality));

    sbus_rc_get_data(&sbus_data);

    int16_t delta = (rand() % 7) - 3;
    simulated_radio_rssi = (int8_t)CONSTRAIN(simulated_radio_rssi + delta, -120, -30);

    quality.rssi = simulated_radio_rssi;
    quality.snr = 12.0f + ((float)(rand() % 60) - 30.0f) / 10.0f;
    quality.packet_loss = sbus_data.signal_loss ? 50.0f : ((float)(rand() % 50)) / 10.0f;
    quality.latency_ms = 10 + (rand() % 20);

    link_manager_update_link_quality(LINK_TYPE_RADIO, &quality);

    if (sbus_data.connected) {
        link_manager_notify_heartbeat(LINK_TYPE_RADIO);
    }
}

static void task_link_manager_update_lte_quality(void)
{
    LinkQuality quality;
    LTEStatus lte_status;

    memset(&quality, 0, sizeof(quality));

    _4g_driver_get_status(&lte_status);

    quality.rssi = lte_status.rssi;
    quality.snr = lte_status.ber != 99 ? (float)(30 - lte_status.ber * 2) : 0.0f;
    quality.packet_loss = lte_status.registered ? ((float)lte_status.ber * 0.5f) : 100.0f;
    quality.latency_ms = lte_status.registered ? 50 + (lte_status.csq < 20 ? 100 : 30) : 0;

    link_manager_update_link_quality(LINK_TYPE_4G, &quality);

    if (_4g_driver_is_connected()) {
        link_manager_notify_heartbeat(LINK_TYPE_4G);
    }
}

static void task_link_manager_check_state_changes(void)
{
    LinkType active_link = link_manager_get_active_link();
    LinkStatus radio_status, lte_status;

    link_manager_get_link_status(LINK_TYPE_RADIO, &radio_status);
    link_manager_get_link_status(LINK_TYPE_4G, &lte_status);

    if (active_link != last_active_link) {
        char text[64];
        snprintf(text, sizeof(text), "Link switched: %s -> %s",
                 link_type_to_string(last_active_link),
                 link_type_to_string(active_link));
        mavlink_send_statustext(MAV_SEVERITY_INFO, text);
        last_active_link = active_link;
    }

    if (radio_status.state != last_radio_state) {
        char text[64];
        snprintf(text, sizeof(text), "Radio link state: %d -> %d (RSSI: %d)",
                 last_radio_state, radio_status.state, radio_status.quality.rssi);
        mavlink_send_statustext(
            radio_status.state == LINK_STATE_CONNECTED ? MAV_SEVERITY_INFO :
            radio_status.state == LINK_STATE_DISCONNECTED ? MAV_SEVERITY_WARNING :
            MAV_SEVERITY_NOTICE, text);
        last_radio_state = radio_status.state;
    }

    if (lte_status.state != last_lte_state) {
        char text[64];
        snprintf(text, sizeof(text), "4G link state: %d -> %d (RSSI: %d)",
                 last_lte_state, lte_status.state, lte_status.quality.rssi);
        mavlink_send_statustext(
            lte_status.state == LINK_STATE_CONNECTED ? MAV_SEVERITY_INFO :
            lte_status.state == LINK_STATE_DISCONNECTED ? MAV_SEVERITY_WARNING :
            MAV_SEVERITY_NOTICE, text);
        last_lte_state = lte_status.state;
    }
}

void task_link_manager_init(void)
{
    link_manager_init();
    _4g_driver_init();

    last_active_link = LINK_TYPE_RADIO;
    last_radio_state = LINK_STATE_DISCONNECTED;
    last_lte_state = LINK_STATE_DISCONNECTED;

    last_link_manager_update = 0;
    last_lte_quality_update = 0;
    last_radio_quality_update = 0;

    xTaskCreate(task_link_manager_main,
                "LinkMgr",
                TASK_LINK_MANAGER_STACK_SIZE,
                NULL,
                TASK_LINK_MANAGER_PRIORITY,
                &task_handle);
}

void task_link_manager_main(void *argument)
{
    UNUSED(argument);

    TickType_t last_wake_time = xTaskGetTickCount();
    const TickType_t period = pdMS_TO_TICKS(1000 / TASK_LINK_MANAGER_FREQ);

    while (1) {
        uint32_t now = HAL_GetTick();

        if (now - last_link_manager_update >= 1000) {
            link_manager_update();
            last_link_manager_update = now;
        }

        if (now - last_lte_quality_update >= 500) {
            _4g_driver_update();
            task_link_manager_update_lte_quality();
            last_lte_quality_update = now;
        }

        if (now - last_radio_quality_update >= 200) {
            task_link_manager_update_radio_quality();
            last_radio_quality_update = now;
        }

        task_link_manager_check_state_changes();

        vTaskDelayUntil(&last_wake_time, period);
    }
}
