#include "command.h"
#include "encoder.h"
#include <stdlib.h>
#include <string.h>

void command_manager_init(command_manager_t* manager)
{
    memset(manager, 0, sizeof(command_manager_t));
    manager->pending_list = NULL;
}

void command_manager_cleanup(command_manager_t* manager)
{
    command_entry_t* entry = manager->pending_list;
    while (entry) {
        command_entry_t* next = entry->next;
        free(entry);
        entry = next;
    }
    manager->pending_list = NULL;
    manager->pending_count = 0;
}

static command_entry_t* command_find_entry(command_manager_t* manager, uint16_t command,
                                          uint8_t target_system, uint8_t target_component)
{
    command_entry_t* entry = manager->pending_list;
    while (entry) {
        if (entry->command == command &&
            entry->target_system == target_system &&
            entry->target_component == target_component &&
            entry->state == COMMAND_STATE_PENDING) {
            return entry;
        }
        entry = entry->next;
    }
    return NULL;
}

int command_send_long(command_manager_t* manager, mavlink_encoder_t* encoder,
                      uint16_t command, uint8_t target_system, uint8_t target_component,
                      float param1, float param2, float param3, float param4,
                      float param5, float param6, float param7,
                      command_ack_cb_t callback, void* user_data, uint64_t now_ms)
{
    command_entry_t* entry = (command_entry_t*)malloc(sizeof(command_entry_t));
    if (!entry) {
        return -1;
    }

    memset(entry, 0, sizeof(command_entry_t));
    entry->command = command;
    entry->target_system = target_system;
    entry->target_component = target_component;
    entry->state = COMMAND_STATE_PENDING;
    entry->retries = 0;
    entry->send_time = now_ms;
    entry->timeout_ms = COMMAND_TIMEOUT_MS;
    entry->callback = callback;
    entry->user_data = user_data;

    mavlink_command_long_t cmd;
    memset(&cmd, 0, sizeof(cmd));
    cmd.param1 = param1;
    cmd.param2 = param2;
    cmd.param3 = param3;
    cmd.param4 = param4;
    cmd.param5 = param5;
    cmd.param6 = param6;
    cmd.param7 = param7;
    cmd.command = command;
    cmd.target_system = target_system;
    cmd.target_component = target_component;
    cmd.confirmation = 0;

    mavlink_message_t msg;
    mavlink_encode_command_long(encoder, &msg, &cmd);

    entry->next = manager->pending_list;
    manager->pending_list = entry;
    manager->pending_count++;
    manager->sent_count++;

    return 0;
}

int command_send_arm(command_manager_t* manager, mavlink_encoder_t* encoder,
                     uint8_t target_system, uint8_t target_component,
                     bool arm, command_ack_cb_t callback, void* user_data, uint64_t now_ms)
{
    return command_send_long(manager, encoder, MAV_CMD_COMPONENT_ARM_DISARM,
                             target_system, target_component,
                             arm ? 1.0f : 0.0f, 0, 0, 0, 0, 0, 0,
                             callback, user_data, now_ms);
}

int command_send_set_mode(command_manager_t* manager, mavlink_encoder_t* encoder,
                          uint8_t target_system, uint8_t target_component,
                          uint8_t base_mode, uint32_t custom_mode,
                          command_ack_cb_t callback, void* user_data, uint64_t now_ms)
{
    return command_send_long(manager, encoder, MAV_CMD_DO_SET_MODE,
                             target_system, target_component,
                             base_mode, (float)(custom_mode & 0xFFFF),
                             (float)((custom_mode >> 16) & 0xFFFF),
                             0, 0, 0, 0,
                             callback, user_data, now_ms);
}

bool command_process_ack(command_manager_t* manager, const mavlink_message_t* msg)
{
    if (mavlink_get_msg_id(msg) != MAVLINK_MSG_ID_COMMAND_ACK) {
        return false;
    }

    mavlink_command_ack_t ack;
    mavlink_msg_command_ack_decode(msg, &ack);

    command_entry_t* entry = command_find_entry(manager, ack.command,
                                                ack.target_system, ack.target_component);
    if (!entry) {
        return false;
    }

    entry->state = (ack.result == MAV_RESULT_ACCEPTED) ? COMMAND_STATE_ACKED : COMMAND_STATE_FAILED;

    if (entry->callback) {
        entry->callback(ack.command, ack.result, ack.result_param1, entry->user_data);
    }

    if (ack.result == MAV_RESULT_ACCEPTED) {
        manager->acked_count++;
    } else {
        manager->failed_count++;
    }

    command_entry_t** prev = &manager->pending_list;
    while (*prev && *prev != entry) {
        prev = &(*prev)->next;
    }
    if (*prev) {
        *prev = entry->next;
        free(entry);
        manager->pending_count--;
    }

    return true;
}

void command_manager_update(command_manager_t* manager, uint64_t now_ms)
{
    command_entry_t* entry = manager->pending_list;
    command_entry_t** prev = &manager->pending_list;

    while (entry) {
        if (entry->state == COMMAND_STATE_PENDING &&
            (now_ms - entry->send_time) > entry->timeout_ms) {
            if (entry->retries < COMMAND_MAX_RETRIES) {
                entry->retries++;
                entry->send_time = now_ms;
            } else {
                entry->state = COMMAND_STATE_TIMEOUT;
                if (entry->callback) {
                    entry->callback(entry->command, MAV_RESULT_FAILED, 0, entry->user_data);
                }
                manager->timeout_count++;
                *prev = entry->next;
                command_entry_t* to_free = entry;
                entry = entry->next;
                free(to_free);
                manager->pending_count--;
                continue;
            }
        }
        prev = &entry->next;
        entry = entry->next;
    }
}

uint32_t command_get_pending_count(const command_manager_t* manager) { return manager->pending_count; }
uint32_t command_get_sent_count(const command_manager_t* manager) { return manager->sent_count; }
uint32_t command_get_acked_count(const command_manager_t* manager) { return manager->acked_count; }
uint32_t command_get_timeout_count(const command_manager_t* manager) { return manager->timeout_count; }
uint32_t command_get_failed_count(const command_manager_t* manager) { return manager->failed_count; }
