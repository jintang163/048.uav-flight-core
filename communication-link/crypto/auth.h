#ifndef CRYPTO_AUTH_H
#define CRYPTO_AUTH_H

#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>

#define SM3_BLOCK_SIZE    64
#define SM3_HASH_SIZE     32
#define HMAC_SM3_SIZE     32

typedef struct {
    uint32_t state[8];
    uint64_t count;
    uint8_t buffer[SM3_BLOCK_SIZE];
} sm3_ctx_t;

typedef struct {
    sm3_ctx_t inner_ctx;
    sm3_ctx_t outer_ctx;
    uint8_t key[SM3_BLOCK_SIZE];
    bool key_set;
} hmac_sm3_ctx_t;

void sm3_init(sm3_ctx_t* ctx);
void sm3_update(sm3_ctx_t* ctx, const uint8_t* data, size_t len);
void sm3_final(sm3_ctx_t* ctx, uint8_t* hash);
void sm3_hash(const uint8_t* data, size_t len, uint8_t* hash);

void hmac_sm3_init(hmac_sm3_ctx_t* ctx, const uint8_t* key, size_t key_len);
void hmac_sm3_update(hmac_sm3_ctx_t* ctx, const uint8_t* data, size_t len);
void hmac_sm3_final(hmac_sm3_ctx_t* ctx, uint8_t* hmac);
void hmac_sm3(const uint8_t* key, size_t key_len,
              const uint8_t* data, size_t data_len,
              uint8_t* hmac);

bool hmac_sm3_verify(const uint8_t* key, size_t key_len,
                     const uint8_t* data, size_t data_len,
                     const uint8_t* expected_hmac, size_t hmac_len);

#endif
