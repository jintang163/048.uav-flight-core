#ifndef __TASK_MOTOR_CONTROL_H__
#define __TASK_MOTOR_CONTROL_H__

#include "types.h"
#include "flight_config.h"
#include "FreeRTOS.h"
#include "task.h"

void task_motor_control_init(void);
void task_motor_control_main(void *argument);

#endif
