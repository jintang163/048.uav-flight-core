#include "link_monitor.h"
#include <string.h>
#include <stdlib.h>
#include <math.h>

int link_monitor_init(link_monitor_t* monitor)
{
    if (!monitor) return -1;
    memset(monitor, 0, sizeof(link_monitor_t));
    monitor->ping_interval_ms = LINK_MONITOR_DEFAULT_INTERVAL_MS;
    monitor->ping_timeout_ms = 2000;
    monitor->window_size_ms = LINK_MONITOR_DEFAULT_WINDOW_MS;
    monitor->max_pending_pings = 10;
    monitor->monitoring = false;
    monitor->sample_index = 0;
    monitor->sample_count = 0;
    monitor->ping_seq = 0;
    monitor->pending_ping_count = 0;
    monitor->stats.link_quality_score = LINK_QUALITY_MAX;
    monitor->stats.signal_strength = -50;
    monitor->stats.signal_quality = 100;
    return 0;
}

int link_monitor_start(link_monitor_t* monitor)
{
    if (!monitor) return -1;
    link_monitor_reset(monitor);
    monitor->monitoring = true;
    return 0;
}

int link_monitor_stop(link_monitor_t* monitor)
{
    if (!monitor) return -1;
    monitor->monitoring = false;
    return 0;
}

int link_monitor_set_interval(link_monitor_t* monitor, uint32_t interval_ms)
{
    if (!monitor || interval_ms == 0) return -1;
    monitor->ping_interval_ms = interval_ms;
    return 0;
}

int link_monitor_set_timeout(link_monitor_t* monitor, uint32_t timeout_ms)
{
    if (!monitor || timeout_ms == 0) return -1;
    monitor->ping_timeout_ms = timeout_ms;
    return 0;
}

int link_monitor_set_window(link_monitor_t* monitor, uint32_t window_ms)
{
    if (!monitor || window_ms == 0) return -1;
    monitor->window_size_ms = window_ms;
    return 0;
}

int link_monitor_add_ping_sent(link_monitor_t* monitor, uint32_t seq)
{
    if (!monitor || !monitor->monitoring) return -1;
    if (monitor->pending_ping_count >= monitor->max_pending_pings) return -1;

    uint32_t idx = monitor->sample_index;
    monitor->samples[idx].seq = seq;
    monitor->samples[idx].timestamp_send = 0;
    monitor->samples[idx].timestamp_recv = 0;
    monitor->samples[idx].rtt_ms = 0;
    monitor->samples[idx].received = false;

    monitor->sample_index = (monitor->sample_index + 1) % LINK_MONITOR_MAX_SAMPLES;
    if (monitor->sample_count < LINK_MONITOR_MAX_SAMPLES) {
        monitor->sample_count++;
    }
    monitor->pending_ping_count++;
    monitor->stats.packets_sent++;
    return 0;
}

int link_monitor_add_ping_received(link_monitor_t* monitor, uint32_t seq, uint64_t timestamp)
{
    if (!monitor || !monitor->monitoring) return -1;

    for (uint32_t i = 0; i < monitor->sample_count; i++) {
        uint32_t idx = (monitor->sample_index + LINK_MONITOR_MAX_SAMPLES - i - 1) % LINK_MONITOR_MAX_SAMPLES;
        if (monitor->samples[idx].seq == seq && !monitor->samples[idx].received) {
            monitor->samples[idx].timestamp_recv = timestamp;
            monitor->samples[idx].received = true;
            monitor->samples[idx].rtt_ms = (uint32_t)(timestamp - monitor->samples[idx].timestamp_send);
            monitor->stats.packets_received++;
            if (monitor->pending_ping_count > 0) {
                monitor->pending_ping_count--;
            }
            return 0;
        }
    }
    return -1;
}

int link_monitor_add_packet_sent(link_monitor_t* monitor, size_t bytes)
{
    if (!monitor) return -1;
    monitor->bytes_sent += bytes;
    return 0;
}

int link_monitor_add_packet_received(link_monitor_t* monitor, size_t bytes)
{
    if (!monitor) return -1;
    monitor->bytes_received += bytes;
    return 0;
}

int link_monitor_add_packet_lost(link_monitor_t* monitor, uint32_t count)
{
    if (!monitor) return -1;
    monitor->stats.packets_lost += count;
    return 0;
}

