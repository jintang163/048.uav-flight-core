#ifndef __TASK_MISSION_H__
#define __TASK_MISSION_H__

#include "types.h"
#include "flight_config.h"
#include "FreeRTOS.h"
#include "task.h"

void task_mission_init(void);
void task_mission_main(void *argument);

#endif
