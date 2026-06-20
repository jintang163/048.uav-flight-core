#include "barometer.h"
#include "main.h"

extern I2C_HandleTypeDef hi2c1;

static Barometer_Data baro_data;

static bool baro_write_reg(uint8_t reg, uint8_t data)
{
    return HAL_I2C_Mem_Write(&hi2c1, BARO_I2C_ADDR << 1, reg, 1, &data, 1, 100) == HAL_OK;
}

static bool baro_read_reg(uint8_t reg, uint8_t *data)
{
    return HAL_I2C_Mem_Read(&hi2c1, BARO_I2C_ADDR << 1, reg, 1, data, 1, 100) == HAL_OK;
}

static bool baro_read_regs(uint8_t reg, uint8_t *data, uint16_t len)
{
    return HAL_I2C_Mem_Read(&hi2c1, BARO_I2C_ADDR << 1, reg, 1, data, len, 100) == HAL_OK;
}

static int16_t baro_read_int16(uint8_t reg)
{
    uint8_t buf[2];
    baro_read_regs(reg, buf, 2);
    return (int16_t)((buf[0] << 8) | buf[1]);
}

static uint16_t baro_read_uint16(uint8_t reg)
{
    uint8_t buf[2];
    baro_read_regs(reg, buf, 2);
    return (uint16_t)((buf[0] << 8) | buf[1]);
}

static bool baro_read_calibration(void)
{
    baro_data.cal.AC1 = baro_read_int16(BARO_REG_AC1);
    baro_data.cal.AC2 = baro_read_int16(BARO_REG_AC2);
    baro_data.cal.AC3 = baro_read_int16(BARO_REG_AC3);
    baro_data.cal.AC4 = baro_read_uint16(BARO_REG_AC4);
    baro_data.cal.AC5 = baro_read_uint16(BARO_REG_AC5);
    baro_data.cal.AC6 = baro_read_uint16(BARO_REG_AC6);
    baro_data.cal.B1 = baro_read_int16(BARO_REG_B1);
    baro_data.cal.B2 = baro_read_int16(BARO_REG_B2);
    baro_data.cal.MB = baro_read_int16(BARO_REG_MB);
    baro_data.cal.MC = baro_read_int16(BARO_REG_MC);
    baro_data.cal.MD = baro_read_int16(BARO_REG_MD);
    baro_data.calibrated = true;
    return true;
}

static int32_t baro_read_raw_temp(void)
{
    uint8_t buf[2];
    baro_write_reg(BARO_REG_CONTROL, BARO_CMD_TEMP);
    HAL_Delay(5);
    baro_read_regs(BARO_REG_DATA, buf, 2);
    return (int32_t)((buf[0] << 8) | buf[1]);
}

static int32_t baro_read_raw_pressure(void)
{
    uint8_t buf[3];
    baro_write_reg(BARO_REG_CONTROL, BARO_CMD_PRES);
    HAL_Delay(5);
    baro_read_regs(BARO_REG_DATA, buf, 3);
    return (int32_t)((buf[0] << 16) | (buf[1] << 8) | buf[2]);
}

static float baro_calculate_temperature(int32_t raw_temp)
{
    long X1 = (raw_temp - baro_data.cal.AC6) * baro_data.cal.AC5 / 32768;
    long X2 = baro_data.cal.MC * 2048 / (X1 + baro_data.cal.MD);
    baro_data.cal.B5 = X1 + X2;
    float temp = (baro_data.cal.B5 + 8) / 16.0f / 10.0f;
    return temp;
}

static float baro_calculate_pressure(int32_t raw_pres)
{
    long B6 = baro_data.cal.B5 - 4000;
    long X1 = (baro_data.cal.B2 * (B6 * B6 / 4096)) / 2048;
    long X2 = baro_data.cal.AC2 * B6 / 2048;
    long X3 = X1 + X2;
    long B3 = (((long)baro_data.cal.AC1 * 4 + X3) + 2) / 4;

    X1 = baro_data.cal.AC3 * B6 / 8192;
    X2 = (baro_data.cal.B1 * (B6 * B6 / 4096)) / 65536;
    X3 = ((X1 + X2) + 2) / 4;
    unsigned long B4 = baro_data.cal.AC4 * (unsigned long)(X3 + 32768) / 32768;
    unsigned long B7 = ((unsigned long)raw_pres - B3) * 50000;

    long p;
    if (B7 < 0x80000000) {
        p = (B7 * 2) / B4;
    } else {
        p = (B7 / B4) * 2;
    }

    X1 = (p / 256) * (p / 256);
    X1 = (X1 * 3038) / 65536;
    X2 = (-7357 * p) / 65536;
    p = p + (X1 + X2 + 3791) / 16;

    return (float)p / 100.0f;
}

static float baro_calculate_altitude(float pressure)
{
    float sea_level_pressure = 1013.25f;
    if (baro_data.base_pressure > 0) {
        sea_level_pressure = baro_data.base_pressure;
    }
    return 44330.0f * (1.0f - powf(pressure / sea_level_pressure, 0.1903f));
}

bool barometer_init(void)
{
    if (!baro_read_calibration()) {
        return false;
    }

    baro_data.base_pressure = 0.0f;
    baro_data.altitude_filtered = 0.0f;
    baro_data.initialized = true;
    baro_data.last_update = HAL_GetTick();

    barometer_read();
    baro_set_base_pressure();

    return true;
}

void barometer_read(void)
{
    if (!baro_data.initialized || !baro_data.calibrated) {
        return;
    }

    baro_data.temperature_raw = baro_read_raw_temp();
    baro_data.temperature = baro_calculate_temperature(baro_data.temperature_raw);

    baro_data.pressure_raw = baro_read_raw_pressure();
    baro_data.pressure = baro_calculate_pressure(baro_data.pressure_raw);

    baro_data.altitude = baro_calculate_altitude(baro_data.pressure);

    baro_data.altitude_filtered = 0.95f * baro_data.altitude_filtered + 0.05f * baro_data.altitude;

    baro_data.last_update = HAL_GetTick();
}

float barometer_get_altitude(void)
{
    return baro_data.altitude_filtered;
}

float barometer_get_pressure(void)
{
    return baro_data.pressure;
}

float barometer_get_temperature(void)
{
    return baro_data.temperature;
}

void barometer_set_base_pressure(void)
{
    baro_data.base_pressure = baro_data.pressure;
}

bool barometer_is_healthy(void)
{
    if (!baro_data.initialized) {
        return false;
    }

    uint32_t now = HAL_GetTick();
    if (now - baro_data.last_update > 500) {
        return false;
    }

    return true;
}
