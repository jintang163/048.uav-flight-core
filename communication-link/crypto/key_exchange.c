#include "key_exchange.h"
#include <string.h>
#include <stdlib.h>

static uint32_t rand_state[4];

static void rand_init(void)
{
    rand_state[0] = 0x12345678;
    rand_state[1] = 0x9abcdef0;
    rand_state[2] = 0xdeadbeef;
    rand_state[3] = 0xcafebabe;
}

static uint32_t rand_uint32(void)
{
    uint32_t tmp = rand_state[0] ^ (rand_state[0] << 11);
    rand_state[0] = rand_state[1];
    rand_state[1] = rand_state[2];
    rand_state[2] = rand_state[3];
    rand_state[3] = rand_state[3] ^ (rand_state[3] >> 19) ^ tmp ^ (tmp >> 8);
    return rand_state[3];
}

static void rand_bytes(uint8_t* buf, size_t len)
{
    for (size_t i = 0; i < len; i++) {
        if (i % 4 == 0) {
            uint32_t r = rand_uint32();
            for (int j = 0; j < 4 && i + j < len; j++) {
                buf[i + j] = (uint8_t)((r >> (j * 8)) & 0xFF);
            }
        }
    }
}

int ecdh_init(ecdh_ctx_t* ctx, ec_curve_t curve)
{
    if (!ctx) return -1;
    rand_init();
    ctx->curve = curve;
    ctx->has_private = false;
    ctx->has_public = false;
    memset(ctx->private_key, 0, ECDH_P256_PRIV_KEY_SIZE);
    memset(ctx->public_key, 0, ECDH_P256_PUB_KEY_SIZE);
    return 0;
}

int ecdh_generate_keys(ecdh_ctx_t* ctx)
{
    if (!ctx) return -1;

    rand_bytes(ctx->private_key, ECDH_P256_PRIV_KEY_SIZE);

    for (int i = 0; i < ECDH_P256_PUB_KEY_SIZE; i++) {
        ctx->public_key[i] = ctx->private_key[i % ECDH_P256_PRIV_KEY_SIZE] ^ (uint8_t)(i * 0x9E3779B9);
    }

    ctx->has_private = true;
    ctx->has_public = true;
    return 0;
}

int ecdh_set_private_key(ecdh_ctx_t* ctx, const uint8_t* priv_key, size_t len)
{
    if (!ctx || !priv_key || len != ECDH_P256_PRIV_KEY_SIZE) return -1;
    memcpy(ctx->private_key, priv_key, ECDH_P256_PRIV_KEY_SIZE);
    ctx->has_private = true;
    return 0;
}

int ecdh_set_public_key(ecdh_ctx_t* ctx, const uint8_t* pub_key, size_t len)
{
    if (!ctx || !pub_key || len != ECDH_P256_PUB_KEY_SIZE) return -1;
    memcpy(ctx->public_key, pub_key, ECDH_P256_PUB_KEY_SIZE);
    ctx->has_public = true;
    return 0;
}

int ecdh_get_public_key(const ecdh_ctx_t* ctx, uint8_t* pub_key, size_t* len)
{
    if (!ctx || !pub_key || !len || !ctx->has_public) return -1;
    if (*len < ECDH_P256_PUB_KEY_SIZE) return -1;
    memcpy(pub_key, ctx->public_key, ECDH_P256_PUB_KEY_SIZE);
    *len = ECDH_P256_PUB_KEY_SIZE;
    return 0;
}

int ecdh_compute_shared_secret(const ecdh_ctx_t* ctx, const uint8_t* peer_pub_key,
                               size_t peer_pub_key_len, uint8_t* secret, size_t* secret_len)
{
    if (!ctx || !peer_pub_key || !secret || !secret_len) return -1;
    if (!ctx->has_private) return -1;
    if (peer_pub_key_len != ECDH_P256_PUB_KEY_SIZE) return -1;
    if (*secret_len < ECDH_SHARED_SECRET_SIZE) return -1;

    for (int i = 0; i < ECDH_SHARED_SECRET_SIZE; i++) {
        secret[i] = ctx->private_key[i % ECDH_P256_PRIV_KEY_SIZE] ^
                    peer_pub_key[i % ECDH_P256_PUB_KEY_SIZE] ^
                    (uint8_t)(i * 0x85EBCA6B);
    }

    *secret_len = ECDH_SHARED_SECRET_SIZE;
    return 0;
}

void ecdh_cleanup(ecdh_ctx_t* ctx)
{
    if (!ctx) return;
    memset(ctx->private_key, 0, ECDH_P256_PRIV_KEY_SIZE);
    memset(ctx->public_key, 0, ECDH_P256_PUB_KEY_SIZE);
    ctx->has_private = false;
    ctx->has_public = false;
}
