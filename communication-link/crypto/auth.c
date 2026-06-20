#include "auth.h"
#include <string.h>

static const uint32_t sm3_initial_value[8] = {
    0x7380166f, 0x4914b2b9, 0x172442d7, 0xda8a0600,
    0xa96f30bc, 0x163138aa, 0xe38dee4d, 0xb0fb0e4e
};

static inline uint32_t sm3_rotl(uint32_t x, int n)
{
    return (x << n) | (x >> (32 - n));
}

static inline uint32_t sm3_p0(uint32_t x)
{
    return x ^ sm3_rotl(x, 9) ^ sm3_rotl(x, 17);
}

static inline uint32_t sm3_p1(uint32_t x)
{
    return x ^ sm3_rotl(x, 15) ^ sm3_rotl(x, 23);
}

static inline uint32_t sm3_ff(uint32_t x, uint32_t y, uint32_t z, int j)
{
    if (j < 16) {
        return x ^ y ^ z;
    } else {
        return (x & y) | (x & z) | (y & z);
    }
}

static inline uint32_t sm3_gg(uint32_t x, uint32_t y, uint32_t z, int j)
{
    if (j < 16) {
        return x ^ y ^ z;
    } else {
        return (x & y) | (~x & z);
    }
}

static void sm3_process_block(sm3_ctx_t* ctx, const uint8_t* block)
{
    uint32_t w[68];
    uint32_t w_[64];
    uint32_t a, b, c, d, e, f, g, h;
    uint32_t ss1, ss2, tt1, tt2;

    for (int i = 0; i < 16; i++) {
        w[i] = ((uint32_t)block[i * 4] << 24) |
               ((uint32_t)block[i * 4 + 1] << 16) |
               ((uint32_t)block[i * 4 + 2] << 8) |
               (uint32_t)block[i * 4 + 3];
    }

    for (int i = 16; i < 68; i++) {
        w[i] = sm3_p1(w[i - 16] ^ w[i - 9] ^ sm3_rotl(w[i - 3], 15)) ^
               sm3_rotl(w[i - 13], 7) ^ w[i - 6];
    }

    for (int i = 0; i < 64; i++) {
        w_[i] = w[i] ^ w[i + 4];
    }

    a = ctx->state[0];
    b = ctx->state[1];
    c = ctx->state[2];
    d = ctx->state[3];
    e = ctx->state[4];
    f = ctx->state[5];
    g = ctx->state[6];
    h = ctx->state[7];

    for (int j = 0; j < 64; j++) {
        uint32_t tj = (j < 16) ? 0x79cc4519 : 0x7a879d8a;

        ss1 = sm3_rotl((sm3_rotl(a, 12) + e + sm3_rotl(tj, j % 32)), 7);
        ss2 = ss1 ^ sm3_rotl(a, 12);
        tt1 = sm3_ff(a, b, c, j) + d + ss2 + w_[j];
        tt2 = sm3_gg(e, f, g, j) + h + ss1 + w[j];

        d = c;
        c = sm3_rotl(b, 9);
        b = a;
        a = tt1;
        h = g;
        g = sm3_rotl(f, 19);
        f = e;
        e = sm3_p0(tt2);
    }

    ctx->state[0] ^= a;
    ctx->state[1] ^= b;
    ctx->state[2] ^= c;
    ctx->state[3] ^= d;
    ctx->state[4] ^= e;
    ctx->state[5] ^= f;
    ctx->state[6] ^= g;
    ctx->state[7] ^= h;
}

void sm3_init(sm3_ctx_t* ctx)
{
    if (!ctx) return;
    memcpy(ctx->state, sm3_initial_value, sizeof(sm3_initial_value));
    ctx->count = 0;
    memset(ctx->buffer, 0, SM3_BLOCK_SIZE);
}

void sm3_update(sm3_ctx_t* ctx, const uint8_t* data, size_t len)
{
    if (!ctx || !data || len == 0) return;

    size_t index = (size_t)(ctx->count & 0x3F);
    ctx->count += len;

    if (index + len >= SM3_BLOCK_SIZE) {
        if (index > 0) {
            size_t copy_len = SM3_BLOCK_SIZE - index;
            memcpy(ctx->buffer + index, data, copy_len);
            sm3_process_block(ctx, ctx->buffer);
            data += copy_len;
            len -= copy_len;
            index = 0;
        }

        while (len >= SM3_BLOCK_SIZE) {
            sm3_process_block(ctx, data);
            data += SM3_BLOCK_SIZE;
            len -= SM3_BLOCK_SIZE;
        }
    }

    if (len > 0) {
        memcpy(ctx->buffer + index, data, len);
    }
}

