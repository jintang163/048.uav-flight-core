#ifndef __BLACKBOX_LOGGER_H__
#define __BLACKBOX_LOGGER_H__

#include "types.h"
#include "flight_config.h"

#define BLACKBOX_FLASH_SIZE       (16 * 1024 * 1024)
#define BLACKBOX_SECTOR_SIZE      4096
#define BLACKBOX_HEADER_MAGIC     0x424C4B58
#define BLACKBOX_VERSION          1
#define BLACKBOX_SAMPLE_RATE      10

#define BLACKBOX_LOG_TYPE_DATA    0x01
#define BLACKBOX_LOG_TYPE_EVENT   0x02

#define BLACKBOX_EVENT_LOW_BATTERY      (1 << 0)
#define BLACKBOX_EVENT_GPS_LOSS         (1 << 1)
#define BLACKBOX_EVENT_RC_LOSS          (1 << 2)
#define BLACKBOX_EVENT_VOLTAGE_DIP      (1 << 3)
#define BLACKBOX_EVENT_MOTOR_FAILURE    (1 << 4)
#define BLACKBOX_EVENT_CRASH            (1 << 5)
#define BLACKBOX_EVENT_FENCE_BREACH     (1 << 6)
#define BLACKBOX_EVENT_ARM              (1 << 7)
#define BLACKBOX_EVENT_DISARM           (1 << 8)
#define BLACKBOX_EVENT_TAKEOFF          (1 << 9)
#define BLACKBOX_EVENT_LAND             (1 << 10)
#define BLACKBOX_EVENT_MODE_CHANGE      (1 << 11)
#define BLACKBOX_EVENT_FAILSAFE         (1 << 12)
#define BLACKBOX_EVENT_OBSTACLE_DETECTED (1 << 13)
#define BLACKBOX_EVENT_OBSTACLE_CLEARED  (1 << 14)

typedef struct __attribute__((packed)) {
    uint32_t magic;
    uint8_t  version;
    uint32_t flight_id;
    uint32_t start_time;
    uint32_t end_time;
    uint32_t total_entries;
    uint32_t data_size;
    uint16_t sample_rate;
    uint8_t  reserved[10];
} BlackboxHeader;

typedef struct __attribute__((packed)) {
    uint32_t timestamp;
    int32_t  lat;
    int32_t  lon;
    int32_t  alt;
    float    roll;
    float    pitch;
    float    yaw;
    float    vx;
    float    vy;
    float    vz;
    float    voltage;
    float    current;
    float    throttle;
    uint16_t rc_channels[8];
    uint16_t motor_pwm[4];
    uint8_t  flight_mode;
    uint8_t  satellites;
    uint8_t  gps_fix_type;
    uint8_t  error_flags;
    uint8_t  reserved[3];
} BlackboxDataEntry;

typedef struct __attribute__((packed)) {
    uint32_t timestamp;
    uint32_t event_type;
    int32_t  param1;
    int32_t  param2;
    float    param3;
    float    param4;
    uint8_t  event_severity;
    char     description[64];
} BlackboxEventEntry;

typedef struct __attribute__((packed)) {
    uint8_t type;
    uint16_t size;
    union {
        BlackboxDataEntry data;
        BlackboxEventEntry event;
    } payload;
} BlackboxLogEntry;

typedef struct {
    uint32_t write_offset;
    uint32_t read_offset;
    uint32_t entry_count;
    uint32_t current_flight_id;
    bool     recording;
    bool     initialized;
    BlackboxHeader header;
    float    last_voltage;
    uint32_t last_gps_time;
    uint32_t last_rc_time;
} BlackboxState;

typedef struct {
    uint32_t total_entries;
    uint32_t total_bytes;
    uint32_t start_time;
    uint32_t end_time;
    uint32_t flight_id;
    bool     is_recording;
} BlackboxInfo;

void blackbox_init(void);
void blackbox_start(void);
void blackbox_stop(void);
bool blackbox_is_recording(void);
void blackbox_log_data(void);
void blackbox_log_event(uint32_t event_type, int32_t p1, int32_t p2, float p3, float p4, const char *desc);
uint32_t blackbox_get_entry_count(void);
uint32_t blackbox_get_flight_id(void);
bool blackbox_read_entry(uint32_t index, BlackboxLogEntry *entry);
bool blackbox_read_range(uint32_t start_index, uint32_t count, uint8_t *buffer, uint32_t *bytes_read);
uint32_t blackbox_get_total_bytes(void);
void blackbox_check_anomalies(void);
void blackbox_reset(void);
void blackbox_get_info(BlackboxInfo *info);
bool blackbox_read_data(uint32_t offset, uint8_t *buffer, uint32_t length);

#endif
