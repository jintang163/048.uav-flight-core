#ifndef __SENSOR_MANAGER_H__
#define __SENSOR_MANAGER_H__

#include "types.h"
#include "flight_config.h"
#include "attitude_estimator.h"

typedef struct {
    IMUData imu;
    GPSPosition gps_pos;
    GPSVelocity gps_vel;
    Vector3f mag;
    float baro_alt;
    float baro_temp;
    float baro_press;
    BatteryState battery;
    RCInput rc;
    uint32_t imu_timestamp;
    uint32_t gps_timestamp;
    uint32_t mag_timestamp;
    uint32_t baro_timestamp;
    uint32_t battery_timestamp;
    uint32_t rc_timestamp;
} SensorData;

void sensor_manager_init(void);
void sensor_manager_read_imu(void);
void sensor_manager_read_gps(void);
void sensor_manager_read_mag(void);
void sensor_manager_read_baro(void);
void sensor_manager_read_battery(void);
void sensor_manager_read_rc(void);
void sensor_manager_get_imu(IMUData *imu);
void sensor_manager_get_gps(GPSPosition *pos, GPSVelocity *vel);
void sensor_manager_get_mag(Vector3f *mag);
void sensor_manager_get_baro(float *alt, float *temp, float *press);
void sensor_manager_get_battery(BatteryState *battery);
void sensor_manager_get_rc(RCInput *rc);
bool sensor_manager_check_imu_timeout(void);
bool sensor_manager_check_gps_timeout(void);
bool sensor_manager_check_mag_timeout(void);
bool sensor_manager_check_baro_timeout(void);
bool sensor_manager_check_rc_timeout(void);
void sensor_manager_calibrate_gyro(void);
void sensor_manager_calibrate_accel(void);
void sensor_manager_calibrate_mag(void);

#endif
