#ifndef __EKF_H__
#define __EKF_H__

#include "types.h"
#include "quaternion.h"

typedef struct {
    Quaternion quat;
    Vector3f gyro_bias;
    float P[7][7];
    float Q[7][7];
    float R_accel[3][3];
    float R_mag[3][3];
    bool initialized;
} EKF;

void ekf_init(EKF *ekf);
void ekf_predict(EKF *ekf, const Vector3f *gyro, float dt);
void ekf_update_accel(EKF *ekf, const Vector3f *accel);
void ekf_update_mag(EKF *ekf, const Vector3f *mag);
void ekf_get_quaternion(const EKF *ekf, Quaternion *quat);
void ekf_get_euler(const EKF *ekf, EulerAngle *euler);
void ekf_get_gyro_bias(const EKF *ekf, Vector3f *bias);
void ekf_reset(EKF *ekf);
void ekf_set_initial_attitude(EKF *ekf, const Vector3f *accel, const Vector3f *mag);

#endif
