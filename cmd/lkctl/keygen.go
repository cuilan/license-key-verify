package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cuilan/license-key-verify/pkg/crypto"
)

func handleKeygen() {
	fs := flag.NewFlagSet("keygen", flag.ExitOnError)
	out := fs.String("out", "keys", "Output directory")
	fs.Parse(os.Args[2:])

	keyPair, err := crypto.GenerateKeyPair()
	if err != nil {
		fatal("Failed to generate key pair: %v", err)
	}

	privateKeyPEM, err := keyPair.PrivateKeyToPEM()
	if err != nil {
		fatal("Failed to serialize private key: %v", err)
	}
	publicKeyPEM, err := keyPair.PublicKeyToPEM()
	if err != nil {
		fatal("Failed to serialize public key: %v", err)
	}

	if err := os.MkdirAll(*out, 0755); err != nil {
		fatal("Failed to create output directory: %v", err)
	}

	privatePath := filepath.Join(*out, "private.pem")
	publicPath := filepath.Join(*out, "public.pem")
	if err := os.WriteFile(privatePath, privateKeyPEM, 0600); err != nil {
		fatal("Failed to save private key: %v", err)
	}
	if err := os.WriteFile(publicPath, publicKeyPEM, 0644); err != nil {
		fatal("Failed to save public key: %v", err)
	}

	fmt.Printf("Keys generated:\n  Private key: %s (keep it secret!)\n  Public key:  %s\n", privatePath, publicPath)
}
