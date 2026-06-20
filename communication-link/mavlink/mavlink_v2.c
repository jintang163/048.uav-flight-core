#include "mavlink_v2.h"
#include <string.h>

#define MAVLINK_PARSE_STATE_UNINIT      0
#define MAVLINK_PARSE_STATE_IDLE        1
#define MAVLINK_PARSE_STATE_GOT_STX     2
#define MAVLINK_PARSE_STATE_GOT_HEADER  3
#define MAVLINK_PARSE_STATE_GOT_PAYLOAD 4
#define MAVLINK_PARSE_STATE_GOT_CRC1    5

typedef struct {
    uint32_t msgid;
    uint8_t crc_extra;
} mavlink_crc_entry_t;

static const mavlink_crc_entry_t mavlink_crc_table[] = {
    {0,   50},
    {1,   124},
    {24,  137},
    {30,  147},
    {33,  104},
    {35,  246},
    {39,  254},
    {42,  28},
    {44,  221},
    {76,  152},
    {77,  143},
    {147, 154},
};

#define MAVLINK_CRC_TABLE_SIZE (sizeof(mavlink_crc_table) / sizeof(mavlink_crc_table[0]))

uint16_t mavlink_get_msg_crc_extra(uint32_t msgid)
{
    for (size_t i = 0; i < MAVLINK_CRC_TABLE_SIZE; i++) {
        if (mavlink_crc_table[i].msgid == msgid) {
            return mavlink_crc_table[i].crc_extra;
        }
    }
    return 0;
}

void mavlink_init_parser(mavlink_parser_t* parser)
{
    memset(parser, 0, sizeof(mavlink_parser_t));
    parser->state = MAVLINK_PARSE_STATE_IDLE;
    parser->checksum.ck_a = 0;
    parser->checksum.ck_b = 0;
}

void mavlink_crc_accumulate(uint8_t b, mavlink_checksum_t* crc)
{
    uint8_t tmp = b ^ crc->ck_a;
    crc->ck_a = crc->ck_b;
    crc->ck_b = tmp;
    tmp = (uint8_t)((tmp << 4) | (tmp >> 4));
    crc->ck_b ^= tmp;
    tmp = (uint8_t)((tmp & 0x0F) << 3);
    crc->ck_b ^= tmp;
    tmp = (uint8_t)((crc->ck_b & 0xFF) >> 4);
    crc->ck_b ^= tmp;
}

uint16_t mavlink_crc_calculate(const uint8_t* data, size_t len, uint16_t crc)
{
    mavlink_checksum_t c;
    c.ck_a = (uint8_t)(crc & 0xFF);
    c.ck_b = (uint8_t)(crc >> 8);
    for (size_t i = 0; i < len; i++) {
        mavlink_crc_accumulate(data[i], &c);
    }
    return (uint16_t)(c.ck_a | ((uint16_t)c.ck_b << 8));
}

static void mavlink_reset_parser(mavlink_parser_t* parser)
{
    parser->state = MAVLINK_PARSE_STATE_IDLE;
    parser->msg_idx = 0;
    parser->checksum.ck_a = 0;
    parser->checksum.ck_b = 0;
    parser->signature_present = false;
}

