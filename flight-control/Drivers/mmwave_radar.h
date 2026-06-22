#ifndef __MMWAVE_RADAR_H__
#define __MMWAVE_RADAR_H__

#include "stm32f4xx_hal.h"

#define MMWAVE_UART_BUF_SIZE    256
#define MMWAVE_MAX_TARGETS      8
#define MMWAVE_DETECTION_RANGE  15.0f

#define MMWAVE_MAGIC_0  0x02
#define MMWAVE_MAGIC_1  0x01
#define MMWAVE_MAGIC_2  0x03
#define MMWAVE_MAGIC_3  0x05

#define MMWAVE_TLV_DETECTED_POINTS  0x01
#define MMWAVE_FRAME_HEADER_LEN     36
#define MMWAVE_TLV_HEADER_LEN       8

#define MMWAVE_TIMEOUT_MS  5000

typedef struct {
    float distance;
    float angle;
    float size;
    float confidence;
    float velocity;
} mmwave_target_t;

typedef struct {
    mmwave_target_t targets[MMWAVE_MAX_TARGETS];
    uint8_t target_count;
    uint32_t timestamp;
} mmwave_data_t;

bool mmwave_init(void);
void mmwave_process_byte(uint8_t byte);
void mmwave_get_data(mmwave_data_t *data);
bool mmwave_is_connected(void);

#endif
