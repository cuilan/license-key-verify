package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cuilan/license-key-verify/pkg/license"
	"github.com/cuilan/license-key-verify/pkg/machine"
)

const (
	Version = "1.0.0"
	Usage   = `lkctl - License Key Control Tool

用法:
  lkctl get <info>           获取机器信息
    lkctl get mac            获取MAC地址
    lkctl get uuid           获取系统UUID
    lkctl get cpuid          获取CPU ID

  lkctl gen [选项] <输出文件>  生成许可证
    --mac <mac>              指定MAC地址
    --uuid <uuid>            指定系统UUID
    --cpuid <cpuid>          指定CPU ID
    --duration <天数>        有效期（天）
    --customer <客户名>      客户名称
    --product <产品名>       产品名称
    --version <版本>         产品版本
    --features <功能列表>    功能列表（逗号分隔）
    --max-users <数量>       最大用户数

  lkctl verify <许可证文件>   验证许可证
  lkctl info <许可证文件>     查看许可证信息

  lkctl keys                生成密钥对
    --output <目录>          输出目录（默认当前目录）

  lkctl --version           显示版本
  lkctl --help              显示帮助
`
)

func main() {
	if len(os.Args) < 2 {
		fmt.Print(Usage)
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "get":
		handleGet()
	case "gen":
		handleGen()
	case "verify":
		handleVerify()
	case "info":
		handleInfo()
	case "keys":
		handleKeys()
	case "--version":
		fmt.Printf("lkctl version %s\n", Version)
	case "--help":
		fmt.Print(Usage)
	default:
		fmt.Printf("未知命令: %s\n", command)
		fmt.Print(Usage)
		os.Exit(1)
	}
}