int link_monitor_set_signal_strength(link_monitor_t* monitor, int32_t strength)
{
    if (!monitor) return -1;
    monitor->stats.signal_strength = strength;
    return 0;
}

int link_monitor_set_signal_quality(link_monitor_t* monitor, int32_t quality)
{
    if (!monitor) return -1;
    monitor->stats.signal_quality = quality;
    return 0;
}

static void link_monitor_calculate_stats(link_monitor_t* monitor)
{
    if (!monitor) return;

    uint32_t rtt_min = 0xFFFFFFFF;
    uint32_t rtt_max = 0;
    uint64_t rtt_sum = 0;
    uint32_t rtt_count = 0;
    uint32_t received = 0;
    uint32_t total = 0;

    for (uint32_t i = 0; i < monitor->sample_count; i++) {
        if (monitor->samples[i].received) {
            received++;
            if (monitor->samples[i].rtt_ms < rtt_min) {
                rtt_min = monitor->samples[i].rtt_ms;
            }
            if (monitor->samples[i].rtt_ms > rtt_max) {
                rtt_max = monitor->samples[i].rtt_ms;
            }
            rtt_sum += monitor->samples[i].rtt_ms;
            rtt_count++;
        }
        total++;
    }

    if (rtt_count > 0) {
        monitor->stats.rtt_min_ms = rtt_min;
        monitor->stats.rtt_max_ms = rtt_max;
        monitor->stats.rtt_avg_ms = (uint32_t)(rtt_sum / rtt_count);
        if (rtt_count > 0) {
            monitor->stats.rtt_last_ms = monitor->samples[(monitor->sample_index + LINK_MONITOR_MAX_SAMPLES - 1) % LINK_MONITOR_MAX_SAMPLES].rtt_ms;
        }
    }

    if (total > 0) {
        monitor->stats.packet_loss_rate = (float)(total - received) / (float)total * 100.0f;
    }

    int32_t rtt_score = 100;
    if (monitor->stats.rtt_avg_ms > 100) {
        rtt_score = 100 - (monitor->stats.rtt_avg_ms - 100) / 10;
        if (rtt_score < 0) rtt_score = 0;
    }

    int32_t loss_score = 100;
    if (monitor->stats.packet_loss_rate > 0) {
        loss_score = 100 - (int32_t)monitor->stats.packet_loss_rate * 5;
        if (loss_score < 0) loss_score = 0;
    }

    int32_t signal_score = monitor->stats.signal_quality;
    if (monitor->stats.signal_strength < -100) {
        signal_score = 0;
    } else if (monitor->stats.signal_strength > -50) {
        signal_score = 100;
    } else {
        signal_score = (monitor->stats.signal_strength + 100) * 2;
    }

    monitor->stats.link_quality_score = (rtt_score * 4 + loss_score * 4 + signal_score * 2) / 10;
    if (monitor->stats.link_quality_score > LINK_QUALITY_MAX) {
        monitor->stats.link_quality_score = LINK_QUALITY_MAX;
    }
    if (monitor->stats.link_quality_score < LINK_QUALITY_MIN) {
        monitor->stats.link_quality_score = LINK_QUALITY_MIN;
    }
}

int link_monitor_get_stats(const link_monitor_t* monitor, link_stats_t* stats)
{
    if (!monitor || !stats) return -1;
    memcpy(stats, &monitor->stats, sizeof(link_stats_t));
    return 0;
}

int32_t link_monitor_get_quality_score(const link_monitor_t* monitor)
{
    return monitor ? monitor->stats.link_quality_score : LINK_QUALITY_MIN;
}

bool link_monitor_is_monitoring(const link_monitor_t* monitor)
{
    return monitor ? monitor->monitoring : false;
}

int link_monitor_reset(link_monitor_t* monitor)
{
    if (!monitor) return -1;
    memset(monitor->samples, 0, sizeof(monitor->samples));
    monitor->sample_index = 0;
    monitor->sample_count = 0;
    monitor->ping_seq = 0;
    monitor->pending_ping_count = 0;
    monitor->last_ping_ms = 0;
    monitor->bytes_sent = 0;
    monitor->bytes_received = 0;
    monitor->last_stats_update = 0;
    memset(&monitor->stats, 0, sizeof(monitor->stats));
    monitor->stats.link_quality_score = LINK_QUALITY_MAX;
    monitor->stats.signal_strength = -50;
    monitor->stats.signal_quality = 100;
    return 0;
}
