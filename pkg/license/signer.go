package license

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/cuilan/license-key-verify/v2/pkg/crypto"
)

// Sign 签发许可证：构造 License 结构并用私钥对内容签名。
// 成功后 License.Signature 已填充，可直接 Encode/Save 输出。
func Sign(options *SignOptions, privateKeyPEM []byte) (*License, error) {
	if options == nil {
		return nil, fmt.Errorf("sign options cannot be nil")
	}
	if options.Fingerprint == "" {
		return nil, fmt.Errorf("fingerprint is required (run 'lkctl fingerprint' or the app with --fingerprint)")
	}

	privateKey, err := crypto.LoadPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %v", err)
	}

	licenseID := options.LicenseID
	if licenseID == "" {
		licenseID, err = generateLicenseID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate license ID: %v", err)
		}
	}

	issuedAt := options.IssuedAt
	if issuedAt.IsZero() {
		issuedAt = time.Now()
	}
	duration := options.Duration
	if duration <= 0 {
		duration = DefaultDuration
	}
	product := options.Product
	if product == "" {
		product = DefaultProduct
	}

	lic := &License{
		LicenseID:   licenseID,
		Customer:    options.Customer,
		Product:     product,
		Fingerprint: options.Fingerprint,
		IssuedAt:    issuedAt,
		ExpiresAt:   issuedAt.Add(duration),
		Features:    options.Features,
	}

	// 签名覆盖除 Signature 外的全部字段（Signature 为空时 omitempty 不输出）
	payload, err := json.Marshal(lic)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal license: %v", err)
	}

	signature, err := crypto.SignData(payload, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign license: %v", err)
	}
	lic.Signature = base64.StdEncoding.EncodeToString(signature)

	return lic, nil
}

// Encode 将许可证序列化为带缩进的 JSON 文件内容。
func (l *License) Encode() ([]byte, error) {
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal license: %v", err)
	}
	return append(data, '\n'), nil
}

// Save 将许可证写入文件。
func (l *License) Save(path string) error {
	data, err := l.Encode()
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write license file: %v", err)
	}
	return nil
}

// generateLicenseID 生成 UUID 样式的许可证 ID
func generateLicenseID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
