#include "gps_ublox.h"
#include "main.h"

extern UART_HandleTypeDef huart1;

static GPS_Data gps_data;
static uint8_t parse_state = 0;
static uint16_t parse_index = 0;
static uint8_t rx_buffer[256];
static UBX_Header rx_header;

static uint16_t ubx_checksum(uint8_t *data, uint16_t length)
{
    uint8_t ck_a = 0, ck_b = 0;
    for (uint16_t i = 0; i < length; i++) {
        ck_a += data[i];
        ck_b += ck_a;
    }
    return (uint16_t)((ck_b << 8) | ck_a);
}

static void gps_ublox_parse_message(void)
{
    if (rx_header.class == UBX_CLASS_NAV && rx_header.id == UBX_NAV_POSLLH) {
        if (rx_header.length == sizeof(UBX_NAV_POSLLH_t)) {
            UBX_NAV_POSLLH_t *posllh = (UBX_NAV_POSLLH_t *)rx_buffer;
            gps_data.position.lat = posllh->lat;
            gps_data.position.lon = posllh->lon;
            gps_data.position.alt = posllh->hMSL;
            gps_data.timestamp = posllh->iTOW;
            gps_data.last_update = HAL_GetTick();
        }
    } else if (rx_header.class == UBX_CLASS_NAV && rx_header.id == UBX_NAV_VELNED) {
        if (rx_header.length == sizeof(UBX_NAV_VELNED_t)) {
            UBX_NAV_VELNED_t *velned = (UBX_NAV_VELNED_t *)rx_buffer;
            gps_data.velocity.vn = (float)velned->velN / 1000.0f;
            gps_data.velocity.ve = (float)velned->velE / 1000.0f;
            gps_data.velocity.vd = (float)velned->velD / 1000.0f;
            gps_data.ground_speed = (float)velned->gSpeed / 1000.0f;
            gps_data.heading = (float)velned->heading / 100000.0f;
        }
    } else if (rx_header.class == UBX_CLASS_NAV && rx_header.id == UBX_NAV_STATUS) {
        if (rx_header.length == sizeof(UBX_NAV_STATUS_t)) {
            UBX_NAV_STATUS_t *status = (UBX_NAV_STATUS_t *)rx_buffer;
            gps_data.fix_type = status->gpsFix;
        }
    }
}

bool gps_ublox_init(void)
{
    gps_data.initialized = true;
    gps_data.last_update = 0;
    gps_data.fix_type = 0;
    gps_data.satellites = 0;
    parse_state = 0;
    parse_index = 0;

    return true;
}

void gps_ublox_process_byte(uint8_t byte)
{
    switch (parse_state) {
        case 0:
            if (byte == UBX_SYNC1) {
                parse_state = 1;
            }
            break;
        case 1:
            if (byte == UBX_SYNC2) {
                parse_state = 2;
                parse_index = 0;
            } else {
                parse_state = 0;
            }
            break;
        case 2:
            ((uint8_t *)&rx_header)[parse_index++] = byte;
            if (parse_index >= sizeof(UBX_Header)) {
                parse_state = 3;
                parse_index = 0;
            }
            break;
        case 3:
            rx_buffer[parse_index++] = byte;
            if (parse_index >= rx_header.length) {
                parse_state = 4;
                parse_index = 0;
            }
            break;
        case 4:
            ((uint8_t *)&rx_header)[parse_index++] = byte;
            if (parse_index >= 2) {
                uint16_t calc_ck = ubx_checksum((uint8_t *)&rx_header.class, rx_header.length + 4);
                uint16_t rx_ck = (rx_buffer[rx_header.length + 1] << 8) | rx_buffer[rx_header.length];
                if (calc_ck == rx_ck) {
                    gps_ublox_parse_message();
                }
                parse_state = 0;
            }
            break;
        default:
            parse_state = 0;
            break;
    }
}

void gps_ublox_get_position(GPSPosition *pos)
{
    *pos = gps_data.position;
}

void gps_ublox_get_velocity(GPSVelocity *vel)
{
    *vel = gps_data.velocity;
}

uint8_t gps_ublox_get_fix_type(void)
{
    return gps_data.fix_type;
}

uint8_t gps_ublox_get_satellites(void)
{
    return gps_data.satellites;
}

float gps_ublox_get_ground_speed(void)
{
    return gps_data.ground_speed;
}

float gps_ublox_get_heading(void)
{
    return gps_data.heading;
}

bool gps_ublox_is_healthy(void)
{
    if (!gps_data.initialized) {
        return false;
    }

    uint32_t now = HAL_GetTick();
    if (now - gps_data.last_update > 5000) {
        return false;
    }

    return gps_data.fix_type >= 3;
}

void gps_ublox_get_data(GPS_Data *data)
{
    *data = gps_data;
}
