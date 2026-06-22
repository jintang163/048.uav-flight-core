#include "sensor_manager.h"
#include "mpu6050.h"
#include "gps_ublox.h"
#include "magnetometer.h"
#include "barometer.h"
#include "sbus_rc.h"
#include "mmwave_radar.h"
#include "stereo_vision.h"

static SensorData sensor_data;

void sensor_manager_init(void)
{
    memset(&sensor_data, 0, sizeof(SensorData));

    mpu6050_init();
    gps_ublox_init();
    magnetometer_init();
    barometer_init();
    sbus_rc_init();
    mmwave_init();
    stereo_vision_init();

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

void sensor_manager_get_position(PositionState *pos)
{
    if (pos != NULL) {
        GPSPosition gps_pos;
        GPSVelocity gps_vel;
        sensor_manager_get_gps(&gps_pos, &gps_vel);

        pos->position = gps_pos;
        pos->velocity = gps_vel;
        pos->altitude = (float)gps_pos.alt / 1000.0f;
        pos->ground_speed = sqrtf(gps_vel.vn * gps_vel.vn + gps_vel.ve * gps_vel.ve);
        pos->heading = atan2f(gps_vel.ve, gps_vel.vn);

        float baro_alt;
        sensor_manager_get_baro(&baro_alt, NULL, NULL);
        if (baro_alt > 0) {
            pos->altitude = baro_alt;
        }
    }
}

float sensor_manager_get_mmwave_distance(uint8_t index, float *angle, float *size, float *confidence)
{
    mmwave_data_t data;
    mmwave_get_data(&data);

    if (index >= data.target_count || index >= MMWAVE_MAX_TARGETS) {
        if (angle) *angle = 0;
        if (size) *size = 0;
        if (confidence) *confidence = 0;
        return 0.0f;
    }

    if (angle) *angle = data.targets[index].angle;
    if (size) *size = data.targets[index].size;
    if (confidence) *confidence = data.targets[index].confidence;
    return data.targets[index].distance;
}

float sensor_manager_get_stereo_distance(uint8_t index, float *angle, float *size, float *confidence)
{
    stereo_data_t data;
    stereo_vision_get_data(&data);

    if (index >= data.obstacle_count || index >= STEREO_MAX_OBSTACLES) {
        if (angle) *angle = 0;
        if (size) *size = 0;
        if (confidence) *confidence = 0;
        return 0.0f;
    }

    if (angle) *angle = data.obstacles[index].angle;
    if (size) *size = data.obstacles[index].size;
    if (confidence) *confidence = data.obstacles[index].confidence;
    return data.obstacles[index].distance;
}
