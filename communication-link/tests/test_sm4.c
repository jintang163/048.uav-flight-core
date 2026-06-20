#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>
#include "../crypto/sm4.h"
#include "../crypto/auth.h"

#define TEST_ASSERT(cond) do { if (!(cond)) { printf("FAIL: %s line %d\n", __FILE__, __LINE__); return -1; } } while(0)

static int test_sm4_ecb_encrypt_decrypt(void)
{
    printf("Testing SM4 ECB mode... ");

    const uint8_t key[SM4_KEY_SIZE] = {
        0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
        0xfe, 0xdc, 0xba, 0x98, 0x76, 0x54, 0x32, 0x10
    };

    const uint8_t plaintext[SM4_BLOCK_SIZE] = {
        0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
        0xfe, 0xdc, 0xba, 0x98, 0x76, 0x54, 0x32, 0x10
    };

    const uint8_t expected_ciphertext[SM4_BLOCK_SIZE] = {
        0x68, 0x1e, 0xdf, 0x34, 0xd2, 0x06, 0x96, 0x5e,
        0x86, 0xb3, 0xe9, 0x4f, 0x53, 0x6e, 0x42, 0x46
    };

    uint8_t ciphertext[SM4_BLOCK_SIZE];
    uint8_t decrypted[SM4_BLOCK_SIZE];
    size_t out_len;
    int ret;

    ret = sm4_ecb_encrypt(key, plaintext, SM4_BLOCK_SIZE, ciphertext, &out_len);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(out_len == SM4_BLOCK_SIZE);
    TEST_ASSERT(memcmp(ciphertext, expected_ciphertext, SM4_BLOCK_SIZE) == 0);

    ret = sm4_ecb_decrypt(key, ciphertext, SM4_BLOCK_SIZE, decrypted, &out_len);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(out_len == SM4_BLOCK_SIZE);
    TEST_ASSERT(memcmp(decrypted, plaintext, SM4_BLOCK_SIZE) == 0);

    printf("OK\n");
    return 0;
}

static int test_sm4_cbc_encrypt_decrypt(void)
{
    printf("Testing SM4 CBC mode... ");

    const uint8_t key[SM4_KEY_SIZE] = {
        0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
        0xfe, 0xdc, 0xba, 0x98, 0x76, 0x54, 0x32, 0x10
    };

    const uint8_t iv[SM4_IV_SIZE] = {
        0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
        0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f
    };

    const uint8_t plaintext[] = "Hello, SM4 CBC Mode Encryption Test!";
    size_t plaintext_len = strlen((char*)plaintext);

    size_t padded_len = ((plaintext_len + SM4_BLOCK_SIZE - 1) / SM4_BLOCK_SIZE) * SM4_BLOCK_SIZE;
    uint8_t* padded = (uint8_t*)malloc(padded_len);
    uint8_t* ciphertext = (uint8_t*)malloc(padded_len);
    uint8_t* decrypted = (uint8_t*)malloc(padded_len + 1);

    memcpy(padded, plaintext, plaintext_len);
    sm4_pkcs7_pad(padded, plaintext_len, SM4_BLOCK_SIZE);

    size_t out_len;
    int ret = sm4_cbc_encrypt(key, iv, padded, padded_len, ciphertext, &out_len);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(out_len == padded_len);

    uint8_t iv2[SM4_IV_SIZE];
    memcpy(iv2, iv, SM4_IV_SIZE);
    ret = sm4_cbc_decrypt(key, iv2, ciphertext, padded_len, decrypted, &out_len);
    TEST_ASSERT(ret == 0);

    size_t unpadded_len;
    ret = sm4_pkcs7_unpad(decrypted, out_len, &unpadded_len);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(unpadded_len == plaintext_len);

    decrypted[unpadded_len] = '\0';
    TEST_ASSERT(strcmp((char*)decrypted, (char*)plaintext) == 0);

    free(padded);
    free(ciphertext);
    free(decrypted);

    printf("OK\n");
    return 0;
}

static int test_sm4_cfb_ofb_modes(void)
{
    printf("Testing SM4 CFB/OFB modes... ");

    const uint8_t key[SM4_KEY_SIZE] = {
        0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
        0xfe, 0xdc, 0xba, 0x98, 0x76, 0x54, 0x32, 0x10
    };

    const uint8_t iv[SM4_IV_SIZE] = {
        0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
        0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f
    };

    const uint8_t plaintext[] = "Stream Cipher Mode Test for CFB and OFB";
    size_t len = strlen((char*)plaintext);

    uint8_t cfb_cipher[256], cfb_plain[256];
    uint8_t ofb_cipher[256], ofb_plain[256];
    size_t out_len;

    int ret = sm4_cfb_encrypt(key, iv, plaintext, len, cfb_cipher, &out_len);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(out_len == len);

    uint8_t iv2[SM4_IV_SIZE];
    memcpy(iv2, iv, SM4_IV_SIZE);
    ret = sm4_cfb_decrypt(key, iv2, cfb_cipher, len, cfb_plain, &out_len);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(out_len == len);
    TEST_ASSERT(memcmp(cfb_plain, plaintext, len) == 0);

    ret = sm4_ofb_encrypt(key, iv, plaintext, len, ofb_cipher, &out_len);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(out_len == len);

    memcpy(iv2, iv, SM4_IV_SIZE);
    ret = sm4_ofb_decrypt(key, iv2, ofb_cipher, len, ofb_plain, &out_len);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(out_len == len);
    TEST_ASSERT(memcmp(ofb_plain, plaintext, len) == 0);

    printf("OK\n");
    return 0;
}

