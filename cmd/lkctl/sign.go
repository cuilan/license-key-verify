package main

import (
	"crypto/rsa"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cuilan/license-key-verify/v2/pkg/crypto"
	"github.com/cuilan/license-key-verify/v2/pkg/license"
)

func handleSign() {
	fs := flag.NewFlagSet("sign", flag.ExitOnError)
	privateKey := fs.String("private-key", "", "Private key PEM file (required)")
	fingerprint := fs.String("fingerprint", "", "Target machine fingerprint (required)")
	duration := fs.Int("duration", 365, "Validity period in days")
	customer := fs.String("customer", "", "Customer name")
	product := fs.String("product", "", "Product name")
	features := fs.String("features", "", "Comma-separated feature list")
	licenseID := fs.String("license-id", "", "Custom license ID")
	out := fs.String("out", "", "Output license file (required)")
	fs.Parse(os.Args[2:])

	if *privateKey == "" || *fingerprint == "" || *out == "" {
		fmt.Println("Usage: lkctl sign --private-key <pem> --fingerprint <fp> --out <file> [options]")
		os.Exit(2)
	}

	privateKeyPEM, err := os.ReadFile(*privateKey)
	if err != nil {
		fatal("Failed to read private key: %v", err)
	}

	privKey, err := crypto.LoadPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		fatal("Invalid private key: %v", err)
	}

	// 配对提示（仅警告，不阻断）：lkctl 是通用工具，未内嵌公钥时 sign 正常工作；
	// 仅当二进制内嵌了公钥（fork 自建场景）且与签名私钥不配对时提醒，
	// 因为签出的许可证无法被内嵌该公钥的应用或本 lkctl verify 校验。
	if license.HasEmbeddedKey() {
		embeddedPEM, err := license.EmbeddedKeyPEM()
		if err == nil {
			if embeddedPub, err := crypto.LoadPublicKeyFromPEM(embeddedPEM); err == nil {
				if pubFromPriv, ok := privKey.Public().(*rsa.PublicKey); ok &&
					(embeddedPub.N.Cmp(pubFromPriv.N) != 0 || embeddedPub.E != pubFromPriv.E) {
					fmt.Fprintln(os.Stderr, "Warning: the embedded public key does not match this private key;")
					fmt.Fprintln(os.Stderr, "         the signed license cannot be verified by apps embedding that key or by this lkctl verify.")
				}
			}
		}
	}

	opts := &license.SignOptions{
		LicenseID:   *licenseID,
		Customer:    *customer,
		Product:     *product,
		Fingerprint: *fingerprint,
		Duration:    time.Duration(*duration) * 24 * time.Hour,
	}
	if *features != "" {
		for _, f := range strings.Split(*features, ",") {
			if f = strings.TrimSpace(f); f != "" {
				opts.Features = append(opts.Features, f)
			}
		}
	}

	lic, err := license.Sign(opts, privateKeyPEM)
	if err != nil {
		fatal("Failed to sign license: %v", err)
	}
	if err := lic.Save(*out); err != nil {
		fatal("Failed to save license: %v", err)
	}

	fmt.Printf("License signed: %s\n", *out)
	fmt.Printf("License ID: %s\n", lic.LicenseID)
	fmt.Printf("Fingerprint: %s\n", lic.Fingerprint)
	fmt.Printf("Expires at: %s\n", lic.ExpiresAt.Format("2006-01-02 15:04:05"))
}
