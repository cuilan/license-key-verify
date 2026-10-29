package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cuilan/license-key-verify/pkg/license"
)

func handleInfo() {
	// 手动解析：文件名与 flag 顺序随意（flag.FlagSet 遇到位置参数会停止解析）
	var publicKey string
	rest := os.Args[2:]
	var args []string
	for i := 0; i < len(rest); i++ {
		arg := rest[i]
		if arg == "--public-key" {
			if i+1 >= len(rest) {
				fmt.Println("--public-key requires a file path")
				os.Exit(2)
			}
			i++
			publicKey = rest[i]
			continue
		}
		if strings.HasPrefix(arg, "-") {
			fmt.Printf("Unknown option: %s\n", arg)
			os.Exit(2)
		}
		args = append(args, arg)
	}
	if len(args) == 0 {
		fmt.Println("Usage: lkctl info <license-file> [--public-key <pem>]")
		os.Exit(2)
	}
	if len(args) > 1 {
		fmt.Println("Only one license file can be specified")
		os.Exit(2)
	}

	applyPublicKeyOverride(publicKey)
	result := license.VerifyFile(args[0])
	if !result.Valid && result.License == nil {
		fmt.Fprintf(os.Stderr, "Failed to read license (signature invalid?): %s\n", result.Error)
		os.Exit(1)
	}
	lic := result.License
	fmt.Printf("License ID:   %s\n", lic.LicenseID)
	fmt.Printf("Customer:     %s\n", lic.Customer)
	fmt.Printf("Product:      %s\n", lic.Product)
	fmt.Printf("Fingerprint:  %s\n", lic.Fingerprint)
	fmt.Printf("Issued At:    %s\n", lic.IssuedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Expires At:   %s\n", lic.ExpiresAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Features:     %s\n", strings.Join(lic.Features, ", "))
}
