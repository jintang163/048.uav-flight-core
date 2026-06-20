#ifndef __TASK_FLIGHT_CONTROL_H__
#define __TASK_FLIGHT_CONTROL_H__

#include "types.h"
#include "flight_config.h"
#include "FreeRTOS.h"
#include "task.h"

void task_flight_control_init(void);
void task_flight_control_main(void *argument);

#endif
