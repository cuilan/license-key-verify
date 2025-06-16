package machine

import (
	"testing"
)

func TestGetMACAddress(t *testing.T) {
	mac, err := GetMACAddress()
	if err != nil {
		t.Errorf("GetMACAddress() error = %v", err)
		return
	}
	if mac == "" {
		t.Error("GetMACAddress() returned empty string")
	}
	t.Logf("MAC Address: %s", mac)
}

func TestGetSystemUUID(t *testing.T) {
	uuid, err := GetSystemUUID()
	if err != nil {
		t.Errorf("GetSystemUUID() error = %v", err)
		return
	}
	if uuid == "" {
		t.Error("GetSystemUUID() returned empty string")
	}
	t.Logf("System UUID: %s", uuid)
}

func TestGetCPUID(t *testing.T) {
	cpuid, err := GetCPUID()
	if err != nil {
		t.Errorf("GetCPUID() error = %v", err)
		return
	}
	if cpuid == "" {
		t.Error("GetCPUID() returned empty string")
	}
	t.Logf("CPU ID: %s", cpuid)
}

func TestGetAllInfo(t *testing.T) {
	info, err := GetAllInfo()
	if err != nil {
		t.Errorf("GetAllInfo() error = %v", err)
		return
	}

	if info.MAC == "" {
		t.Error("GetAllInfo() returned empty MAC")
	}
	if info.UUID == "" {
		t.Error("GetAllInfo() returned empty UUID")
	}
	if info.CPUID == "" {
		t.Error("GetAllInfo() returned empty CPUID")
	}

	t.Logf("Machine Info: MAC=%s, UUID=%s, CPUID=%s", info.MAC, info.UUID, info.CPUID)
}
