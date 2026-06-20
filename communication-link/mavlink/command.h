#ifndef MAVLINK_COMMAND_H
#define MAVLINK_COMMAND_H

#include "mavlink_v2.h"
#include "messages/common.h"
#include <stdint.h>
#include <stdbool.h>

#define COMMAND_TIMEOUT_MS           5000
#define COMMAND_MAX_RETRIES          3

typedef enum {
    COMMAND_STATE_IDLE = 0,
    COMMAND_STATE_PENDING = 1,
    COMMAND_STATE_ACKED = 2,
    COMMAND_STATE_TIMEOUT = 3,
    COMMAND_STATE_FAILED = 4
} command_state_t;

typedef struct command_entry command_entry_t;

typedef void (*command_ack_cb_t)(uint16_t command, uint8_t result,
                                 int32_t result_param1, void* user_data);

struct command_entry {
    uint16_t command;
    uint8_t target_system;
    uint8_t target_component;
    command_state_t state;
    uint8_t retries;
    uint64_t send_time;
    uint64_t timeout_ms;
    command_ack_cb_t callback;
    void* user_data;
    command_entry_t* next;
};

typedef struct {
    command_entry_t* pending_list;
    uint32_t pending_count;
    uint32_t sent_count;
    uint32_t acked_count;
    uint32_t timeout_count;
    uint32_t failed_count;
} command_manager_t;

void command_manager_init(command_manager_t* manager);
void command_manager_cleanup(command_manager_t* manager);

int command_send_long(command_manager_t* manager, mavlink_encoder_t* encoder,
                      uint16_t command, uint8_t target_system, uint8_t target_component,
                      float param1, float param2, float param3, float param4,
                      float param5, float param6, float param7,
                      command_ack_cb_t callback, void* user_data, uint64_t now_ms);

int command_send_arm(command_manager_t* manager, mavlink_encoder_t* encoder,
                     uint8_t target_system, uint8_t target_component,
                     bool arm, command_ack_cb_t callback, void* user_data, uint64_t now_ms);

int command_send_set_mode(command_manager_t* manager, mavlink_encoder_t* encoder,
                          uint8_t target_system, uint8_t target_component,
                          uint8_t base_mode, uint32_t custom_mode,
                          command_ack_cb_t callback, void* user_data, uint64_t now_ms);

bool command_process_ack(command_manager_t* manager, const mavlink_message_t* msg);
void command_manager_update(command_manager_t* manager, uint64_t now_ms);

uint32_t command_get_pending_count(const command_manager_t* manager);
uint32_t command_get_sent_count(const command_manager_t* manager);
uint32_t command_get_acked_count(const command_manager_t* manager);
uint32_t command_get_timeout_count(const command_manager_t* manager);
uint32_t command_get_failed_count(const command_manager_t* manager);

#endif
