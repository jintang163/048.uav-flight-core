#include "task_blackbox.h"
#include "blackbox_logger.h"
#include "flight_controller.h"
#include "sensor_manager.h"

static TaskHandle_t task_handle = NULL;
static uint32_t log_counter = 0;

void task_blackbox_init(void)
{
    blackbox_init();

    xTaskCreate(task_blackbox_main,
                "Blackbox",
                TASK_BLACKBOX_STACK_SIZE,
                NULL,
                TASK_BLACKBOX_PRIORITY,
                &task_handle);
}

void task_blackbox_main(void *argument)
{
    UNUSED(argument);

    TickType_t last_wake_time = xTaskGetTickCount();
    const TickType_t period = pdMS_TO_TICKS(1000 / TASK_BLACKBOX_FREQ);

    bool was_armed = false;

    while (1) {
        bool is_armed = flight_controller_is_armed();

        if (is_armed && !was_armed) {
            blackbox_start();
        } else if (!is_armed && was_armed) {
            blackbox_stop();
        }
        was_armed = is_armed;

        if (blackbox_is_recording()) {
            blackbox_log_data();
            log_counter++;

            if (log_counter % TASK_BLACKBOX_FREQ == 0) {
                blackbox_check_anomalies();
            }
        }

        vTaskDelayUntil(&last_wake_time, period);
    }
}
