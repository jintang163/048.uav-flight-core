#include "sm4.h"
#include <string.h>
#include <stdio.h>

static const uint8_t sm4_sbox[256] = {
    0xd6, 0x90, 0xe9, 0xfe, 0xcc, 0xe1, 0x3d, 0xb7, 0x16, 0xb6, 0x14, 0xc2, 0x28, 0xfb, 0x2c, 0x05,
    0x2b, 0x67, 0x9a, 0x76, 0x2a, 0xbe, 0x04, 0xc3, 0xaa, 0x44, 0x13, 0x26, 0x49, 0x86, 0x06, 0x99,
    0x9c, 0x42, 0x50, 0xf4, 0x91, 0xef, 0x98, 0x7a, 0x33, 0x54, 0x0b, 0x43, 0xed, 0xcf, 0xac, 0x62,
    0xe4, 0xb3, 0x1c, 0xa9, 0xc9, 0x08, 0xe8, 0x95, 0x80, 0xdf, 0x94, 0xfa, 0x75, 0x8f, 0x3f, 0xa6,
    0x47, 0x07, 0xa7, 0xfc, 0xf3, 0x73, 0x17, 0xba, 0x83, 0x59, 0x3c, 0x19, 0xe6, 0x85, 0x4f, 0xa8,
    0x68, 0x6b, 0x81, 0xb2, 0x71, 0x64, 0xda, 0x8b, 0xf8, 0xeb, 0x0f, 0x4b, 0x70, 0x56, 0x9d, 0x35,
    0x1e, 0x24, 0x0e, 0x5e, 0x63, 0x58, 0xd1, 0xa2, 0x25, 0x22, 0x7c, 0x3b, 0x01, 0x21, 0x78, 0x87,
    0xd4, 0x00, 0x46, 0x57, 0x9f, 0xd3, 0x27, 0x52, 0x4c, 0x36, 0x02, 0xe7, 0xa0, 0xc4, 0xc8, 0x9e,
    0xea, 0xbf, 0x8a, 0xd2, 0x40, 0xc7, 0x38, 0xb5, 0xa3, 0xf7, 0xf2, 0xce, 0xf9, 0x61, 0x15, 0xa1,
    0xe0, 0xae, 0x5d, 0xa4, 0x9b, 0x34, 0x1a, 0x55, 0xad, 0x93, 0x32, 0x30, 0xf5, 0x8c, 0xb1, 0xe3,
    0x1d, 0xf6, 0xe2, 0x2e, 0x82, 0x66, 0xca, 0x60, 0xc0, 0x29, 0x23, 0xab, 0x0d, 0x53, 0x4e, 0x6f,
    0xd5, 0xdb, 0x37, 0x45, 0xde, 0xfd, 0x8e, 0x2f, 0x03, 0xff, 0x6a, 0x72, 0x6d, 0x6c, 0x5b, 0x51,
    0x8d, 0x1b, 0xaf, 0x92, 0xbb, 0xdd, 0xbc, 0x7f, 0x11, 0xd9, 0x5c, 0x41, 0x1f, 0x10, 0x5a, 0xd8,
    0x0a, 0xc1, 0x31, 0x88, 0xa5, 0xcd, 0x7b, 0xbd, 0x2d, 0x74, 0xd0, 0x12, 0xb8, 0xe5, 0xb4, 0xb0,
    0x89, 0x69, 0x97, 0x4a, 0x0c, 0x96, 0x77, 0x7e, 0x65, 0xb9, 0xf1, 0x09, 0xc5, 0x6e, 0xc6, 0x84,
    0x18, 0xf0, 0x7d, 0xec, 0x3a, 0xdc, 0x4d, 0x20, 0x79, 0xe1, 0x5f, 0x3e, 0xd7, 0xcb, 0x39, 0x48
};

static const uint32_t sm4_fk[4] = {
    0xa3b1bac6, 0x56aa3350, 0x677d9197, 0xb27022dc
};

static const uint32_t sm4_ck[32] = {
    0x00070e15, 0x1c232a31, 0x383f464d, 0x545b6269,
    0x70777e85, 0x8c939aa1, 0xa8afb6bd, 0xc4cbd2d9,
    0xe0e7eef5, 0xfc030a11, 0x181f262d, 0x343b4249,
    0x50575e65, 0x6c737a81, 0x888f969d, 0xa4abb2b9,
    0xc0c7ced5, 0xdce3eaf1, 0xf8ff060d, 0x141b2229,
    0x30373e45, 0x4c535a61, 0x686f767d, 0x848b9299,
    0xa0a7aeb5, 0xbcc3cad1, 0xd8dfe6ed, 0xf4fb0209,
    0x10171e25, 0x2c333a41, 0x484f565d, 0x646b7279
};

