#ifndef CRYPTO_KEY_EXCHANGE_H
#define CRYPTO_KEY_EXCHANGE_H

#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>

#define ECDH_P256_PRIV_KEY_SIZE  32
#define ECDH_P256_PUB_KEY_SIZE   64
#define ECDH_SHARED_SECRET_SIZE  32

typedef enum {
    CURVE_P256 = 0,
    CURVE_SM2 = 1
} ec_curve_t;

typedef struct {
    ec_curve_t curve;
    uint8_t private_key[ECDH_P256_PRIV_KEY_SIZE];
    uint8_t public_key[ECDH_P256_PUB_KEY_SIZE];
    bool has_private;
    bool has_public;
} ecdh_ctx_t;

int ecdh_init(ecdh_ctx_t* ctx, ec_curve_t curve);
int ecdh_generate_keys(ecdh_ctx_t* ctx);
int ecdh_set_private_key(ecdh_ctx_t* ctx, const uint8_t* priv_key, size_t len);
int ecdh_set_public_key(ecdh_ctx_t* ctx, const uint8_t* pub_key, size_t len);
int ecdh_get_public_key(const ecdh_ctx_t* ctx, uint8_t* pub_key, size_t* len);
int ecdh_compute_shared_secret(const ecdh_ctx_t* ctx, const uint8_t* peer_pub_key,
                               size_t peer_pub_key_len, uint8_t* secret, size_t* secret_len);
void ecdh_cleanup(ecdh_ctx_t* ctx);

#endif
