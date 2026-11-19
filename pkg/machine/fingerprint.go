// Package machine 采集机器指纹。
// 指纹由「稳定标识（平台 UUID / machine-id）+ 所有物理网卡 MAC（排序后拼接）」
// 经 SHA-256 计算得到，同一台机器重复计算结果一致，不同机器结果不同。
//
// 「物理网卡」是刻意收紧的口径（见 isPhysicalInterface）：
//   - 只统计挂在真实总线上的网卡，虚拟网卡一律排除。容器启停产生的 veth、
//     Docker 网桥、bond/VLAN、tun/tap、VPN 隧道都不会让指纹抖动；
//   - 统计时不看 UP/DOWN 状态。插上一块新网卡（哪怕没插网线、没配 IP）
//     指纹立刻变化，许可证随之失效——这是"锁死硬件"的预期行为。
//
// 容器内场景：
//
//	默认宿主标识路径在容器中不可见。如果容器通过 bind mount 暴露了宿主硬件信息，
//	优先读取 /host/* 下的路径（与 Dockerfile 建议的 --mount 配合）：
//
//	  docker run \
//	    --mount type=bind,source=/etc/machine-id,target=/host/etc/machine-id,readonly \
//	    --mount type=bind,source=/sys/class/dmi/id,target=/host/sys/class/dmi/id,readonly \
//	    --network host \
//	    your-image
//
//	这样容器内指纹 = 宿主裸机指纹，docker save 到新宿主机后指纹必然不同。
package machine

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// fingerprintVersion 指纹算法版本号，参与哈希计算。
// 修改采集逻辑时递增，避免新旧版本指纹混淆。
//
//	lkfp-v1  统计所有非回环且 UP 的网卡（含 veth / docker0 等虚拟网卡）
//	lkfp-v2  只统计物理网卡，且不看 UP/DOWN（见 isPhysicalInterface）
const fingerprintVersion = "lkfp-v2"

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
			"/sys/class/dmi/id/product_uuid",                   // 容器内 DMI
			hostMountPrefix + "/etc/machine-id",                // 宿主 machine-id（通过 bind mount）
			"/etc/machine-id",                                  // 容器自身
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

// MACAddresses 返回所有物理网卡的 MAC 地址（小写、排序），保证顺序稳定。
//
// 与 v1 的区别：
//   - 只看物理网卡（isPhysicalInterface），虚拟网卡不参与；
//   - 不看 UP/DOWN，网卡插上但没启用也算。
//
// 这样容器重建、Docker 启停、VPN 连接都不会改变指纹，
// 而物理上新增/拆除/更换一块网卡会立刻改变指纹。
func MACAddresses() ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %v", err)
	}
	return interfaceMACs(interfaces, isPhysicalInterface), nil
}

// interfaceMACs 从网卡列表中挑出物理网卡的 MAC。
// 拆成纯函数便于测试（注入 isPhysical，不需要真实网卡）。
func interfaceMACs(ifaces []net.Interface, isPhysical func(name string, iface net.Interface) bool) []string {
	var macs []string
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if !isPhysical(iface.Name, iface) {
			continue
		}
		mac := strings.ToLower(iface.HardwareAddr.String())
		if mac == "" || mac == "00:00:00:00:00:00" {
			continue
		}
		macs = append(macs, mac)
	}
	sort.Strings(macs)
	return macs
}

// isPhysicalInterface 判断是否是物理网卡。
//
// Linux 的判据是 /sys/class/net/<name>/device 这个符号链接是否存在：
// 只有挂在真实总线（PCI / USB / platform）上的网卡才有，它指向网卡的总线设备。
// veth、docker0、br-*、bond 主设备、VLAN 子接口、tun/tap、virbr、tailscale0、
// wg0 这类虚拟设备都没有该链接，因此一律被排除——容器启停不会让指纹抖动。
//
// 注意边界情况：
//   - bond 主设备被排除，但从设备（eth0/eth1…）各自都被统计，
//     所以给 bond 增加一块从网卡同样会改变指纹；
//   - VLAN 子接口被排除，物理父接口会被统计；
//   - 容器内（除非 privileged 且 /sys 完整映射）通常读不到 device 链接，
//     此时退化为「只用稳定标识」，与 v1 的容器行为一致。
//
// Windows / macOS 没有 /sys，退回按网卡名的虚拟设备前缀排除（见各自列表）。
func isPhysicalInterface(name string, iface net.Interface) bool {
	switch runtime.GOOS {
	case "linux":
		_, err := os.Stat(filepath.Join("/sys/class/net", name, "device"))
		return err == nil
	case "windows":
		return !hasAnyPrefix(strings.ToLower(name), windowsVirtualPrefixes)
	case "darwin":
		return !hasAnyPrefix(strings.ToLower(name), darwinVirtualPrefixes)
	default:
		return true
	}
}

// windowsVirtualPrefixes Windows 上虚拟网卡的名字前缀（小写比较）。
// Hyper-V / WSL / VPN / 虚拟化软件创建的适配器都带这些前缀。
var windowsVirtualPrefixes = []string{
	"vethernet", "loopback", "virtual", "hyper-v", "vmware", "virtualbox",
	"tap-", "npcap", "tailscale", "zerotier", "wireguard", "bluetooth", "wsl",
}

// darwinVirtualPrefixes macOS 上虚拟网卡的名字前缀（小写比较）。
// en0/en1（有线/无线）是真机网卡，其余多为隧道或桥接设备。
var darwinVirtualPrefixes = []string{
	"utun", "bridge", "awdl", "llw", "ap", "gif", "stf", "p2p", "anpi", "vmenet",
}

// hasAnyPrefix 判断 s 是否以 prefixes 中任一项开头。
func hasAnyPrefix(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
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
