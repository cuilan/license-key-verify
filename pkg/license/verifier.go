package license

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/cuilan/license-key-verify/pkg/crypto"
	"github.com/cuilan/license-key-verify/pkg/machine"
)

// Verifier 许可证验证器
type Verifier struct {
	publicKey *rsa.PublicKey
	aesKey    []byte
}

// NewVerifier 创建新的验证器
func NewVerifier(publicKeyPEM []byte, aesKey []byte) (*Verifier, error) {
	publicKey, err := crypto.LoadPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key: %v", err)
	}

	return &Verifier{
		publicKey: publicKey,
		aesKey:    aesKey,
	}, nil
}

// NewVerifierFromFiles 从文件创建验证器
func NewVerifierFromFiles(publicKeyPath, aesKeyPath string) (*Verifier, error) {
	// 读取公钥
	publicKeyPEM, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %v", err)
	}

	// 读取AES密钥
	aesKeyEncoded, err := os.ReadFile(aesKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read AES key file: %v", err)
	}

	aesKey, err := crypto.DecodeBase64(string(aesKeyEncoded))
	if err != nil {
		return nil, fmt.Errorf("failed to decode AES key: %v", err)
	}

	return NewVerifier(publicKeyPEM, aesKey)
}

// VerifyFile 验证许可证文件
func (v *Verifier) VerifyFile(filePath string) (*VerificationResult, error) {
	// 读取许可证文件
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return &VerificationResult{
			Valid:      false,
			Error:      fmt.Sprintf("failed to read license file: %v", err),
			VerifiedAt: time.Now(),
		}, nil
	}

	return v.Verify(fileData)
}

// Verify 验证许可证数据
func (v *Verifier) Verify(fileData []byte) (*VerificationResult, error) {
	result := &VerificationResult{
		VerifiedAt: time.Now(),
	}

	// 解析许可证文件
	var licenseFile LicenseFile
	err := json.Unmarshal(fileData, &licenseFile)
	if err != nil {
		result.Error = fmt.Sprintf("failed to parse license file: %v", err)
		return result, nil
	}

	// 检查文件格式版本
	if licenseFile.Version != FileFormatVersion {
		result.Error = fmt.Sprintf("unsupported file format version: %s", licenseFile.Version)
		return result, nil
	}

	// 解码数据和签名
	encryptedData, err := crypto.DecodeBase64(licenseFile.Data)
	if err != nil {
		result.Error = fmt.Sprintf("failed to decode license data: %v", err)
		return result, nil
	}

	signature, err := crypto.DecodeBase64(licenseFile.Signature)
	if err != nil {
		result.Error = fmt.Sprintf("failed to decode signature: %v", err)
		return result, nil
	}

	// 验证签名
	err = crypto.VerifySignature(encryptedData, signature, v.publicKey)
	if err != nil {
		result.Error = fmt.Sprintf("signature verification failed: %v", err)
		return result, nil
	}

	// 解密许可证数据
	licenseData, err := crypto.DecryptAES(encryptedData, v.aesKey)
	if err != nil {
		result.Error = fmt.Sprintf("failed to decrypt license data: %v", err)
		return result, nil
	}

	// 解析许可证
	var license License
	err = json.Unmarshal(licenseData, &license)
	if err != nil {
		result.Error = fmt.Sprintf("failed to parse license: %v", err)
		return result, nil
	}

	result.License = &license

	// 检查时间有效性
	now := time.Now()
	if now.Before(license.IssuedAt) {
		result.Error = "license is not yet valid"
		return result, nil
	}

	if now.After(license.ExpiresAt) {
		result.Error = "license has expired"
		result.ExpiresIn = 0
		return result, nil
	}

	result.ExpiresIn = int64(license.ExpiresAt.Sub(now).Seconds())

	// 获取当前机器信息
	machineInfo, err := machine.GetAllInfo()
	if err != nil {
		result.Error = fmt.Sprintf("failed to get machine info: %v", err)
		return result, nil
	}

	result.MachineInfo.MAC = machineInfo.MAC
	result.MachineInfo.UUID = machineInfo.UUID
	result.MachineInfo.CPUID = machineInfo.CPUID

	// 验证机器信息
	machineMatched := true

	if license.MAC != "" && license.MAC != machineInfo.MAC {
		machineMatched = false
	}

	if license.UUID != "" && license.UUID != machineInfo.UUID {
		machineMatched = false
	}

	if license.CPUID != "" && license.CPUID != machineInfo.CPUID {
		machineMatched = false
	}

	result.MachineInfo.Matched = machineMatched

	if !machineMatched {
		result.Error = "machine information does not match"
		return result, nil
	}

	// 验证通过
	result.Valid = true
	return result, nil
}

// GetLicenseInfo 获取许可证信息（不验证机器信息）
func (v *Verifier) GetLicenseInfo(filePath string) (*License, error) {
	// 读取许可证文件
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read license file: %v", err)
	}

	// 解析许可证文件
	var licenseFile LicenseFile
	err = json.Unmarshal(fileData, &licenseFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse license file: %v", err)
	}

	// 解码数据和签名
	encryptedData, err := crypto.DecodeBase64(licenseFile.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode license data: %v", err)
	}

	signature, err := crypto.DecodeBase64(licenseFile.Signature)
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature: %v", err)
	}

	// 验证签名
	err = crypto.VerifySignature(encryptedData, signature, v.publicKey)
	if err != nil {
		return nil, fmt.Errorf("signature verification failed: %v", err)
	}

	// 解密许可证数据
	licenseData, err := crypto.DecryptAES(encryptedData, v.aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt license data: %v", err)
	}

	// 解析许可证
	var license License
	err = json.Unmarshal(licenseData, &license)
	if err != nil {
		return nil, fmt.Errorf("failed to parse license: %v", err)
	}

	return &license, nil
}

// QuickVerify 快速验证（仅返回是否有效）
func (v *Verifier) QuickVerify(filePath string) bool {
	result, err := v.VerifyFile(filePath)
	if err != nil {
		return false
	}
	return result.Valid
}
