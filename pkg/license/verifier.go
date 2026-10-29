package license

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cuilan/license-key-verify/pkg/crypto"
	"github.com/cuilan/license-key-verify/pkg/machine"
)

var (
	keyMu     sync.RWMutex
	globalKey *rsa.PublicKey
)

// UsePublicKeyPEM 用指定的 PEM 公钥覆盖内置公钥。
// 直接引用本模块（未 fork 替换 key.go）的项目，必须在首次验证前调用。
func UsePublicKeyPEM(publicKeyPEM []byte) error {
	key, err := crypto.LoadPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		return fmt.Errorf("failed to load public key: %v", err)
	}
	keyMu.Lock()
	globalKey = key
	keyMu.Unlock()
	return nil
}

// currentKey 返回当前生效的验证公钥（内置或 UsePublicKeyPEM 设置的）。
func currentKey() (*rsa.PublicKey, error) {
	keyMu.RLock()
	defer keyMu.RUnlock()
	if globalKey == nil {
		return nil, errors.New("no public key available: embedded key is invalid, call license.UsePublicKeyPEM first")
	}
	return globalKey, nil
}

// HasEmbeddedKey 返回当前二进制是否内嵌了可用公钥。
// fork 本仓库后未执行 lkctl embed 时为 false，验证与签发防呆检查都会失败。
func HasEmbeddedKey() bool {
	key, err := currentKey()
	return err == nil && key != nil
}

// EmbeddedKeyPEM 返回内嵌公钥的 PEM 编码（未内嵌时返回错误）。
func EmbeddedKeyPEM() ([]byte, error) {
	key, err := currentKey()
	if err != nil {
		return nil, err
	}
	return crypto.PublicKeyToPEM(key)
}

// VerifyFile 验证许可证文件（使用内置公钥）。
func VerifyFile(path string) Result {
	data, err := os.ReadFile(path)
	if err != nil {
		return failResult(fmt.Sprintf("failed to read license file: %v", err))
	}
	return Verify(data)
}

// Verify 验证许可证数据（使用内置公钥）。
// 依次校验：JSON 解析 → RSA 签名 → 有效期 → 机器指纹。
func Verify(data []byte) Result {
	result := Result{VerifiedAt: time.Now().Format(time.RFC3339)}

	var lic License
	if err := json.Unmarshal(data, &lic); err != nil {
		result.Error = fmt.Sprintf("failed to parse license file: %v", err)
		return result
	}

	// 还原被签名的字节：置空签名后重新序列化（与签发侧完全一致）
	signature, err := base64.StdEncoding.DecodeString(lic.Signature)
	if err != nil {
		result.Error = fmt.Sprintf("failed to decode signature: %v", err)
		return result
	}
	lic.Signature = ""
	payload, err := json.Marshal(&lic)
	if err != nil {
		result.Error = fmt.Sprintf("failed to re-serialize license: %v", err)
		return result
	}

	key, err := currentKey()
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if err := crypto.VerifySignature(payload, signature, key); err != nil {
		result.Error = err.Error()
		return result
	}

	// 签名通过，内容可信
	result.License = &lic

	now := time.Now()
	if now.Before(lic.IssuedAt) {
		result.Error = "license is not yet valid"
		return result
	}
	if now.After(lic.ExpiresAt) {
		result.Error = "license has expired"
		result.ExpiresIn = 0
		return result
	}
	result.ExpiresIn = int64(lic.ExpiresAt.Sub(now).Seconds())

	// 机器指纹校验（空指纹视为不匹配）
	fp, err := machine.Fingerprint()
	if err != nil {
		result.Error = fmt.Sprintf("failed to compute machine fingerprint: %v", err)
		return result
	}
	result.Fingerprint = fp
	result.Matched = lic.Fingerprint == fp
	if !result.Matched {
		result.Error = "machine fingerprint does not match"
		return result
	}

	result.Valid = true
	return result
}

// failResult 构造失败结果（无法确定指纹时使用）。
func failResult(msg string) Result {
	return Result{VerifiedAt: time.Now().Format(time.RFC3339), Error: msg}
}
