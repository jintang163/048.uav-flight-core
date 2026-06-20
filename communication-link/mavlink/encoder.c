#include "encoder.h"
#include <string.h>

void mavlink_encoder_init(mavlink_encoder_t* encoder, uint8_t system_id, uint8_t component_id)
{
    encoder->system_id = system_id;
    encoder->component_id = component_id;
    encoder->seq = 0;
    encoder->signing = NULL;
}

void mavlink_encoder_set_signing(mavlink_encoder_t* encoder, mavlink_signing_t* signing)
{
    encoder->signing = signing;
}

static void mavlink_fill_header(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                                uint32_t msgid, uint8_t payload_len)
{
    msg->header.magic = MAVLINK_STX_V2;
    msg->header.payload_len = payload_len;
    msg->header.incompat_flags = 0;
    msg->header.compat_flags = 0;
    msg->header.seq = encoder->seq++;
    msg->header.sysid = encoder->system_id;
    msg->header.compid = encoder->component_id;
    msg->header.msgid = msgid;
    msg->signature_present = false;
}

static uint16_t mavlink_finalize_message(mavlink_encoder_t* encoder, mavlink_message_t* msg)
{
    if (encoder->signing) {
        mavlink_sign_message(msg, encoder->signing);
    }
    return MAVLINK_HEADER_LEN_V2 + msg->header.payload_len + MAVLINK_CHECKSUM_LEN +
           (msg->signature_present ? MAVLINK_SIGNATURE_LEN : 0);
}

uint16_t mavlink_encode_heartbeat(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                                  const mavlink_heartbeat_t* data)
{
    mavlink_fill_header(encoder, msg, MAVLINK_MSG_ID_HEARTBEAT, MAVLINK_MSG_HEARTBEAT_LEN);
    mavlink_msg_heartbeat_encode(encoder->system_id, encoder->component_id, msg, data);
    return mavlink_finalize_message(encoder, msg);
}

uint16_t mavlink_encode_sys_status(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                                   const mavlink_sys_status_t* data)
{
    mavlink_fill_header(encoder, msg, MAVLINK_MSG_ID_SYS_STATUS, MAVLINK_MSG_SYS_STATUS_LEN);
    mavlink_msg_sys_status_encode(encoder->system_id, encoder->component_id, msg, data);
    return mavlink_finalize_message(encoder, msg);
}

uint16_t mavlink_encode_gps_raw_int(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                                    const mavlink_gps_raw_int_t* data)
{
    mavlink_fill_header(encoder, msg, MAVLINK_MSG_ID_GPS_RAW_INT, MAVLINK_MSG_GPS_RAW_INT_LEN);
    mavlink_msg_gps_raw_int_encode(encoder->system_id, encoder->component_id, msg, data);
    return mavlink_finalize_message(encoder, msg);
}

uint16_t mavlink_encode_attitude(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                                 const mavlink_attitude_t* data)
{
    mavlink_fill_header(encoder, msg, MAVLINK_MSG_ID_ATTITUDE, MAVLINK_MSG_ATTITUDE_LEN);
    mavlink_msg_attitude_encode(encoder->system_id, encoder->component_id, msg, data);
    return mavlink_finalize_message(encoder, msg);
}

uint16_t mavlink_encode_global_position_int(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                                            const mavlink_global_position_int_t* data)
{
    mavlink_fill_header(encoder, msg, MAVLINK_MSG_ID_GLOBAL_POSITION_INT, MAVLINK_MSG_GLOBAL_POSITION_INT_LEN);
    mavlink_msg_global_position_int_encode(encoder->system_id, encoder->component_id, msg, data);
    return mavlink_finalize_message(encoder, msg);
}

uint16_t mavlink_encode_rc_channels_raw(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                                        const mavlink_rc_channels_raw_t* data)
{
    mavlink_fill_header(encoder, msg, MAVLINK_MSG_ID_RC_CHANNELS_RAW, MAVLINK_MSG_RC_CHANNELS_RAW_LEN);
    mavlink_msg_rc_channels_raw_encode(encoder->system_id, encoder->component_id, msg, data);
    return mavlink_finalize_message(encoder, msg);
}

