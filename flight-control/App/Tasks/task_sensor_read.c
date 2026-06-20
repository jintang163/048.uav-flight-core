#include "task_sensor_read.h"
#include "sensor_manager.h"
#include "mpu6050.h"
#include "gps_ublox.h"
#include "magnetometer.h"
#include "barometer.h"
#include "sbus_rc.h"

static TaskHandle_t task_handle = NULL;

void task_sensor_read_init(void)
{
    sensor_manager_init();

    xTaskCreate(task_sensor_read_main,
                "SensorRead",
                TASK_SENSOR_READ_STACK_SIZE,
                NULL,
                TASK_SENSOR_READ_PRIORITY,
                &task_handle);
}

void task_sensor_read_main(void *argument)
{
    UNUSED(argument);

    TickType_t last_wake_time = xTaskGetTickCount();
    const TickType_t period = pdMS_TO_TICKS(1000 / TASK_SENSOR_READ_FREQ);

    uint32_t gps_counter = 0;
    uint32_t mag_counter = 0;
    uint32_t baro_counter = 0;
    uint32_t battery_counter = 0;

    while (1) {
        sensor_manager_read_imu();

        if (gps_counter++ >= (TASK_SENSOR_READ_FREQ / 10)) {
            sensor_manager_read_gps();
            gps_counter = 0;
        }

        if (mag_counter++ >= (TASK_SENSOR_READ_FREQ / 100)) {
            sensor_manager_read_mag();
            mag_counter = 0;
        }

        if (baro_counter++ >= (TASK_SENSOR_READ_FREQ / 50)) {
            sensor_manager_read_baro();
            baro_counter = 0;
        }

        if (battery_counter++ >= (TASK_SENSOR_READ_FREQ / 10)) {
            sensor_manager_read_battery();
            battery_counter = 0;
        }

        sensor_manager_read_rc();

        vTaskDelayUntil(&last_wake_time, period);
    }
}
