#ifndef __GPS_UBLOX_H__
#define __GPS_UBLOX_H__

#include "types.h"
#include "flight_config.h"

#define UBX_SYNC1 0xB5
#define UBX_SYNC2 0x62

#define UBX_CLASS_NAV 0x01
#define UBX_CLASS_CFG 0x06

#define UBX_NAV_POSLLH 0x02
#define UBX_NAV_VELNED 0x12
#define UBX_NAV_STATUS 0x03
#define UBX_NAV_SOL 0x06

#define UBX_CFG_PRT 0x00
#define UBX_CFG_MSG 0x01
#define UBX_CFG_RATE 0x08

typedef struct {
    uint8_t class;
    uint8_t id;
    uint16_t length;
} UBX_Header;

typedef struct {
    uint32_t iTOW;
    int32_t lon;
    int32_t lat;
    int32_t height;
    int32_t hMSL;
    uint32_t hAcc;
    uint32_t vAcc;
} UBX_NAV_POSLLH_t;

typedef struct {
    uint32_t iTOW;
    int32_t velN;
    int32_t velE;
    int32_t velD;
    uint32_t speed;
    uint32_t gSpeed;
    int32_t heading;
    uint32_t sAcc;
    uint32_t cAcc;
} UBX_NAV_VELNED_t;

typedef struct {
    uint32_t iTOW;
    uint8_t gpsFix;
    uint8_t flags;
    uint8_t fixStat;
    uint8_t flags2;
    uint32_t ttff;
    uint32_t msss;
} UBX_NAV_STATUS_t;

typedef struct {
    GPSPosition position;
    GPSVelocity velocity;
    uint8_t fix_type;
    uint8_t satellites;
    float hdop;
    float vdop;
    float ground_speed;
    float heading;
    uint32_t timestamp;
    bool initialized;
    uint32_t last_update;
} GPS_Data;

bool gps_ublox_init(void);
void gps_ublox_process_byte(uint8_t byte);
void gps_ublox_get_position(GPSPosition *pos);
void gps_ublox_get_velocity(GPSVelocity *vel);
uint8_t gps_ublox_get_fix_type(void);
uint8_t gps_ublox_get_satellites(void);
float gps_ublox_get_ground_speed(void);
float gps_ublox_get_heading(void);
bool gps_ublox_is_healthy(void);
void gps_ublox_get_data(GPS_Data *data);

#endif