uint16_t mavlink_encode_mission_item(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                                     const mavlink_mission_item_t* data)
{
    mavlink_fill_header(encoder, msg, MAVLINK_MSG_ID_MISSION_ITEM, MAVLINK_MSG_MISSION_ITEM_LEN);
    mavlink_msg_mission_item_encode(encoder->system_id, encoder->component_id, msg, data);
    return mavlink_finalize_message(encoder, msg);
}

uint16_t mavlink_encode_mission_current(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                                        const mavlink_mission_current_t* data)
{
    mavlink_fill_header(encoder, msg, MAVLINK_MSG_ID_MISSION_CURRENT, MAVLINK_MSG_MISSION_CURRENT_LEN);
    mavlink_msg_mission_current_encode(encoder->system_id, encoder->component_id, msg, data);
    return mavlink_finalize_message(encoder, msg);
}

uint16_t mavlink_encode_mission_count(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                                      const mavlink_mission_count_t* data)
{
    mavlink_fill_header(encoder, msg, MAVLINK_MSG_ID_MISSION_COUNT, MAVLINK_MSG_MISSION_COUNT_LEN);
    mavlink_msg_mission_count_encode(encoder->system_id, encoder->component_id, msg, data);
    return mavlink_finalize_message(encoder, msg);
}

uint16_t mavlink_encode_command_long(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                                     const mavlink_command_long_t* data)
{
    mavlink_fill_header(encoder, msg, MAVLINK_MSG_ID_COMMAND_LONG, MAVLINK_MSG_COMMAND_LONG_LEN);
    mavlink_msg_command_long_encode(encoder->system_id, encoder->component_id, msg, data);
    return mavlink_finalize_message(encoder, msg);
}

uint16_t mavlink_encode_command_ack(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                                    const mavlink_command_ack_t* data)
{
    mavlink_fill_header(encoder, msg, MAVLINK_MSG_ID_COMMAND_ACK, MAVLINK_MSG_COMMAND_ACK_LEN);
    mavlink_msg_command_ack_encode(encoder->system_id, encoder->component_id, msg, data);
    return mavlink_finalize_message(encoder, msg);
}

uint16_t mavlink_encode_battery_status(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                                       const mavlink_battery_status_t* data)
{
    mavlink_fill_header(encoder, msg, MAVLINK_MSG_ID_BATTERY_STATUS, MAVLINK_MSG_BATTERY_STATUS_LEN);
    mavlink_msg_battery_status_encode(encoder->system_id, encoder->component_id, msg, data);
    return mavlink_finalize_message(encoder, msg);
}

#define DEFINE_MSG_ENCODE_DECODE(msg_name, msg_type) \
void mavlink_msg_##msg_name##_encode(uint8_t system_id, uint8_t component_id, \
                                     mavlink_message_t* msg, const mavlink_##msg_type##_t* data) \
{ \
    (void)system_id; (void)component_id; \
    memcpy(msg->payload, data, sizeof(mavlink_##msg_type##_t)); \
} \
void mavlink_msg_##msg_name##_decode(const mavlink_message_t* msg, mavlink_##msg_type##_t* data) \
{ \
    memcpy(data, msg->payload, sizeof(mavlink_##msg_type##_t)); \
}

DEFINE_MSG_ENCODE_DECODE(heartbeat, heartbeat)
DEFINE_MSG_ENCODE_DECODE(sys_status, sys_status)
DEFINE_MSG_ENCODE_DECODE(gps_raw_int, gps_raw_int)
DEFINE_MSG_ENCODE_DECODE(attitude, attitude)
DEFINE_MSG_ENCODE_DECODE(global_position_int, global_position_int)
DEFINE_MSG_ENCODE_DECODE(rc_channels_raw, rc_channels_raw)
DEFINE_MSG_ENCODE_DECODE(mission_item, mission_item)
DEFINE_MSG_ENCODE_DECODE(mission_current, mission_current)
DEFINE_MSG_ENCODE_DECODE(mission_count, mission_count)
DEFINE_MSG_ENCODE_DECODE(command_long, command_long)
DEFINE_MSG_ENCODE_DECODE(command_ack, command_ack)
DEFINE_MSG_ENCODE_DECODE(battery_status, battery_status)
