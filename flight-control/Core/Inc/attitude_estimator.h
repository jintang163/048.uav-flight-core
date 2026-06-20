#ifndef __ATTITUDE_ESTIMATOR_H__
#define __ATTITUDE_ESTIMATOR_H__

#include "types.h"
#include "flight_config.h"

typedef struct {
    Quaternion quat;
    Vector3f gyro_bias;
    float P[16];
    uint32_t last_update;
    bool initialized;
} EKFState;

typedef struct {
    Vector3f accel;
    Vector3f gyro;
    Vector3f mag;
    uint32_t timestamp;
} IMUData;

void attitude_estimator_init(void);
void attitude_estimator_update(IMUData *imu_data, float dt);
void attitude_estimator_get_quaternion(Quaternion *quat);
void attitude_estimator_get_euler(EulerAngle *euler);
void attitude_estimator_get_angular_velocity(Vector3f *gyro);
void attitude_estimator_get_linear_accel(Vector3f *accel);
float attitude_estimator_get_yaw_rate(void);
void attitude_estimator_reset(void);

#endif
