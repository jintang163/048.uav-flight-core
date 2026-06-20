#ifndef MAVLINK_MISSION_H
#define MAVLINK_MISSION_H

#include "mavlink_v2.h"
#include "messages/common.h"
#include <stdint.h>
#include <stdbool.h>

#define MISSION_TIMEOUT_MS           10000
#define MISSION_MAX_RETRIES          3
#define MISSION_MAX_ITEMS            256

typedef enum {
    MISSION_STATE_IDLE = 0,
    MISSION_STATE_UPLOADING = 1,
    MISSION_STATE_DOWNLOADING = 2,
    MISSION_STATE_COMPLETE = 3,
    MISSION_STATE_FAILED = 4
} mission_state_t;

typedef struct {
    uint16_t seq;
    uint8_t frame;
    uint16_t command;
    uint8_t current;
    uint8_t autocontinue;
    float param1;
    float param2;
    float param3;
    float param4;
    double x;
    double y;
    double z;
} mission_item_t;

typedef struct {
    mission_item_t items[MISSION_MAX_ITEMS];
    uint16_t count;
    uint16_t current_seq;
    uint8_t mission_type;
} mission_t;

typedef void (*mission_progress_cb_t)(uint16_t current, uint16_t total, void* user_data);
typedef void (*mission_complete_cb_t)(bool success, void* user_data);

typedef struct {
    mission_state_t state;
    mission_t* mission;
    uint8_t target_system;
    uint8_t target_component;
    uint16_t transfer_index;
    uint8_t retries;
    uint64_t last_action_time;
    uint64_t timeout_ms;
    mission_progress_cb_t progress_cb;
    mission_complete_cb_t complete_cb;
    void* user_data;
    uint32_t upload_count;
    uint32_t download_count;
} mission_manager_t;

void mission_init(mission_t* mission);
int mission_add_item(mission_t* mission, const mission_item_t* item);
void mission_clear(mission_t* mission);
bool mission_get_item(const mission_t* mission, uint16_t seq, mission_item_t* out);
uint16_t mission_get_count(const mission_t* mission);
uint16_t mission_get_current(const mission_t* mission);
void mission_set_current(mission_t* mission, uint16_t seq);

void mission_manager_init(mission_manager_t* manager);
int mission_start_upload(mission_manager_t* manager, mission_t* mission,
                         uint8_t target_system, uint8_t target_component,
                         mission_progress_cb_t progress_cb, mission_complete_cb_t complete_cb,
                         void* user_data, uint64_t now_ms);
int mission_start_download(mission_manager_t* manager, mission_t* mission,
                           uint8_t target_system, uint8_t target_component,
                           mission_progress_cb_t progress_cb, mission_complete_cb_t complete_cb,
                           void* user_data, uint64_t now_ms);

bool mission_process_message(mission_manager_t* manager, const mavlink_message_t* msg,
                             mavlink_encoder_t* encoder, uint64_t now_ms);
void mission_manager_update(mission_manager_t* manager, mavlink_encoder_t* encoder, uint64_t now_ms);

mission_state_t mission_get_state(const mission_manager_t* manager);

#endif
