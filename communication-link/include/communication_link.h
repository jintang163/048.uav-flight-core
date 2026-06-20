#ifndef COMMUNICATION_LINK_H
#define COMMUNICATION_LINK_H

#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>

#define CL_VERSION_MAJOR 1
#define CL_VERSION_MINOR 0
#define CL_VERSION_PATCH 0

#define COMMUNICATION_LINK_VERSION_MAJOR CL_VERSION_MAJOR
#define COMMUNICATION_LINK_VERSION_MINOR CL_VERSION_MINOR
#define COMMUNICATION_LINK_VERSION_PATCH CL_VERSION_PATCH

#define CL_OK                    0
#define CL_ERROR                -1
#define CL_INVALID_PARAM        -2
#define CL_TIMEOUT              -3
#define CL_NO_DATA              -4
#define CL_BUFFER_TOO_SMALL     -5
#define CL_CRYPTO_ERROR         -6
#define CL_CONNECTION_CLOSED    -7
#define CL_CHECKSUM_ERROR       -8
#define CL_SIGNATURE_ERROR      -9

typedef struct {
    uint8_t major;
    uint8_t minor;
    uint8_t patch;
} cl_version_t;

typedef void (*cl_log_cb_t)(int level, const char* fmt, ...);

void cl_init(void);
void cl_cleanup(void);
cl_version_t cl_get_version(void);
void cl_set_log_callback(cl_log_cb_t cb);

#endif
