#include "task_tracking.h"
#include "tracking_manager.h"
#include "flight_controller.h"
#include "sensor_manager.h"
#include "flight_config.h"
#include "types.h"
#include <math.h>

static TaskHandle_t task_handle = NULL;

void task_tracking_init(void)
{
    tracking_manager_init();

    xTaskCreate(task_tracking_main,
                "Tracking",
                TASK_TRACKING_STACK_SIZE,
                NULL,
                TASK_TRACKING_PRIORITY,
                &task_handle);
}

void task_tracking_main(void *argument)
{
    UNUSED(argument);

    TickType_t last_wake_time = xTaskGetTickCount();
    const TickType_t period = pdMS_TO_TICKS(1000 / TASK_TRACKING_FREQ);
    float dt = 1.0f / TASK_TRACKING_FREQ;

    while (1) {
        GPSPosition gps_pos;
        GPSVelocity gps_vel;
        sensor_manager_get_gps(&gps_pos, &gps_vel);

        FlightMode mode = flight_controller_get_mode();

        if (mode == FLIGHT_MODE_TRACKING) {
            tracking_manager_update(dt);

            TrackingState state = tracking_manager_get_state();

            if (state == TRACKING_STATE_TRACKING ||
                state == TRACKING_STATE_SEARCHING ||
                state == TRACKING_STATE_LOCKING) {
                ControlCommand tracking_cmd;
                float target_vel_n, target_vel_e, target_vel_d, target_yaw_rate;

                tracking_manager_get_velocity_commands(&target_vel_n,
                                                        &target_vel_e,
                                                        &target_vel_d,
                                                        &target_yaw_rate);

                float vel_error_n = target_vel_n - gps_vel.vn;
                float vel_error_e = target_vel_e - gps_vel.ve;

                tracking_cmd.roll = CONSTRAIN(vel_error_e * 0.5f, -MAX_TILT_ANGLE, MAX_TILT_ANGLE);
                tracking_cmd.pitch = CONSTRAIN(-vel_error_n * 0.5f, -MAX_TILT_ANGLE, MAX_TILT_ANGLE);
                tracking_cmd.yaw = target_yaw_rate;

                float baro_alt;
                sensor_manager_get_baro(&baro_alt, NULL, NULL);

                float baro_vel = 0.0f;
                float alt_error = -target_vel_d * 0.5f;
                if (tracking_manager_is_searching()) {
                    alt_error = 0.0f;
                }
                tracking_cmd.throttle = CONSTRAIN(0.5f + alt_error, 0.3f, 0.8f);

                flight_controller_set_mavlink_command(&tracking_cmd);
            }
        } else {
            tracking_manager_reset();
        }

        vTaskDelayUntil(&last_wake_time, period);
    }
}
