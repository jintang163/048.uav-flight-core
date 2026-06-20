#ifndef MAVLINK_PARSER_H
#define MAVLINK_PARSER_H

#include "mavlink_v2.h"
#include "messages/common.h"

typedef enum {
    MAVLINK_PARSE_OK = 0,
    MAVLINK_PARSE_INCOMPLETE = 1,
    MAVLINK_PARSE_ERROR_CRC = -1,
    MAVLINK_PARSE_ERROR_SIGNATURE = -2,
    MAVLINK_PARSE_ERROR_UNKNOWN_MSG = -3
} mavlink_parse_result_t;

typedef void (*mavlink_msg_handler_t)(const mavlink_message_t* msg, void* user_data);

typedef struct {
    mavlink_parser_t parser;
    mavlink_signing_t* signing;
    mavlink_msg_handler_t handler;
    void* user_data;
    uint32_t received_count;
    uint32_t error_count;
} mavlink_parser_ctx_t;

void mavlink_parser_ctx_init(mavlink_parser_ctx_t* ctx);
void mavlink_parser_set_signing(mavlink_parser_ctx_t* ctx, mavlink_signing_t* signing);
void mavlink_parser_set_handler(mavlink_parser_ctx_t* ctx, mavlink_msg_handler_t handler, void* user_data);

mavlink_parse_result_t mavlink_parser_feed(mavlink_parser_ctx_t* ctx, const uint8_t* data, size_t len,
                                           mavlink_message_t* out_msg);
mavlink_parse_result_t mavlink_parser_parse_buffer(const uint8_t* buf, size_t buf_len,
                                                   mavlink_message_t* msg, mavlink_signing_t* signing);

bool mavlink_decode_heartbeat(const mavlink_message_t* msg, mavlink_heartbeat_t* data);
bool mavlink_decode_sys_status(const mavlink_message_t* msg, mavlink_sys_status_t* data);
bool mavlink_decode_gps_raw_int(const mavlink_message_t* msg, mavlink_gps_raw_int_t* data);
bool mavlink_decode_attitude(const mavlink_message_t* msg, mavlink_attitude_t* data);
bool mavlink_decode_global_position_int(const mavlink_message_t* msg, mavlink_global_position_int_t* data);
bool mavlink_decode_rc_channels_raw(const mavlink_message_t* msg, mavlink_rc_channels_raw_t* data);
bool mavlink_decode_mission_item(const mavlink_message_t* msg, mavlink_mission_item_t* data);
bool mavlink_decode_mission_current(const mavlink_message_t* msg, mavlink_mission_current_t* data);
bool mavlink_decode_mission_count(const mavlink_message_t* msg, mavlink_mission_count_t* data);
bool mavlink_decode_command_long(const mavlink_message_t* msg, mavlink_command_long_t* data);
bool mavlink_decode_command_ack(const mavlink_message_t* msg, mavlink_command_ack_t* data);
bool mavlink_decode_battery_status(const mavlink_message_t* msg, mavlink_battery_status_t* data);

uint32_t mavlink_parser_get_received_count(const mavlink_parser_ctx_t* ctx);
uint32_t mavlink_parser_get_error_count(const mavlink_parser_ctx_t* ctx);
void mavlink_parser_reset_stats(mavlink_parser_ctx_t* ctx);

#endif