bool mavlink_parse_byte(mavlink_parser_t* parser, uint8_t byte, mavlink_message_t* message)
{
    bool msg_received = false;

    switch (parser->state) {
    case MAVLINK_PARSE_STATE_UNINIT:
    case MAVLINK_PARSE_STATE_IDLE:
        if (byte == MAVLINK_STX_V2) {
            parser->state = MAVLINK_PARSE_STATE_GOT_STX;
            parser->msg_idx = 1;
            parser->header[0] = byte;
            parser->checksum.ck_a = 0;
            parser->checksum.ck_b = 0;
            parser->signature_present = false;
            mavlink_crc_accumulate(byte, &parser->checksum);
        }
        break;

    case MAVLINK_PARSE_STATE_GOT_STX:
        parser->header[parser->msg_idx++] = byte;
        mavlink_crc_accumulate(byte, &parser->checksum);
        if (parser->msg_idx >= MAVLINK_HEADER_LEN_V2) {
            memcpy(&message->header, parser->header, MAVLINK_HEADER_LEN_V2);
            if (message->header.incompat_flags & 0x01) {
                parser->signature_present = true;
            }
            parser->state = MAVLINK_PARSE_STATE_GOT_HEADER;
            parser->msg_idx = 0;
        }
        break;

    case MAVLINK_PARSE_STATE_GOT_HEADER:
        parser->payload[parser->msg_idx++] = byte;
        mavlink_crc_accumulate(byte, &parser->checksum);
        if (parser->msg_idx >= message->header.payload_len) {
            parser->state = MAVLINK_PARSE_STATE_GOT_PAYLOAD;
            parser->msg_idx = 0;
        }
        break;

    case MAVLINK_PARSE_STATE_GOT_PAYLOAD: {
        uint16_t crc_calc = (uint16_t)(parser->checksum.ck_a | ((uint16_t)parser->checksum.ck_b << 8));
        uint16_t crc_extra = mavlink_get_msg_crc_extra(message->header.msgid);
        mavlink_checksum_t crc_tmp;
        crc_tmp.ck_a = (uint8_t)(crc_calc & 0xFF);
        crc_tmp.ck_b = (uint8_t)(crc_calc >> 8);
        mavlink_crc_accumulate((uint8_t)(crc_extra & 0xFF), &crc_tmp);
        crc_calc = (uint16_t)(crc_tmp.ck_a | ((uint16_t)crc_tmp.ck_b << 8));

        if (parser->msg_idx == 0) {
            if (byte != (uint8_t)(crc_calc & 0xFF)) {
                mavlink_reset_parser(parser);
                parser->parse_error = 1;
                break;
            }
            parser->msg_idx = 1;
        } else if (parser->msg_idx == 1) {
            if (byte != (uint8_t)(crc_calc >> 8)) {
                mavlink_reset_parser(parser);
                parser->parse_error = 2;
                break;
            }
            message->checksum = crc_calc;
            memcpy(message->payload, parser->payload, message->header.payload_len);
            message->signature_present = parser->signature_present;

            if (parser->signature_present) {
                parser->state = MAVLINK_PARSE_STATE_GOT_CRC1;
                parser->msg_idx = 0;
            } else {
                msg_received = true;
                mavlink_reset_parser(parser);
            }
        }
        break;
    }

    case MAVLINK_PARSE_STATE_GOT_CRC1:
        message->signature[parser->msg_idx++] = byte;
        if (parser->msg_idx >= MAVLINK_SIGNATURE_LEN) {
            msg_received = true;
            mavlink_reset_parser(parser);
        }
        break;

    default:
        mavlink_reset_parser(parser);
        break;
    }

    return msg_received;
}

uint16_t mavlink_msg_to_buf(const mavlink_message_t* msg, uint8_t* buf, size_t buf_len)
{
    size_t total_len = MAVLINK_HEADER_LEN_V2 + msg->header.payload_len + MAVLINK_CHECKSUM_LEN;
    if (msg->signature_present) {
        total_len += MAVLINK_SIGNATURE_LEN;
    }
    if (buf_len < total_len) {
        return 0;
    }

    memcpy(buf, &msg->header, MAVLINK_HEADER_LEN_V2);
    memcpy(buf + MAVLINK_HEADER_LEN_V2, msg->payload, msg->header.payload_len);

    mavlink_checksum_t crc;
    crc.ck_a = 0;
    crc.ck_b = 0;
    for (size_t i = 0; i < MAVLINK_HEADER_LEN_V2; i++) {
        mavlink_crc_accumulate(buf[i], &crc);
    }
    for (size_t i = 0; i < msg->header.payload_len; i++) {
        mavlink_crc_accumulate(msg->payload[i], &crc);
    }
    uint16_t crc_extra = mavlink_get_msg_crc_extra(msg->header.msgid);
    mavlink_crc_accumulate((uint8_t)crc_extra, &crc);

    buf[MAVLINK_HEADER_LEN_V2 + msg->header.payload_len] = crc.ck_a;
    buf[MAVLINK_HEADER_LEN_V2 + msg->header.payload_len + 1] = crc.ck_b;

    size_t offset = MAVLINK_HEADER_LEN_V2 + msg->header.payload_len + MAVLINK_CHECKSUM_LEN;
    if (msg->signature_present) {
        memcpy(buf + offset, msg->signature, MAVLINK_SIGNATURE_LEN);
    }

    return (uint16_t)total_len;
}

