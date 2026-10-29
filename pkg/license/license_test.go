package license

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cuilan/license-key-verify/pkg/crypto"
	"github.com/cuilan/license-key-verify/pkg/machine"
)

// setupTestKey 生成测试密钥对并设置为当前验证公钥。
func setupTestKey(t *testing.T) (privateKeyPEM []byte) {
	t.Helper()
	keyPair, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}
	privateKeyPEM, err = keyPair.PrivateKeyToPEM()
	if err != nil {
		t.Fatalf("PrivateKeyToPEM() error = %v", err)
	}
	publicKeyPEM, err := keyPair.PublicKeyToPEM()
	if err != nil {
		t.Fatalf("PublicKeyToPEM() error = %v", err)
	}
	t.Cleanup(func() {
		// 回到"未内嵌公钥"基线（仓库不含演示密钥）
		keyMu.Lock()
		globalKey = nil
		keyMu.Unlock()
	})
	if err := UsePublicKeyPEM(publicKeyPEM); err != nil {
		t.Fatalf("UsePublicKeyPEM() error = %v", err)
	}
	return privateKeyPEM
}

func signTestLicense(t *testing.T, privateKeyPEM []byte, opts *SignOptions) *License {
	t.Helper()
	lic, err := Sign(opts, privateKeyPEM)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}
	return lic
}

func TestSignAndVerifyRoundTrip(t *testing.T) {
	privateKeyPEM := setupTestKey(t)

	localFP, err := machine.Fingerprint()
	if err != nil {
		t.Skipf("machine fingerprint unavailable: %v", err)
	}

	lic := signTestLicense(t, privateKeyPEM, &SignOptions{
		Customer:    "测试客户",
		Product:     "test-product",
		Fingerprint: localFP,
		Features:    []string{"full"},
		Duration:    24 * time.Hour,
	})

	data, err := lic.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	result := Verify(data)
	if !result.Valid {
		t.Fatalf("Verify() invalid, error = %s", result.Error)
	}
	if result.License.Customer != "测试客户" || result.License.Product != "test-product" {
		t.Errorf("license fields mismatch: %+v", result.License)
	}
	if !result.Matched {
		t.Error("fingerprint should match")
	}
	if result.ExpiresIn <= 0 || result.ExpiresIn > 24*3600 {
		t.Errorf("ExpiresIn out of range: %d", result.ExpiresIn)
	}
	if lic.LicenseID == "" {
		t.Error("license ID should be auto-generated")
	}
}

func TestVerifyRejectsTamperedLicense(t *testing.T) {
	privateKeyPEM := setupTestKey(t)

	lic := signTestLicense(t, privateKeyPEM, &SignOptions{
		Fingerprint: ComputeTestFingerprint(t),
		Duration:    time.Hour,
	})

	// 篡改客户名
	lic.Customer = "attacker"
	data, err := lic.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	result := Verify(data)
	if result.Valid {
		t.Fatal("tampered license should be rejected")
	}
	if !strings.Contains(result.Error, "signature") {
		t.Errorf("error should mention signature, got: %s", result.Error)
	}
	if result.License != nil {
		t.Error("license should not be populated for invalid signature")
	}
}

func TestVerifyRejectsWrongFingerprint(t *testing.T) {
	privateKeyPEM := setupTestKey(t)

	lic := signTestLicense(t, privateKeyPEM, &SignOptions{
		Fingerprint: strings.Repeat("a", 64), // 别的机器
		Duration:    time.Hour,
	})

	data, err := lic.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	result := Verify(data)
	if result.Valid {
		t.Fatal("license with foreign fingerprint should be rejected")
	}
	if result.Matched {
		t.Error("Matched should be false")
	}
	if !strings.Contains(result.Error, "fingerprint") {
		t.Errorf("error should mention fingerprint, got: %s", result.Error)
	}
}

