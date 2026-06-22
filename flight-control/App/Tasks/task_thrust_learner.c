#include "task_thrust_learner.h"
#include "thrust_learner.h"
#include "flight_config.h"
#include <string.h>

static TaskHandle_t task_handle = NULL;

void task_thrust_learner_init(void)
{
    thrust_learner_init();

    xTaskCreate(task_thrust_learner_main,
                "ThrustLearn",
                THRUST_LEARNER_TASK_STACK_SIZE,
                NULL,
                THRUST_LEARNER_TASK_PRIORITY,
                &task_handle);
}

void task_thrust_learner_main(void *argument)
{
    UNUSED(argument);

    TickType_t last_wake_time = xTaskGetTickCount();
    const TickType_t period = pdMS_TO_TICKS(1000 / THRUST_LEARNER_TASK_FREQ);
    float dt = 1.0f / (float)THRUST_LEARNER_TASK_FREQ;

    while (1) {
        vTaskDelayUntil(&last_wake_time, period);

        thrust_learner_update(dt);
    }
}
