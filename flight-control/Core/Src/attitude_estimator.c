#include "attitude_estimator.h"
#include "ekf.h"
#include "sensor_manager.h"
#include "quaternion.h"

static EKF ekf;
static AttitudeState attitude_state;
static Vector3f last_gyro;
static uint32_t last_update_time;

void attitude_estimator_init(void)
{
    ekf_init(&ekf);
    memset(&attitude_state, 0, sizeof(AttitudeState));
    quaternion_init(&attitude_state.quat);
    last_gyro.x = 0;
    last_gyro.y = 0;
    last_gyro.z = 0;
    last_update_time = HAL_GetTick();
}

void attitude_estimator_update(IMUData *imu_data, float dt)
{
    if (dt <= 0.0f) {
        return;
    }

    if (!ekf.initialized) {
        Vector3f mag;
        sensor_manager_get_mag(&mag);

        if (fabsf(imu_data->accel.x) > 0.1f ||
            fabsf(imu_data->accel.y) > 0.1f ||
            fabsf(imu_data->accel.z) > 0.1f) {
            ekf_set_initial_attitude(&ekf, &imu_data->accel, &mag);
        }
        return;
    }

    Vector3f gyro_corrected;
    gyro_corrected.x = imu_data->gyro.x * (1.0f - GYRO_LPF_ALPHA) + last_gyro.x * GYRO_LPF_ALPHA;
    gyro_corrected.y = imu_data->gyro.y * (1.0f - GYRO_LPF_ALPHA) + last_gyro.y * GYRO_LPF_ALPHA;
    gyro_corrected.z = imu_data->gyro.z * (1.0f - GYRO_LPF_ALPHA) + last_gyro.z * GYRO_LPF_ALPHA;
    last_gyro = gyro_corrected;

    ekf_predict(&ekf, &gyro_corrected, dt);

    float accel_mag = sqrtf(imu_data->accel.x * imu_data->accel.x +
                           imu_data->accel.y * imu_data->accel.y +
                           imu_data->accel.z * imu_data->accel.z);

    if (accel_mag > 8.0f && accel_mag < 12.0f) {
        ekf_update_accel(&ekf, &imu_data->accel);
    }

    Vector3f mag;
    sensor_manager_get_mag(&mag);
    float mag_mag = sqrtf(mag.x * mag.x + mag.y * mag.y + mag.z * mag.z);
    if (mag_mag > 0.1f) {
        ekf_update_mag(&ekf, &mag);
    }

    ekf_get_quaternion(&ekf, &attitude_state.quat);
    ekf_get_euler(&ekf, &attitude_state.euler);
    attitude_state.angular_velocity = gyro_corrected;
    attitude_state.linear_accel = imu_data->accel;
    attitude_state.yaw_rate = gyro_corrected.z;

    last_update_time = HAL_GetTick();
}

void attitude_estimator_get_quaternion(Quaternion *quat)
{
    *quat = attitude_state.quat;
}

void attitude_estimator_get_euler(EulerAngle *euler)
{
    *euler = attitude_state.euler;
}

void attitude_estimator_get_angular_velocity(Vector3f *gyro)
{
    *gyro = attitude_state.angular_velocity;
}

void attitude_estimator_get_linear_accel(Vector3f *accel)
{
    *accel = attitude_state.linear_accel;
}

float attitude_estimator_get_yaw_rate(void)
{
    return attitude_state.yaw_rate;
}

void attitude_estimator_reset(void)
{
    ekf_reset(&ekf);
    quaternion_init(&attitude_state.quat);
    attitude_state.euler.roll = 0;
    attitude_state.euler.pitch = 0;
    attitude_state.euler.yaw = 0;
    attitude_state.angular_velocity.x = 0;
    attitude_state.angular_velocity.y = 0;
    attitude_state.angular_velocity.z = 0;
    attitude_state.linear_accel.x = 0;
    attitude_state.linear_accel.y = 0;
    attitude_state.linear_accel.z = 0;
    attitude_state.yaw_rate = 0;
}
