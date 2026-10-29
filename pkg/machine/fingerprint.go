// Package machine 采集机器指纹。
// 指纹由「稳定标识（平台 UUID / machine-id）+ 所有物理网卡 MAC（排序后拼接）」
// 经 SHA-256 计算得到，同一台机器重复计算结果一致，不同机器结果不同。
//
// 容器内场景：
//   默认宿主标识路径在容器中不可见。如果容器通过 bind mount 暴露了宿主硬件信息，
//   优先读取 /host/* 下的路径（与 Dockerfile 建议的 --mount 配合）：
//
//     docker run \
//       --mount type=bind,source=/etc/machine-id,target=/host/etc/machine-id,readonly \
//       --mount type=bind,source=/sys/class/dmi/id,target=/host/sys/class/dmi/id,readonly \
//       --network host \
//       your-image
//
//   这样容器内指纹 = 宿主裸机指纹，docker save 到新宿主机后指纹必然不同。
package machine

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
)

// fingerprintVersion 指纹算法版本号，参与哈希计算。
// 修改采集逻辑时递增，避免新旧版本指纹混淆。
const fingerprintVersion = "lkfp-v1"

// 宿主信息 bind mount 路径前缀。容器启动时可通过 --mount 把宿主硬件信息挂到 /host/ 下。
// 例如 --mount type=bind,source=/etc/machine-id,target=/host/etc/machine-id,readonly
const hostMountPrefix = "/host"

// StableID 返回机器的稳定标识。
// Linux 下优先查宿主 bind mount 路径（/host/*），再回退容器内路径：
//
//	/host/sys/class/dmi/id/product_uuid  (宿主 DMI，最稳定)
//	/sys/class/dmi/id/product_uuid       (容器内 DMI，通常不可读)
//	/host/etc/machine-id                 (宿主 machine-id)
//	/etc/machine-id                      (容器自身，可能随容器重建变化)
//	/var/lib/dbus/machine-id
//
// 不使用 /proc/sys/kernel/random/uuid 之类的随机值作为回退。
func StableID() (string, error) {
	switch runtime.GOOS {
	case "linux":
		paths := []string{
			hostMountPrefix + "/sys/class/dmi/id/product_uuid", // 宿主 DMI（通过 bind mount）
			"/sys/class/dmi/id/product_uuid",                  // 容器内 DMI
			hostMountPrefix + "/etc/machine-id",               // 宿主 machine-id（通过 bind mount）
			"/etc/machine-id",                                 // 容器自身
			"/var/lib/dbus/machine-id",
		}
		for _, path := range paths {
			if data, err := os.ReadFile(path); err == nil {
				if id := strings.ToLower(strings.TrimSpace(string(data))); id != "" {
					return id, nil
				}
			}
		}
		return "", fmt.Errorf("no stable machine id found (tried product_uuid, machine-id, host-mount variants)")
	case "darwin":
		output, err := exec.Command("system_profiler", "SPHardwareDataType").Output()
		if err != nil {
			return "", fmt.Errorf("failed to execute system_profiler: %v", err)
		}
		for _, line := range strings.Split(string(output), "\n") {
			if strings.Contains(line, "Hardware UUID:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					if id := strings.TrimSpace(parts[1]); id != "" {
						return strings.ToLower(id), nil
					}
				}
			}
		}
		return "", fmt.Errorf("hardware UUID not found in system_profiler output")
	case "windows":
		output, err := exec.Command("powershell", "-NoProfile", "-Command",
			"(Get-CimInstance -ClassName Win32_ComputerSystemProduct).UUID").Output()
		if err != nil {
			return "", fmt.Errorf("failed to execute powershell: %v", err)
		}
		if id := strings.ToLower(strings.TrimSpace(string(output))); id != "" {
			return id, nil
		}
		return "", fmt.Errorf("system UUID is empty")
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// MACAddresses 返回所有 UP 状态、非回环网卡的 MAC 地址（小写、排序），保证顺序稳定。
//
// 容器内如果用了 --network host，net.Interfaces() 直接返回宿主物理网卡的 MAC——
// 这时容器内指纹 = 宿主裸机指纹，docker save 到新宿主机后必然不同。
// 如果用的是容器默认 bridge 网络，拿到的是容器 veth pair 的 MAC，每次 docker run 可能不同。
func MACAddresses() ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %v", err)
	}

	var macs []string
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		mac := strings.ToLower(iface.HardwareAddr.String())
		if mac != "" && mac != "00:00:00:00:00:00" {
			macs = append(macs, mac)
		}
	}
	sort.Strings(macs)
	return macs, nil
}

// Fingerprint 计算当前机器的指纹（64 位十六进制 SHA-256）。
// 稳定标识与 MAC 至少需要一项可用。
func Fingerprint() (string, error) {
	stableID, _ := StableID()
	macs, _ := MACAddresses()
	if stableID == "" && len(macs) == 0 {
		return "", fmt.Errorf("failed to collect machine fingerprint: no stable id and no MAC address")
	}
	return ComputeFingerprint(stableID, macs), nil
}

// ComputeFingerprint 由稳定标识与 MAC 列表计算指纹（纯函数，便于测试）。
func ComputeFingerprint(stableID string, macs []string) string {
	payload := fingerprintVersion + "|" + stableID + "|" + strings.Join(macs, ",")
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}
