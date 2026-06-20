#include "mission.h"
#include "encoder.h"
#include <string.h>

void mission_init(mission_t* mission)
{
    memset(mission, 0, sizeof(mission_t));
    mission->count = 0;
    mission->current_seq = 0;
    mission->mission_type = 0;
}

int mission_add_item(mission_t* mission, const mission_item_t* item)
{
    if (mission->count >= MISSION_MAX_ITEMS) {
        return -1;
    }
    memcpy(&mission->items[mission->count], item, sizeof(mission_item_t));
    mission->items[mission->count].seq = mission->count;
    mission->count++;
    return 0;
}

void mission_clear(mission_t* mission)
{
    mission->count = 0;
    mission->current_seq = 0;
}

bool mission_get_item(const mission_t* mission, uint16_t seq, mission_item_t* out)
{
    if (seq >= mission->count) {
        return false;
    }
    memcpy(out, &mission->items[seq], sizeof(mission_item_t));
    return true;
}

uint16_t mission_get_count(const mission_t* mission) { return mission->count; }
uint16_t mission_get_current(const mission_t* mission) { return mission->current_seq; }
void mission_set_current(mission_t* mission, uint16_t seq) { mission->current_seq = seq; }

void mission_manager_init(mission_manager_t* manager)
{
    memset(manager, 0, sizeof(mission_manager_t));
    manager->state = MISSION_STATE_IDLE;
    manager->timeout_ms = MISSION_TIMEOUT_MS;
}

int mission_start_upload(mission_manager_t* manager, mission_t* mission,
                         uint8_t target_system, uint8_t target_component,
                         mission_progress_cb_t progress_cb, mission_complete_cb_t complete_cb,
                         void* user_data, uint64_t now_ms)
{
    if (manager->state != MISSION_STATE_IDLE) {
        return -1;
    }
    manager->state = MISSION_STATE_UPLOADING;
    manager->mission = mission;
    manager->target_system = target_system;
    manager->target_component = target_component;
    manager->transfer_index = 0;
    manager->retries = 0;
    manager->last_action_time = now_ms;
    manager->progress_cb = progress_cb;
    manager->complete_cb = complete_cb;
    manager->user_data = user_data;
    return 0;
}

int mission_start_download(mission_manager_t* manager, mission_t* mission,
                           uint8_t target_system, uint8_t target_component,
                           mission_progress_cb_t progress_cb, mission_complete_cb_t complete_cb,
                           void* user_data, uint64_t now_ms)
{
    if (manager->state != MISSION_STATE_IDLE) {
        return -1;
    }
    manager->state = MISSION_STATE_DOWNLOADING;
    manager->mission = mission;
    manager->target_system = target_system;
    manager->target_component = target_component;
    manager->transfer_index = 0;
    manager->retries = 0;
    manager->last_action_time = now_ms;
    manager->progress_cb = progress_cb;
    manager->complete_cb = complete_cb;
    manager->user_data = user_data;
    return 0;
}

bool mission_process_message(mission_manager_t* manager, const mavlink_message_t* msg,
                             mavlink_encoder_t* encoder, uint64_t now_ms)
{
    (void)encoder;
    uint32_t msgid = mavlink_get_msg_id(msg);

    switch (msgid) {
    case MAVLINK_MSG_ID_MISSION_COUNT: {
        mavlink_mission_count_t count;
        mavlink_msg_mission_count_decode(msg, &count);
        if (manager->state == MISSION_STATE_DOWNLOADING) {
            manager->mission->count = count.count;
            manager->mission->mission_type = count.mission_type;
            manager->last_action_time = now_ms;
            manager->retries = 0;
            if (manager->progress_cb) {
                manager->progress_cb(0, count.count, manager->user_data);
            }
        }
        return true;
    }
    case MAVLINK_MSG_ID_MISSION_ITEM: {
        mavlink_mission_item_t item;
        mavlink_msg_mission_item_decode(msg, &item);
        if (manager->state == MISSION_STATE_DOWNLOADING && item.seq < MISSION_MAX_ITEMS) {
            mission_item_t mitem;
            memset(&mitem, 0, sizeof(mitem));
            mitem.seq = item.seq;
            mitem.frame = item.frame;
            mitem.command = item.command;
            mitem.current = item.current;
            mitem.autocontinue = item.autocontinue;
            mitem.param1 = item.param1;
            mitem.param2 = item.param2;
            mitem.param3 = item.param3;
            mitem.param4 = item.param4;
            mitem.x = item.x;
            mitem.y = item.y;
            mitem.z = item.z;
            memcpy(&manager->mission->items[item.seq], &mitem, sizeof(mission_item_t));
            manager->transfer_index = item.seq + 1;
            manager->last_action_time = now_ms;
            manager->retries = 0;
            if (manager->progress_cb) {
                manager->progress_cb(manager->transfer_index, manager->mission->count, manager->user_data);
            }
            if (manager->transfer_index >= manager->mission->count) {
                manager->state = MISSION_STATE_COMPLETE;
                manager->download_count++;
                if (manager->complete_cb) {
                    manager->complete_cb(true, manager->user_data);
                }
            }
        } else if (manager->state == MISSION_STATE_UPLOADING) {
            manager->transfer_index++;
            manager->last_action_time = now_ms;
            manager->retries = 0;
            if (manager->progress_cb) {
                manager->progress_cb(manager->transfer_index, manager->mission->count, manager->user_data);
            }
            if (manager->transfer_index >= manager->mission->count) {
                manager->state = MISSION_STATE_COMPLETE;
                manager->upload_count++;
                if (manager->complete_cb) {
                    manager->complete_cb(true, manager->user_data);
                }
            }
        }
        return true;
    }
    case MAVLINK_MSG_ID_MISSION_CURRENT: {
        mavlink_mission_current_t curr;
        mavlink_msg_mission_current_decode(msg, &curr);
        mission_set_current(manager->mission, curr.seq);
        return true;
    }
    default:
        break;
    }
    return false;
}

void mission_manager_update(mission_manager_t* manager, mavlink_encoder_t* encoder, uint64_t now_ms)
{
    if (manager->state == MISSION_STATE_IDLE || manager->state == MISSION_STATE_COMPLETE ||
        manager->state == MISSION_STATE_FAILED) {
        return;
    }

    if ((now_ms - manager->last_action_time) > manager->timeout_ms) {
        if (manager->retries < MISSION_MAX_RETRIES) {
            manager->retries++;
            manager->last_action_time = now_ms;
        } else {
            manager->state = MISSION_STATE_FAILED;
            if (manager->complete_cb) {
                manager->complete_cb(false, manager->user_data);
            }
        }
        return;
    }

    if (manager->state == MISSION_STATE_UPLOADING && manager->transfer_index < manager->mission->count) {
        mavlink_mission_count_t count;
        memset(&count, 0, sizeof(count));
        count.count = manager->mission->count;
        count.target_system = manager->target_system;
        count.target_component = manager->target_component;
        count.mission_type = manager->mission->mission_type;
        mavlink_message_t msg;
        mavlink_encode_mission_count(encoder, &msg, &count);
    }
}

mission_state_t mission_get_state(const mission_manager_t* manager)
{
    return manager->state;
}
