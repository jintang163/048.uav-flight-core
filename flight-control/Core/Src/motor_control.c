#include "motor_control.h"
#include "pwm_output.h"
#include "flight_config.h"

static MotorControlData motor_data;

void motor_control_init(void)
{
    memset(&motor_data, 0, sizeof(MotorControlData));

    for (int i = 0; i < 4; i++) {
        motor_data.output[i] = 0.0f;
        motor_data.armed = false;
    }
}

void motor_control_update(float throttle, float roll, float pitch, float yaw)
{
    throttle = CONSTRAIN(throttle, 0.0f, 1.0f);
    roll = CONSTRAIN(roll, -1.0f, 1.0f);
    pitch = CONSTRAIN(pitch, -1.0f, 1.0f);
    yaw = CONSTRAIN(yaw, -1.0f, 1.0f);

    motor_data.throttle = throttle;
    motor_data.roll = roll;
    motor_data.pitch = pitch;
    motor_data.yaw = yaw;

    if (!motor_data.armed) {
        for (int i = 0; i < 4; i++) {
            motor_data.output[i] = 0.0f;
        }
        pwm_output_update((float *)motor_data.output);
        return;
    }

    motor_mix(throttle, roll, pitch, yaw, (float *)motor_data.output);

    for (int i = 0; i < 4; i++) {
        motor_data.output[i] = CONSTRAIN(motor_data.output[i], 0.0f, 1.0f);
    }

    pwm_output_update((float *)motor_data.output);
}

void motor_mix(float throttle, float roll, float pitch, float yaw, float *motors)
{
    /*
     * 四轴X型电机混控算法
     *
     * 电机布局（机头朝上，X型）:
     *
     *        机头
     *   M1        M2
     *     \      /
     *       \  /
     *        /\
     *       /  \
     *     /      \
     *   M4        M3
     *
     * 旋转方向：
     *   M1: 逆时针 (CCW) - 产生正升力，正yaw力矩
     *   M2: 顺时针 (CW)  - 产生正升力，负yaw力矩
     *   M3: 逆时针 (CCW) - 产生正升力，正yaw力矩
     *   M4: 顺时针 (CW)  - 产生正升力，负yaw力矩
     *
     * 控制输入范围：
     *   throttle: [0, 1]  - 油门
     *   roll:     [-1, 1] - 横滚（左负右正）
     *   pitch:    [-1, 1] - 俯仰（后负前正）
     *   yaw:      [-1, 1] - 偏航（左负右正）
     *
     * 力矩分配系数：
     *   roll_coeff:  横滚力矩系数 = 0.5
     *   pitch_coeff: 俯仰力矩系数 = 0.5
     *   yaw_coeff:   偏航力矩系数 = 0.5
     *
     * 混控公式（X型）：
     *   M1 = throttle - roll_coeff * roll + pitch_coeff * pitch + yaw_coeff * yaw
     *   M2 = throttle + roll_coeff * roll + pitch_coeff * pitch - yaw_coeff * yaw
     *   M3 = throttle + roll_coeff * roll - pitch_coeff * pitch + yaw_coeff * yaw
     *   M4 = throttle - roll_coeff * roll - pitch_coeff * pitch - yaw_coeff * yaw
     *
     * 说明：
     *   - 右滚(roll>0): M2, M3 加速，M1, M4 减速
     *   - 左滚(roll<0): M1, M4 加速，M2, M3 减速
     *   - 前俯(pitch>0): M1, M2 加速，M3, M4 减速
     *   - 后仰(pitch<0): M3, M4 加速，M1, M2 减速
     *   - 右偏(yaw>0): M1, M3 加速，M2, M4 减速
     *   - 左偏(yaw<0): M2, M4 加速，M1, M3 减速
     */

    float roll_coeff = 0.5f;
    float pitch_coeff = 0.5f;
    float yaw_coeff = 0.5f;

    motors[MOTOR_M1] = throttle - roll_coeff * roll + pitch_coeff * pitch + yaw_coeff * yaw;
    motors[MOTOR_M2] = throttle + roll_coeff * roll + pitch_coeff * pitch - yaw_coeff * yaw;
    motors[MOTOR_M3] = throttle + roll_coeff * roll - pitch_coeff * pitch + yaw_coeff * yaw;
    motors[MOTOR_M4] = throttle - roll_coeff * roll - pitch_coeff * pitch - yaw_coeff * yaw;

    float max_motor = motors[0];
    for (int i = 1; i < 4; i++) {
        if (motors[i] > max_motor) {
            max_motor = motors[i];
        }
    }

    if (max_motor > 1.0f && throttle > MOTOR_MIN_THROTTLE) {
        float scale = (1.0f - throttle) / (max_motor - throttle);
        for (int i = 0; i < 4; i++) {
            motors[i] = throttle + (motors[i] - throttle) * scale;
        }
    }

    for (int i = 0; i < 4; i++) {
        motors[i] = CONSTRAIN(motors[i], 0.0f, 1.0f);
    }
}

void motor_control_arm(void)
{
    motor_data.armed = true;
}

void motor_control_disarm(void)
{
    motor_data.armed = false;
    for (int i = 0; i < 4; i++) {
        motor_data.output[i] = 0.0f;
    }
    pwm_output_update((float *)motor_data.output);
}

bool motor_control_is_armed(void)
{
    return motor_data.armed;
}

void motor_control_get_output(float *output)
{
    for (int i = 0; i < 4; i++) {
        output[i] = motor_data.output[i];
    }
}

void motor_control_stop_all(void)
{
    for (int i = 0; i < 4; i++) {
        motor_data.output[i] = 0.0f;
    }
    pwm_output_update((float *)motor_data.output);
}

void motor_control_calibrate_esc(void)
{
    if (!motor_data.armed) {
        for (int i = 0; i < 4; i++) {
            motor_data.output[i] = 1.0f;
        }
        pwm_output_update((float *)motor_data.output);
        HAL_Delay(2000);

        for (int i = 0; i < 4; i++) {
            motor_data.output[i] = 0.0f;
        }
        pwm_output_update((float *)motor_data.output);
        HAL_Delay(2000);
    }
}
