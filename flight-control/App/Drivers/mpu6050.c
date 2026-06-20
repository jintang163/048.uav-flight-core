#include "mpu6050.h"
#include "main.h"

extern I2C_HandleTypeDef hi2c1;

static MPU6050_Data mpu6050_data;

static bool mpu6050_write_reg(uint8_t reg, uint8_t data)
{
    return HAL_I2C_Mem_Write(&hi2c1, MPU6050_I2C_ADDR << 1, reg, 1, &data, 1, 100) == HAL_OK;
}

static bool mpu6050_read_reg(uint8_t reg, uint8_t *data)
{
    return HAL_I2C_Mem_Read(&hi2c1, MPU6050_I2C_ADDR << 1, reg, 1, data, 1, 100) == HAL_OK;
}

static bool mpu6050_read_regs(uint8_t reg, uint8_t *data, uint16_t len)
{
    return HAL_I2C_Mem_Read(&hi2c1, MPU6050_I2C_ADDR << 1, reg, 1, data, len, 100) == HAL_OK;
}

bool mpu6050_init(void)
{
    uint8_t who_am_i = 0;

    if (!mpu6050_read_reg(MPU6050_REG_WHO_AM_I, &who_am_i)) {
        return false;
    }

    if (who_am_i != MPU6050_WHO_AM_I_VALUE) {
        return false;
    }

    if (!mpu6050_write_reg(MPU6050_REG_PWR_MGMT_1, 0x80)) {
        return false;
    }
    HAL_Delay(100);

    if (!mpu6050_write_reg(MPU6050_REG_PWR_MGMT_1, 0x01)) {
        return false;
    }

    if (!mpu6050_write_reg(MPU6050_REG_SMPLRT_DIV, 0x00)) {
        return false;
    }

    if (!mpu6050_write_reg(MPU6050_REG_CONFIG, 0x03)) {
        return false;
    }

    if (!mpu6050_write_reg(MPU6050_REG_GYRO_CONFIG, 0x18)) {
        return false;
    }

    if (!mpu6050_write_reg(MPU6050_REG_ACCEL_CONFIG, 0x18)) {
        return false;
    }

    mpu6050_data.gyro_bias.x = GYRO_BIAS_X;
    mpu6050_data.gyro_bias.y = GYRO_BIAS_Y;
    mpu6050_data.gyro_bias.z = GYRO_BIAS_Z;

    mpu6050_data.accel_bias.x = ACCEL_BIAS_X;
    mpu6050_data.accel_bias.y = ACCEL_BIAS_Y;
    mpu6050_data.accel_bias.z = ACCEL_BIAS_Z;

    mpu6050_data.initialized = true;
    mpu6050_data.last_update = HAL_GetTick();

    return true;
}

void mpu6050_read(void)
{
    if (!mpu6050_data.initialized) {
        return;
    }

    uint8_t raw_data[14];

    if (!mpu6050_read_regs(MPU6050_REG_ACCEL_XOUT_H, raw_data, 14)) {
        return;
    }

    mpu6050_data.accel_x_raw = (int16_t)((raw_data[0] << 8) | raw_data[1]);
    mpu6050_data.accel_y_raw = (int16_t)((raw_data[2] << 8) | raw_data[3]);
    mpu6050_data.accel_z_raw = (int16_t)((raw_data[4] << 8) | raw_data[5]);
    mpu6050_data.temp_raw = (int16_t)((raw_data[6] << 8) | raw_data[7]);
    mpu6050_data.gyro_x_raw = (int16_t)((raw_data[8] << 8) | raw_data[9]);
    mpu6050_data.gyro_y_raw = (int16_t)((raw_data[10] << 8) | raw_data[11]);
    mpu6050_data.gyro_z_raw = (int16_t)((raw_data[12] << 8) | raw_data[13]);

    mpu6050_data.accel_x = (float)mpu6050_data.accel_x_raw / IMU_ACCEL_LSB - mpu6050_data.accel_bias.x;
    mpu6050_data.accel_y = (float)mpu6050_data.accel_y_raw / IMU_ACCEL_LSB - mpu6050_data.accel_bias.y;
    mpu6050_data.accel_z = (float)mpu6050_data.accel_z_raw / IMU_ACCEL_LSB - mpu6050_data.accel_bias.z;

    mpu6050_data.gyro_x = ((float)mpu6050_data.gyro_x_raw / IMU_GYRO_LSB - mpu6050_data.gyro_bias.x) * (float)M_PI / 180.0f;
    mpu6050_data.gyro_y = ((float)mpu6050_data.gyro_y_raw / IMU_GYRO_LSB - mpu6050_data.gyro_bias.y) * (float)M_PI / 180.0f;
    mpu6050_data.gyro_z = ((float)mpu6050_data.gyro_z_raw / IMU_GYRO_LSB - mpu6050_data.gyro_bias.z) * (float)M_PI / 180.0f;

    mpu6050_data.temperature = (float)mpu6050_data.temp_raw / 340.0f + 36.53f;

    mpu6050_data.last_update = HAL_GetTick();
}

void mpu6050_get_accel(Vector3f *accel)
{
    accel->x = mpu6050_data.accel_x;
    accel->y = mpu6050_data.accel_y;
    accel->z = mpu6050_data.accel_z;
}

void mpu6050_get_gyro(Vector3f *gyro)
{
    gyro->x = mpu6050_data.gyro_x;
    gyro->y = mpu6050_data.gyro_y;
    gyro->z = mpu6050_data.gyro_z;
}

float mpu6050_get_temperature(void)
{
    return mpu6050_data.temperature;
}

void mpu6050_calibrate_gyro(void)
{
    const int samples = 1000;
    Vector3f sum = {0, 0, 0};

    for (int i = 0; i < samples; i++) {
        mpu6050_read();
        sum.x += (float)mpu6050_data.gyro_x_raw / IMU_GYRO_LSB;
        sum.y += (float)mpu6050_data.gyro_y_raw / IMU_GYRO_LSB;
        sum.z += (float)mpu6050_data.gyro_z_raw / IMU_GYRO_LSB;
        HAL_Delay(1);
    }

    mpu6050_data.gyro_bias.x = sum.x / samples;
    mpu6050_data.gyro_bias.y = sum.y / samples;
    mpu6050_data.gyro_bias.z = sum.z / samples;
}

void mpu6050_calibrate_accel(void)
{
    const int samples = 1000;
    Vector3f sum = {0, 0, 0};

    for (int i = 0; i < samples; i++) {
        mpu6050_read();
        sum.x += (float)mpu6050_data.accel_x_raw / IMU_ACCEL_LSB;
        sum.y += (float)mpu6050_data.accel_y_raw / IMU_ACCEL_LSB;
        sum.z += (float)mpu6050_data.accel_z_raw / IMU_ACCEL_LSB;
        HAL_Delay(1);
    }

    mpu6050_data.accel_bias.x = sum.x / samples;
    mpu6050_data.accel_bias.y = sum.y / samples;
    mpu6050_data.accel_bias.z = sum.z / samples - 1.0f;
}

bool mpu6050_is_healthy(void)
{
    if (!mpu6050_data.initialized) {
        return false;
    }

    uint32_t now = HAL_GetTick();
    if (now - mpu6050_data.last_update > 100) {
        return false;
    }

    return true;
}

void mpu6050_get_raw_data(MPU6050_Data *data)
{
    *data = mpu6050_data;
}