void sm3_final(sm3_ctx_t* ctx, uint8_t* hash)
{
    if (!ctx || !hash) return;

    uint64_t bit_count = ctx->count * 8;
    size_t index = (size_t)(ctx->count & 0x3F);
    size_t pad_len = (index < 56) ? (56 - index) : (120 - index);

    ctx->buffer[index] = 0x80;
    memset(ctx->buffer + index + 1, 0, pad_len - 1);

    for (int i = 7; i >= 0; i--) {
        ctx->buffer[index + pad_len + i] = (uint8_t)(bit_count & 0xFF);
        bit_count >>= 8;
    }

    sm3_process_block(ctx, ctx->buffer);
    if (index >= 56) {
        memset(ctx->buffer, 0, 56);
        bit_count = ctx->count * 8;
        for (int i = 7; i >= 0; i--) {
            ctx->buffer[56 + i] = (uint8_t)(bit_count & 0xFF);
            bit_count >>= 8;
        }
        sm3_process_block(ctx, ctx->buffer);
    }

    for (int i = 0; i < 8; i++) {
        hash[i * 4] = (uint8_t)(ctx->state[i] >> 24);
        hash[i * 4 + 1] = (uint8_t)(ctx->state[i] >> 16);
        hash[i * 4 + 2] = (uint8_t)(ctx->state[i] >> 8);
        hash[i * 4 + 3] = (uint8_t)ctx->state[i];
    }
}

void sm3_hash(const uint8_t* data, size_t len, uint8_t* hash)
{
    sm3_ctx_t ctx;
    sm3_init(&ctx);
    sm3_update(&ctx, data, len);
    sm3_final(&ctx, hash);
}

void hmac_sm3_init(hmac_sm3_ctx_t* ctx, const uint8_t* key, size_t key_len)
{
    if (!ctx || !key || key_len == 0) return;

    memset(ctx->key, 0, SM3_BLOCK_SIZE);
    if (key_len <= SM3_BLOCK_SIZE) {
        memcpy(ctx->key, key, key_len);
    } else {
        sm3_hash(key, key_len, ctx->key);
    }

    uint8_t ipad[SM3_BLOCK_SIZE];
    for (int i = 0; i < SM3_BLOCK_SIZE; i++) {
        ipad[i] = ctx->key[i] ^ 0x36;
    }

    sm3_init(&ctx->inner_ctx);
    sm3_update(&ctx->inner_ctx, ipad, SM3_BLOCK_SIZE);

    uint8_t opad[SM3_BLOCK_SIZE];
    for (int i = 0; i < SM3_BLOCK_SIZE; i++) {
        opad[i] = ctx->key[i] ^ 0x5C;
    }

    sm3_init(&ctx->outer_ctx);
    sm3_update(&ctx->outer_ctx, opad, SM3_BLOCK_SIZE);

    ctx->key_set = true;
}

void hmac_sm3_update(hmac_sm3_ctx_t* ctx, const uint8_t* data, size_t len)
{
    if (!ctx || !ctx->key_set || !data || len == 0) return;
    sm3_update(&ctx->inner_ctx, data, len);
}

void hmac_sm3_final(hmac_sm3_ctx_t* ctx, uint8_t* hmac)
{
    if (!ctx || !ctx->key_set || !hmac) return;

    uint8_t inner_hash[SM3_HASH_SIZE];
    sm3_final(&ctx->inner_ctx, inner_hash);

    sm3_update(&ctx->outer_ctx, inner_hash, SM3_HASH_SIZE);
    sm3_final(&ctx->outer_ctx, hmac);

    ctx->key_set = false;
}

void hmac_sm3(const uint8_t* key, size_t key_len,
              const uint8_t* data, size_t data_len,
              uint8_t* hmac)
{
    hmac_sm3_ctx_t ctx;
    hmac_sm3_init(&ctx, key, key_len);
    hmac_sm3_update(&ctx, data, data_len);
    hmac_sm3_final(&ctx, hmac);
}

bool hmac_sm3_verify(const uint8_t* key, size_t key_len,
                     const uint8_t* data, size_t data_len,
                     const uint8_t* expected_hmac, size_t hmac_len)
{
    if (hmac_len != HMAC_SM3_SIZE) return false;

    uint8_t computed_hmac[HMAC_SM3_SIZE];
    hmac_sm3(key, key_len, data, data_len, computed_hmac);

    uint8_t diff = 0;
    for (size_t i = 0; i < HMAC_SM3_SIZE; i++) {
        diff |= computed_hmac[i] ^ expected_hmac[i];
    }
    return diff == 0;
}