static inline uint32_t sm4_sbox_substitute(uint32_t x)
{
    uint32_t result = 0;
    result |= (uint32_t)sm4_sbox[(x >> 24) & 0xFF] << 24;
    result |= (uint32_t)sm4_sbox[(x >> 16) & 0xFF] << 16;
    result |= (uint32_t)sm4_sbox[(x >> 8) & 0xFF] << 8;
    result |= (uint32_t)sm4_sbox[x & 0xFF];
    return result;
}

static inline uint32_t sm4_rotl32(uint32_t x, int n)
{
    return (x << n) | (x >> (32 - n));
}

static inline uint32_t sm4_tau(uint32_t x)
{
    return sm4_sbox_substitute(x);
}

static inline uint32_t sm4_L(uint32_t x)
{
    return x ^ sm4_rotl32(x, 2) ^ sm4_rotl32(x, 10) ^ sm4_rotl32(x, 18) ^ sm4_rotl32(x, 24);
}

static inline uint32_t sm4_L_prime(uint32_t x)
{
    return x ^ sm4_rotl32(x, 13) ^ sm4_rotl32(x, 23);
}

static inline uint32_t sm4_t(uint32_t x)
{
    return sm4_L(sm4_tau(x));
}

static inline uint32_t sm4_t_prime(uint32_t x)
{
    return sm4_L_prime(sm4_tau(x));
}

static void sm4_key_expansion(const uint8_t* key, uint32_t* rk)
{
    uint32_t k[4];
    k[0] = ((uint32_t)key[0] << 24) | ((uint32_t)key[1] << 16) | ((uint32_t)key[2] << 8) | key[3];
    k[1] = ((uint32_t)key[4] << 24) | ((uint32_t)key[5] << 16) | ((uint32_t)key[6] << 8) | key[7];
    k[2] = ((uint32_t)key[8] << 24) | ((uint32_t)key[9] << 16) | ((uint32_t)key[10] << 8) | key[11];
    k[3] = ((uint32_t)key[12] << 24) | ((uint32_t)key[13] << 16) | ((uint32_t)key[14] << 8) | key[15];

    k[0] ^= sm4_fk[0];
    k[1] ^= sm4_fk[1];
    k[2] ^= sm4_fk[2];
    k[3] ^= sm4_fk[3];

    for (int i = 0; i < SM4_ROUNDS; i++) {
        rk[i] = k[(i + 1) % 4] ^ sm4_t_prime(k[(i + 2) % 4] ^ k[(i + 3) % 4] ^ sm4_ck[i]);
        k[i % 4] = rk[i];
    }
}

static void sm4_block_encrypt(const uint32_t* rk, const uint8_t* input, uint8_t* output)
{
    uint32_t x[4];
    x[0] = ((uint32_t)input[0] << 24) | ((uint32_t)input[1] << 16) | ((uint32_t)input[2] << 8) | input[3];
    x[1] = ((uint32_t)input[4] << 24) | ((uint32_t)input[5] << 16) | ((uint32_t)input[6] << 8) | input[7];
    x[2] = ((uint32_t)input[8] << 24) | ((uint32_t)input[9] << 16) | ((uint32_t)input[10] << 8) | input[11];
    x[3] = ((uint32_t)input[12] << 24) | ((uint32_t)input[13] << 16) | ((uint32_t)input[14] << 8) | input[15];

    for (int i = 0; i < SM4_ROUNDS; i++) {
        uint32_t tmp = x[(i + 1) % 4] ^ x[(i + 2) % 4] ^ x[(i + 3) % 4] ^ rk[i];
        x[i % 4] ^= sm4_t(tmp);
    }

    uint32_t y[4] = {x[3], x[2], x[1], x[0]};

    output[0] = (uint8_t)(y[0] >> 24);
    output[1] = (uint8_t)(y[0] >> 16);
    output[2] = (uint8_t)(y[0] >> 8);
    output[3] = (uint8_t)y[0];
    output[4] = (uint8_t)(y[1] >> 24);
    output[5] = (uint8_t)(y[1] >> 16);
    output[6] = (uint8_t)(y[1] >> 8);
    output[7] = (uint8_t)y[1];
    output[8] = (uint8_t)(y[2] >> 24);
    output[9] = (uint8_t)(y[2] >> 16);
    output[10] = (uint8_t)(y[2] >> 8);
    output[11] = (uint8_t)y[2];
    output[12] = (uint8_t)(y[3] >> 24);
    output[13] = (uint8_t)(y[3] >> 16);
    output[14] = (uint8_t)(y[3] >> 8);
    output[15] = (uint8_t)y[3];
}

