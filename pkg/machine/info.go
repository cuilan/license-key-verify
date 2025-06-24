package machine

import (
	"crypto/md5"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// MachineInfo 机器信息结构体
type MachineInfo struct {
	MAC   string `json:"mac"`
	UUID  string `json:"uuid"`
	CPUID string `json:"cpuid"`
}

// GetMACAddress 获取MAC地址
func GetMACAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("获取网络接口失败: %v", err)
	}

	for _, iface := range interfaces {
		// 过滤掉回环接口和虚拟接口
		if iface.Flags&net.FlagLoopback == 0 && iface.Flags&net.FlagUp != 0 {
			mac := iface.HardwareAddr.String()
			if mac != "" && mac != "00:00:00:00:00:00" {
				return mac, nil
			}
		}
	}

	return "", fmt.Errorf("未找到有效的MAC地址")
}

// GetSystemUUID 获取系统UUID
func GetSystemUUID() (string, error) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("system_profiler", "SPHardwareDataType")
	case "linux":
		// 尝试多种方式获取UUID
		if _, err := os.Stat("/sys/class/dmi/id/product_uuid"); err == nil {
			cmd = exec.Command("cat", "/sys/class/dmi/id/product_uuid")
		} else if _, err := os.Stat("/proc/sys/kernel/random/uuid"); err == nil {
			cmd = exec.Command("cat", "/proc/sys/kernel/random/uuid")
		} else {
			return "", fmt.Errorf("无法获取系统UUID")
		}
	case "windows":
		// 使用PowerShell替代已弃用的wmic命令
		cmd = exec.Command("powershell", "-Command", "Get-CimInstance -ClassName Win32_ComputerSystemProduct | Select-Object -ExpandProperty UUID")
	default:
		return "", fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("执行命令失败: %v", err)
	}

	uuid := parseUUID(string(output), runtime.GOOS)
	if uuid == "" {
		return "", fmt.Errorf("无法解析UUID")
	}

	return uuid, nil
}

// GetCPUID 获取CPU ID
func GetCPUID() (string, error) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("sysctl", "-n", "machdep.cpu.brand_string")
	case "linux":
		cmd = exec.Command("cat", "/proc/cpuinfo")
	case "windows":
		// 使用PowerShell替代已弃用的wmic命令
		cmd = exec.Command("powershell", "-Command", "Get-CimInstance -ClassName Win32_Processor | Select-Object -ExpandProperty ProcessorId")
	default:
		return "", fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("执行命令失败: %v", err)
	}

	cpuid := parseCPUID(string(output), runtime.GOOS)
	if cpuid == "" {
		return "", fmt.Errorf("无法解析CPU ID")
	}

	// 生成CPU ID的MD5哈希值作为唯一标识
	hash := md5.Sum([]byte(cpuid))
	return fmt.Sprintf("%x", hash), nil
}

// parseUUID 解析UUID
func parseUUID(output, os string) string {
	lines := strings.Split(output, "\n")

	switch os {
	case "darwin":
		for _, line := range lines {
			if strings.Contains(line, "Hardware UUID:") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					return strings.TrimSpace(parts[1])
				}
			}
		}
	case "linux":
		return strings.TrimSpace(output)
	case "windows":
		// PowerShell输出格式直接返回UUID值
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.Contains(line, "UUID") && !strings.Contains(line, "---") {
				return line
			}
		}
	}

	return ""
}

// parseCPUID 解析CPU ID
func parseCPUID(output, os string) string {
	switch os {
	case "darwin":
		return strings.TrimSpace(output)
	case "linux":
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "processor") && strings.Contains(line, "0") {
				return strings.TrimSpace(line)
			}
		}
	case "windows":
		// PowerShell输出格式直接返回ProcessorId值
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.Contains(line, "ProcessorId") && !strings.Contains(line, "---") {
				return line
			}
		}
	}

	// 如果无法解析，返回原始输出的前100个字符
	if len(output) > 100 {
		return output[:100]
	}
	return output
}

// GetAllInfo 获取所有机器信息
func GetAllInfo() (*MachineInfo, error) {
	info := &MachineInfo{}

	mac, err := GetMACAddress()
	if err == nil {
		info.MAC = mac
	}

	uuid, err := GetSystemUUID()
	if err == nil {
		info.UUID = uuid
	}

	cpuid, err := GetCPUID()
	if err == nil {
		info.CPUID = cpuid
	}

	return info, nil
}
