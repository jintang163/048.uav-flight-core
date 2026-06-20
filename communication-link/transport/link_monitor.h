#ifndef LINK_MONITOR_H
#define LINK_MONITOR_H

#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>

#define LINK_MONITOR_MAX_SAMPLES    100
#define LINK_MONITOR_DEFAULT_INTERVAL_MS    1000
#define LINK_MONITOR_DEFAULT_WINDOW_MS    10000
#define LINK_QUALITY_MAX    100
#define LINK_QUALITY_MIN    0

typedef struct {
    uint32_t rtt_min_ms;
    uint32_t rtt_max_ms;
    uint32_t rtt_avg_ms;
    uint32_t rtt_last_ms;
    float packet_loss_rate;
    uint32_t packets_sent;
    uint32_t packets_received;
    uint32_t packets_lost;
    uint32_t bytes_sent_per_sec;
    uint32_t bytes_received_per_sec;
    int32_t signal_strength;
    int32_t signal_quality;
    int32_t link_quality_score;
} link_stats_t;

typedef struct {
    uint32_t seq;
    uint64_t timestamp_send;
    uint64_t timestamp_recv;
    uint32_t rtt_ms;
    bool received;
} ping_sample_t;

typedef struct {
    ping_sample_t samples[LINK_MONITOR_MAX_SAMPLES];
    uint32_t sample_index;
    uint32_t sample_count;
    uint32_t ping_interval_ms;
    uint32_t ping_timeout_ms;
    uint32_t window_size_ms;
    uint32_t last_ping_ms;
    uint32_t ping_seq;
    uint32_t pending_ping_count;
    uint32_t max_pending_pings;
    bool monitoring;
    link_stats_t stats;
    uint64_t bytes_sent;
    uint64_t bytes_received;
    uint64_t last_stats_update;
} link_monitor_t;

typedef void (*link_monitor_cb_t)(const link_stats_t* stats, void* user_data);

int link_monitor_init(link_monitor_t* monitor);
int link_monitor_start(link_monitor_t* monitor);
int link_monitor_stop(link_monitor_t* monitor);
int link_monitor_set_interval(link_monitor_t* monitor, uint32_t interval_ms);
int link_monitor_set_timeout(link_monitor_t* monitor, uint32_t timeout_ms);
int link_monitor_set_window(link_monitor_t* monitor, uint32_t window_ms);
int link_monitor_add_ping_sent(link_monitor_t* monitor, uint32_t seq);
int link_monitor_add_ping_received(link_monitor_t* monitor, uint32_t seq, uint64_t timestamp);
int link_monitor_add_packet_sent(link_monitor_t* monitor, size_t bytes);
int link_monitor_add_packet_received(link_monitor_t* monitor, size_t bytes);
int link_monitor_add_packet_lost(link_monitor_t* monitor, uint32_t count);
int link_monitor_set_signal_strength(link_monitor_t* monitor, int32_t strength);
int link_monitor_set_signal_quality(link_monitor_t* monitor, int32_t quality);
int link_monitor_get_stats(const link_monitor_t* monitor, link_stats_t* stats);
int32_t link_monitor_get_quality_score(const link_monitor_t* monitor);
bool link_monitor_is_monitoring(const link_monitor_t* monitor);
int link_monitor_reset(link_monitor_t* monitor);

#endif
