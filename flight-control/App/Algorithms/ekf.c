#include "ekf.h"
#include "flight_config.h"
#include "coordinate.h"

void ekf_init(EKF *ekf)
{
    quaternion_init(&ekf->quat);

    ekf->gyro_bias.x = 0.0f;
    ekf->gyro_bias.y = 0.0f;
    ekf->gyro_bias.z = 0.0f;

    for (int i = 0; i < 7; i++) {
        for (int j = 0; j < 7; j++) {
            ekf->P[i][j] = 0.0f;
            ekf->Q[i][j] = 0.0f;
        }
    }

    for (int i = 0; i < 4; i++) {
        ekf->P[i][i] = 0.01f;
        ekf->Q[i][i] = EKF_PROCESS_NOISE_QUAT;
    }
    for (int i = 4; i < 7; i++) {
        ekf->P[i][i] = 0.001f;
        ekf->Q[i][i] = EKF_PROCESS_NOISE_GYRO_BIAS;
    }

    for (int i = 0; i < 3; i++) {
        for (int j = 0; j < 3; j++) {
            ekf->R_accel[i][j] = 0.0f;
            ekf->R_mag[i][j] = 0.0f;
        }
        ekf->R_accel[i][i] = EKF_MEASUREMENT_NOISE_ACCEL;
        ekf->R_mag[i][i] = EKF_MEASUREMENT_NOISE_MAG;
    }

    ekf->initialized = false;
}

void ekf_set_initial_attitude(EKF *ekf, const Vector3f *accel, const Vector3f *mag)
{
    float ax = accel->x;
    float ay = accel->y;
    float az = accel->z;

    float pitch = atan2f(-ax, sqrtf(ay * ay + az * az));
    float roll = atan2f(ay, az);

    float cr = cosf(roll);
    float sr = sinf(roll);
    float cp = cosf(pitch);
    float sp = sinf(pitch);

    float mx = mag->x * cp + mag->z * sp;
    float my = mag->x * sr * sp + mag->y * cr - mag->z * sr * cp;
    float yaw = atan2f(-my, mx);

    quaternion_from_euler(&ekf->quat, roll, pitch, yaw);

    ekf->gyro_bias.x = GYRO_BIAS_X;
    ekf->gyro_bias.y = GYRO_BIAS_Y;
    ekf->gyro_bias.z = GYRO_BIAS_Z;

    ekf->initialized = true;
}

void ekf_predict(EKF *ekf, const Vector3f *gyro, float dt)
{
    if (!ekf->initialized) {
        return;
    }

    float gx = gyro->x - ekf->gyro_bias.x;
    float gy = gyro->y - ekf->gyro_bias.y;
    float gz = gyro->z - ekf->gyro_bias.z;

    float qw = ekf->quat.w;
    float qx = ekf->quat.x;
    float qy = ekf->quat.y;
    float qz = ekf->quat.z;

    float dqw = 0.5f * (-qx * gx - qy * gy - qz * gz);
    float dqx = 0.5f * (qw * gx + qy * gz - qz * gy);
    float dqy = 0.5f * (qw * gy - qx * gz + qz * gx);
    float dqz = 0.5f * (qw * gz + qx * gy - qy * gx);

    ekf->quat.w += dqw * dt;
    ekf->quat.x += dqx * dt;
    ekf->quat.y += dqy * dt;
    ekf->quat.z += dqz * dt;
    quaternion_normalize(&ekf->quat);

    float F[7][7];
    for (int i = 0; i < 7; i++) {
        for (int j = 0; j < 7; j++) {
            F[i][j] = 0.0f;
        }
        F[i][i] = 1.0f;
    }

    F[0][1] = -0.5f * gx * dt;
    F[0][2] = -0.5f * gy * dt;
    F[0][3] = -0.5f * gz * dt;
    F[0][4] = 0.5f * qx * dt;
    F[0][5] = 0.5f * qy * dt;
    F[0][6] = 0.5f * qz * dt;

    F[1][0] = 0.5f * gx * dt;
    F[1][2] = 0.5f * gz * dt;
    F[1][3] = -0.5f * gy * dt;
    F[1][4] = -0.5f * qw * dt;
    F[1][5] = 0.5f * qz * dt;
    F[1][6] = -0.5f * qy * dt;

    F[2][0] = 0.5f * gy * dt;
    F[2][1] = -0.5f * gz * dt;
    F[2][3] = 0.5f * gx * dt;
    F[2][4] = -0.5f * qz * dt;
    F[2][5] = -0.5f * qw * dt;
    F[2][6] = 0.5f * qx * dt;

    F[3][0] = 0.5f * gz * dt;
    F[3][1] = 0.5f * gy * dt;
    F[3][2] = -0.5f * gx * dt;
    F[3][4] = 0.5f * qy * dt;
    F[3][5] = -0.5f * qx * dt;
    F[3][6] = -0.5f * qw * dt;

    float P_temp[7][7];
    for (int i = 0; i < 7; i++) {
        for (int j = 0; j < 7; j++) {
            P_temp[i][j] = 0.0f;
            for (int k = 0; k < 7; k++) {
                P_temp[i][j] += F[i][k] * ekf->P[k][j];
            }
        }
    }

    for (int i = 0; i < 7; i++) {
        for (int j = 0; j < 7; j++) {
            ekf->P[i][j] = 0.0f;
            for (int k = 0; k < 7; k++) {
                ekf->P[i][j] += P_temp[i][k] * F[j][k];
            }
            ekf->P[i][j] += ekf->Q[i][j] * dt;
        }
    }
}

