#ifndef MAVLINK_V2_H
#define MAVLINK_V2_H

#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>

#define MAVLINK_STX_V2             0xFD
#define MAVLINK_HEADER_LEN_V2      10
#define MAVLINK_CHECKSUM_LEN       2
#define MAVLINK_SIGNATURE_LEN      13
#define MAVLINK_MAX_PAYLOAD_LEN    255
#define MAVLINK_MAX_PACKET_LEN     (MAVLINK_HEADER_LEN_V2 + MAVLINK_MAX_PAYLOAD_LEN + MAVLINK_CHECKSUM_LEN + MAVLINK_SIGNATURE_LEN)

#define MAVLINK_SYS_ID_ALL         0
#define MAVLINK_COMP_ID_ALL        0

#define MAVLINK_TYPE_UINT8_T       0
#define MAVLINK_TYPE_INT8_T        1
#define MAVLINK_TYPE_UINT16_T      2
#define MAVLINK_TYPE_INT16_T       3
#define MAVLINK_TYPE_UINT32_T      4
#define MAVLINK_TYPE_INT32_T       5
#define MAVLINK_TYPE_UINT64_T      6
#define MAVLINK_TYPE_INT64_T       7
#define MAVLINK_TYPE_FLOAT         9
#define MAVLINK_TYPE_DOUBLE        10
#define MAVLINK_TYPE_CHAR          13

typedef struct __attribute__((packed)) {
    uint8_t magic;
    uint8_t payload_len;
    uint8_t incompat_flags;
    uint8_t compat_flags;
    uint8_t seq;
    uint8_t sysid;
    uint8_t compid;
    uint32_t msgid:24;
} mavlink_header_t;

typedef struct {
    uint8_t header[MAVLINK_HEADER_LEN_V2];
    uint8_t payload[MAVLINK_MAX_PAYLOAD_LEN];
} mavlink_buffer_t;

typedef struct {
    mavlink_header_t header;
    uint8_t payload[MAVLINK_MAX_PAYLOAD_LEN];
    uint16_t checksum;
    bool signature_present;
    uint8_t signature[MAVLINK_SIGNATURE_LEN];
} mavlink_message_t;

typedef struct {
    uint8_t ck_a;
    uint8_t ck_b;
} mavlink_checksum_t;

typedef struct {
    uint8_t state;
    uint16_t msg_idx;
    uint8_t header[MAVLINK_HEADER_LEN_V2];
    uint8_t payload[MAVLINK_MAX_PAYLOAD_LEN + MAVLINK_CHECKSUM_LEN + MAVLINK_SIGNATURE_LEN];
    mavlink_checksum_t checksum;
    bool signature_present;
    uint8_t parse_error;
} mavlink_parser_t;

typedef struct {
    uint32_t link_id;
    uint64_t timestamp;
    uint8_t secret_key[32];
    bool enabled;
} mavlink_signing_t;

void mavlink_init_parser(mavlink_parser_t* parser);
bool mavlink_parse_byte(mavlink_parser_t* parser, uint8_t byte, mavlink_message_t* message);
uint16_t mavlink_crc_calculate(const uint8_t* data, size_t len, uint16_t crc);
void mavlink_crc_accumulate(uint8_t b, mavlink_checksum_t* crc);
uint16_t mavlink_get_msg_crc_extra(uint32_t msgid);

void mavlink_signing_init(mavlink_signing_t* signing, const uint8_t* secret_key, uint32_t link_id);
bool mavlink_sign_message(mavlink_message_t* msg, mavlink_signing_t* signing);
bool mavlink_verify_signature(const mavlink_message_t* msg, const mavlink_signing_t* signing);

uint16_t mavlink_msg_to_buf(const mavlink_message_t* msg, uint8_t* buf, size_t buf_len);
bool mavlink_buf_to_msg(const uint8_t* buf, size_t buf_len, mavlink_message_t* msg);

static inline uint8_t mavlink_get_system_id(const mavlink_message_t* msg) { return msg->header.sysid; }
static inline uint8_t mavlink_get_component_id(const mavlink_message_t* msg) { return msg->header.compid; }
static inline uint32_t mavlink_get_msg_id(const mavlink_message_t* msg) { return msg->header.msgid; }
static inline uint8_t mavlink_get_payload_len(const mavlink_message_t* msg) { return msg->header.payload_len; }

#endif
