#include "task_motor_control.h"
#include "motor_control.h"
#include "pwm_output.h"

static TaskHandle_t task_handle = NULL;

void task_motor_control_init(void)
{
    motor_control_init();
    pwm_output_init();

    xTaskCreate(task_motor_control_main,
                "MotorCtrl",
                TASK_MOTOR_CONTROL_STACK_SIZE,
                NULL,
                TASK_MOTOR_CONTROL_PRIORITY,
                &task_handle);
}

void task_motor_control_main(void *argument)
{
    UNUSED(argument);

    TickType_t last_wake_time = xTaskGetTickCount();
    const TickType_t period = pdMS_TO_TICKS(1000 / TASK_MOTOR_CONTROL_FREQ);

    while (1) {
        motor_control_update();
        pwm_output_update();

        vTaskDelayUntil(&last_wake_time, period);
    }
}