void ekf_update_accel(EKF *ekf, const Vector3f *accel)
{
    if (!ekf->initialized) {
        return;
    }

    float qw = ekf->quat.w;
    float qx = ekf->quat.x;
    float qy = ekf->quat.y;
    float qz = ekf->quat.z;

    float hx = 2.0f * (qx * qz - qw * qy);
    float hy = 2.0f * (qw * qx + qy * qz);
    float hz = qw * qw - qx * qx - qy * qy + qz * qz;

    float ax_norm = accel->x;
    float ay_norm = accel->y;
    float az_norm = accel->z;
    float accel_mag = sqrtf(ax_norm * ax_norm + ay_norm * ay_norm + az_norm * az_norm);

    if (accel_mag > 0.1f) {
        ax_norm /= accel_mag;
        ay_norm /= accel_mag;
        az_norm /= accel_mag;
    }

    float y[3];
    y[0] = ax_norm - hx;
    y[1] = ay_norm - hy;
    y[2] = az_norm - hz;

    float H[3][7];
    for (int i = 0; i < 3; i++) {
        for (int j = 0; j < 7; j++) {
            H[i][j] = 0.0f;
        }
    }

    H[0][0] = -2.0f * qy;
    H[0][1] = 2.0f * qz;
    H[0][2] = -2.0f * qw;
    H[0][3] = 2.0f * qx;

    H[1][0] = 2.0f * qx;
    H[1][1] = 2.0f * qw;
    H[1][2] = 2.0f * qz;
    H[1][3] = 2.0f * qy;

    H[2][0] = 2.0f * qw;
    H[2][1] = -2.0f * qx;
    H[2][2] = -2.0f * qy;
    H[2][3] = 2.0f * qz;

    float S[3][3];
    for (int i = 0; i < 3; i++) {
        for (int j = 0; j < 3; j++) {
            S[i][j] = ekf->R_accel[i][j];
            for (int k = 0; k < 7; k++) {
                S[i][j] += H[i][k] * ekf->P[k][j];
            }
        }
    }

    float S_inv[3][3];
    float det = S[0][0] * (S[1][1] * S[2][2] - S[1][2] * S[2][1]) -
                S[0][1] * (S[1][0] * S[2][2] - S[1][2] * S[2][0]) +
                S[0][2] * (S[1][0] * S[2][1] - S[1][1] * S[2][0]);

    if (fabsf(det) < 1e-10f) {
        return;
    }

    float inv_det = 1.0f / det;
    S_inv[0][0] = (S[1][1] * S[2][2] - S[1][2] * S[2][1]) * inv_det;
    S_inv[0][1] = (S[0][2] * S[2][1] - S[0][1] * S[2][2]) * inv_det;
    S_inv[0][2] = (S[0][1] * S[1][2] - S[0][2] * S[1][1]) * inv_det;
    S_inv[1][0] = (S[1][2] * S[2][0] - S[1][0] * S[2][2]) * inv_det;
    S_inv[1][1] = (S[0][0] * S[2][2] - S[0][2] * S[2][0]) * inv_det;
    S_inv[1][2] = (S[0][2] * S[1][0] - S[0][0] * S[1][2]) * inv_det;
    S_inv[2][0] = (S[1][0] * S[2][1] - S[1][1] * S[2][0]) * inv_det;
    S_inv[2][1] = (S[0][1] * S[2][0] - S[0][0] * S[2][1]) * inv_det;
    S_inv[2][2] = (S[0][0] * S[1][1] - S[0][1] * S[1][0]) * inv_det;

    float K[7][3];
    for (int i = 0; i < 7; i++) {
        for (int j = 0; j < 3; j++) {
            K[i][j] = 0.0f;
            for (int k = 0; k < 7; k++) {
                K[i][j] += ekf->P[i][k] * H[j][k];
            }
        }
    }

    for (int i = 0; i < 7; i++) {
        for (int j = 0; j < 3; j++) {
            float temp = K[i][j];
            K[i][j] = 0.0f;
            for (int k = 0; k < 3; k++) {
                K[i][j] += temp * S_inv[k][j];
            }
        }
    }

    float dx[7];
    for (int i = 0; i < 7; i++) {
        dx[i] = 0.0f;
        for (int j = 0; j < 3; j++) {
            dx[i] += K[i][j] * y[j];
        }
    }

    Quaternion dq;
    dq.w = 1.0f - 0.125f * (dx[1] * dx[1] + dx[2] * dx[2] + dx[3] * dx[3]);
    dq.x = 0.5f * dx[1];
    dq.y = 0.5f * dx[2];
    dq.z = 0.5f * dx[3];
    quaternion_normalize(&dq);

    Quaternion new_q;
    quaternion_multiply(&new_q, &ekf->quat, &dq);
    ekf->quat = new_q;
    quaternion_normalize(&ekf->quat);

    ekf->gyro_bias.x += dx[4];
    ekf->gyro_bias.y += dx[5];
    ekf->gyro_bias.z += dx[6];

    float KH[7][7];
    for (int i = 0; i < 7; i++) {
        for (int j = 0; j < 7; j++) {
            KH[i][j] = 0.0f;
            for (int k = 0; k < 3; k++) {
                KH[i][j] += K[i][k] * H[k][j];
            }
        }
    }

    for (int i = 0; i < 7; i++) {
        for (int j = 0; j < 7; j++) {
            ekf->P[i][j] = ekf->P[i][j] - KH[i][j] * ekf->P[i][j];
        }
    }
}

