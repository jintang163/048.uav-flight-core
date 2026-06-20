#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "../include/communication_link.h"
#include "../crypto/sm4.h"
#include "../crypto/auth.h"
#include "../crypto/crypto_manager.h"
#include "../crypto/key_exchange.h"
#include "../crypto/tls_manager.h"

static void print_hex(const char* label, const uint8_t* data, size_t len)
{
    printf("%s: ", label);
    for (size_t i = 0; i < len; i++) {
        printf("%02X", data[i]);
    }
    printf("\n");
}

int main(int argc, char* argv[])
{
    (void)argc;
    (void)argv;

    printf("=== 通信链路示例 - 加密通信\n\n");

    uint8_t key[SM4_KEY_SIZE];
    uint8_t iv[SM4_IV_SIZE];
    const uint8_t plaintext[] = "这是一条需要加密的敏感数据: 无人机控制指令";
    size_t plaintext_len = strlen((char*)plaintext);

    printf("原始数据: %s\n", plaintext);
    printf("数据长度: %zu 字节\n\n", plaintext_len);

    for (int i = 0; i < SM4_KEY_SIZE; i++) {
        key[i] = (uint8_t)(rand() & 0xFF);
        iv[i] = (uint8_t)(rand() & 0xFF);
    }

    print_hex("密钥", key, SM4_KEY_SIZE);
    print_hex("初始向量", iv, SM4_IV_SIZE);
    printf("\n");

    crypto_manager_t crypto;
    crypto_manager_init(&crypto, CRYPTO_ALGO_SM4_CBC);
    crypto_manager_set_key(&crypto, key, SM4_KEY_SIZE);
    crypto_manager_set_iv(&crypto, iv, SM4_IV_SIZE);

    printf("=== SM4-CBC 加密测试\n");

    uint8_t ciphertext[1024];
    size_t ciphertext_len = sizeof(ciphertext);

    int ret = crypto_manager_encrypt(&crypto, plaintext, plaintext_len,
                                      ciphertext, &ciphertext_len);
    if (ret == 0) {
        print_hex("加密结果", ciphertext, ciphertext_len);
        printf("加密后长度: %zu 字节\n", ciphertext_len);

        uint8_t decrypted[1024];
        size_t decrypted_len = sizeof(decrypted);

        crypto_manager_set_iv(&crypto, iv, SM4_IV_SIZE);
        ret = crypto_manager_decrypt(&crypto, ciphertext, ciphertext_len,
                                    decrypted, &decrypted_len);
        if (ret == 0) {
            decrypted[decrypted_len] = '\0';
            printf("解密结果: %s\n", decrypted);
            printf("解密后长度: %zu 字节\n", decrypted_len);
            printf("数据完整性: %s\n\n",
                   strcmp((char*)decrypted, (char*)plaintext) == 0 ? "验证通过" : "验证失败");
        }
    }

    printf("=== HMAC-SM3 消息认证测试\n");

    uint8_t hmac[HMAC_SM3_SIZE];
    size_t hmac_len = sizeof(hmac);

    ret = crypto_manager_sign(&crypto, plaintext, plaintext_len, hmac, &hmac_len);
    if (ret == 0) {
        print_hex("HMAC签名", hmac, HMAC_SM3_SIZE);

        ret = crypto_manager_verify(&crypto, plaintext, plaintext_len, hmac, HMAC_SM3_SIZE);
        printf("HMAC验证: %s\n", ret == 0 ? "通过" : "失败");
    }

    printf("\n=== ECDH 密钥交换测试\n");

    ecdh_ctx_t alice, bob;
    ecdh_init(&alice, CURVE_P256);
    ecdh_init(&bob, CURVE_P256);

    ecdh_generate_keys(&alice);
    ecdh_generate_keys(&bob);

    uint8_t alice_pub[ECDH_P256_PUB_KEY_SIZE];
    uint8_t bob_pub[ECDH_P256_PUB_KEY_SIZE];
    size_t pub_len = ECDH_P256_PUB_KEY_SIZE;

    ecdh_get_public_key(&alice, alice_pub, &pub_len);
    ecdh_get_public_key(&bob, bob_pub, &pub_len);

    print_hex("Alice 公钥", alice_pub, 16);
    print_hex("Bob 公钥", bob_pub, 16);

    uint8_t alice_secret[ECDH_SHARED_SECRET_SIZE];
    uint8_t bob_secret[ECDH_SHARED_SECRET_SIZE];
    size_t secret_len = ECDH_SHARED_SECRET_SIZE;

    ecdh_compute_shared_secret(&alice, bob_pub, ECDH_P256_PUB_KEY_SIZE,
                                  alice_secret, &secret_len);
    ecdh_compute_shared_secret(&bob, alice_pub, ECDH_P256_PUB_KEY_SIZE,
                              bob_secret, &secret_len);

    print_hex("Alice 共享密钥", alice_secret, 16);
    print_hex("Bob 共享密钥", bob_secret, 16);

    printf("共享密钥一致性: %s\n",
           memcmp(alice_secret, bob_secret, ECDH_SHARED_SECRET_SIZE) == 0 ? "一致" : "不一致");

    printf("\n=== 加密统计信息\n");
    printf("已加密字节: %llu\n", (unsigned long long)crypto_manager_get_encrypted_bytes(&crypto));
    printf("已解密字节: %llu\n", (unsigned long long)crypto_manager_get_decrypted_bytes(&crypto));
    printf("错误次数: %u\n", crypto_manager_get_error_count(&crypto));

    crypto_manager_cleanup(&crypto);
    ecdh_cleanup(&alice);
    ecdh_cleanup(&bob);

    printf("\n=== 加密通信示例完成 ===\n");

    return 0;
}
