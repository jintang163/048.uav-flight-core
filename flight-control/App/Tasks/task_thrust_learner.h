#ifndef __TASK_THRUST_LEARNER_H__
#define __TASK_THRUST_LEARNER_H__

#include "main.h"

#define THRUST_LEARNER_TASK_FREQ       TASK_THRUST_LEARNER_FREQ
#define THRUST_LEARNER_TASK_STACK_SIZE TASK_THRUST_LEARNER_STACK_SIZE
#define THRUST_LEARNER_TASK_PRIORITY   (tskIDLE_PRIORITY + TASK_THRUST_LEARNER_PRIORITY)

void task_thrust_learner_init(void);
void task_thrust_learner_main(void *argument);

#endif
