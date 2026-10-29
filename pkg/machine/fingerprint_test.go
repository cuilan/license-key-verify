package machine

import (
	"strings"
	"testing"
)

func TestComputeFingerprintDeterministic(t *testing.T) {
	a := ComputeFingerprint("uuid-1234", []string{"aa:bb:cc:dd:ee:ff", "11:22:33:44:55:66"})
	b := ComputeFingerprint("uuid-1234", []string{"aa:bb:cc:dd:ee:ff", "11:22:33:44:55:66"})
	if a != b {
		t.Errorf("fingerprint not deterministic: %s != %s", a, b)
	}
	if len(a) != 64 {
		t.Errorf("fingerprint length = %d, want 64 hex chars", len(a))
	}
}

func TestComputeFingerprintSensitive(t *testing.T) {
	base := ComputeFingerprint("uuid-1234", []string{"aa:bb:cc:dd:ee:ff"})

	diffID := ComputeFingerprint("uuid-5678", []string{"aa:bb:cc:dd:ee:ff"})
	if diffID == base {
		t.Error("different stable id should produce different fingerprint")
	}

	diffMAC := ComputeFingerprint("uuid-1234", []string{"11:22:33:44:55:66"})
	if diffMAC == base {
		t.Error("different MAC should produce different fingerprint")
	}

	// MAC 顺序不影响结果（采集侧已排序，这里验证拼接顺序敏感但调用方约定一致）
	reversed := ComputeFingerprint("uuid-1234", []string{"bb:bb", "aa:aa"})
	ordered := ComputeFingerprint("uuid-1234", []string{"aa:aa", "bb:bb"})
	if reversed == ordered {
		t.Error("MAC order should affect fingerprint (caller must sort)")
	}
}

func TestFingerprintFormat(t *testing.T) {
	fp, err := Fingerprint()
	if err != nil {
		t.Skipf("machine fingerprint unavailable in this environment: %v", err)
	}
	if len(fp) != 64 || strings.ToLower(fp) != fp {
		t.Errorf("fingerprint format invalid: %q", fp)
	}
	t.Logf("fingerprint: %s", fp)
}
