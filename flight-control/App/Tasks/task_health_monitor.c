#include "task_health_monitor.h"
#include "sensor_manager.h"
#include "flight_controller.h"
#include "mpu6050.h"
#include "gps_ublox.h"
#include "magnetometer.h"
#include "barometer.h"
#include "sbus_rc.h"

static TaskHandle_t task_handle = NULL;
static HealthStatus health_status;

void task_health_monitor_init(void)
{
    memset(&health_status, 0, sizeof(HealthStatus));
    health_status.error_flags = 0;

    xTaskCreate(task_health_monitor_main,
                "HealthMon",
                TASK_HEALTH_MONITOR_STACK_SIZE,
                NULL,
                TASK_HEALTH_MONITOR_PRIORITY,
                &task_handle);
}

void task_health_monitor_main(void *argument)
{
    UNUSED(argument);

    TickType_t last_wake_time = xTaskGetTickCount();
    const TickType_t period = pdMS_TO_TICKS(1000 / TASK_HEALTH_MONITOR_FREQ);

    bool failsafe_triggered = false;
    uint32_t failsafe_start_time = 0;

    while (1) {
        health_status.imu_ok = mpu6050_is_healthy() && !sensor_manager_check_imu_timeout();
        health_status.gps_ok = gps_ublox_is_healthy() && !sensor_manager_check_gps_timeout();
        health_status.mag_ok = magnetometer_is_healthy() && !sensor_manager_check_mag_timeout();
        health_status.baro_ok = barometer_is_healthy() && !sensor_manager_check_baro_timeout();
        health_status.rc_ok = sbus_rc_is_connected() && !sensor_manager_check_rc_timeout();

        BatteryState battery;
        sensor_manager_get_battery(&battery);
        health_status.battery_ok = battery.voltage > LOW_BATTERY_LAND_TRIGGER_VOLTAGE;

        uint32_t new_error_flags = 0;
        if (!health_status.imu_ok) new_error_flags |= ERROR_FLAG_IMU_TIMEOUT;
        if (!health_status.gps_ok) new_error_flags |= ERROR_FLAG_GPS_TIMEOUT;
        if (!health_status.mag_ok) new_error_flags |= ERROR_FLAG_MAG_TIMEOUT;
        if (!health_status.baro_ok) new_error_flags |= ERROR_FLAG_BARO_TIMEOUT;
        if (!health_status.rc_ok) new_error_flags |= ERROR_FLAG_RC_SIGNAL_LOSS;
        if (!health_status.battery_ok) new_error_flags |= ERROR_FLAG_LOW_BATTERY;

        health_status.error_flags = new_error_flags;

        if (!health_status.rc_ok || battery.voltage < LOW_BATTERY_RTL_TRIGGER_VOLTAGE) {
            if (!failsafe_triggered) {
                failsafe_triggered = true;
                failsafe_start_time = xTaskGetTickCount();

                if (flight_controller_is_armed()) {
                    if (health_status.gps_ok && flight_controller_is_home_set()) {
                        flight_controller_set_mode(FLIGHT_MODE_RTL);
                        flight_controller_trigger_failsafe();
                        mavlink_send_statustext(2, "Failsafe: Returning to launch");
                    } else {
                        flight_controller_set_mode(FLIGHT_MODE_LAND);
                        flight_controller_trigger_failsafe();
                        mavlink_send_statustext(2, "Failsafe: Landing");
                    }
                }
            }
        } else {
            if (failsafe_triggered) {
                uint32_t elapsed = xTaskGetTickCount() - failsafe_start_time;
                if (elapsed > 5000 && health_status.rc_ok && battery.voltage > LOW_BATTERY_RTL_TRIGGER_VOLTAGE + 0.5f) {
                    failsafe_triggered = false;
                    flight_controller_clear_failsafe();
                    mavlink_send_statustext(4, "Failsafe cleared");
                }
            }
        }

        if (battery.voltage < LOW_BATTERY_LAND_TRIGGER_VOLTAGE && flight_controller_is_armed()) {
            flight_controller_set_mode(FLIGHT_MODE_LAND);
            mavlink_send_statustext(1, "Critical battery: Landing");
        }

        vTaskDelayUntil(&last_wake_time, period);
    }
}

void task_health_monitor_get_status(HealthStatus *status)
{
    *status = health_status;
}
