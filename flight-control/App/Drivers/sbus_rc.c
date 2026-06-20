#include "sbus_rc.h"
#include "main.h"

extern UART_HandleTypeDef huart2;

static SBUS_Data sbus_data;
static uint8_t rx_buffer[SBUS_FRAME_SIZE];
static uint8_t rx_index = 0;
static uint32_t last_byte_time = 0;

bool sbus_rc_init(void)
{
    for (int i = 0; i < SBUS_CHANNEL_COUNT; i++) {
        sbus_data.channels[i] = SBUS_CHANNEL_MID;
    }
    sbus_data.signal_loss = false;
    sbus_data.failsafe = false;
    sbus_data.initialized = true;
    sbus_data.last_update = 0;
    rx_index = 0;

    return true;
}

void sbus_rc_process_byte(uint8_t byte)
{
    uint32_t now = HAL_GetTick();

    if (now - last_byte_time > 5) {
        rx_index = 0;
    }
    last_byte_time = now;

    if (rx_index == 0) {
        if (byte == SBUS_HEADER_BYTE) {
            rx_buffer[rx_index++] = byte;
        }
    } else if (rx_index < SBUS_FRAME_SIZE) {
        rx_buffer[rx_index++] = byte;

        if (rx_index == SBUS_FRAME_SIZE) {
            sbus_rc_process_frame(rx_buffer);
            rx_index = 0;
        }
    } else {
        rx_index = 0;
    }
}

void sbus_rc_process_frame(uint8_t *frame)
{
    if (frame[0] != SBUS_HEADER_BYTE) {
        return;
    }

    sbus_data.channels[0]  = (uint16_t)((frame[1] | frame[2] << 8) & 0x07FF);
    sbus_data.channels[1]  = (uint16_t)((frame[2] >> 3 | frame[3] << 5) & 0x07FF);
    sbus_data.channels[2]  = (uint16_t)((frame[3] >> 6 | frame[4] << 2 | frame[5] << 10) & 0x07FF);
    sbus_data.channels[3]  = (uint16_t)((frame[5] >> 1 | frame[6] << 7) & 0x07FF);
    sbus_data.channels[4]  = (uint16_t)((frame[6] >> 4 | frame[7] << 4) & 0x07FF);
    sbus_data.channels[5]  = (uint16_t)((frame[7] >> 7 | frame[8] << 1 | frame[9] << 9) & 0x07FF);
    sbus_data.channels[6]  = (uint16_t)((frame[9] >> 2 | frame[10] << 6) & 0x07FF);
    sbus_data.channels[7]  = (uint16_t)((frame[10] >> 5 | frame[11] << 3) & 0x07FF);
    sbus_data.channels[8]  = (uint16_t)((frame[12] | frame[13] << 8) & 0x07FF);
    sbus_data.channels[9]  = (uint16_t)((frame[13] >> 3 | frame[14] << 5) & 0x07FF);
    sbus_data.channels[10] = (uint16_t)((frame[14] >> 6 | frame[15] << 2 | frame[16] << 10) & 0x07FF);
    sbus_data.channels[11] = (uint16_t)((frame[16] >> 1 | frame[17] << 7) & 0x07FF);
    sbus_data.channels[12] = (uint16_t)((frame[17] >> 4 | frame[18] << 4) & 0x07FF);
    sbus_data.channels[13] = (uint16_t)((frame[18] >> 7 | frame[19] << 1 | frame[20] << 9) & 0x07FF);
    sbus_data.channels[14] = (uint16_t)((frame[20] >> 2 | frame[21] << 6) & 0x07FF);
    sbus_data.channels[15] = (uint16_t)((frame[21] >> 5 | frame[22] << 3) & 0x07FF);

    uint8_t flags = frame[23];
    sbus_data.signal_loss = (flags & SBUS_FLAG_SIGNAL_LOSS) != 0;
    sbus_data.failsafe = (flags & SBUS_FLAG_FAILSAFE) != 0;

    for (int i = 0; i < SBUS_CHANNEL_COUNT; i++) {
        sbus_data.channels[i] = CONSTRAIN(sbus_data.channels[i], SBUS_CHANNEL_MIN, SBUS_CHANNEL_MAX);
    }

    sbus_data.last_update = HAL_GetTick();
}

uint16_t sbus_rc_get_channel(uint8_t channel)
{
    if (channel >= SBUS_CHANNEL_COUNT) {
        return SBUS_CHANNEL_MID;
    }
    return sbus_data.channels[channel];
}

void sbus_rc_get_channels(uint16_t *channels)
{
    for (int i = 0; i < SBUS_CHANNEL_COUNT; i++) {
        channels[i] = sbus_data.channels[i];
    }
}

float sbus_rc_get_channel_normalized(uint8_t channel)
{
    if (channel >= SBUS_CHANNEL_COUNT) {
        return 0.0f;
    }

    uint16_t value = sbus_data.channels[channel];
    float normalized = (float)(value - SBUS_CHANNEL_MID) / (float)(SBUS_CHANNEL_MAX - SBUS_CHANNEL_MID) * 2.0f;
    return CONSTRAIN(normalized, -1.0f, 1.0f);
}

bool sbus_rc_is_connected(void)
{
    if (!sbus_data.initialized) {
        return false;
    }

    uint32_t now = HAL_GetTick();
    if (now - sbus_data.last_update > RC_SIGNAL_LOSS_TIMEOUT) {
        return false;
    }

    return !sbus_data.signal_loss && !sbus_data.failsafe;
}

bool sbus_rc_has_signal_loss(void)
{
    return sbus_data.signal_loss || sbus_data.failsafe;
}

void sbus_rc_get_data(SBUS_Data *data)
{
    *data = sbus_data;
}
