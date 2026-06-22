#include "mmwave_radar.h"
#include <string.h>

static mmwave_data_t mmwave_data;
static uint8_t uart_buf[MMWAVE_UART_BUF_SIZE];
static uint8_t parse_state = 0;
static uint16_t buf_idx = 0;
static uint32_t frame_total_len = 0;
static uint32_t num_tlvs = 0;
static uint32_t tlv_type = 0;
static uint32_t tlv_len = 0;
static uint32_t tlv_bytes_read = 0;
static uint8_t frame_header_buf[MMWAVE_FRAME_HEADER_LEN];
static uint8_t tlv_header_buf[MMWAVE_TLV_HEADER_LEN];
static uint32_t last_update = 0;

bool mmwave_init(void)
{
    memset(&mmwave_data, 0, sizeof(mmwave_data_t));
    memset(uart_buf, 0, MMWAVE_UART_BUF_SIZE);
    memset(frame_header_buf, 0, MMWAVE_FRAME_HEADER_LEN);
    memset(tlv_header_buf, 0, MMWAVE_TLV_HEADER_LEN);
    parse_state = 0;
    buf_idx = 0;
    frame_total_len = 0;
    num_tlvs = 0;
    tlv_type = 0;
    tlv_len = 0;
    tlv_bytes_read = 0;
    last_update = 0;
    return true;
}

void mmwave_process_byte(uint8_t byte)
{
    switch (parse_state) {
        case 0:
            if (byte == MMWAVE_MAGIC_0) {
                parse_state = 1;
            }
            break;
        case 1:
            if (byte == MMWAVE_MAGIC_1) {
                parse_state = 2;
            } else {
                parse_state = 0;
            }
            break;
        case 2:
            if (byte == MMWAVE_MAGIC_2) {
                parse_state = 3;
            } else {
                parse_state = 0;
            }
            break;
        case 3:
            if (byte == MMWAVE_MAGIC_3) {
                parse_state = 4;
                buf_idx = 0;
            } else {
                parse_state = 0;
            }
            break;
        case 4:
            frame_header_buf[buf_idx++] = byte;
            if (buf_idx >= MMWAVE_FRAME_HEADER_LEN) {
                frame_total_len = *(uint32_t *)&frame_header_buf[4];
                num_tlvs = *(uint32_t *)&frame_header_buf[24];
                buf_idx = 0;
                if (num_tlvs > 0) {
                    parse_state = 5;
                } else {
                    parse_state = 0;
                }
            }
            break;
        case 5:
            tlv_header_buf[buf_idx++] = byte;
            if (buf_idx >= MMWAVE_TLV_HEADER_LEN) {
                tlv_type = *(uint32_t *)&tlv_header_buf[0];
                tlv_len = *(uint32_t *)&tlv_header_buf[4];
                tlv_bytes_read = 0;
                buf_idx = 0;
                if (tlv_len > 0 && tlv_len <= MMWAVE_UART_BUF_SIZE) {
                    parse_state = 6;
                } else {
                    num_tlvs--;
                    if (num_tlvs == 0) {
                        parse_state = 0;
                    } else {
                        parse_state = 5;
                    }
                }
            }
            break;
        case 6:
            uart_buf[tlv_bytes_read++] = byte;
            if (tlv_bytes_read >= tlv_len) {
                if (tlv_type == MMWAVE_TLV_DETECTED_POINTS) {
                    uint16_t point_size = sizeof(mmwave_target_t);
                    uint8_t count = tlv_len / point_size;
                    if (count > MMWAVE_MAX_TARGETS) {
                        count = MMWAVE_MAX_TARGETS;
                    }
                    mmwave_data.target_count = count;
                    for (uint8_t i = 0; i < count; i++) {
                        memcpy(&mmwave_data.targets[i], &uart_buf[i * point_size], point_size);
                    }
                    mmwave_data.timestamp = HAL_GetTick();
                    last_update = HAL_GetTick();
                }
                num_tlvs--;
                if (num_tlvs == 0) {
                    parse_state = 0;
                } else {
                    parse_state = 5;
                    buf_idx = 0;
                }
            }
            break;
        default:
            parse_state = 0;
            break;
    }
}

void mmwave_get_data(mmwave_data_t *data)
{
    *data = mmwave_data;
}

bool mmwave_is_connected(void)
{
    if (last_update == 0) {
        return false;
    }
    return (HAL_GetTick() - last_update) < MMWAVE_TIMEOUT_MS;
}