static void sm4_block_decrypt(const uint32_t* rk, const uint8_t* input, uint8_t* output)
{
    uint32_t rk_rev[SM4_ROUNDS];
    for (int i = 0; i < SM4_ROUNDS; i++) {
        rk_rev[i] = rk[SM4_ROUNDS - 1 - i];
    }
    sm4_block_encrypt(rk_rev, input, output);
}

int sm4_init(sm4_ctx_t* ctx, const uint8_t* key, size_t key_len,
             sm4_mode_t mode, sm4_direction_t direction)
{
    if (!ctx || !key || key_len != SM4_KEY_SIZE) {
        return -1;
    }

    sm4_key_expansion(key, ctx->rk);
    ctx->mode = mode;
    ctx->direction = direction;
    ctx->iv_set = false;
    memset(ctx->iv, 0, SM4_IV_SIZE);

    return 0;
}

int sm4_set_iv(sm4_ctx_t* ctx, const uint8_t* iv, size_t iv_len)
{
    if (!ctx || !iv || iv_len != SM4_IV_SIZE) {
        return -1;
    }

    memcpy(ctx->iv, iv, SM4_IV_SIZE);
    ctx->iv_set = true;

    return 0;
}

int sm4_pkcs7_pad(uint8_t* data, size_t data_len, size_t block_size)
{
    if (!data || block_size == 0 || block_size > 256) {
        return -1;
    }

    size_t pad_len = block_size - (data_len % block_size);
    for (size_t i = 0; i < pad_len; i++) {
        data[data_len + i] = (uint8_t)pad_len;
    }

    return (int)pad_len;
}

int sm4_pkcs7_unpad(const uint8_t* data, size_t data_len, size_t* output_len)
{
    if (!data || !output_len || data_len == 0) {
        return -1;
    }

    uint8_t pad_len = data[data_len - 1];
    if (pad_len == 0 || pad_len > data_len || pad_len > 16) {
        return -1;
    }

    for (size_t i = 0; i < pad_len; i++) {
        if (data[data_len - 1 - i] != pad_len) {
            return -1;
        }
    }

    *output_len = data_len - pad_len;
    return 0;
}

int sm4_ecb_encrypt(const uint8_t* key, const uint8_t* input, size_t input_len,
                    uint8_t* output, size_t* output_len)
{
    if (!key || !input || !output || !output_len) {
        return -1;
    }

    uint32_t rk[SM4_ROUNDS];
    sm4_key_expansion(key, rk);

    size_t blocks = input_len / SM4_BLOCK_SIZE;
    for (size_t i = 0; i < blocks; i++) {
        sm4_block_encrypt(rk, input + i * SM4_BLOCK_SIZE, output + i * SM4_BLOCK_SIZE);
    }

    *output_len = blocks * SM4_BLOCK_SIZE;
    return 0;
}

int sm4_ecb_decrypt(const uint8_t* key, const uint8_t* input, size_t input_len,
                    uint8_t* output, size_t* output_len)
{
    if (!key || !input || !output || !output_len || input_len % SM4_BLOCK_SIZE != 0) {
        return -1;
    }

    uint32_t rk[SM4_ROUNDS];
    sm4_key_expansion(key, rk);

    size_t blocks = input_len / SM4_BLOCK_SIZE;
    for (size_t i = 0; i < blocks; i++) {
        sm4_block_decrypt(rk, input + i * SM4_BLOCK_SIZE, output + i * SM4_BLOCK_SIZE);
    }

    *output_len = blocks * SM4_BLOCK_SIZE;
    return 0;
}

int sm4_cbc_encrypt(const uint8_t* key, const uint8_t* iv,
                    const uint8_t* input, size_t input_len,
                    uint8_t* output, size_t* output_len)
{
    if (!key || !iv || !input || !output || !output_len || input_len % SM4_BLOCK_SIZE != 0) {
        return -1;
    }

    uint32_t rk[SM4_ROUNDS];
    sm4_key_expansion(key, rk);

    uint8_t last_block[SM4_BLOCK_SIZE];
    memcpy(last_block, iv, SM4_BLOCK_SIZE);

    size_t blocks = input_len / SM4_BLOCK_SIZE;
    for (size_t i = 0; i < blocks; i++) {
        uint8_t block[SM4_BLOCK_SIZE];
        for (int j = 0; j < SM4_BLOCK_SIZE; j++) {
            block[j] = input[i * SM4_BLOCK_SIZE + j] ^ last_block[j];
        }
        sm4_block_encrypt(rk, block, output + i * SM4_BLOCK_SIZE);
        memcpy(last_block, output + i * SM4_BLOCK_SIZE, SM4_BLOCK_SIZE);
    }

    *output_len = blocks * SM4_BLOCK_SIZE;
    return 0;
}

