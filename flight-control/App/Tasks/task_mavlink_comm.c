#include "task_mavlink_comm.h"
#include "mavlink_handler.h"
#include "flight_controller.h"
#include "task_attitude_estimation.h"
#include "sensor_manager.h"

static TaskHandle_t task_handle = NULL;

void task_mavlink_comm_init(void)
{
    mavlink_handler_init();

    xTaskCreate(task_mavlink_comm_main,
                "MAVLink",
                TASK_MAVLINK_COMM_STACK_SIZE,
                NULL,
                TASK_MAVLINK_COMM_PRIORITY,
                &task_handle);
}

void task_mavlink_comm_main(void *argument)
{
    UNUSED(argument);

    TickType_t last_wake_time = xTaskGetTickCount();
    const TickType_t period = pdMS_TO_TICKS(1000 / TASK_MAVLINK_COMM_FREQ);

    uint32_t heartbeat_counter = 0;
    uint32_t attitude_counter = 0;
    uint32_t position_counter = 0;
    uint32_t battery_counter = 0;
    uint32_t rc_counter = 0;
    uint32_t sys_status_counter = 0;

    while (1) {
        mavlink_handler_update();

        ControlCommand mavlink_cmd;
        if (mavlink_get_command(&mavlink_cmd)) {
            flight_controller_set_mavlink_command(&mavlink_cmd);
        }

        if (heartbeat_counter++ >= (TASK_MAVLINK_COMM_FREQ / 1)) {
            mavlink_send_heartbeat();
            heartbeat_counter = 0;
        }

        if (attitude_counter++ >= (TASK_MAVLINK_COMM_FREQ / 50)) {
            mavlink_send_attitude();
            attitude_counter = 0;
        }

        if (position_counter++ >= (TASK_MAVLINK_COMM_FREQ / 10)) {
            mavlink_send_global_position_int();
            mavlink_send_local_position_ned();
            mavlink_send_altitude();
            mavlink_send_vfr_hud();
            position_counter = 0;
        }

        if (battery_counter++ >= (TASK_MAVLINK_COMM_FREQ / 1)) {
            mavlink_send_battery_status();
            battery_counter = 0;
        }

        if (rc_counter++ >= (TASK_MAVLINK_COMM_FREQ / 10)) {
            mavlink_send_rc_channels_raw();
            rc_counter = 0;
        }

        if (sys_status_counter++ >= (TASK_MAVLINK_COMM_FREQ / 1)) {
            mavlink_send_sys_status();
            sys_status_counter = 0;
        }

        vTaskDelayUntil(&last_wake_time, period);
    }
}
