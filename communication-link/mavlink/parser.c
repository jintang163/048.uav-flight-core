#include "parser.h"
#include <string.h>

void mavlink_parser_ctx_init(mavlink_parser_ctx_t* ctx)
{
    mavlink_init_parser(&ctx->parser);
    ctx->signing = NULL;
    ctx->handler = NULL;
    ctx->user_data = NULL;
    ctx->received_count = 0;
    ctx->error_count = 0;
}

void mavlink_parser_set_signing(mavlink_parser_ctx_t* ctx, mavlink_signing_t* signing)
{
    ctx->signing = signing;
}

void mavlink_parser_set_handler(mavlink_parser_ctx_t* ctx, mavlink_msg_handler_t handler, void* user_data)
{
    ctx->handler = handler;
    ctx->user_data = user_data;
}

mavlink_parse_result_t mavlink_parser_feed(mavlink_parser_ctx_t* ctx, const uint8_t* data, size_t len,
                                           mavlink_message_t* out_msg)
{
    mavlink_parse_result_t result = MAVLINK_PARSE_INCOMPLETE;

    for (size_t i = 0; i < len; i++) {
        mavlink_message_t msg;
        if (mavlink_parse_byte(&ctx->parser, data[i], &msg)) {
            if (ctx->signing && msg.signature_present) {
                if (!mavlink_verify_signature(&msg, ctx->signing)) {
                    ctx->error_count++;
                    if (out_msg) {
                        memcpy(out_msg, &msg, sizeof(mavlink_message_t));
                    }
                    return MAVLINK_PARSE_ERROR_SIGNATURE;
                }
            }
            ctx->received_count++;
            if (out_msg) {
                memcpy(out_msg, &msg, sizeof(mavlink_message_t));
            }
            if (ctx->handler) {
                ctx->handler(&msg, ctx->user_data);
            }
            result = MAVLINK_PARSE_OK;
            break;
        }
    }

    if (ctx->parser.parse_error != 0) {
        ctx->error_count++;
        result = MAVLINK_PARSE_ERROR_CRC;
    }

    return result;
}

mavlink_parse_result_t mavlink_parser_parse_buffer(const uint8_t* buf, size_t buf_len,
                                                   mavlink_message_t* msg, mavlink_signing_t* signing)
{
    if (!mavlink_buf_to_msg(buf, buf_len, msg)) {
        return MAVLINK_PARSE_ERROR_CRC;
    }
    if (signing && msg->signature_present) {
        if (!mavlink_verify_signature(msg, signing)) {
            return MAVLINK_PARSE_ERROR_SIGNATURE;
        }
    }
    return MAVLINK_PARSE_OK;
}

#define DEFINE_DECODE_FUNC(msg_name, msg_type) \
bool mavlink_decode_##msg_name(const mavlink_message_t* msg, mavlink_##msg_type##_t* data) \
{ \
    if (mavlink_get_msg_id(msg) != MAVLINK_MSG_ID_##msg_name) { \
        return false; \
    } \
    if (mavlink_get_payload_len(msg) < sizeof(mavlink_##msg_type##_t)) { \
        return false; \
    } \
    mavlink_msg_##msg_name##_decode(msg, data); \
    return true; \
}

DEFINE_DECODE_FUNC(heartbeat, heartbeat)
DEFINE_DECODE_FUNC(sys_status, sys_status)
DEFINE_DECODE_FUNC(gps_raw_int, gps_raw_int)
DEFINE_DECODE_FUNC(attitude, attitude)
DEFINE_DECODE_FUNC(global_position_int, global_position_int)
DEFINE_DECODE_FUNC(rc_channels_raw, rc_channels_raw)
DEFINE_DECODE_FUNC(mission_item, mission_item)
DEFINE_DECODE_FUNC(mission_current, mission_current)
DEFINE_DECODE_FUNC(mission_count, mission_count)
DEFINE_DECODE_FUNC(command_long, command_long)
DEFINE_DECODE_FUNC(command_ack, command_ack)
DEFINE_DECODE_FUNC(battery_status, battery_status)

uint32_t mavlink_parser_get_received_count(const mavlink_parser_ctx_t* ctx)
{
    return ctx->received_count;
}

uint32_t mavlink_parser_get_error_count(const mavlink_parser_ctx_t* ctx)
{
    return ctx->error_count;
}

void mavlink_parser_reset_stats(mavlink_parser_ctx_t* ctx)
{
    ctx->received_count = 0;
    ctx->error_count = 0;
}