func handleGet() {
	if len(os.Args) < 3 {
		fmt.Println("用法: lkctl get <mac|uuid|cpuid>")
		os.Exit(1)
	}

	infoType := os.Args[2]

	switch infoType {
	case "mac":
		mac, err := machine.GetMACAddress()
		if err != nil {
			fmt.Printf("获取MAC地址失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(mac)

	case "uuid":
		uuid, err := machine.GetSystemUUID()
		if err != nil {
			fmt.Printf("获取系统UUID失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(uuid)

	case "cpuid":
		cpuid, err := machine.GetCPUID()
		if err != nil {
			fmt.Printf("获取CPU ID失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(cpuid)

	default:
		fmt.Printf("未知信息类型: %s\n", infoType)
		fmt.Println("支持的类型: mac, uuid, cpuid")
		os.Exit(1)
	}
}

func handleGen() {
	fs := flag.NewFlagSet("gen", flag.ExitOnError)

	var (
		mac      = fs.String("mac", "", "MAC地址")
		uuid     = fs.String("uuid", "", "系统UUID")
		cpuid    = fs.String("cpuid", "", "CPU ID")
		duration = fs.Int("duration", 365, "有效期（天）")
		customer = fs.String("customer", "", "客户名称")
		product  = fs.String("product", "", "产品名称")
		version  = fs.String("version", "", "产品版本")
		features = fs.String("features", "", "功能列表（逗号分隔）")
		maxUsers = fs.Int("max-users", 0, "最大用户数")
	)

	fs.Parse(os.Args[2:])

	args := fs.Args()
	if len(args) == 0 {
		fmt.Println("用法: lkctl gen [选项] <输出文件>")
		os.Exit(1)
	}

	outputFile := args[0]

	// 创建生成器
	generator, err := license.NewGenerator()
	if err != nil {
		fmt.Printf("创建生成器失败: %v\n", err)
		os.Exit(1)
	}

	// 设置生成选项
	options := &license.GenerateOptions{
		ProductName:  *product,
		Version:      *version,
		CustomerName: *customer,
		MAC:          *mac,
		UUID:         *uuid,
		CPUID:        *cpuid,
		Duration:     time.Duration(*duration) * 24 * time.Hour,
		MaxUsers:     *maxUsers,
	}

	if *features != "" {
		options.Features = strings.Split(*features, ",")
	}

	// 生成许可证
	lic, err := generator.Generate(options)
	if err != nil {
		fmt.Printf("生成许可证失败: %v\n", err)
		os.Exit(1)
	}

	// 保存到文件
	err = generator.SaveToFile(lic, outputFile)
	if err != nil {
		fmt.Printf("保存许可证失败: %v\n", err)
		os.Exit(1)
	}

	// 保存密钥文件
	keyDir := "keys"
	os.MkdirAll(keyDir, 0755)

	err = generator.SaveKeys(
		keyDir+"/private.pem",
		keyDir+"/public.pem",
		keyDir+"/aes.key",
	)
	if err != nil {
		fmt.Printf("保存密钥失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("许可证已生成: %s\n", outputFile)
	fmt.Printf("密钥已保存到: %s/\n", keyDir)
	fmt.Printf("许可证ID: %s\n", lic.ID)
	fmt.Printf("有效期至: %s\n", lic.ExpiresAt.Format("2006-01-02 15:04:05"))
}

func handleVerify() {
	if len(os.Args) < 3 {
		fmt.Println("用法: lkctl verify <许可证文件>")
		os.Exit(1)
	}

	licenseFile := os.Args[2]

	// 创建验证器
	verifier, err := license.NewVerifierFromFiles("keys/public.pem", "keys/aes.key")
	if err != nil {
		fmt.Printf("创建验证器失败: %v\n", err)
		fmt.Println("请确保密钥文件存在: keys/public.pem, keys/aes.key")
		os.Exit(1)
	}

	// 验证许可证
	result, err := verifier.VerifyFile(licenseFile)
	if err != nil {
		fmt.Printf("验证失败: %v\n", err)
		os.Exit(1)
	}

	// 输出结果
	if result.Valid {
		fmt.Println("✓ 许可证验证通过")
		fmt.Printf("许可证ID: %s\n", result.License.ID)
		fmt.Printf("产品名称: %s\n", result.License.ProductName)
		fmt.Printf("客户名称: %s\n", result.License.CustomerName)
		fmt.Printf("有效期至: %s\n", result.License.ExpiresAt.Format("2006-01-02 15:04:05"))

		days := result.ExpiresIn / (24 * 3600)
		fmt.Printf("剩余天数: %d 天\n", days)
	} else {
		fmt.Println("✗ 许可证验证失败")
		fmt.Printf("错误: %s\n", result.Error)
	}

	// 输出机器信息匹配状态
	if result.MachineInfo.Matched {
		fmt.Println("✓ 机器信息匹配")
	} else {
		fmt.Println("✗ 机器信息不匹配")
		fmt.Printf("当前MAC: %s\n", result.MachineInfo.MAC)
		fmt.Printf("当前UUID: %s\n", result.MachineInfo.UUID)
		fmt.Printf("当前CPUID: %s\n", result.MachineInfo.CPUID)
	}
}

func handleInfo() {
	if len(os.Args) < 3 {
		fmt.Println("用法: lkctl info <许可证文件>")
		os.Exit(1)
	}

	licenseFile := os.Args[2]

	// 创建验证器
	verifier, err := license.NewVerifierFromFiles("keys/public.pem", "keys/aes.key")
	if err != nil {
		fmt.Printf("创建验证器失败: %v\n", err)
		fmt.Println("请确保密钥文件存在: keys/public.pem, keys/aes.key")
		os.Exit(1)
	}

	// 获取许可证信息
	lic, err := verifier.GetLicenseInfo(licenseFile)
	if err != nil {
		fmt.Printf("获取许可证信息失败: %v\n", err)
		os.Exit(1)
	}

	// 输出信息
	data, err := json.MarshalIndent(lic, "", "  ")
	if err != nil {
		fmt.Printf("序列化许可证信息失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(data))
}

func handleKeys() {
	fs := flag.NewFlagSet("keys", flag.ExitOnError)
	output := fs.String("output", ".", "输出目录")
	fs.Parse(os.Args[2:])

	// 创建生成器
	generator, err := license.NewGenerator()
	if err != nil {
		fmt.Printf("创建生成器失败: %v\n", err)
		os.Exit(1)
	}

	// 确保输出目录存在
	err = os.MkdirAll(*output, 0755)
	if err != nil {
		fmt.Printf("创建输出目录失败: %v\n", err)
		os.Exit(1)
	}

	// 保存密钥
	privateKeyPath := *output + "/private.pem"
	publicKeyPath := *output + "/public.pem"
	aesKeyPath := *output + "/aes.key"

	err = generator.SaveKeys(privateKeyPath, publicKeyPath, aesKeyPath)
	if err != nil {
		fmt.Printf("保存密钥失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("密钥已生成:\n")
	fmt.Printf("  私钥: %s\n", privateKeyPath)
	fmt.Printf("  公钥: %s\n", publicKeyPath)
	fmt.Printf("  AES密钥: %s\n", aesKeyPath)
}
