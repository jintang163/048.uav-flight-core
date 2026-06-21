#ifndef __TASK_BLACKBOX_H__
#define __TASK_BLACKBOX_H__

#include "types.h"
#include "flight_config.h"

#define TASK_BLACKBOX_FREQ         10
#define TASK_BLACKBOX_STACK_SIZE   512
#define TASK_BLACKBOX_PRIORITY     2

void task_blackbox_init(void);
void task_blackbox_main(void *argument);

#endif