int sm4_cbc_decrypt(const uint8_t* key, const uint8_t* iv,
                    const uint8_t* input, size_t input_len,
                    uint8_t* output, size_t* output_len)
{
    if (!key || !iv || !input || !output || !output_len || input_len % SM4_BLOCK_SIZE != 0) {
        return -1;
    }

    uint32_t rk[SM4_ROUNDS];
    sm4_key_expansion(key, rk);

    uint8_t last_block[SM4_BLOCK_SIZE];
    memcpy(last_block, iv, SM4_BLOCK_SIZE);

    size_t blocks = input_len / SM4_BLOCK_SIZE;
    for (size_t i = 0; i < blocks; i++) {
        uint8_t block[SM4_BLOCK_SIZE];
        sm4_block_decrypt(rk, input + i * SM4_BLOCK_SIZE, block);
        for (int j = 0; j < SM4_BLOCK_SIZE; j++) {
            output[i * SM4_BLOCK_SIZE + j] = block[j] ^ last_block[j];
        }
        memcpy(last_block, input + i * SM4_BLOCK_SIZE, SM4_BLOCK_SIZE);
    }

    *output_len = blocks * SM4_BLOCK_SIZE;
    return 0;
}

int sm4_cfb_encrypt(const uint8_t* key, const uint8_t* iv,
                    const uint8_t* input, size_t input_len,
                    uint8_t* output, size_t* output_len)
{
    if (!key || !iv || !input || !output || !output_len) {
        return -1;
    }

    uint32_t rk[SM4_ROUNDS];
    sm4_key_expansion(key, rk);

    uint8_t shift_reg[SM4_BLOCK_SIZE];
    memcpy(shift_reg, iv, SM4_BLOCK_SIZE);

    for (size_t i = 0; i < input_len; i++) {
        if (i % SM4_BLOCK_SIZE == 0) {
            sm4_block_encrypt(rk, shift_reg, shift_reg);
        }
        output[i] = input[i] ^ shift_reg[i % SM4_BLOCK_SIZE];
        shift_reg[i % SM4_BLOCK_SIZE] = output[i];
    }

    *output_len = input_len;
    return 0;
}

int sm4_cfb_decrypt(const uint8_t* key, const uint8_t* iv,
                    const uint8_t* input, size_t input_len,
                    uint8_t* output, size_t* output_len)
{
    if (!key || !iv || !input || !output || !output_len) {
        return -1;
    }

    uint32_t rk[SM4_ROUNDS];
    sm4_key_expansion(key, rk);

    uint8_t shift_reg[SM4_BLOCK_SIZE];
    memcpy(shift_reg, iv, SM4_BLOCK_SIZE);

    for (size_t i = 0; i < input_len; i++) {
        if (i % SM4_BLOCK_SIZE == 0) {
            sm4_block_encrypt(rk, shift_reg, shift_reg);
        }
        uint8_t cipher_byte = input[i];
        output[i] = input[i] ^ shift_reg[i % SM4_BLOCK_SIZE];
        shift_reg[i % SM4_BLOCK_SIZE] = cipher_byte;
    }

    *output_len = input_len;
    return 0;
}

int sm4_ofb_encrypt(const uint8_t* key, const uint8_t* iv,
                    const uint8_t* input, size_t input_len,
                    uint8_t* output, size_t* output_len)
{
    if (!key || !iv || !input || !output || !output_len) {
        return -1;
    }

    uint32_t rk[SM4_ROUNDS];
    sm4_key_expansion(key, rk);

    uint8_t shift_reg[SM4_BLOCK_SIZE];
    memcpy(shift_reg, iv, SM4_BLOCK_SIZE);

    for (size_t i = 0; i < input_len; i++) {
        if (i % SM4_BLOCK_SIZE == 0) {
            sm4_block_encrypt(rk, shift_reg, shift_reg);
        }
        output[i] = input[i] ^ shift_reg[i % SM4_BLOCK_SIZE];
    }

    *output_len = input_len;
    return 0;
}

