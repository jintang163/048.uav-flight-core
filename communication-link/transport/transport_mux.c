#include "transport_mux.h"
#include "link_monitor.h"
#include <string.h>
#include <stdlib.h>

static int mux_transport_connect_impl(transport_t* transport)
{
    transport_mux_t* mux = (transport_mux_t*)transport;
    if (!mux) return -1;

    for (uint8_t i = 0; i < mux->link_count; i++) {
        if (mux->links[i].enabled && mux->links[i].transport) {
            transport_connect(mux->links[i].transport);
        }
    }

    mux->base.state = TRANSPORT_STATE_CONNECTED;
    if (mux->base.on_state_change) {
        mux->base.on_state_change(transport, TRANSPORT_STATE_CONNECTED, mux->base.user_data);
    }
    return 0;
}

static int mux_transport_disconnect_impl(transport_t* transport)
{
    transport_mux_t* mux = (transport_mux_t*)transport;
    if (!mux) return -1;

    for (uint8_t i = 0; i < mux->link_count; i++) {
        if (mux->links[i].transport) {
            transport_disconnect(mux->links[i].transport);
        }
    }

    mux->base.state = TRANSPORT_STATE_DISCONNECTED;
    if (mux->base.on_state_change) {
        mux->base.on_state_change(transport, TRANSPORT_STATE_DISCONNECTED, mux->base.user_data);
    }
    return 0;
}

static int mux_select_link(transport_mux_t* mux)
{
    if (!mux || mux->link_count == 0) return -1;

    int32_t best_link = -1;
    int32_t best_score = -1;

    for (uint8_t i = 0; i < mux->link_count; i++) {
        if (!mux->links[i].enabled) continue;
        if (mux->links[i].status == MUX_LINK_STATUS_FAILED) continue;

        int32_t score = mux->links[i].quality_score + mux->links[i].priority * 10;
        if (score > best_score) {
            best_score = score;
            best_link = i;
        }
    }

    return best_link;
}

static int mux_transport_send_impl(transport_t* transport, const uint8_t* data, size_t len, uint32_t timeout_ms)
{
    transport_mux_t* mux = (transport_mux_t*)transport;
    if (!mux || !data || len == 0) return -1;

    int32_t link_index;
    int ret = -1;

    switch (mux->mode) {
    case MUX_MODE_PRIMARY:
        if (mux->primary_link >= 0 && mux->primary_link < mux->link_count) {
            link_index = mux->primary_link;
            if (mux->links[link_index].status == MUX_LINK_STATUS_FAILED &&
                mux->failover_threshold > 0) {
                link_index = mux_select_link(mux);
            }
        } else {
            link_index = mux_select_link(mux);
        }
        break;

    case MUX_MODE_FAILOVER:
        if (mux->active_link >= 0 && mux->active_link < mux->link_count &&
            mux->links[mux->active_link].status != MUX_LINK_STATUS_FAILED) {
            link_index = mux->active_link;
        } else {
            link_index = mux_select_link(mux);
            mux->active_link = link_index;
        }
        break;

    case MUX_MODE_REDUNDANT:
        ret = 0;
        for (uint8_t i = 0; i < mux->link_count; i++) {
            if (mux->links[i].enabled && mux->links[i].status != MUX_LINK_STATUS_FAILED &&
                mux->links[i].transport) {
                int r = transport_send(mux->links[i].transport, data, len, timeout_ms);
                if (r > 0) {
                    mux->links[i].bytes_sent += (uint64_t)r;
                    ret = r;
                } else if (r < 0) {
                    mux->links[i].failures++;
                    if (mux->links[i].failures > mux->failover_threshold) {
                        mux->links[i].status = MUX_LINK_STATUS_FAILED;
                    }
                }
            }
        }
        return ret;

    case MUX_MODE_LOAD_BALANCE:
    default:
        link_index = mux_select_link(mux);
        break;
    }

    if (link_index < 0 || link_index >= mux->link_count) return -1;
    if (!mux->links[link_index].enabled || !mux->links[link_index].transport) return -1;

    ret = transport_send(mux->links[link_index].transport, data, len, timeout_ms);
    if (ret > 0) {
        mux->links[link_index].bytes_sent += (uint64_t)ret;
        mux->links[link_index].last_used_ms = 0;
        mux->active_link = link_index;
    } else if (ret < 0) {
        mux->links[link_index].failures++;
        if (mux->links[link_index].failures > mux->failover_threshold) {
            mux->links[link_index].status = MUX_LINK_STATUS_FAILED;
        }
    }

    return ret;
}

