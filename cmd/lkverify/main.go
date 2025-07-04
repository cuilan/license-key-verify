package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cuilan/license-key-verify/pkg/license"
)

var (
	Version = "dev" // will be set by -ldflags during build
)

const (
	Usage = `
  ┌─[LICENSE]──────────────┐
  │                        │
  │    .--------------.    │
  │   /    V A L I D   \   │
  │  |    (stamped)     |  │
  │   \________________/   │
  │                        │
  └────────────────────────┘

  lkverify - License Key Verification Tool

  Usage:
  lkverify <license-file> [options]

  Options:
    --keys-dir <directory>  Specify the directory for key files (default: keys)
    --public-key <file>     Specify the path to the public key file (overrides --keys-dir)
    --aes-key <file>        Specify the path to the AES key file (overrides --keys-dir)
    --json                  Output results in JSON format
    --quiet                 Quiet mode, only outputs exit code
    --version               Show version
    --help                  Show this help message
  
  Exit Codes:
    0  License is valid
    1  License is invalid or other error
    2  Argument error
  
  Examples:
    lkverify license.lic
    lkverify license.lic --json
    lkverify license.lic --keys-dir ./mykeys
    lkverify license.lic --public-key /path/to/public.pem --aes-key /path/to/aes.key
`
)

type Config struct {
	LicenseFile   string
	KeysDir       string
	PublicKeyPath string
	AESKeyPath    string
	JSONOutput    bool
	Quiet         bool
}

func main() {
	config := parseArgs()

	// 确定密钥文件路径
	publicKeyPath := config.PublicKeyPath
	if publicKeyPath == "" {
		publicKeyPath = config.KeysDir + "/public.pem"
	}
	aesKeyPath := config.AESKeyPath
	if aesKeyPath == "" {
		aesKeyPath = config.KeysDir + "/aes.key"
	}

	// 创建验证器
	verifier, err := license.NewVerifierFromFiles(publicKeyPath, aesKeyPath)
	if err != nil {
		if !config.Quiet {
			fmt.Fprintf(os.Stderr, "Failed to create verifier: %v\n", err)
			fmt.Fprintf(os.Stderr, "Please make sure the key files exist: %s, %s\n", publicKeyPath, aesKeyPath)
		}
		os.Exit(1)
	}

	// 验证许可证
	result, err := verifier.VerifyFile(config.LicenseFile)
	if err != nil {
		if !config.Quiet {
			fmt.Fprintf(os.Stderr, "Verification failed: %v\n", err)
		}
		os.Exit(1)
	}

	// 输出结果
	if config.JSONOutput {
		// JSON格式输出
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			if !config.Quiet {
				fmt.Fprintf(os.Stderr, "Failed to serialize result: %v\n", err)
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
				fmt.Fprintf(os.Stderr, "--keys-dir requires a directory\n")
				os.Exit(2)
			}
			i++
			config.KeysDir = args[i]
		case "--public-key":
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "--public-key requires a file path\n")
				os.Exit(2)
			}
			i++
			config.PublicKeyPath = args[i]
		case "--aes-key":
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "--aes-key requires a file path\n")
				os.Exit(2)
			}
			i++
			config.AESKeyPath = args[i]
		default:
			if arg[0] == '-' {
				fmt.Fprintf(os.Stderr, "Unknown option: %s\n", arg)
				os.Exit(2)
			} else {
				if config.LicenseFile == "" {
					config.LicenseFile = arg
				} else {
					fmt.Fprintf(os.Stderr, "Only one license file can be specified\n")
					os.Exit(2)
				}
			}
		}
	}

	if config.LicenseFile == "" {
		fmt.Fprintf(os.Stderr, "A license file must be specified\n")
		fmt.Print(Usage)
		os.Exit(2)
	}

	return config
}

func printResult(result *license.VerificationResult) {
	if result.Valid {
		fmt.Println("✓ License verification passed")

		if result.License != nil {
			fmt.Printf("License ID: %s\n", result.License.ID)
			fmt.Printf("Product Name: %s\n", result.License.ProductName)

			if result.License.CustomerName != "" {
				fmt.Printf("Customer Name: %s\n", result.License.CustomerName)
			}

			fmt.Printf("Issued At: %s\n", result.License.IssuedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("Expires At: %s\n", result.License.ExpiresAt.Format("2006-01-02 15:04:05"))

			if result.ExpiresIn > 0 {
				days := result.ExpiresIn / (24 * 3600)
				hours := (result.ExpiresIn % (24 * 3600)) / 3600
				fmt.Printf("Expires In: %d days %d hours\n", days, hours)
			}

			if len(result.License.Features) > 0 {
				fmt.Printf("Features: %v\n", result.License.Features)
			}

			if result.License.MaxUsers > 0 {
				fmt.Printf("Max Users: %d\n", result.License.MaxUsers)
			}

			if result.License.Notes != "" {
				fmt.Printf("Notes: %s\n", result.License.Notes)
			}
		}

		// 机器信息匹配状态
		if result.MachineInfo.Matched {
			fmt.Println("✓ Machine information matched")
		} else {
			fmt.Println("⚠ Machine information partially matched")
		}

	} else {
		fmt.Println("✗ License verification failed")
		fmt.Printf("Error: %s\n", result.Error)

		// 显示当前机器信息以便调试
		if result.MachineInfo.MAC != "" {
			fmt.Printf("Current MAC: %s\n", result.MachineInfo.MAC)
		}
		if result.MachineInfo.UUID != "" {
			fmt.Printf("Current UUID: %s\n", result.MachineInfo.UUID)
		}
		if result.MachineInfo.CPUID != "" {
			fmt.Printf("Current CPUID: %s\n", result.MachineInfo.CPUID)
		}
	}
}
