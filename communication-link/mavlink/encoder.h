#ifndef MAVLINK_ENCODER_H
#define MAVLINK_ENCODER_H

#include "mavlink_v2.h"
#include "messages/common.h"

typedef struct {
    uint8_t system_id;
    uint8_t component_id;
    uint8_t seq;
    mavlink_signing_t* signing;
} mavlink_encoder_t;

void mavlink_encoder_init(mavlink_encoder_t* encoder, uint8_t system_id, uint8_t component_id);
void mavlink_encoder_set_signing(mavlink_encoder_t* encoder, mavlink_signing_t* signing);

uint16_t mavlink_encode_heartbeat(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                                  const mavlink_heartbeat_t* data);
uint16_t mavlink_encode_sys_status(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                                   const mavlink_sys_status_t* data);
uint16_t mavlink_encode_gps_raw_int(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                                    const mavlink_gps_raw_int_t* data);
uint16_t mavlink_encode_attitude(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                                 const mavlink_attitude_t* data);
uint16_t mavlink_encode_global_position_int(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                                            const mavlink_global_position_int_t* data);
uint16_t mavlink_encode_rc_channels_raw(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                                        const mavlink_rc_channels_raw_t* data);
uint16_t mavlink_encode_mission_item(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                                     const mavlink_mission_item_t* data);
uint16_t mavlink_encode_mission_current(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                                        const mavlink_mission_current_t* data);
uint16_t mavlink_encode_mission_count(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                                      const mavlink_mission_count_t* data);
uint16_t mavlink_encode_command_long(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                                     const mavlink_command_long_t* data);
uint16_t mavlink_encode_command_ack(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                                    const mavlink_command_ack_t* data);
uint16_t mavlink_encode_battery_status(mavlink_encoder_t* encoder, mavlink_message_t* msg,
                                       const mavlink_battery_status_t* data);

#endif
