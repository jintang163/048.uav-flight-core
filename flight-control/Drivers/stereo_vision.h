#ifndef __STEREO_VISION_H__
#define __STEREO_VISION_H__

#include "stm32f4xx_hal.h"

#define STEREO_MAX_OBSTACLES     8
#define STEREO_DETECTION_RANGE   15.0f
#define STEREO_UART_BUF_SIZE     256

#define STEREO_HEADER_0  0xAA
#define STEREO_HEADER_1  0x55

#define STEREO_TIMEOUT_MS  5000

#define STEREO_OBSTACLE_SIZE  17

typedef struct {
    float distance;
    float angle;
    float size;
    float confidence;
    uint8_t direction;
} stereo_obstacle_t;

typedef struct {
    stereo_obstacle_t obstacles[STEREO_MAX_OBSTACLES];
    uint8_t obstacle_count;
    uint32_t timestamp;
} stereo_data_t;

bool stereo_vision_init(void);
void stereo_vision_process_byte(uint8_t byte);
void stereo_vision_get_data(stereo_data_t *data);
bool stereo_vision_is_connected(void);

#endif
