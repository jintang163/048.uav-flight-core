#ifndef __THRUST_LEARNER_H__
#define __THRUST_LEARNER_H__

#include "types.h"
#include "flight_config.h"

typedef struct {
    float throttle;
    float accel_z;
    float roll;
    float pitch;
    float motor_pwm[4];
    float voltage;
    uint32_t timestamp;
} ThrustSample;

typedef struct {
    float throttle;
    float thrust_N;
    float motor_rpm_avg;
} ThrustCurvePoint;

typedef struct {
    float roll_p;
    float roll_i;
    float roll_d;
    float rate_roll_p;
    float rate_roll_i;
    float rate_roll_d;
    float pitch_p;
    float pitch_i;
    float pitch_d;
    float rate_pitch_p;
    float rate_pitch_i;
    float rate_pitch_d;
    float yaw_p;
    float yaw_i;
    float yaw_d;
    float rate_yaw_p;
    float rate_yaw_i;
    float rate_yaw_d;
    float alt_p;
    float alt_i;
    float alt_d;
} PIDGainSet;

typedef struct {
    bool estimating;
    uint32_t start_time;
    float avg_accel_z;
    float avg_throttle;
    float hover_throttle;
    float estimated_weight_kg;
    uint32_t sample_count;
} WeightEstimateState;

typedef enum {
    LS_IDLE = 0,
    LS_WEIGHT_ESTIMATION = 1,
    LS_DATA_COLLECTING = 2,
    LS_MODEL_OPTIMIZING = 3,
    LS_APPLIED = 4
} LearningState;

void thrust_learner_init(void);
void thrust_learner_update(float dt);
void thrust_learner_start_weight_estimation(void);
float thrust_learner_get_estimated_weight(void);
void thrust_learner_get_pid_gains(PIDGainSet *gains);
void thrust_learner_get_thrust_curve(ThrustCurvePoint *points, uint8_t start_index, uint8_t count, uint8_t *out_count);
LearningState thrust_learner_get_state(void);
float thrust_learner_get_hover_throttle(void);
void thrust_learner_trigger_optimization(void);
uint32_t thrust_learner_get_sample_count(void);

#endif
