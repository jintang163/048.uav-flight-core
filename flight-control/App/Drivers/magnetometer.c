#include "magnetometer.h"
#include "main.h"

extern I2C_HandleTypeDef hi2c1;

static Magnetometer_Data mag_data;

static bool mag_write_reg(uint8_t reg, uint8_t data)
{
    return HAL_I2C_Mem_Write(&hi2c1, MAG_I2C_ADDR << 1, reg, 1, &data, 1, 100) == HAL_OK;
}

static bool mag_read_reg(uint8_t reg, uint8_t *data)
{
    return HAL_I2C_Mem_Read(&hi2c1, MAG_I2C_ADDR << 1, reg, 1, data, 1, 100) == HAL_OK;
}

static bool mag_read_regs(uint8_t reg, uint8_t *data, uint16_t len)
{
    return HAL_I2C_Mem_Read(&hi2c1, MAG_I2C_ADDR << 1, reg, 1, data, len, 100) == HAL_OK;
}

bool magnetometer_init(void)
{
    if (!mag_write_reg(MAG_REG_CONFIG_A, 0x78)) {
        return false;
    }
    HAL_Delay(10);

    if (!mag_write_reg(MAG_REG_CONFIG_B, 0x20)) {
        return false;
    }
    HAL_Delay(10);

    if (!mag_write_reg(MAG_REG_MODE, 0x00)) {
        return false;
    }
    HAL_Delay(10);

    mag_data.bias.x = 0.0f;
    mag_data.bias.y = 0.0f;
    mag_data.bias.z = 0.0f;
    mag_data.scale.x = 1.0f;
    mag_data.scale.y = 1.0f;
    mag_data.scale.z = 1.0f;

    mag_data.initialized = true;
    mag_data.last_update = HAL_GetTick();

    return true;
}

void magnetometer_read(void)
{
    if (!mag_data.initialized) {
        return;
    }

    uint8_t raw_data[6];

    if (!mag_read_regs(MAG_REG_DATA_X_H, raw_data, 6)) {
        return;
    }

    mag_data.x_raw = (int16_t)((raw_data[0] << 8) | raw_data[1]);
    mag_data.z_raw = (int16_t)((raw_data[2] << 8) | raw_data[3]);
    mag_data.y_raw = (int16_t)((raw_data[4] << 8) | raw_data[5]);

    mag_data.x = ((float)mag_data.x_raw - mag_data.bias.x) * mag_data.scale.x;
    mag_data.y = ((float)mag_data.y_raw - mag_data.bias.y) * mag_data.scale.y;
    mag_data.z = ((float)mag_data.z_raw - mag_data.bias.z) * mag_data.scale.z;

    float mag_norm = sqrtf(mag_data.x * mag_data.x + mag_data.y * mag_data.y + mag_data.z * mag_data.z);
    if (mag_norm > 0.0f) {
        mag_data.x /= mag_norm;
        mag_data.y /= mag_norm;
        mag_data.z /= mag_norm;
    }

    mag_data.heading = atan2f(mag_data.y, mag_data.x);

    mag_data.last_update = HAL_GetTick();
}

void magnetometer_get_data(Vector3f *mag)
{
    mag->x = mag_data.x;
    mag->y = mag_data.y;
    mag->z = mag_data.z;
}

float magnetometer_get_heading(void)
{
    return mag_data.heading;
}

void magnetometer_calibrate(void)
{
    const int samples = 1000;
    Vector3f max_val = {-10000, -10000, -10000};
    Vector3f min_val = {10000, 10000, 10000};

    for (int i = 0; i < samples; i++) {
        magnetometer_read();
        max_val.x = fmaxf(max_val.x, (float)mag_data.x_raw);
        max_val.y = fmaxf(max_val.y, (float)mag_data.y_raw);
        max_val.z = fmaxf(max_val.z, (float)mag_data.z_raw);
        min_val.x = fminf(min_val.x, (float)mag_data.x_raw);
        min_val.y = fminf(min_val.y, (float)mag_data.y_raw);
        min_val.z = fminf(min_val.z, (float)mag_data.z_raw);
        HAL_Delay(10);
    }

    mag_data.bias.x = (max_val.x + min_val.x) / 2.0f;
    mag_data.bias.y = (max_val.y + min_val.y) / 2.0f;
    mag_data.bias.z = (max_val.z + min_val.z) / 2.0f;

    Vector3f span = {
        (max_val.x - min_val.x) / 2.0f,
        (max_val.y - min_val.y) / 2.0f,
        (max_val.z - min_val.z) / 2.0f
    };

    float avg_span = (span.x + span.y + span.z) / 3.0f;

    if (span.x > 0 && span.y > 0 && span.z > 0) {
        mag_data.scale.x = avg_span / span.x;
        mag_data.scale.y = avg_span / span.y;
        mag_data.scale.z = avg_span / span.z;
    }
}

bool magnetometer_is_healthy(void)
{
    if (!mag_data.initialized) {
        return false;
    }

    uint32_t now = HAL_GetTick();
    if (now - mag_data.last_update > 200) {
        return false;
    }

    return true;
}