static int test_sm4_pkcs7_padding(void)
{
    printf("Testing SM4 PKCS7 padding... ");

    uint8_t buf[32];

    for (size_t i = 0; i < 16; i++) {
        memset(buf, 0, sizeof(buf));
        for (size_t j = 0; j < i; j++) buf[j] = 0xAA;

        size_t pad_len = sm4_pkcs7_pad(buf, i, 16);
        TEST_ASSERT(pad_len == 16 - i);

        for (size_t j = i; j < 16; j++) {
            TEST_ASSERT(buf[j] == (uint8_t)(16 - i));
        }

        size_t unpadded_len;
        int ret = sm4_pkcs7_unpad(buf, 16, &unpadded_len);
        TEST_ASSERT(ret == 0);
        TEST_ASSERT(unpadded_len == i);
    }

    printf("OK\n");
    return 0;
}

static int test_sm4_self_test(void)
{
    printf("Testing SM4 self-test... ");
    int ret = sm4_self_test();
    TEST_ASSERT(ret == 0);
    printf("OK\n");
    return 0;
}

static int test_sm3_hash(void)
{
    printf("Testing SM3 hash... ");

    const uint8_t msg1[] = "abc";
    uint8_t hash1[SM3_HASH_SIZE];

    sm3_hash(msg1, strlen((char*)msg1), hash1);

    const uint8_t expected1[SM3_HASH_SIZE] = {
        0x66, 0xc7, 0xf0, 0xf4, 0x62, 0xee, 0xed, 0xd9,
        0xd1, 0xf2, 0xd4, 0x6b, 0xdc, 0x10, 0xe4, 0xe2,
        0x41, 0x67, 0xc4, 0x87, 0x5c, 0xf2, 0xf7, 0xa2,
        0x29, 0x7d, 0xa0, 0x2b, 0x8f, 0x4b, 0xa8, 0xe0
    };

    TEST_ASSERT(memcmp(hash1, expected1, SM3_HASH_SIZE) == 0);

    printf("OK\n");
    return 0;
}

static int test_hmac_sm3(void)
{
    printf("Testing HMAC-SM3... ");

    const uint8_t key[] = "secret_key";
    const uint8_t data[] = "Hello, HMAC-SM3!";
    uint8_t hmac[HMAC_SM3_SIZE];
    uint8_t hmac2[HMAC_SM3_SIZE];

    hmac_sm3(key, strlen((char*)key), data, strlen((char*)data), hmac);
    hmac_sm3(key, strlen((char*)key), data, strlen((char*)data), hmac2);

    TEST_ASSERT(memcmp(hmac, hmac2, HMAC_SM3_SIZE) == 0);

    bool verify_ok = hmac_sm3_verify(key, strlen((char*)key),
                                     data, strlen((char*)data),
                                     hmac, HMAC_SM3_SIZE);
    TEST_ASSERT(verify_ok);

    hmac[0] ^= 0x01;
    verify_ok = hmac_sm3_verify(key, strlen((char*)key),
                                data, strlen((char*)data),
                                hmac, HMAC_SM3_SIZE);
    TEST_ASSERT(!verify_ok);

    printf("OK\n");
    return 0;
}

static int test_large_data_encryption(void)
{
    printf("Testing large data encryption... ");

    const uint8_t key[SM4_KEY_SIZE] = {0};
    const uint8_t iv[SM4_IV_SIZE] = {0};

    size_t data_len = 4096;
    uint8_t* plaintext = (uint8_t*)malloc(data_len);
    uint8_t* ciphertext = (uint8_t*)malloc(data_len + SM4_BLOCK_SIZE);
    uint8_t* decrypted = (uint8_t*)malloc(data_len + SM4_BLOCK_SIZE);

    for (size_t i = 0; i < data_len; i++) {
        plaintext[i] = (uint8_t)(i & 0xFF);
    }

    size_t pad_len = SM4_BLOCK_SIZE - (data_len % SM4_BLOCK_SIZE);
    size_t padded_len = data_len + pad_len;
    uint8_t* padded = (uint8_t*)malloc(padded_len);
    memcpy(padded, plaintext, data_len);
    sm4_pkcs7_pad(padded, data_len, SM4_BLOCK_SIZE);

    size_t enc_len, dec_len;
    int ret = sm4_cbc_encrypt(key, iv, padded, padded_len, ciphertext, &enc_len);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(enc_len == padded_len);

    uint8_t iv2[SM4_IV_SIZE];
    memcpy(iv2, iv, SM4_IV_SIZE);
    ret = sm4_cbc_decrypt(key, iv2, ciphertext, enc_len, decrypted, &dec_len);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(dec_len == padded_len);

    size_t unpadded_len;
    ret = sm4_pkcs7_unpad(decrypted, dec_len, &unpadded_len);
    TEST_ASSERT(ret == 0);
    TEST_ASSERT(unpadded_len == data_len);
    TEST_ASSERT(memcmp(decrypted, plaintext, data_len) == 0);

    free(plaintext);
    free(ciphertext);
    free(decrypted);
    free(padded);

    printf("OK\n");
    return 0;
}

int main(void)
{
    printf("\n=== SM4 Crypto Unit Tests ===\n\n");

    int ret = 0;
    ret |= test_sm4_self_test();
    ret |= test_sm4_ecb_encrypt_decrypt();
    ret |= test_sm4_cbc_encrypt_decrypt();
    ret |= test_sm4_cfb_ofb_modes();
    ret |= test_sm4_pkcs7_padding();
    ret |= test_sm3_hash();
    ret |= test_hmac_sm3();
    ret |= test_large_data_encryption();

    printf("\n=== SM4 Crypto Tests %s ===\n\n", ret == 0 ? "PASSED" : "FAILED");

    return ret;
}
