#include "task_attitude_estimation.h"
#include "attitude_estimator.h"
#include "sensor_manager.h"

static TaskHandle_t task_handle = NULL;
static AttitudeState attitude_state;

void task_attitude_estimation_init(void)
{
    attitude_estimator_init();

    memset(&attitude_state, 0, sizeof(AttitudeState));
    quaternion_init(&attitude_state.quat);

    xTaskCreate(task_attitude_estimation_main,
                "AttitudeEst",
                TASK_ATTITUDE_ESTIMATION_STACK_SIZE,
                NULL,
                TASK_ATTITUDE_ESTIMATION_PRIORITY,
                &task_handle);
}

void task_attitude_estimation_main(void *argument)
{
    UNUSED(argument);

    TickType_t last_wake_time = xTaskGetTickCount();
    const TickType_t period = pdMS_TO_TICKS(1000 / TASK_ATTITUDE_ESTIMATION_FREQ);
    float dt = 1.0f / TASK_ATTITUDE_ESTIMATION_FREQ;

    while (1) {
        IMUData imu_data;
        sensor_manager_get_imu(&imu_data);

        attitude_estimator_update(&imu_data, dt);

        attitude_estimator_get_quaternion(&attitude_state.quat);
        attitude_estimator_get_euler(&attitude_state.euler);
        attitude_estimator_get_angular_velocity(&attitude_state.angular_velocity);
        attitude_estimator_get_linear_accel(&attitude_state.linear_accel);
        attitude_state.yaw_rate = attitude_estimator_get_yaw_rate();

        vTaskDelayUntil(&last_wake_time, period);
    }
}

void task_attitude_estimation_get_state(AttitudeState *state)
{
    *state = attitude_state;
}
