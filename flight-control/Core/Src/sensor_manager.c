#include "sensor_manager.h"
#include "mpu6050.h"
#include "gps_ublox.h"
#include "magnetometer.h"
#include "barometer.h"
#include "sbus_rc.h"

static SensorData sensor_data;

void sensor_manager_init(void)
{
    memset(&sensor_data, 0, sizeof(SensorData));

    mpu6050_init();
    gps_ublox_init();
    magnetometer_init();
    barometer_init();
    sbus_rc_init();

    sensor_data.imu_timestamp = HAL_GetTick();
    sensor_data.gps_timestamp = HAL_GetTick();
    sensor_data.mag_timestamp = HAL_GetTick();
    sensor_data.baro_timestamp = HAL_GetTick();
    sensor_data.battery_timestamp = HAL_GetTick();
    sensor_data.rc_timestamp = HAL_GetTick();
}

void sensor_manager_read_imu(void)
{
    mpu6050_read();

    mpu6050_get_accel(&sensor_data.imu.accel);
    mpu6050_get_gyro(&sensor_data.imu.gyro);
    sensor_data.imu.timestamp = HAL_GetTick();

    sensor_data.imu_timestamp = HAL_GetTick();
}

void sensor_manager_read_gps(void)
{
    sensor_data.gps_timestamp = HAL_GetTick();
}

void sensor_manager_read_mag(void)
{
    magnetometer_read();
    magnetometer_get_data(&sensor_data.mag);
    sensor_data.mag_timestamp = HAL_GetTick();
}

void sensor_manager_read_baro(void)
{
    barometer_read();
    sensor_data.baro_alt = barometer_get_altitude();
    sensor_data.baro_temp = barometer_get_temperature();
    sensor_data.baro_press = barometer_get_pressure();
    sensor_data.baro_timestamp = HAL_GetTick();
}

void sensor_manager_read_battery(void)
{
    sensor_data.battery.voltage = 16.8f;
    sensor_data.battery.current = 5.0f;
    sensor_data.battery.capacity_used = 500.0f;
    sensor_data.battery.battery_percent =
        ((sensor_data.battery.voltage / BATTERY_CELL_COUNT - BATTERY_CELL_MIN_VOLTAGE) /
         (BATTERY_CELL_MAX_VOLTAGE - BATTERY_CELL_MIN_VOLTAGE)) * 100.0f;
    sensor_data.battery_percent = CONSTRAIN(sensor_data.battery.battery_percent, 0.0f, 100.0f);
    sensor_data.battery_timestamp = HAL_GetTick();
}

void sensor_manager_read_rc(void)
{
    uint16_t channels[16];
    sbus_rc_get_channels(channels);

    for (int i = 0; i < 16; i++) {
        sensor_data.rc.channels[i] = channels[i];
    }
    sensor_data.rc.connected = sbus_rc_is_connected();
    sensor_data.rc.last_update = HAL_GetTick();
    sensor_data.rc_timestamp = HAL_GetTick();
}

void sensor_manager_get_imu(IMUData *imu)
{
    *imu = sensor_data.imu;
}

void sensor_manager_get_gps(GPSPosition *pos, GPSVelocity *vel)
{
    gps_ublox_get_position(pos);
    gps_ublox_get_velocity(vel);
}

void sensor_manager_get_mag(Vector3f *mag)
{
    *mag = sensor_data.mag;
}

void sensor_manager_get_baro(float *alt, float *temp, float *press)
{
    if (alt) *alt = sensor_data.baro_alt;
    if (temp) *temp = sensor_data.baro_temp;
    if (press) *press = sensor_data.baro_press;
}

void sensor_manager_get_battery(BatteryState *battery)
{
    *battery = sensor_data.battery;
}

void sensor_manager_get_rc(RCInput *rc)
{
    *rc = sensor_data.rc;
}

bool sensor_manager_check_imu_timeout(void)
{
    return (HAL_GetTick() - sensor_data.imu_timestamp) > 100;
}

bool sensor_manager_check_gps_timeout(void)
{
    return (HAL_GetTick() - sensor_data.gps_timestamp) > 5000;
}

bool sensor_manager_check_mag_timeout(void)
{
    return (HAL_GetTick() - sensor_data.mag_timestamp) > 200;
}

bool sensor_manager_check_baro_timeout(void)
{
    return (HAL_GetTick() - sensor_data.baro_timestamp) > 500;
}

bool sensor_manager_check_rc_timeout(void)
{
    return (HAL_GetTick() - sensor_data.rc_timestamp) > RC_SIGNAL_LOSS_TIMEOUT;
}

void sensor_manager_calibrate_gyro(void)
{
    mpu6050_calibrate_gyro();
}

void sensor_manager_calibrate_accel(void)
{
    mpu6050_calibrate_accel();
}

void sensor_manager_calibrate_mag(void)
{
    magnetometer_calibrate();
}