int sm4_ofb_decrypt(const uint8_t* key, const uint8_t* iv,
                    const uint8_t* input, size_t input_len,
                    uint8_t* output, size_t* output_len)
{
    return sm4_ofb_encrypt(key, iv, input, input_len, output, output_len);
}

int sm4_encrypt_update(sm4_ctx_t* ctx, const uint8_t* input, size_t input_len,
                       uint8_t* output, size_t* output_len)
{
    if (!ctx || !input || !output || !output_len) {
        return -1;
    }

    int ret = 0;
    switch (ctx->mode) {
    case SM4_MODE_ECB:
        ret = sm4_ecb_encrypt((uint8_t*)ctx->rk, input, input_len, output, output_len);
        break;
    case SM4_MODE_CBC:
        if (!ctx->iv_set) return -1;
        ret = sm4_cbc_encrypt((uint8_t*)ctx->rk, ctx->iv, input, input_len, output, output_len);
        break;
    case SM4_MODE_CFB:
        if (!ctx->iv_set) return -1;
        ret = sm4_cfb_encrypt((uint8_t*)ctx->rk, ctx->iv, input, input_len, output, output_len);
        break;
    case SM4_MODE_OFB:
        if (!ctx->iv_set) return -1;
        ret = sm4_ofb_encrypt((uint8_t*)ctx->rk, ctx->iv, input, input_len, output, output_len);
        break;
    default:
        ret = -1;
        break;
    }
    return ret;
}

int sm4_decrypt_update(sm4_ctx_t* ctx, const uint8_t* input, size_t input_len,
                       uint8_t* output, size_t* output_len)
{
    if (!ctx || !input || !output || !output_len) {
        return -1;
    }

    int ret = 0;
    switch (ctx->mode) {
    case SM4_MODE_ECB:
        ret = sm4_ecb_decrypt((uint8_t*)ctx->rk, input, input_len, output, output_len);
        break;
    case SM4_MODE_CBC:
        if (!ctx->iv_set) return -1;
        ret = sm4_cbc_decrypt((uint8_t*)ctx->rk, ctx->iv, input, input_len, output, output_len);
        break;
    case SM4_MODE_CFB:
        if (!ctx->iv_set) return -1;
        ret = sm4_cfb_decrypt((uint8_t*)ctx->rk, ctx->iv, input, input_len, output, output_len);
        break;
    case SM4_MODE_OFB:
        if (!ctx->iv_set) return -1;
        ret = sm4_ofb_decrypt((uint8_t*)ctx->rk, ctx->iv, input, input_len, output, output_len);
        break;
    default:
        ret = -1;
        break;
    }
    return ret;
}

int sm4_encrypt_final(sm4_ctx_t* ctx, uint8_t* output, size_t* output_len)
{
    (void)ctx;
    (void)output;
    if (output_len) *output_len = 0;
    return 0;
}

int sm4_decrypt_final(sm4_ctx_t* ctx, uint8_t* output, size_t* output_len)
{
    (void)ctx;
    (void)output;
    if (output_len) *output_len = 0;
    return 0;
}

void sm4_self_test(void)
{
    uint8_t key[SM4_KEY_SIZE] = {
        0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
        0xfe, 0xdc, 0xba, 0x98, 0x76, 0x54, 0x32, 0x10
    };
    uint8_t plain[SM4_BLOCK_SIZE] = {
        0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
        0xfe, 0xdc, 0xba, 0x98, 0x76, 0x54, 0x32, 0x10
    };
    uint8_t cipher[SM4_BLOCK_SIZE];
    uint8_t decrypted[SM4_BLOCK_SIZE];
    size_t out_len;

    printf("SM4 Self Test:\n");
    printf("  Key: ");
    for (int i = 0; i < SM4_KEY_SIZE; i++) printf("%02x", key[i]);
    printf("\n  Plain: ");
    for (int i = 0; i < SM4_BLOCK_SIZE; i++) printf("%02x", plain[i]);

    sm4_ecb_encrypt(key, plain, SM4_BLOCK_SIZE, cipher, &out_len);
    printf("\n  Cipher: ");
    for (int i = 0; i < SM4_BLOCK_SIZE; i++) printf("%02x", cipher[i]);

    sm4_ecb_decrypt(key, cipher, SM4_BLOCK_SIZE, decrypted, &out_len);
    printf("\n  Decrypted: ");
    for (int i = 0; i < SM4_BLOCK_SIZE; i++) printf("%02x", decrypted[i]);
    printf("\n  Result: %s\n", memcmp(plain, decrypted, SM4_BLOCK_SIZE) == 0 ? "PASS" : "FAIL");
}
