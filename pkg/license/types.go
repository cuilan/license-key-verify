package license

import (
	"time"
)

// License 许可证结构体
type License struct {
	// 基本信息
	ID          string `json:"id"`           // 许可证ID
	ProductName string `json:"product_name"` // 产品名称
	Version     string `json:"version"`      // 版本

	// 机器信息
	MAC   string `json:"mac"`   // MAC地址
	UUID  string `json:"uuid"`  // 系统UUID
	CPUID string `json:"cpuid"` // CPU ID

	// 时间信息
	IssuedAt  time.Time `json:"issued_at"`  // 签发时间
	ExpiresAt time.Time `json:"expires_at"` // 过期时间

	// 功能限制
	Features []string `json:"features"`  // 允许的功能列表
	MaxUsers int      `json:"max_users"` // 最大用户数

	// 其他信息
	CustomerName string                 `json:"customer_name"` // 客户名称
	Notes        string                 `json:"notes"`         // 备注
	Extra        map[string]interface{} `json:"extra"`         // 扩展字段
}

// LicenseFile 许可证文件结构
type LicenseFile struct {
	Data      string `json:"data"`      // 加密的许可证数据
	Signature string `json:"signature"` // 数字签名
	Algorithm string `json:"algorithm"` // 加密算法
	Version   string `json:"version"`   // 文件格式版本
}

// VerificationResult 验证结果
type VerificationResult struct {
	Valid       bool      `json:"valid"`       // 是否有效
	License     *License  `json:"license"`     // 许可证信息
	Error       string    `json:"error"`       // 错误信息
	VerifiedAt  time.Time `json:"verified_at"` // 验证时间
	ExpiresIn   int64     `json:"expires_in"`  // 剩余有效期（秒）
	MachineInfo struct {
		MAC     string `json:"mac"`     // 当前机器MAC
		UUID    string `json:"uuid"`    // 当前机器UUID
		CPUID   string `json:"cpuid"`   // 当前机器CPU ID
		Matched bool   `json:"matched"` // 机器信息是否匹配
	} `json:"machine_info"`
}

// GenerateOptions 生成许可证选项
type GenerateOptions struct {
	// 基本信息
	ProductName  string
	Version      string
	CustomerName string
	Notes        string

	// 机器信息
	MAC   string
	UUID  string
	CPUID string

	// 时间设置
	Duration time.Duration // 有效期长度

	// 功能设置
	Features []string
	MaxUsers int

	// 扩展字段
	Extra map[string]interface{}
}

// Constants
const (
	DefaultProductName = "License Key Verify Tool"
	DefaultVersion     = "1.0.0"
	FileFormatVersion  = "1.0"
	DefaultAlgorithm   = "AES256-GCM+RSA2048"
)