void ekf_update_mag(EKF *ekf, const Vector3f *mag)
{
    if (!ekf->initialized) {
        return;
    }

    float qw = ekf->quat.w;
    float qx = ekf->quat.x;
    float qy = ekf->quat.y;
    float qz = ekf->quat.z;

    float mx = mag->x;
    float my = mag->y;
    float mz = mag->z;

    float mag_mag = sqrtf(mx * mx + my * my + mz * mz);
    if (mag_mag < 0.1f) {
        return;
    }
    mx /= mag_mag;
    my /= mag_mag;
    mz /= mag_mag;

    float hx = mx * (qw * qw + qx * qx - qy * qy - qz * qz) +
               my * 2.0f * (qx * qy + qw * qz) +
               mz * 2.0f * (qx * qz - qw * qy);
    float hy = mx * 2.0f * (qx * qy - qw * qz) +
               my * (qw * qw - qx * qx + qy * qy - qz * qz) +
               mz * 2.0f * (qy * qz + qw * qx);
    float hz = mx * 2.0f * (qx * qz + qw * qy) +
               my * 2.0f * (qy * qz - qw * qx) +
               mz * (qw * qw - qx * qx - qy * qy + qz * qz);

    float y[3];
    y[0] = mx - hx;
    y[1] = my - hy;
    y[2] = mz - hz;

    float H[3][7];
    for (int i = 0; i < 3; i++) {
        for (int j = 0; j < 7; j++) {
            H[i][j] = 0.0f;
        }
    }

    H[0][0] = 2.0f * mx * qw + 2.0f * my * qz - 2.0f * mz * qy;
    H[0][1] = 2.0f * mx * qx + 2.0f * my * qy + 2.0f * mz * qz;
    H[0][2] = -2.0f * mx * qy + 2.0f * my * qx - 2.0f * mz * qw;
    H[0][3] = -2.0f * mx * qz + 2.0f * my * qw + 2.0f * mz * qx;

    H[1][0] = -2.0f * mx * qz + 2.0f * my * qw + 2.0f * mz * qx;
    H[1][1] = 2.0f * mx * qy - 2.0f * my * qx + 2.0f * mz * qw;
    H[1][2] = 2.0f * mx * qx + 2.0f * my * qy + 2.0f * mz * qz;
    H[1][3] = -2.0f * mx * qw - 2.0f * my * qz + 2.0f * mz * qy;

    H[2][0] = 2.0f * mx * qy - 2.0f * my * qx + 2.0f * mz * qw;
    H[2][1] = 2.0f * mx * qz - 2.0f * my * qw - 2.0f * mz * qx;
    H[2][2] = 2.0f * mx * qw + 2.0f * my * qz - 2.0f * mz * qy;
    H[2][3] = 2.0f * mx * qx + 2.0f * my * qy + 2.0f * mz * qz;

    float S[3][3];
    for (int i = 0; i < 3; i++) {
        for (int j = 0; j < 3; j++) {
            S[i][j] = ekf->R_mag[i][j];
            for (int k = 0; k < 7; k++) {
                S[i][j] += H[i][k] * ekf->P[k][j];
            }
        }
    }

    float S_inv[3][3];
    float det = S[0][0] * (S[1][1] * S[2][2] - S[1][2] * S[2][1]) -
                S[0][1] * (S[1][0] * S[2][2] - S[1][2] * S[2][0]) +
                S[0][2] * (S[1][0] * S[2][1] - S[1][1] * S[2][0]);

    if (fabsf(det) < 1e-10f) {
        return;
    }

    float inv_det = 1.0f / det;
    S_inv[0][0] = (S[1][1] * S[2][2] - S[1][2] * S[2][1]) * inv_det;
    S_inv[0][1] = (S[0][2] * S[2][1] - S[0][1] * S[2][2]) * inv_det;
    S_inv[0][2] = (S[0][1] * S[1][2] - S[0][2] * S[1][1]) * inv_det;
    S_inv[1][0] = (S[1][2] * S[2][0] - S[1][0] * S[2][2]) * inv_det;
    S_inv[1][1] = (S[0][0] * S[2][2] - S[0][2] * S[2][0]) * inv_det;
    S_inv[1][2] = (S[0][2] * S[1][0] - S[0][0] * S[1][2]) * inv_det;
    S_inv[2][0] = (S[1][0] * S[2][1] - S[1][1] * S[2][0]) * inv_det;
    S_inv[2][1] = (S[0][1] * S[2][0] - S[0][0] * S[2][1]) * inv_det;
    S_inv[2][2] = (S[0][0] * S[1][1] - S[0][1] * S[1][0]) * inv_det;

    float K[7][3];
    for (int i = 0; i < 7; i++) {
        for (int j = 0; j < 3; j++) {
            K[i][j] = 0.0f;
            for (int k = 0; k < 7; k++) {
                K[i][j] += ekf->P[i][k] * H[j][k];
            }
        }
    }

    for (int i = 0; i < 7; i++) {
        for (int j = 0; j < 3; j++) {
            float temp = K[i][j];
            K[i][j] = 0.0f;
            for (int k = 0; k < 3; k++) {
                K[i][j] += temp * S_inv[k][j];
            }
        }
    }

    float dx[7];
    for (int i = 0; i < 7; i++) {
        dx[i] = 0.0f;
        for (int j = 0; j < 3; j++) {
            dx[i] += K[i][j] * y[j];
        }
    }

    Quaternion dq;
    dq.w = 1.0f - 0.125f * (dx[1] * dx[1] + dx[2] * dx[2] + dx[3] * dx[3]);
    dq.x = 0.5f * dx[1];
    dq.y = 0.5f * dx[2];
    dq.z = 0.5f * dx[3];
    quaternion_normalize(&dq);

    Quaternion new_q;
    quaternion_multiply(&new_q, &ekf->quat, &dq);
    ekf->quat = new_q;
    quaternion_normalize(&ekf->quat);

    ekf->gyro_bias.x += dx[4];
    ekf->gyro_bias.y += dx[5];
    ekf->gyro_bias.z += dx[6];

    float KH[7][7];
    for (int i = 0; i < 7; i++) {
        for (int j = 0; j < 7; j++) {
            KH[i][j] = 0.0f;
            for (int k = 0; k < 3; k++) {
                KH[i][j] += K[i][k] * H[k][j];
            }
        }
    }

    for (int i = 0; i < 7; i++) {
        for (int j = 0; j < 7; j++) {
            ekf->P[i][j] = ekf->P[i][j] - KH[i][j] * ekf->P[i][j];
        }
    }
}

void ekf_get_quaternion(const EKF *ekf, Quaternion *quat)
{
    *quat = ekf->quat;
}

void ekf_get_euler(const EKF *ekf, EulerAngle *euler)
{
    quaternion_to_euler(&ekf->quat, euler);
}

void ekf_get_gyro_bias(const EKF *ekf, Vector3f *bias)
{
    *bias = ekf->gyro_bias;
}

void ekf_reset(EKF *ekf)
{
    ekf_init(ekf);
}
