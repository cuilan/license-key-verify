package machine

import (
	"net"
	"reflect"
	"strings"
	"testing"
)

// fakeIface 构造一个用于测试的网卡描述。
func fakeIface(name, mac string, flags net.Flags) net.Interface {
	hw, _ := net.ParseMAC(mac)
	return net.Interface{Name: name, HardwareAddr: hw, Flags: flags}
}

// TestInterfaceMACsFiltersVirtual 验证虚拟网卡被排除、物理网卡（含 DOWN）被保留。
func TestInterfaceMACsFiltersVirtual(t *testing.T) {
	physical := map[string]bool{
		"eth0": true,
		"eth1": true, // 新插的网卡，处于 DOWN 也要算
	}
	ifaces := []net.Interface{
		fakeIface("lo", "00:00:00:00:00:00", net.FlagLoopback|net.FlagUp),
		fakeIface("eth0", "AA:BB:CC:DD:EE:01", net.FlagUp),
		fakeIface("eth1", "AA:BB:CC:DD:EE:02", 0), // DOWN
		fakeIface("docker0", "02:42:AC:11:00:01", net.FlagUp),
		fakeIface("veth1a2b3c", "7E:11:22:33:44:55", net.FlagUp),
		fakeIface("br-9f8e7d", "02:42:00:11:22:33", net.FlagUp),
	}

	got := interfaceMACs(ifaces, func(name string, _ net.Interface) bool {
		return physical[name]
	})
	want := []string{"aa:bb:cc:dd:ee:01", "aa:bb:cc:dd:ee:02"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("interfaceMACs = %v, want %v", got, want)
	}
}

// TestInterfaceMACsSkipsEmptyMAC 验证没有 MAC 的物理网卡被跳过（如某些隧道/PPP 设备）。
func TestInterfaceMACsSkipsEmptyMAC(t *testing.T) {
	ifaces := []net.Interface{
		{Name: "eth0", Flags: net.FlagUp}, // HardwareAddr 为空
		fakeIface("eth1", "00:00:00:00:00:00", net.FlagUp),
		fakeIface("eth2", "AA:BB:CC:DD:EE:03", net.FlagUp),
	}
	got := interfaceMACs(ifaces, func(string, net.Interface) bool { return true })
	want := []string{"aa:bb:cc:dd:ee:03"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("interfaceMACs = %v, want %v", got, want)
	}
}

// TestInterfaceMACsAddingNICChangesSet 验证"新增物理网卡必然改变 MAC 集合"
// ——这是许可证失效的核心保障。
func TestInterfaceMACsAddingNICChangesSet(t *testing.T) {
	all := func(string, net.Interface) bool { return true }
	before := interfaceMACs([]net.Interface{
		fakeIface("eth0", "AA:BB:CC:DD:EE:01", net.FlagUp),
	}, all)
	after := interfaceMACs([]net.Interface{
		fakeIface("eth0", "AA:BB:CC:DD:EE:01", net.FlagUp),
		fakeIface("eth1", "AA:BB:CC:DD:EE:02", 0),
	}, all)
	if reflect.DeepEqual(before, after) {
		t.Fatal("adding a physical NIC must change the MAC set")
	}
}

// TestVirtualPrefixes 验证 Windows/macOS 的虚拟网卡前缀表确实命中常见设备名。
func TestVirtualPrefixes(t *testing.T) {
	for _, name := range []string{"vEthernet (WSL)", "Loopback Pseudo-Interface 1", "VMware Network Adapter VMnet8"} {
		if !hasAnyPrefix(strings.ToLower(name), windowsVirtualPrefixes) {
			t.Errorf("windows virtual adapter %q not filtered", name)
		}
	}
	for _, name := range []string{"utun0", "bridge0", "awdl0"} {
		if !hasAnyPrefix(strings.ToLower(name), darwinVirtualPrefixes) {
			t.Errorf("darwin virtual adapter %q not filtered", name)
		}
	}
	if !hasAnyPrefix("eth0", []string{"eth"}) {
		t.Error("hasAnyPrefix should match eth0 against eth")
	}
	if hasAnyPrefix("eth0", []string{"x", "y"}) {
		t.Error("hasAnyPrefix should not match eth0 against x/y")
	}
}

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