static int mux_transport_recv_impl(transport_t* transport, uint8_t* data, size_t len, size_t* recv_len, uint32_t timeout_ms)
{
    transport_mux_t* mux = (transport_mux_t*)transport;
    if (!mux || !data || len == 0 || !recv_len) return -1;

    for (uint8_t i = 0; i < mux->link_count; i++) {
        if (!mux->links[i].enabled || !mux->links[i].transport) continue;
        if (mux->links[i].status == MUX_LINK_STATUS_FAILED) continue;

        int ret = transport_recv(mux->links[i].transport, data, len, recv_len, 0);
        if (ret == 0 && *recv_len > 0) {
            mux->links[i].bytes_received += (uint64_t)*recv_len;
            mux->links[i].status = MUX_LINK_STATUS_ACTIVE;
            mux->links[i].failures = 0;
            mux->links[i].last_used_ms = 0;
            return 0;
        }
    }

    if (mux->active_link >= 0 && mux->active_link < mux->link_count &&
        mux->links[mux->active_link].enabled &&
        mux->links[mux->active_link].transport) {
        int ret = transport_recv(mux->links[mux->active_link].transport, data, len, recv_len, timeout_ms);
        if (ret == 0 && *recv_len > 0) {
            mux->links[mux->active_link].bytes_received += (uint64_t)*recv_len;
            mux->links[mux->active_link].status = MUX_LINK_STATUS_ACTIVE;
            mux->links[mux->active_link].failures = 0;
        }
        return ret;
    }

    *recv_len = 0;
    return 0;
}

int transport_mux_init(transport_mux_t* mux, mux_mode_t mode)
{
    if (!mux) return -1;
    memset(mux, 0, sizeof(transport_mux_t));
    transport_init(&mux->base, TRANSPORT_TYPE_MUX);
    mux->base.connect = mux_transport_connect_impl;
    mux->base.disconnect = mux_transport_disconnect_impl;
    mux->base.send = mux_transport_send_impl;
    mux->base.recv = mux_transport_recv_impl;
    mux->mode = mode;
    mux->link_count = 0;
    mux->primary_link = -1;
    mux->active_link = -1;
    mux->failover_threshold = 5;
    mux->failback_delay_ms = 30000;
    mux->last_failover_ms = 0;
    return 0;
}

int transport_mux_add_link(transport_mux_t* mux, transport_t* transport, uint32_t priority)
{
    if (!mux || !transport) return -1;
    if (mux->link_count >= MUX_MAX_LINKS) return -1;

    mux->links[mux->link_count].transport = transport;
    mux->links[mux->link_count].priority = priority;
    mux->links[mux->link_count].status = MUX_LINK_STATUS_UNKNOWN;
    mux->links[mux->link_count].quality_score = 50;
    mux->links[mux->link_count].enabled = true;
    mux->links[mux->link_count].bytes_sent = 0;
    mux->links[mux->link_count].bytes_received = 0;
    mux->links[mux->link_count].failures = 0;
    mux->links[mux->link_count].last_used_ms = 0;

    if (mux->primary_link < 0) {
        mux->primary_link = mux->link_count;
        mux->active_link = mux->link_count;
    }

    mux->link_count++;
    return 0;
}

int transport_mux_remove_link(transport_mux_t* mux, uint8_t index)
{
    if (!mux || index >= mux->link_count) return -1;

    for (uint8_t i = index; i < mux->link_count - 1; i++) {
        memcpy(&mux->links[i], &mux->links[i + 1], sizeof(mux_link_t));
    }
    mux->link_count--;

    if (mux->primary_link == index) {
        mux->primary_link = mux->link_count > 0 ? 0 : -1;
    } else if (mux->primary_link > index) {
        mux->primary_link--;
    }

    if (mux->active_link == index) {
        mux->active_link = mux->link_count > 0 ? 0 : -1;
    } else if (mux->active_link > index) {
        mux->active_link--;
    }

    return 0;
}

int transport_mux_set_primary(transport_mux_t* mux, uint8_t index)
{
    if (!mux || index >= mux->link_count) return -1;
    mux->primary_link = index;
    return 0;
}

int transport_mux_set_mode(transport_mux_t* mux, mux_mode_t mode)
{
    if (!mux) return -1;
    mux->mode = mode;
    return 0;
}

int transport_mux_enable_link(transport_mux_t* mux, uint8_t index, bool enable)
{
    if (!mux || index >= mux->link_count) return -1;
    mux->links[index].enabled = enable;
    if (!enable && mux->links[index].transport) {
        transport_disconnect(mux->links[index].transport);
    } else if (enable && mux->links[index].transport) {
        transport_connect(mux->links[index].transport);
    }
    return 0;
}

int transport_mux_get_link_status(const transport_mux_t* mux, uint8_t index, mux_link_status_t* status)
{
    if (!mux || index >= mux->link_count || !status) return -1;
    *status = mux->links[index].status;
    return 0;
}

int transport_mux_get_link_quality(const transport_mux_t* mux, uint8_t index, int32_t* quality)
{
    if (!mux || index >= mux->link_count || !quality) return -1;
    *quality = mux->links[index].quality_score;
    return 0;
}

int32_t transport_mux_get_active_link(const transport_mux_t* mux)
{
    return mux ? mux->active_link : -1;
}

int transport_mux_set_failover_threshold(transport_mux_t* mux, uint32_t threshold)
{
    if (!mux) return -1;
    mux->failover_threshold = threshold;
    return 0;
}

int transport_mux_set_failback_delay(transport_mux_t* mux, uint32_t delay_ms)
{
    if (!mux) return -1;
    mux->failback_delay_ms = delay_ms;
    return 0;
}

int transport_mux_manual_switch(transport_mux_t* mux, uint8_t index)
{
    if (!mux || index >= mux->link_count) return -1;
    if (!mux->links[index].enabled) return -1;
    if (mux->links[index].status == MUX_LINK_STATUS_FAILED) return -1;
    mux->active_link = index;
    mux->last_failover_ms = 0;
    return 0;
}
