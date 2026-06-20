#ifndef __TASK_FORMATION_H__
#define __TASK_FORMATION_H__

#include "types.h"
#include "flight_config.h"
#include "FreeRTOS.h"
#include "task.h"

void task_formation_init(void);
void task_formation_main(void *argument);

#endif
