package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cuilan/license-key-verify/v2/pkg/license"
)

// applyPublicKeyOverride 当指定 --public-key 时用它覆盖内置公钥（调试/测试用）。
func applyPublicKeyOverride(path string) {
	if path == "" {
		return
	}
	pemData, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read public key: %v\n", err)
		os.Exit(1)
	}
	if err := license.UsePublicKeyPEM(pemData); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid public key: %v\n", err)
		os.Exit(1)
	}
}

// printResult 以人类可读格式输出验证结果。
func printResult(result license.Result) {
	if result.Valid {
		fmt.Println("✓ License verification passed")
		lic := result.License
		fmt.Printf("License ID:  %s\n", lic.LicenseID)
		if lic.Customer != "" {
			fmt.Printf("Customer:    %s\n", lic.Customer)
		}
		fmt.Printf("Product:     %s\n", lic.Product)
		fmt.Printf("Expires At:  %s\n", lic.ExpiresAt.Format("2006-01-02 15:04:05"))
		days := result.ExpiresIn / (24 * 3600)
		hours := (result.ExpiresIn % (24 * 3600)) / 3600
		fmt.Printf("Expires In:  %d days %d hours\n", days, hours)
		if len(lic.Features) > 0 {
			fmt.Printf("Features:    %v\n", lic.Features)
		}
		fmt.Println("✓ Machine fingerprint matched")
		return
	}

	fmt.Println("✗ License verification failed")
	fmt.Printf("Error: %s\n", result.Error)
	if result.Fingerprint != "" && !result.Matched {
		fmt.Printf("Local fingerprint: %s\n", result.Fingerprint)
	}
}

// printResultJSON 以 JSON 格式输出验证结果。
func printResultJSON(result license.Result) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to serialize result: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}
