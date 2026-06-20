#ifndef __BAROMETER_H__
#define __BAROMETER_H__

#include "types.h"
#include "flight_config.h"

#define BARO_REG_AC1 0xAA
#define BARO_REG_AC2 0xAC
#define BARO_REG_AC3 0xAE
#define BARO_REG_AC4 0xB0
#define BARO_REG_AC5 0xB2
#define BARO_REG_AC6 0xB4
#define BARO_REG_B1  0xB6
#define BARO_REG_B2  0xB8
#define BARO_REG_MB  0xBA
#define BARO_REG_MC  0xBC
#define BARO_REG_MD  0xBE

#define BARO_REG_CONTROL 0xF4
#define BARO_REG_DATA    0xF6

#define BARO_CMD_TEMP 0x2E
#define BARO_CMD_PRES 0x34

typedef struct {
    int16_t AC1, AC2, AC3;
    uint16_t AC4, AC5, AC6;
    int16_t B1, B2;
    int16_t MB, MC, MD;
    long B5;
} Baro_Calibration;

typedef struct {
    int32_t pressure_raw;
    int32_t temperature_raw;
    float pressure;
    float temperature;
    float altitude;
    float altitude_filtered;
    float base_pressure;
    Baro_Calibration cal;
    bool calibrated;
    bool initialized;
    uint32_t last_update;
} Barometer_Data;

bool barometer_init(void);
void barometer_read(void);
float barometer_get_altitude(void);
float barometer_get_pressure(void);
float barometer_get_temperature(void);
void barometer_set_base_pressure(void);
bool barometer_is_healthy(void);

#endif
