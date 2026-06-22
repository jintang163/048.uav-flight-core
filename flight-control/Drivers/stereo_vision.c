#include "stereo_vision.h"
#include <string.h>

static stereo_data_t stereo_data;
static uint8_t uart_buf[STEREO_UART_BUF_SIZE];
static uint8_t parse_state = 0;
static uint16_t buf_idx = 0;
static uint16_t payload_len = 0;
static uint8_t obstacle_count = 0;
static uint16_t data_bytes_read = 0;
static uint8_t header_buf[3];
static uint32_t last_update = 0;

bool stereo_vision_init(void)
{
    memset(&stereo_data, 0, sizeof(stereo_data_t));
    memset(uart_buf, 0, STEREO_UART_BUF_SIZE);
    memset(header_buf, 0, 3);
    parse_state = 0;
    buf_idx = 0;
    payload_len = 0;
    obstacle_count = 0;
    data_bytes_read = 0;
    last_update = 0;
    return true;
}

void stereo_vision_process_byte(uint8_t byte)
{
    switch (parse_state) {
        case 0:
            if (byte == STEREO_HEADER_0) {
                parse_state = 1;
            }
            break;
        case 1:
            if (byte == STEREO_HEADER_1) {
                parse_state = 2;
                buf_idx = 0;
            } else {
                parse_state = 0;
            }
            break;
        case 2:
            header_buf[buf_idx++] = byte;
            if (buf_idx >= 3) {
                payload_len = *(uint16_t *)&header_buf[0];
                obstacle_count = header_buf[2];
                data_bytes_read = 0;
                buf_idx = 0;
                if (obstacle_count > STEREO_MAX_OBSTACLES) {
                    obstacle_count = STEREO_MAX_OBSTACLES;
                }
                if (payload_len > 0 && payload_len <= STEREO_UART_BUF_SIZE) {
                    parse_state = 3;
                } else {
                    parse_state = 0;
                }
            }
            break;
        case 3:
            uart_buf[data_bytes_read++] = byte;
            if (data_bytes_read >= payload_len) {
                stereo_data.obstacle_count = obstacle_count;
                for (uint8_t i = 0; i < obstacle_count; i++) {
                    uint16_t offset = i * STEREO_OBSTACLE_SIZE;
                    memcpy(&stereo_data.obstacles[i].distance, &uart_buf[offset], 4);
                    memcpy(&stereo_data.obstacles[i].angle, &uart_buf[offset + 4], 4);
                    memcpy(&stereo_data.obstacles[i].size, &uart_buf[offset + 8], 4);
                    memcpy(&stereo_data.obstacles[i].confidence, &uart_buf[offset + 12], 4);
                    stereo_data.obstacles[i].direction = uart_buf[offset + 16];
                }
                stereo_data.timestamp = HAL_GetTick();
                last_update = HAL_GetTick();
                parse_state = 0;
            }
            break;
        default:
            parse_state = 0;
            break;
    }
}

void stereo_vision_get_data(stereo_data_t *data)
{
    *data = stereo_data;
}

bool stereo_vision_is_connected(void)
{
    if (last_update == 0) {
        return false;
    }
    return (HAL_GetTick() - last_update) < STEREO_TIMEOUT_MS;
}
