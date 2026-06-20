#include "fragment.h"
#include <string.h>
#include <stdlib.h>

int fragment_manager_init(fragment_manager_t* manager, uint8_t mtu, uint32_t timeout_ms)
{
    if (!manager || mtu < FRAGMENT_HEADER_SIZE + 1) return -1;
    memset(manager, 0, sizeof(fragment_manager_t));
    manager->next_packet_id = 1;
    manager->mtu = mtu;
    manager->timeout_ms = timeout_ms;
    manager->packets_fragmented = 0;
    manager->packets_reassembled = 0;
    manager->fragments_sent = 0;
    manager->fragments_received = 0;
    manager->error_count = 0;
    return 0;
}

int fragment_manager_set_mtu(fragment_manager_t* manager, uint8_t mtu)
{
    if (!manager || mtu < FRAGMENT_HEADER_SIZE + 1) return -1;
    manager->mtu = mtu;
    return 0;
}

int fragment_manager_set_timeout(fragment_manager_t* manager, uint32_t timeout_ms)
{
    if (!manager) return -1;
    manager->timeout_ms = timeout_ms;
    return 0;
}

int fragment_manager_split(fragment_manager_t* manager, const uint8_t* data, size_t len,
                           uint8_t* fragments, size_t* fragment_count, size_t max_fragments)
{
    if (!manager || !data || len == 0 || !fragments || !fragment_count) return -1;
    if (len > FRAGMENT_MAX_DATA_SIZE) return -1;

    uint8_t max_data_per_fragment = manager->mtu - FRAGMENT_HEADER_SIZE;
    uint8_t total_fragments = (uint8_t)((len + max_data_per_fragment - 1) / max_data_per_fragment);

    if (total_fragments > FRAGMENT_MAX_FRAGMENTS) return -1;
    if (total_fragments > max_fragments) return -1;

    uint16_t packet_id = manager->next_packet_id++;
    if (manager->next_packet_id == 0) {
        manager->next_packet_id = 1;
    }

    size_t remaining = len;
    size_t offset = 0;

    for (uint8_t i = 0; i < total_fragments; i++) {
        fragment_packet_t* pkt = (fragment_packet_t*)(fragments + i * manager->mtu);
        size_t frag_data_len = (remaining > max_data_per_fragment) ? max_data_per_fragment : remaining;

        pkt->packet_id = packet_id;
        pkt->fragment_index = i;
        pkt->fragment_total = total_fragments;
        memcpy(pkt->data, data + offset, frag_data_len);

        offset += frag_data_len;
        remaining -= frag_data_len;
    }

    *fragment_count = total_fragments;
    manager->packets_fragmented++;
    manager->fragments_sent += total_fragments;
    return 0;
}

int fragment_reassembly_init(fragment_reassembly_t* reassembly, uint8_t mtu, uint32_t timeout_ms)
{
    if (!reassembly || mtu < FRAGMENT_HEADER_SIZE + 1) return -1;
    memset(reassembly, 0, sizeof(fragment_reassembly_t));
    reassembly->mtu = mtu;
    reassembly->timeout_ms = timeout_ms;
    reassembly->status = FRAGMENT_STATUS_INIT;
    return 0;
}

static bool fragment_test_bit(const uint8_t* mask, uint8_t bit)
{
    return (mask[bit / 8] & (1 << (bit % 8))) != 0;
}

static void fragment_set_bit(uint8_t* mask, uint8_t bit)
{
    mask[bit / 8] |= (1 << (bit % 8));
}

int fragment_reassembly_add(fragment_reassembly_t* reassembly, const fragment_packet_t* fragment)
{
    if (!reassembly || !fragment) return -1;
    if (fragment->fragment_total > FRAGMENT_MAX_FRAGMENTS) return -1;
    if (fragment->fragment_index >= fragment->fragment_total) return -1;

    uint8_t max_data_per_fragment = reassembly->mtu - FRAGMENT_HEADER_SIZE;

    if (reassembly->status == FRAGMENT_STATUS_INIT) {
        reassembly->packet_id = fragment->packet_id;
        reassembly->fragment_total = fragment->fragment_total;
        reassembly->total_size = 0;
        reassembly->received_size = 0;
        memset(reassembly->fragment_mask, 0, sizeof(reassembly->fragment_mask));
        reassembly->first_fragment_time = 0;
        reassembly->last_fragment_time = 0;
        reassembly->status = FRAGMENT_STATUS_IN_PROGRESS;
    } else if (reassembly->status == FRAGMENT_STATUS_IN_PROGRESS) {
        if (reassembly->packet_id != fragment->packet_id) {
            return -1;
        }
        if (reassembly->fragment_total != fragment->fragment_total) {
            return -1;
        }
    } else {
        return -1;
    }

    if (fragment_test_bit(reassembly->fragment_mask, fragment->fragment_index)) {
        return 0;
    }

    size_t data_offset = fragment->fragment_index * max_data_per_fragment;
    size_t data_len = max_data_per_fragment;
    if (fragment->fragment_index == fragment->fragment_total - 1) {
        if (reassembly->total_size == 0) {
            return -1;
        }
        data_len = reassembly->total_size - data_offset;
    }

    memcpy(reassembly->buffer + data_offset, fragment->data, data_len);
    fragment_set_bit(reassembly->fragment_mask, fragment->fragment_index);
    reassembly->received_size += data_len;
    reassembly->last_fragment_time = 0;

    uint8_t received_count = 0;
    for (uint8_t i = 0; i < reassembly->fragment_total; i++) {
        if (fragment_test_bit(reassembly->fragment_mask, i)) {
            received_count++;
        }
    }

    if (received_count == reassembly->fragment_total) {
        reassembly->status = FRAGMENT_STATUS_COMPLETE;
    }

    return 0;
}

int fragment_reassembly_is_complete(const fragment_reassembly_t* reassembly)
{
    if (!reassembly) return -1;
    return (reassembly->status == FRAGMENT_STATUS_COMPLETE) ? 1 : 0;
}

int fragment_reassembly_get_data(const fragment_reassembly_t* reassembly, uint8_t* data, size_t* len)
{
    if (!reassembly || !data || !len) return -1;
    if (reassembly->status != FRAGMENT_STATUS_COMPLETE) return -1;
    if (*len < reassembly->received_size) return -1;

    memcpy(data, reassembly->buffer, reassembly->received_size);
    *len = reassembly->received_size;
    return 0;
}

fragment_status_t fragment_reassembly_get_status(const fragment_reassembly_t* reassembly)
{
    return reassembly ? reassembly->status : FRAGMENT_STATUS_ERROR;
}

int fragment_reassembly_reset(fragment_reassembly_t* reassembly)
{
    if (!reassembly) return -1;
    memset(reassembly->buffer, 0, FRAGMENT_MAX_DATA_SIZE);
    reassembly->total_size = 0;
    reassembly->received_size = 0;
    reassembly->packet_id = 0;
    reassembly->fragment_total = 0;
    memset(reassembly->fragment_mask, 0, sizeof(reassembly->fragment_mask));
    reassembly->first_fragment_time = 0;
    reassembly->last_fragment_time = 0;
    reassembly->status = FRAGMENT_STATUS_INIT;
    return 0;
}

bool fragment_manager_is_fragment_packet(const uint8_t* data, size_t len)
{
    if (!data || len < FRAGMENT_HEADER_SIZE) return false;
    const fragment_packet_t* pkt = (const fragment_packet_t*)data;
    if (pkt->fragment_total < 2) return false;
    if (pkt->fragment_index >= pkt->fragment_total) return false;
    return true;
}