bool mavlink_buf_to_msg(const uint8_t* buf, size_t buf_len, mavlink_message_t* msg)
{
    if (buf_len < MAVLINK_HEADER_LEN_V2 + MAVLINK_CHECKSUM_LEN) {
        return false;
    }
    if (buf[0] != MAVLINK_STX_V2) {
        return false;
    }

    memcpy(&msg->header, buf, MAVLINK_HEADER_LEN_V2);
    size_t payload_len = msg->header.payload_len;
    size_t expected_len = MAVLINK_HEADER_LEN_V2 + payload_len + MAVLINK_CHECKSUM_LEN;
    bool has_sig = (msg->header.incompat_flags & 0x01) != 0;
    if (has_sig) {
        expected_len += MAVLINK_SIGNATURE_LEN;
    }
    if (buf_len < expected_len) {
        return false;
    }

    memcpy(msg->payload, buf + MAVLINK_HEADER_LEN_V2, payload_len);
    msg->checksum = (uint16_t)(buf[MAVLINK_HEADER_LEN_V2 + payload_len] |
                              ((uint16_t)buf[MAVLINK_HEADER_LEN_V2 + payload_len + 1] << 8));
    msg->signature_present = has_sig;
    if (has_sig) {
        memcpy(msg->signature, buf + MAVLINK_HEADER_LEN_V2 + payload_len + MAVLINK_CHECKSUM_LEN,
               MAVLINK_SIGNATURE_LEN);
    }

    mavlink_checksum_t crc;
    crc.ck_a = 0;
    crc.ck_b = 0;
    for (size_t i = 0; i < MAVLINK_HEADER_LEN_V2; i++) {
        mavlink_crc_accumulate(buf[i], &crc);
    }
    for (size_t i = 0; i < payload_len; i++) {
        mavlink_crc_accumulate(buf[MAVLINK_HEADER_LEN_V2 + i], &crc);
    }
    uint16_t crc_extra = mavlink_get_msg_crc_extra(msg->header.msgid);
    mavlink_crc_accumulate((uint8_t)crc_extra, &crc);
    uint16_t calc_crc = (uint16_t)(crc.ck_a | ((uint16_t)crc.ck_b << 8));

    return (calc_crc == msg->checksum);
}

void mavlink_signing_init(mavlink_signing_t* signing, const uint8_t* secret_key, uint32_t link_id)
{
    signing->link_id = link_id;
    signing->timestamp = 0;
    signing->enabled = true;
    if (secret_key) {
        memcpy(signing->secret_key, secret_key, 32);
    } else {
        memset(signing->secret_key, 0, 32);
    }
}

static void mavlink_hmac_sha256_simple(const uint8_t* key, size_t key_len,
                                       const uint8_t* data, size_t data_len,
                                       uint8_t* out)
{
    (void)key;
    (void)key_len;
    (void)data;
    (void)data_len;
    memset(out, 0, 32);
    for (size_t i = 0; i < data_len && i < 32; i++) {
        out[i] = data[i] ^ key[i % key_len];
    }
}

bool mavlink_sign_message(mavlink_message_t* msg, mavlink_signing_t* signing)
{
    if (!signing || !signing->enabled) {
        return false;
    }

    uint8_t hash[32];
    uint8_t data[MAVLINK_HEADER_LEN_V2 + MAVLINK_MAX_PAYLOAD_LEN + 8];
    size_t data_len = 0;

    memcpy(data, &msg->header, MAVLINK_HEADER_LEN_V2);
    data_len += MAVLINK_HEADER_LEN_V2;
    memcpy(data + data_len, msg->payload, msg->header.payload_len);
    data_len += msg->header.payload_len;
    memcpy(data + data_len, &signing->timestamp, 8);
    data_len += 8;

    mavlink_hmac_sha256_simple(signing->secret_key, 32, data, data_len, hash);

    msg->signature[0] = (uint8_t)(signing->link_id & 0xFF);
    msg->signature[1] = (uint8_t)((signing->timestamp >> 0) & 0xFF);
    msg->signature[2] = (uint8_t)((signing->timestamp >> 8) & 0xFF);
    msg->signature[3] = (uint8_t)((signing->timestamp >> 16) & 0xFF);
    msg->signature[4] = (uint8_t)((signing->timestamp >> 24) & 0xFF);
    msg->signature[5] = (uint8_t)((signing->timestamp >> 32) & 0xFF);
    msg->signature[6] = (uint8_t)((signing->timestamp >> 40) & 0xFF);
    memcpy(msg->signature + 7, hash, 6);

    msg->header.incompat_flags |= 0x01;
    msg->signature_present = true;
    signing->timestamp++;

    return true;
}

bool mavlink_verify_signature(const mavlink_message_t* msg, const mavlink_signing_t* signing)
{
    if (!signing || !signing->enabled || !msg->signature_present) {
        return false;
    }

    uint8_t hash[32];
    uint8_t data[MAVLINK_HEADER_LEN_V2 + MAVLINK_MAX_PAYLOAD_LEN + 8];
    size_t data_len = 0;

    uint64_t timestamp = 0;
    memcpy(&timestamp, msg->signature + 1, 6);

    memcpy(data, &msg->header, MAVLINK_HEADER_LEN_V2);
    data_len += MAVLINK_HEADER_LEN_V2;
    memcpy(data + data_len, msg->payload, msg->header.payload_len);
    data_len += msg->header.payload_len;
    memcpy(data + data_len, &timestamp, 8);
    data_len += 8;

    mavlink_hmac_sha256_simple(signing->secret_key, 32, data, data_len, hash);

    return (memcmp(msg->signature + 7, hash, 6) == 0);
}
