#include "quaternion.h"

void quaternion_init(Quaternion *q)
{
    q->w = 1.0f;
    q->x = 0.0f;
    q->y = 0.0f;
    q->z = 0.0f;
}

void quaternion_from_euler(Quaternion *q, float roll, float pitch, float yaw)
{
    float cx = cosf(roll * 0.5f);
    float sx = sinf(roll * 0.5f);
    float cy = cosf(pitch * 0.5f);
    float sy = sinf(pitch * 0.5f);
    float cz = cosf(yaw * 0.5f);
    float sz = sinf(yaw * 0.5f);

    q->w = cx * cy * cz + sx * sy * sz;
    q->x = sx * cy * cz - cx * sy * sz;
    q->y = cx * sy * cz + sx * cy * sz;
    q->z = cx * cy * sz - sx * sy * cz;
}

void quaternion_to_euler(const Quaternion *q, EulerAngle *euler)
{
    float sinr_cosp = 2.0f * (q->w * q->x + q->y * q->z);
    float cosr_cosp = 1.0f - 2.0f * (q->x * q->x + q->y * q->y);
    euler->roll = atan2f(sinr_cosp, cosr_cosp);

    float sinp = 2.0f * (q->w * q->y - q->z * q->x);
    if (fabsf(sinp) >= 1.0f) {
        euler->pitch = copysignf(M_PI / 2.0f, sinp);
    } else {
        euler->pitch = asinf(sinp);
    }

    float siny_cosp = 2.0f * (q->w * q->z + q->x * q->y);
    float cosy_cosp = 1.0f - 2.0f * (q->y * q->y + q->z * q->z);
    euler->yaw = atan2f(siny_cosp, cosy_cosp);
}

void quaternion_multiply(Quaternion *result, const Quaternion *q1, const Quaternion *q2)
{
    result->w = q1->w * q2->w - q1->x * q2->x - q1->y * q2->y - q1->z * q2->z;
    result->x = q1->w * q2->x + q1->x * q2->w + q1->y * q2->z - q1->z * q2->y;
    result->y = q1->w * q2->y - q1->x * q2->z + q1->y * q2->w + q1->z * q2->x;
    result->z = q1->w * q2->z + q1->x * q2->y - q1->y * q2->x + q1->z * q2->w;
}

void quaternion_conjugate(Quaternion *result, const Quaternion *q)
{
    result->w = q->w;
    result->x = -q->x;
    result->y = -q->y;
    result->z = -q->z;
}

void quaternion_inverse(Quaternion *result, const Quaternion *q)
{
    float norm_sq = q->w * q->w + q->x * q->x + q->y * q->y + q->z * q->z;
    if (norm_sq > 0.0f) {
        float inv_norm = 1.0f / norm_sq;
        result->w = q->w * inv_norm;
        result->x = -q->x * inv_norm;
        result->y = -q->y * inv_norm;
        result->z = -q->z * inv_norm;
    } else {
        quaternion_init(result);
    }
}

float quaternion_norm(const Quaternion *q)
{
    return sqrtf(q->w * q->w + q->x * q->x + q->y * q->y + q->z * q->z);
}

void quaternion_normalize(Quaternion *q)
{
    float norm = quaternion_norm(q);
    if (norm > 0.0f) {
        float inv_norm = 1.0f / norm;
        q->w *= inv_norm;
        q->x *= inv_norm;
        q->y *= inv_norm;
        q->z *= inv_norm;
    }
}

void quaternion_derivative(Quaternion *dq, const Quaternion *q, const Vector3f *omega)
{
    Quaternion omega_q;
    omega_q.w = 0.0f;
    omega_q.x = omega->x * 0.5f;
    omega_q.y = omega->y * 0.5f;
    omega_q.z = omega->z * 0.5f;

    quaternion_multiply(dq, q, &omega_q);
}

void quaternion_integrate(Quaternion *q, const Vector3f *omega, float dt)
{
    Quaternion dq;
    quaternion_derivative(&dq, q, omega);

    q->w += dq.w * dt;
    q->x += dq.x * dt;
    q->y += dq.y * dt;
    q->z += dq.z * dt;

    quaternion_normalize(q);
}

void quaternion_rotate_vector(Vector3f *result, const Quaternion *q, const Vector3f *v)
{
    Quaternion q_conj, qv, q_result;

    qv.w = 0.0f;
    qv.x = v->x;
    qv.y = v->y;
    qv.z = v->z;

    quaternion_conjugate(&q_conj, q);
    quaternion_multiply(&q_result, q, &qv);
    quaternion_multiply(&q_result, &q_result, &q_conj);

    result->x = q_result.x;
    result->y = q_result.y;
    result->z = q_result.z;
}

void quaternion_from_axis_angle(Quaternion *q, const Vector3f *axis, float angle)
{
    float half_angle = angle * 0.5f;
    float sin_half = sinf(half_angle);
    float cos_half = cosf(half_angle);

    q->w = cos_half;
    q->x = axis->x * sin_half;
    q->y = axis->y * sin_half;
    q->z = axis->z * sin_half;
}

float quaternion_dot(const Quaternion *q1, const Quaternion *q2)
{
    return q1->w * q2->w + q1->x * q2->x + q1->y * q2->y + q1->z * q2->z;
}

void quaternion_slerp(Quaternion *result, const Quaternion *q1, const Quaternion *q2, float t)
{
    float dot = quaternion_dot(q1, q2);

    Quaternion q2_local = *q2;
    if (dot < 0.0f) {
        q2_local.w = -q2_local.w;
        q2_local.x = -q2_local.x;
        q2_local.y = -q2_local.y;
        q2_local.z = -q2_local.z;
        dot = -dot;
    }

    if (dot > 0.9995f) {
        result->w = q1->w + t * (q2_local.w - q1->w);
        result->x = q1->x + t * (q2_local.x - q1->x);
        result->y = q1->y + t * (q2_local.y - q1->y);
        result->z = q1->z + t * (q2_local.z - q1->z);
        quaternion_normalize(result);
        return;
    }

    float theta_0 = acosf(dot);
    float sin_theta_0 = sinf(theta_0);
    float theta = theta_0 * t;
    float sin_theta = sinf(theta);

    float s0 = cosf(theta) - dot * sin_theta / sin_theta_0;
    float s1 = sin_theta / sin_theta_0;

    result->w = s0 * q1->w + s1 * q2_local.w;
    result->x = s0 * q1->x + s1 * q2_local.x;
    result->y = s0 * q1->y + s1 * q2_local.y;
    result->z = s0 * q1->z + s1 * q2_local.z;
}

float quaternion_get_roll(const Quaternion *q)
{
    float sinr_cosp = 2.0f * (q->w * q->x + q->y * q->z);
    float cosr_cosp = 1.0f - 2.0f * (q->x * q->x + q->y * q->y);
    return atan2f(sinr_cosp, cosr_cosp);
}

float quaternion_get_pitch(const Quaternion *q)
{
    float sinp = 2.0f * (q->w * q->y - q->z * q->x);
    if (fabsf(sinp) >= 1.0f) {
        return copysignf(M_PI / 2.0f, sinp);
    } else {
        return asinf(sinp);
    }
}

float quaternion_get_yaw(const Quaternion *q)
{
    float siny_cosp = 2.0f * (q->w * q->z + q->x * q->y);
    float cosy_cosp = 1.0f - 2.0f * (q->y * q->y + q->z * q->z);
    return atan2f(siny_cosp, cosy_cosp);
}
