package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cuilan/license-key-verify/pkg/license"
)

var (
	Version = "dev" // 将在构建时通过 -ldflags 设置
)

const (
	Usage = `lkverify - License Key Verification Tool

用法:
  lkverify <许可证文件> [选项]

选项:
  --keys-dir <目录>     指定密钥文件目录（默认: keys）
  --json               以JSON格式输出结果
  --quiet              安静模式，只输出退出码
  --version            显示版本
  --help               显示帮助

退出码:
  0  许可证有效
  1  许可证无效或其他错误
  2  参数错误

示例:
  lkverify license.lic
  lkverify license.lic --json
  lkverify license.lic --keys-dir ./mykeys
`
)

type Config struct {
	LicenseFile string
	KeysDir     string
	JSONOutput  bool
	Quiet       bool
}

func main() {
	config := parseArgs()

	// 创建验证器
	publicKeyPath := config.KeysDir + "/public.pem"
	aesKeyPath := config.KeysDir + "/aes.key"

	verifier, err := license.NewVerifierFromFiles(publicKeyPath, aesKeyPath)
	if err != nil {
		if !config.Quiet {
			fmt.Fprintf(os.Stderr, "创建验证器失败: %v\n", err)
			fmt.Fprintf(os.Stderr, "请确保密钥文件存在: %s, %s\n", publicKeyPath, aesKeyPath)
		}
		os.Exit(1)
	}

	// 验证许可证
	result, err := verifier.VerifyFile(config.LicenseFile)
	if err != nil {
		if !config.Quiet {
			fmt.Fprintf(os.Stderr, "验证失败: %v\n", err)
		}
		os.Exit(1)
	}

	// 输出结果
	if config.JSONOutput {
		// JSON格式输出
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			if !config.Quiet {
				fmt.Fprintf(os.Stderr, "序列化结果失败: %v\n", err)
			}
			os.Exit(1)
		}
		fmt.Println(string(data))
	} else if !config.Quiet {
		// 人类可读格式输出
		printResult(result)
	}

	// 设置退出码
	if result.Valid {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

func parseArgs() *Config {
	config := &Config{
		KeysDir: "keys",
	}

	args := os.Args[1:]

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "--help":
			fmt.Print(Usage)
			os.Exit(0)
		case "--version":
			fmt.Printf("lkverify version %s\n", Version)
			os.Exit(0)
		case "--json":
			config.JSONOutput = true
		case "--quiet":
			config.Quiet = true
		case "--keys-dir":
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "--keys-dir 需要指定目录\n")
				os.Exit(2)
			}
			i++
			config.KeysDir = args[i]
		default:
			if arg[0] == '-' {
				fmt.Fprintf(os.Stderr, "未知选项: %s\n", arg)
				os.Exit(2)
			} else {
				if config.LicenseFile == "" {
					config.LicenseFile = arg
				} else {
					fmt.Fprintf(os.Stderr, "只能指定一个许可证文件\n")
					os.Exit(2)
				}
			}
		}
	}

	if config.LicenseFile == "" {
		fmt.Fprintf(os.Stderr, "必须指定许可证文件\n")
		fmt.Print(Usage)
		os.Exit(2)
	}

	return config
}

func printResult(result *license.VerificationResult) {
	if result.Valid {
		fmt.Println("✓ 许可证验证通过")

		if result.License != nil {
			fmt.Printf("许可证ID: %s\n", result.License.ID)
			fmt.Printf("产品名称: %s\n", result.License.ProductName)

			if result.License.CustomerName != "" {
				fmt.Printf("客户名称: %s\n", result.License.CustomerName)
			}

			fmt.Printf("签发时间: %s\n", result.License.IssuedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("有效期至: %s\n", result.License.ExpiresAt.Format("2006-01-02 15:04:05"))

			if result.ExpiresIn > 0 {
				days := result.ExpiresIn / (24 * 3600)
				hours := (result.ExpiresIn % (24 * 3600)) / 3600
				fmt.Printf("剩余时间: %d 天 %d 小时\n", days, hours)
			}

			if len(result.License.Features) > 0 {
				fmt.Printf("功能列表: %v\n", result.License.Features)
			}

			if result.License.MaxUsers > 0 {
				fmt.Printf("最大用户数: %d\n", result.License.MaxUsers)
			}

			if result.License.Notes != "" {
				fmt.Printf("备注: %s\n", result.License.Notes)
			}
		}

		// 机器信息匹配状态
		if result.MachineInfo.Matched {
			fmt.Println("✓ 机器信息匹配")
		} else {
			fmt.Println("⚠ 机器信息部分匹配")
		}

	} else {
		fmt.Println("✗ 许可证验证失败")
		fmt.Printf("错误: %s\n", result.Error)

		// 显示当前机器信息以便调试
		if result.MachineInfo.MAC != "" {
			fmt.Printf("当前MAC: %s\n", result.MachineInfo.MAC)
		}
		if result.MachineInfo.UUID != "" {
			fmt.Printf("当前UUID: %s\n", result.MachineInfo.UUID)
		}
		if result.MachineInfo.CPUID != "" {
			fmt.Printf("当前CPUID: %s\n", result.MachineInfo.CPUID)
		}
	}
}
