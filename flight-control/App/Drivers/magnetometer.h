#ifndef __MAGNETOMETER_H__
#define __MAGNETOMETER_H__

#include "types.h"
#include "flight_config.h"

#define MAG_REG_CONFIG_A 0x00
#define MAG_REG_CONFIG_B 0x01
#define MAG_REG_MODE     0x02
#define MAG_REG_DATA_X_H 0x03
#define MAG_REG_DATA_X_L 0x04
#define MAG_REG_DATA_Z_H 0x05
#define MAG_REG_DATA_Z_L 0x06
#define MAG_REG_DATA_Y_H 0x07
#define MAG_REG_DATA_Y_L 0x08
#define MAG_REG_STATUS   0x09

typedef struct {
    int16_t x_raw;
    int16_t y_raw;
    int16_t z_raw;
    float x;
    float y;
    float z;
    Vector3f bias;
    Vector3f scale;
    float heading;
    bool initialized;
    uint32_t last_update;
} Magnetometer_Data;

bool magnetometer_init(void);
void magnetometer_read(void);
void magnetometer_get_data(Vector3f *mag);
float magnetometer_get_heading(void);
void magnetometer_calibrate(void);
bool magnetometer_is_healthy(void);

#endif
