package main

import (
	"fmt"
	"os"

	"github.com/cuilan/license-key-verify/pkg/crypto"
	"github.com/cuilan/license-key-verify/pkg/license"
)

// =================================================================
// 关键安全区：将您的公钥和AES密钥内容作为常量嵌入到代码中
// 建议从您安全保存的密钥文件中读取内容，然后粘贴到这里。
// =================================================================

const (
	// publicKeyPEM 应该包含 '-----BEGIN PUBLIC KEY-----' 和 '-----END PUBLIC KEY-----' 的完整内容。
	//
	// !!! 这是一个示例公钥，请务必替换为您自己的公钥。!!!
	// 您可以使用 `lkctl gen` 生成一对密钥，然后将 public.pem 的内容粘贴到这里。
	publicKeyPEM = `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqP4C3nN5Z5vB74rVb5q1
... (your public key content here) ...
-----END PUBLIC KEY-----
`

	// aesKeyBase64 应该是从您生成的 aes.key 文件中复制的 Base64 编码字符串。
	//
	// !!! 这是一个示例AES密钥，请务必替换为您自己的主AES密钥。!!!
	aesKeyBase64 = `YOUR_BASE64_ENCODED_AES_KEY_HERE`
)

// createVerifierFromEmbeddedKeys 从硬编码的密钥创建一个安全的验证器实例。
// 这个验证器只信任内嵌的密钥，从而防止用户使用自己的密钥进行自签发。
func createVerifierFromEmbeddedKeys() (*license.Verifier, error) {
	// 1. 从Base64字符串解码AES密钥
	aesKey, err := crypto.DecodeBase64(aesKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode embedded AES key: %w", err)
	}

	// 2. 使用加载好的密钥创建验证器实例
	// NewVerifier 函数确保验证器只使用给定的密钥。
	return license.NewVerifier([]byte(publicKeyPEM), aesKey)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <path-to-license.lic>\n", os.Args[0])
		os.Exit(1)
	}
	licenseFile := os.Args[1]

	// 初始化一个只信任内嵌密钥的安全验证器
	verifier, err := createVerifierFromEmbeddedKeys()
	if err != nil {
		// 如果密钥在编译时就是错误的，这里会立即失败。
		fmt.Printf("FATAL: Could not initialize license verifier: %v\n", err)
		os.Exit(1)
	}

	// 使用这个安全的验证器来校验许可证文件
	result, err := verifier.VerifyFile(licenseFile)
	if err != nil {
		fmt.Printf("License verification failed with an error: %v\n", err)
		os.Exit(1)
	}

	if result.Valid {
		fmt.Println("✅ License is VALID.")
		// 在这里运行您的应用程序核心逻辑...
	} else {
		fmt.Printf("❌ License is INVALID: %s\n", result.Error)
		os.Exit(1)
	}
}
