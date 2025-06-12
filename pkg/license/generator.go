package license

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"license-key-verify/pkg/crypto"
)

// Generator 许可证生成器
type Generator struct {
	privateKey *rsa.PrivateKey
	aesKey     []byte
}

// NewGenerator 创建新的生成器
func NewGenerator() (*Generator, error) {
	// 生成RSA密钥对
	keyPair, err := crypto.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("生成密钥对失败: %v", err)
	}

	// 生成AES密钥
	aesKey, err := crypto.GenerateAESKey()
	if err != nil {
		return nil, fmt.Errorf("生成AES密钥失败: %v", err)
	}

	return &Generator{
		privateKey: keyPair.PrivateKey,
		aesKey:     aesKey,
	}, nil
}

// NewGeneratorWithKeys 使用指定密钥创建生成器
func NewGeneratorWithKeys(privateKeyPEM []byte, aesKey []byte) (*Generator, error) {
	privateKey, err := crypto.LoadPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("加载私钥失败: %v", err)
	}

	return &Generator{
		privateKey: privateKey,
		aesKey:     aesKey,
	}, nil
}

// Generate 生成许可证
func (g *Generator) Generate(options *GenerateOptions) (*License, error) {
	if options == nil {
		return nil, fmt.Errorf("生成选项不能为空")
	}

	// 生成许可证ID
	licenseID, err := g.generateLicenseID()
	if err != nil {
		return nil, fmt.Errorf("生成许可证ID失败: %v", err)
	}

	// 设置默认值
	if options.ProductName == "" {
		options.ProductName = DefaultProductName
	}
	if options.Version == "" {
		options.Version = DefaultVersion
	}
	if options.Duration == 0 {
		options.Duration = 365 * 24 * time.Hour // 默认1年
	}

	now := time.Now()
	license := &License{
		ID:           licenseID,
		ProductName:  options.ProductName,
		Version:      options.Version,
		MAC:          options.MAC,
		UUID:         options.UUID,
		CPUID:        options.CPUID,
		IssuedAt:     now,
		ExpiresAt:    now.Add(options.Duration),
		Features:     options.Features,
		MaxUsers:     options.MaxUsers,
		CustomerName: options.CustomerName,
		Notes:        options.Notes,
		Extra:        options.Extra,
	}

	return license, nil
}

// SaveToFile 将许可证保存到文件
func (g *Generator) SaveToFile(license *License, filePath string) error {
	// 序列化许可证
	licenseData, err := json.Marshal(license)
	if err != nil {
		return fmt.Errorf("序列化许可证失败: %v", err)
	}

	// 加密许可证数据
	encryptedData, err := crypto.EncryptAES(licenseData, g.aesKey)
	if err != nil {
		return fmt.Errorf("加密许可证数据失败: %v", err)
	}

	// 对加密数据进行签名
	signature, err := crypto.SignData(encryptedData, g.privateKey)
	if err != nil {
		return fmt.Errorf("签名失败: %v", err)
	}

	// 创建许可证文件
	licenseFile := &LicenseFile{
		Data:      crypto.EncodeBase64(encryptedData),
		Signature: crypto.EncodeBase64(signature),
		Algorithm: DefaultAlgorithm,
		Version:   FileFormatVersion,
	}

	// 序列化许可证文件
	fileData, err := json.MarshalIndent(licenseFile, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化许可证文件失败: %v", err)
	}

	// 写入文件
	err = os.WriteFile(filePath, fileData, 0644)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

// GetPublicKey 获取公钥
func (g *Generator) GetPublicKey() *rsa.PublicKey {
	return &g.privateKey.PublicKey
}

// GetPublicKeyPEM 获取PEM格式的公钥
func (g *Generator) GetPublicKeyPEM() ([]byte, error) {
	keyPair := &crypto.KeyPair{
		PrivateKey: g.privateKey,
		PublicKey:  &g.privateKey.PublicKey,
	}
	return keyPair.PublicKeyToPEM()
}

// GetAESKey 获取AES密钥
func (g *Generator) GetAESKey() []byte {
	return g.aesKey
}

// generateLicenseID 生成许可证ID
func (g *Generator) generateLicenseID() (string, error) {
	// 生成16字节随机数
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	// 格式化为UUID样式
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		randomBytes[0:4],
		randomBytes[4:6],
		randomBytes[6:8],
		randomBytes[8:10],
		randomBytes[10:16],
	), nil
}

// SaveKeys 保存密钥到文件
func (g *Generator) SaveKeys(privateKeyPath, publicKeyPath, aesKeyPath string) error {
	// 保存私钥
	keyPair := &crypto.KeyPair{
		PrivateKey: g.privateKey,
		PublicKey:  &g.privateKey.PublicKey,
	}

	privateKeyPEM, err := keyPair.PrivateKeyToPEM()
	if err != nil {
		return fmt.Errorf("转换私钥失败: %v", err)
	}

	err = os.WriteFile(privateKeyPath, privateKeyPEM, 0600)
	if err != nil {
		return fmt.Errorf("保存私钥失败: %v", err)
	}

	// 保存公钥
	publicKeyPEM, err := keyPair.PublicKeyToPEM()
	if err != nil {
		return fmt.Errorf("转换公钥失败: %v", err)
	}

	err = os.WriteFile(publicKeyPath, publicKeyPEM, 0644)
	if err != nil {
		return fmt.Errorf("保存公钥失败: %v", err)
	}

	// 保存AES密钥
	aesKeyEncoded := crypto.EncodeBase64(g.aesKey)
	err = os.WriteFile(aesKeyPath, []byte(aesKeyEncoded), 0600)
	if err != nil {
		return fmt.Errorf("保存AES密钥失败: %v", err)
	}

	return nil
}
