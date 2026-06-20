#ifndef __TASK_HEALTH_MONITOR_H__
#define __TASK_HEALTH_MONITOR_H__

#include "types.h"
#include "flight_config.h"
#include "FreeRTOS.h"
#include "task.h"

void task_health_monitor_init(void);
void task_health_monitor_main(void *argument);
void task_health_monitor_get_status(HealthStatus *status);

#endif
