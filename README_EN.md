# License Key Verify Tool

[![CI](https://github.com/cuilan/license-key-verify/actions/workflows/ci.yml/badge.svg)](https://github.com/cuilan/license-key-verify/actions/workflows/ci.yml)
[![Release](https://github.com/cuilan/license-key-verify/actions/workflows/release.yml/badge.svg)](https://github.com/cuilan/license-key-verify/actions/workflows/release.yml)
[![Docker](https://github.com/cuilan/license-key-verify/actions/workflows/docker.yml/badge.svg)](https://github.com/cuilan/license-key-verify/actions/workflows/docker.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/cuilan/license-key-verify)](https://goreportcard.com/report/github.com/cuilan/license-key-verify)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A comprehensive license key generation and verification tool with machine binding and digital signature support.

[中文文档](README.md) | **English**

## Features

- ✅ **Machine Information**: Get MAC address, system UUID, CPU ID
- ✅ **License Generation**: Generate encrypted license files with machine binding
- ✅ **License Verification**: Verify license validity and machine matching
- ✅ **Digital Signature**: RSA+AES hybrid encryption for security
- ✅ **Cross-Platform**: Support Windows, macOS, Linux
- ✅ **Command Line Tools**: Easy-to-use CLI interface
- ✅ **SDK Integration**: Can be integrated as a library into other projects

## Quick Start

### 1. Build Project

```bash
# Build all binaries
make build

# Or build for all platforms
make build-all
```

### 2. Get Machine Information

```bash
# Get MAC address
lkctl get mac

# Get system UUID
lkctl get uuid

# Get CPU ID
lkctl get cpuid

# Get all machine information
lkctl get all
```

### 3. Generate License

```bash
# Generate a license. If key files do not exist, they will be automatically
# generated and saved to the directory specified by --keys-dir (default: keys/)
lkctl gen \
  --keys-dir ./mykeys \
  --mac "00:11:22:33:44:55" \
  --uuid "12345678-1234-1234-1234-123456789012" \
  --cpuid "abcdef1234567890" \
  --customer "Example Customer" \
  --product "Example Product" \
  --duration 365 \
  license.lic

# Generate a license using existing keys
lkctl gen \
  --private-key ./mykeys/private.pem \
  --aes-key ./mykeys/aes.key \
  --customer "Another Customer" \
  license2.lic
```

### 4. Verify License

```bash
# Verify using lkctl
lkctl verify license.lic

# Verify using lkverify
lkverify license.lic

# JSON output
lkverify license.lic --json
```

## Command Line Tools

### lkctl Tool

`lkctl` is the main command line tool providing complete license management functionality.

#### Get Machine Information

```bash
lkctl get mac     # Get MAC address
lkctl get uuid    # Get system UUID
lkctl get cpuid   # Get CPU ID
lkctl get all     # Get all machine information
```

#### Generate Key Pair

```bash
lkctl keys --output <directory>
```
> **Note**: The `lkctl gen` command also generates keys automatically if they are not provided. The `keys` command is useful when you only need to generate key files.

#### Generate License

```bash
lkctl gen [options] <output_file>

Options:
  --mac <mac>              Specify MAC address
  --uuid <uuid>            Specify system UUID
  --cpuid <cpuid>          Specify CPU ID
  --duration <days>        Validity period (days)
  --customer <name>        Customer name
  --product <name>         Product name
  --version <version>      Product version
  --features <list>        Feature list (comma-separated)
  --max-users <number>     Maximum number of users
  --keys-dir <dir>         Directory to save new keys (default: keys)
  --private-key <file>     Path to the private key file for signing. If not provided, a new one is generated.
  --aes-key <file>         Path to the AES key file for encryption. If not provided, a new one is generated.
```

#### Verify License

```bash
lkctl verify <license_file>   # Verify license
lkctl info <license_file>     # View license information
```

### lkverify Tool

`lkverify` is a specialized verification tool suitable for integration into other programs.

```bash
lkverify <license_file> [options]

Options:
  --keys-dir <directory>   Specify key file directory (default: keys)
  --public-key <file>      Path to the public key file (overrides --keys-dir)
  --aes-key <file>         Path to the AES key file (overrides --keys-dir)
  --json                   Output results in JSON format
  --quiet                  Quiet mode, only output exit code

Exit codes:
  0  License is valid
  1  License is invalid or other error
  2  Parameter error
```

## Using in Other Projects

### Download Dependency

```bash
go get github.com/cuilan/license-key-verify
```

### Resolve Module Path Issues

If you encounter module path mismatch errors, try the following solutions:

```bash
# Clean Go module cache
go clean -modcache

# Re-download dependency
go get github.com/cuilan/license-key-verify@latest

# Or use specific version
go get github.com/cuilan/license-key-verify@main
```

If you still have issues, use replace directive:

```bash
# Add replace directive in your project
go mod edit -replace github.com/cuilan/license-key-verify=github.com/cuilan/license-key-verify@main
go mod tidy
```

### Local Development Mode

If you need to develop and test locally:

```bash
# Clone project locally
git clone https://github.com/cuilan/license-key-verify.git

# Use local module in your project
go mod edit -replace github.com/cuilan/license-key-verify=../path/to/license-key-verify
go mod tidy
```

### Integration in Go Projects

#### 1. Initialize Your Go Project

```bash
# Create new project directory
mkdir my-licensed-app
cd my-licensed-app

# Initialize Go module
go mod init my-licensed-app

# Download license-key-verify library
go get github.com/cuilan/license-key-verify
```

#### 2. Using as Go Library

```go
package main

import (
    "fmt"
    "time"
    
    "github.com/cuilan/license-key-verify/pkg/license"
    "github.com/cuilan/license-key-verify/pkg/machine"
)

func main() {
    // 1. Generate license
    generator, err := license.NewGenerator()
    if err != nil {
        panic(err)
    }
    
    // Get machine information
    machineInfo, err := machine.GetAllInfo()
    if err != nil {
        panic(err)
    }
    
    options := &license.GenerateOptions{
        ProductName:  "My Product",
        CustomerName: "Customer Name",
        MAC:          machineInfo.MAC,
        UUID:         machineInfo.UUID,
        CPUID:        machineInfo.CPUID,
        Duration:     30 * 24 * time.Hour, // 30 days
        Features:     []string{"feature1", "feature2"},
        MaxUsers:     10,
    }
    
    lic, err := generator.Generate(options)
    if err != nil {
        panic(err)
    }
    
    // Save license
    err = generator.SaveToFile(lic, "license.lic")
    if err != nil {
        panic(err)
    }
    
    // 2. Verify license
    verifier, err := license.NewVerifierFromFiles("keys/public.pem", "keys/aes.key")
    if err != nil {
        panic(err)
    }
    
    result, err := verifier.VerifyFile("license.lic")
    if err != nil {
        panic(err)
    }
    
    if result.Valid {
        fmt.Println("License verification passed")
        fmt.Printf("Days remaining: %d\n", result.ExpiresIn/(24*3600))
    } else {
        fmt.Printf("License verification failed: %s\n", result.Error)
    }
}
```

### Secure Integration Mode (Recommended)

The previous examples demonstrated loading keys from external files (`keys/public.pem`, `keys/aes.key`) for simplicity. However, this approach carries a security risk: **if your customers can replace these key files, they can use the `lkctl` tool to issue valid licenses for themselves**.

To completely eliminate this risk, it is strongly recommended to **hard-code the public key and AES key contents directly into your application**. This way, your program will only trust the keys you compiled into it, and no external keys can pass verification.

Here is a more secure integration example (`examples/secure-integration/main.go`):

```go
package main

import (
	"fmt"
	"os"
	
	"github.com/cuilan/license-key-verify/pkg/crypto"
	"github.com/cuilan/license-key-verify/pkg/license"
)

// =================================================================
// Critical Security Zone: Embed your public key and AES key as constants.
// =================================================================

const (
	// publicKeyPEM should contain the full content of '-----BEGIN PUBLIC KEY-----' and '-----END PUBLIC KEY-----'.
	// !!! This is an example key. Be sure to replace it with your own public key. !!!
	publicKeyPEM = `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqP4C3nN5Z5vB74rVb5q1
... (your public key content here) ...
-----END PUBLIC KEY-----
`

	// aesKeyBase64 should be the Base64 encoded string from your aes.key file.
	// !!! This is an example key. Be sure to replace it with your own AES key. !!!
	aesKeyBase64 = `YOUR_BASE64_ENCODED_AES_KEY_HERE`
)

// createVerifierFromEmbeddedKeys creates a secure verifier instance from hard-coded keys.
func createVerifierFromEmbeddedKeys() (*license.Verifier, error) {
	// 1. Decode the AES key from the Base64 string.
	aesKey, err := crypto.DecodeBase64(aesKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode embedded AES key: %w", err)
	}

	// 2. Create the verifier instance using the loaded keys.
	return license.NewVerifier([]byte(publicKeyPEM), aesKey)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <path-to-license.lic>\n", os.Args[0])
		os.Exit(1)
	}
	licenseFile := os.Args[1]

	// Initialize a secure verifier that only trusts the embedded keys.
	verifier, err := createVerifierFromEmbeddedKeys()
	if err != nil {
		fmt.Printf("FATAL: Could not initialize license verifier: %v\n", err)
		os.Exit(1)
	}

	// Verify the license file using the secure verifier.
	result, err := verifier.VerifyFile(licenseFile)
	if err != nil {
		fmt.Printf("License verification failed with an error: %v\n", err)
		os.Exit(1)
	}

	if result.Valid {
		fmt.Println("✅ License is VALID.")
		// Your core application logic goes here...
	} else {
		fmt.Printf("❌ License is INVALID: %s\n", result.Error)
		os.Exit(1)
	}
}
```

#### 4. Build and Run

```bash
# Build your application
go build -o my-app

# Run (requires key files and license file to be ready)
./my-app license.lic
```

## Build and Deployment

### Local Build

```bash
# Build current platform
make build

# Build all platforms
make build-all

# Run tests
make test

# Generate examples
make demo
```

### GitHub Actions Manual Build

The project uses manually triggered GitHub Actions workflows:

1. Visit [GitHub Actions](https://github.com/cuilan/license-key-verify/actions)
2. Select "CI" workflow
3. Click "Run workflow" to manually trigger build
4. Choose whether to run tests and build all platforms

For detailed instructions, see: [Manual Trigger Guide](docs/manual-trigger-guide.md)

### System Installation

```bash
# Install to /usr/local/bin
sudo make install

# Uninstall
sudo make uninstall
```

## Security Features

1. **Hybrid Encryption**: AES-256-GCM symmetric encryption + RSA-2048 asymmetric signature
2. **Machine Binding**: Bind through MAC address, UUID, CPU ID
3. **Tamper-Proof**: Digital signature ensures license file integrity
4. **Time Validation**: Support license expiration validation
5. **Feature Control**: Support authorization by feature modules

## License File Format

License files use JSON format with the following fields:

```json
{
  "data": "Encrypted license data (Base64 encoded)",
  "signature": "Digital signature (Base64 encoded)",
  "algorithm": "Encryption algorithm identifier",
  "version": "File format version"
}
```

## Docker Support

### Using Pre-built Images

```bash
# Pull latest image
docker pull ghcr.io/cuilan/license-key-verify:latest

# Run container
docker run --rm ghcr.io/cuilan/license-key-verify:latest --help

# Mount local directory for license operations
docker run --rm -v $(pwd):/workspace \
  ghcr.io/cuilan/license-key-verify:latest \
  get mac
```

### Local Build

```bash
# Build image
docker build -t license-key-verify .

# Run container
docker run --rm license-key-verify --help
```

## Development and Contributing

### Development Requirements

- Go 1.23+
- Make tool

### Code Standards

```bash
# Format code
make fmt

# Code check
make lint

# Run tests
make test
```

### Contributing

1. Fork the project
2. Create feature branch
3. Commit changes
4. Create Pull Request
5. Wait for code review

## FAQ

### Q: How to verify license on a new machine?

A: License files are bound to machine information and can only be verified on the corresponding machine. To use on a new machine, you need to regenerate the license.

### Q: Can license files be copied to other machines?

A: No. License files contain machine binding information and verification will fail on non-matching machines.

### Q: How to backup and restore keys?

A: When using `lkctl gen`, you can specify the directory for new keys with the `--keys-dir` argument (defaults to `keys/`). It's recommended to backup this entire directory. Private keys are used for license generation; public keys and AES keys are used for verification. Keep your private key secure.

### Q: Does it support offline verification?

A: Yes. The verification process is completely offline and doesn't require network connection.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with Go standard library only
- Uses RSA and AES encryption algorithms
- Inspired by modern software licensing practices

## Support

- 📖 [Documentation](docs/)
- 🐛 [Issue Tracker](https://github.com/cuilan/license-key-verify/issues)
- 💬 [Discussions](https://github.com/cuilan/license-key-verify/discussions)
- 📧 Contact: [Create an issue](https://github.com/cuilan/license-key-verify/issues/new) 