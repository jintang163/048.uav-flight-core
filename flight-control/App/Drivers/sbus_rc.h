#ifndef __SBUS_RC_H__
#define __SBUS_RC_H__

#include "types.h"
#include "flight_config.h"

#define SBUS_FRAME_SIZE 25
#define SBUS_HEADER_BYTE 0x0F
#define SBUS_FOOTER_BYTE 0x00
#define SBUS_CHANNEL_COUNT 16

#define SBUS_FLAG_SIGNAL_LOSS 0x04
#define SBUS_FLAG_FAILSAFE    0x08

typedef struct {
    uint16_t channels[SBUS_CHANNEL_COUNT];
    bool signal_loss;
    bool failsafe;
    uint32_t last_update;
    bool initialized;
} SBUS_Data;

bool sbus_rc_init(void);
void sbus_rc_process_byte(uint8_t byte);
void sbus_rc_process_frame(uint8_t *frame);
uint16_t sbus_rc_get_channel(uint8_t channel);
void sbus_rc_get_channels(uint16_t *channels);
float sbus_rc_get_channel_normalized(uint8_t channel);
bool sbus_rc_is_connected(void);
bool sbus_rc_has_signal_loss(void);
void sbus_rc_get_data(SBUS_Data *data);

#endif