func TestVerifyRejectsExpiredLicense(t *testing.T) {
	privateKeyPEM := setupTestKey(t)

	lic := signTestLicense(t, privateKeyPEM, &SignOptions{
		Fingerprint: ComputeTestFingerprint(t),
		IssuedAt:    time.Now().Add(-48 * time.Hour),
		Duration:    24 * time.Hour, // 已过期 24 小时
	})

	data, err := lic.Encode()
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	result := Verify(data)
	if result.Valid {
		t.Fatal("expired license should be rejected")
	}
	if !strings.Contains(result.Error, "expired") {
		t.Errorf("error should mention expiry, got: %s", result.Error)
	}
	if result.ExpiresIn != 0 {
		t.Errorf("ExpiresIn should be 0 for expired license, got %d", result.ExpiresIn)
	}
}

func TestSignRequiresFingerprint(t *testing.T) {
	privateKeyPEM := setupTestKey(t)
	if _, err := Sign(&SignOptions{Product: "x"}, privateKeyPEM); err == nil {
		t.Fatal("Sign() without fingerprint should fail")
	}
}

func TestVerifyFileAndSave(t *testing.T) {
	privateKeyPEM := setupTestKey(t)
	localFP, err := machine.Fingerprint()
	if err != nil {
		t.Skipf("machine fingerprint unavailable: %v", err)
	}

	lic := signTestLicense(t, privateKeyPEM, &SignOptions{
		Customer:    "file-test",
		Fingerprint: localFP,
		Duration:    time.Hour,
	})

	path := filepath.Join(t.TempDir(), "license.lic")
	if err := lic.Save(path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	result := VerifyFile(path)
	if !result.Valid {
		t.Fatalf("VerifyFile() invalid, error = %s", result.Error)
	}

	// 文件内容应为符合规范的明文 JSON（包含 signature 字段）
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	for _, key := range []string{`"license_id"`, `"fingerprint"`, `"signature"`, `"expires_at"`} {
		if !strings.Contains(string(raw), key) {
			t.Errorf("license file missing field %s", key)
		}
	}
}

func TestDefaultLicensePathEnv(t *testing.T) {
	t.Setenv("LICENSE_FILE", "/tmp/some/license.lic")
	if got := DefaultLicensePath(); got != "/tmp/some/license.lic" {
		t.Errorf("DefaultLicensePath() = %q, want env value", got)
	}
}

// ComputeTestFingerprint 返回测试用本机指纹（采集失败时返回占位值）。
func ComputeTestFingerprint(t *testing.T) string {
	t.Helper()
	fp, err := machine.Fingerprint()
	if err != nil {
		return strings.Repeat("0", 64) // 本机无指纹：验证会在签名后因指纹不匹配而失败
	}
	return fp
}

func TestEmbeddedKeyLifecycle(t *testing.T) {
	// 基线：无内嵌公钥（模拟未 embed 的全新 clone）
	keyMu.Lock()
	globalKey = nil
	keyMu.Unlock()

	if HasEmbeddedKey() {
		t.Fatal("HasEmbeddedKey() should be false before UsePublicKeyPEM")
	}
	if _, err := EmbeddedKeyPEM(); err == nil {
		t.Fatal("EmbeddedKeyPEM() should fail before UsePublicKeyPEM")
	}

	// 注入公钥后可用，且 PEM 可被还原
	keyPair, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}
	publicKeyPEM, err := keyPair.PublicKeyToPEM()
	if err != nil {
		t.Fatalf("PublicKeyToPEM() error = %v", err)
	}
	if err := UsePublicKeyPEM(publicKeyPEM); err != nil {
		t.Fatalf("UsePublicKeyPEM() error = %v", err)
	}
	if !HasEmbeddedKey() {
		t.Fatal("HasEmbeddedKey() should be true after UsePublicKeyPEM")
	}
	pem, err := EmbeddedKeyPEM()
	if err != nil {
		t.Fatalf("EmbeddedKeyPEM() error = %v", err)
	}
	if _, err := crypto.LoadPublicKeyFromPEM(pem); err != nil {
		t.Fatalf("embedded PEM should be loadable: %v", err)
	}
}
