#ifndef __TASK_ATTITUDE_ESTIMATION_H__
#define __TASK_ATTITUDE_ESTIMATION_H__

#include "types.h"
#include "flight_config.h"
#include "FreeRTOS.h"
#include "task.h"

void task_attitude_estimation_init(void);
void task_attitude_estimation_main(void *argument);
void task_attitude_estimation_get_state(AttitudeState *state);

#endif
