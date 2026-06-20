#ifndef __QUATERNION_H__
#define __QUATERNION_H__

#include "types.h"

void quaternion_init(Quaternion *q);
void quaternion_from_euler(Quaternion *q, float roll, float pitch, float yaw);
void quaternion_to_euler(const Quaternion *q, EulerAngle *euler);
void quaternion_multiply(Quaternion *result, const Quaternion *q1, const Quaternion *q2);
void quaternion_conjugate(Quaternion *result, const Quaternion *q);
void quaternion_inverse(Quaternion *result, const Quaternion *q);
float quaternion_norm(const Quaternion *q);
void quaternion_normalize(Quaternion *q);
void quaternion_derivative(Quaternion *dq, const Quaternion *q, const Vector3f *omega);
void quaternion_integrate(Quaternion *q, const Vector3f *omega, float dt);
void quaternion_rotate_vector(Vector3f *result, const Quaternion *q, const Vector3f *v);
void quaternion_from_axis_angle(Quaternion *q, const Vector3f *axis, float angle);
float quaternion_dot(const Quaternion *q1, const Quaternion *q2);
void quaternion_slerp(Quaternion *result, const Quaternion *q1, const Quaternion *q2, float t);
float quaternion_get_roll(const Quaternion *q);
float quaternion_get_pitch(const Quaternion *q);
float quaternion_get_yaw(const Quaternion *q);

#endif
