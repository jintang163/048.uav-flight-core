#ifndef __TASK_LINK_MANAGER_H__
#define __TASK_LINK_MANAGER_H__

#include "types.h"
#include "flight_config.h"
#include "FreeRTOS.h"
#include "task.h"

void task_link_manager_init(void);
void task_link_manager_main(void *argument);

#endif
