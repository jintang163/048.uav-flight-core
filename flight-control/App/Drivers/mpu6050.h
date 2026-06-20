#ifndef __MPU6050_H__
#define __MPU6050_H__

#include "types.h"
#include "flight_config.h"

#define MPU6050_REG_SMPLRT_DIV     0x19
#define MPU6050_REG_CONFIG         0x1A
#define MPU6050_REG_GYRO_CONFIG    0x1B
#define MPU6050_REG_ACCEL_CONFIG   0x1C
#define MPU6050_REG_ACCEL_XOUT_H   0x3B
#define MPU6050_REG_ACCEL_XOUT_L   0x3C
#define MPU6050_REG_ACCEL_YOUT_H   0x3D
#define MPU6050_REG_ACCEL_YOUT_L   0x3E
#define MPU6050_REG_ACCEL_ZOUT_H   0x3F
#define MPU6050_REG_ACCEL_ZOUT_L   0x40
#define MPU6050_REG_TEMP_OUT_H     0x41
#define MPU6050_REG_TEMP_OUT_L     0x42
#define MPU6050_REG_GYRO_XOUT_H    0x43
#define MPU6050_REG_GYRO_XOUT_L    0x44
#define MPU6050_REG_GYRO_YOUT_H    0x45
#define MPU6050_REG_GYRO_YOUT_L    0x46
#define MPU6050_REG_GYRO_ZOUT_H    0x47
#define MPU6050_REG_GYRO_ZOUT_L    0x48
#define MPU6050_REG_PWR_MGMT_1     0x6B
#define MPU6050_REG_WHO_AM_I       0x75

#define MPU6050_WHO_AM_I_VALUE     0x68

typedef struct {
    int16_t accel_x_raw;
    int16_t accel_y_raw;
    int16_t accel_z_raw;
    int16_t temp_raw;
    int16_t gyro_x_raw;
    int16_t gyro_y_raw;
    int16_t gyro_z_raw;
    float accel_x;
    float accel_y;
    float accel_z;
    float gyro_x;
    float gyro_y;
    float gyro_z;
    float temperature;
    Vector3f accel_bias;
    Vector3f gyro_bias;
    bool initialized;
    uint32_t last_update;
} MPU6050_Data;

bool mpu6050_init(void);
void mpu6050_read(void);
void mpu6050_get_accel(Vector3f *accel);
void mpu6050_get_gyro(Vector3f *gyro);
float mpu6050_get_temperature(void);
void mpu6050_calibrate_gyro(void);
void mpu6050_calibrate_accel(void);
bool mpu6050_is_healthy(void);
void mpu6050_get_raw_data(MPU6050_Data *data);

#endif
