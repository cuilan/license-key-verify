package license

import (
	"time"
)

// License 许可证内容（明文 JSON + 签名，无加密层）。
// 签名覆盖除 Signature 外的所有字段：序列化时 Signature 为空（omitempty 不输出），
// 验证时将其置空后重新序列化即可得到相同的被签名字节。
type License struct {
	LicenseID   string    `json:"license_id"`
	Customer    string    `json:"customer"`
	Product     string    `json:"product"`
	Fingerprint string    `json:"fingerprint"`
	IssuedAt    time.Time `json:"issued_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	Features    []string  `json:"features,omitempty"`
	Signature   string    `json:"signature,omitempty"` // Base64 编码的 RSA-SHA256 签名
}

// Result 验证结果
type Result struct {
	Valid       bool     `json:"valid"`             // 是否有效（签名 + 时间 + 指纹全部通过）
	License     *License `json:"license,omitempty"` // 许可证内容（签名通过后填充）
	Error       string   `json:"error,omitempty"`   // 失败原因
	VerifiedAt  string   `json:"verified_at"`       // 验证时间（RFC3339）
	ExpiresIn   int64    `json:"expires_in"`        // 剩余有效期（秒），已过期为 0
	Fingerprint string   `json:"fingerprint"`       // 当前机器指纹
	Matched     bool     `json:"matched"`           // 指纹是否匹配
}

// SignOptions 签发选项
type SignOptions struct {
	LicenseID   string        // 可选，留空自动生成
	Customer    string        // 客户名称
	Product     string        // 产品名称
	Fingerprint string        // 目标机器指纹（必填，lkctl fingerprint 或应用 --fingerprint 获取）
	Features    []string      // 授权的功能列表
	Duration    time.Duration // 有效期长度，默认 365 天
	IssuedAt    time.Time     // 签发时间，默认当前时间
}

// 常量
const (
	DefaultProduct  = "License Key Verify Tool"
	DefaultDuration = 365 * 24 * time.Hour
)
