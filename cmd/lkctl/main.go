package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cuilan/license-key-verify/pkg/crypto"
	"github.com/cuilan/license-key-verify/pkg/license"
	"github.com/cuilan/license-key-verify/pkg/machine"
)

var (
	Version = "dev" // will be set by -ldflags during build
)

const (
	Usage = `
  ┌─[NEW]──────────────────┐
  │                        │
  │  .--. -> [########]    │
  │ /  _ \                 │
  │ \ ` + "`" + `' /   (key)         │
  │  ` + "`" + `--'                  │
  └────────────────────────┘

  lkctl - License Key Control Tool

  Usage:
    lkctl get <info>            Get machine information
    lkctl get mac               Get MAC address
    lkctl get uuid              Get system UUID
    lkctl get cpuid             Get CPU ID
    lkctl get all               Get all machine information

  lkctl gen [options] <output-file> Generate a license
    --mac <mac>                 Specify MAC address
    --uuid <uuid>               Specify system UUID
    --cpuid <cpuid>             Specify CPU ID
    --duration <days>           Validity period (days)
    --customer <name>           Customer name
    --product <name>            Product name
    --version <version>         Product version
    --features <list>           Comma-separated list of features
    --max-users <count>         Maximum number of users
    --keys-dir <dir>            Directory for key files (default: keys)
    --private-key <file>        Path to private key file. If not provided, a new one is generated.
    --aes-key <file>            Path to AES key file. If not provided, a new one is generated.

  lkctl verify <license-file>   Verify a license
  lkctl info <license-file>     Show license information

  lkctl keys                    Generate a new key pair
    --output <dir>              Output directory (default: current directory)

  lkctl --version               Show version
  lkctl --help                  Show this help message
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
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Print(Usage)
		os.Exit(1)
	}
}

func handleGet() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: lkctl get <mac|uuid|cpuid|all>")
		os.Exit(1)
	}

	infoType := os.Args[2]

	switch infoType {
	case "mac":
		mac, err := machine.GetMACAddress()
		if err != nil {
			fmt.Printf("Failed to get MAC address: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(mac)

	case "uuid":
		uuid, err := machine.GetSystemUUID()
		if err != nil {
			fmt.Printf("Failed to get system UUID: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(uuid)

	case "cpuid":
		cpuid, err := machine.GetCPUID()
		if err != nil {
			fmt.Printf("Failed to get CPU ID: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(cpuid)

	case "all":
		mac, err := machine.GetMACAddress()
		if err != nil {
			fmt.Printf("MAC: <error: %v>\n", err)
		} else {
			fmt.Printf("MAC: %s\n", mac)
		}

		uuid, err := machine.GetSystemUUID()
		if err != nil {
			fmt.Printf("UUID: <error: %v>\n", err)
		} else {
			fmt.Printf("UUID: %s\n", uuid)
		}

		cpuid, err := machine.GetCPUID()
		if err != nil {
			fmt.Printf("CPUID: <error: %v>\n", err)
		} else {
			fmt.Printf("CPUID: %s\n", cpuid)
		}

	default:
		fmt.Printf("Unknown info type: %s\n", infoType)
		fmt.Println("Supported types: mac, uuid, cpuid, all")
		os.Exit(1)
	}
}

func handleGen() {
	fs := flag.NewFlagSet("gen", flag.ExitOnError)

	var (
		mac      = fs.String("mac", "", "MAC address")
		uuid     = fs.String("uuid", "", "System UUID")
		cpuid    = fs.String("cpuid", "", "CPU ID")
		duration = fs.Int("duration", 365, "Validity period (days)")
		customer = fs.String("customer", "", "Customer name")
		product  = fs.String("product", "", "Product name")
		version  = fs.String("version", "", "Product version")
		features = fs.String("features", "", "Comma-separated list of features")
		maxUsers = fs.Int("max-users", 0, "Maximum number of users")
		keysDir  = fs.String("keys-dir", "keys", "Directory to save newly generated key files")
		privKey  = fs.String("private-key", "", "Path to private key file. If not provided, a new one is generated.")
		aesKey   = fs.String("aes-key", "", "Path to AES key file. If not provided, a new one is generated.")
	)

	fs.Parse(os.Args[2:])

	args := fs.Args()
	if len(args) == 0 {
		fmt.Println("Usage: lkctl gen [options] <output-file>")
		os.Exit(1)
	}

	outputFile := args[0]

	var (
		generator        *license.Generator
		privateKeyPEM    []byte
		aesKeyBytes      []byte
		err              error
		generatedPrivKey bool
		generatedAesKey  bool
	)

	// Handle private key
	if *privKey != "" {
		privateKeyPEM, err = os.ReadFile(*privKey)
		if err != nil {
			fmt.Printf("Failed to read private key: %v\n", err)
			os.Exit(1)
		}
	} else {
		keyPair, err := crypto.GenerateKeyPair()
		if err != nil {
			fmt.Printf("Failed to generate key pair: %v\n", err)
			os.Exit(1)
		}
		privateKeyPEM, err = keyPair.PrivateKeyToPEM()
		if err != nil {
			fmt.Printf("Failed to convert private key to PEM: %v\n", err)
			os.Exit(1)
		}
		generatedPrivKey = true
	}

	// Handle AES key
	if *aesKey != "" {
		aesKeyFileBytes, err := os.ReadFile(*aesKey)
		if err != nil {
			fmt.Printf("Failed to read AES key file: %v\n", err)
			os.Exit(1)
		}
		aesKeyBytes, err = crypto.DecodeBase64(string(aesKeyFileBytes))
		if err != nil {
			fmt.Printf("Failed to decode AES key: %v\n", err)
			os.Exit(1)
		}
	} else {
		aesKeyBytes, err = crypto.GenerateAESKey()
		if err != nil {
			fmt.Printf("Failed to generate AES key: %v\n", err)
			os.Exit(1)
		}
		generatedAesKey = true
	}

	generator, err = license.NewGeneratorWithKeys(privateKeyPEM, aesKeyBytes)
	if err != nil {
		fmt.Printf("Failed to create generator with keys: %v\n", err)
		os.Exit(1)
	}

	// Set generation options
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

	// Generate license
	lic, err := generator.Generate(options)
	if err != nil {
		fmt.Printf("Failed to generate license: %v\n", err)
		os.Exit(1)
	}

	// Save to file
	err = generator.SaveToFile(lic, outputFile)
	if err != nil {
		fmt.Printf("Failed to save license: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("License generated: %s\n", outputFile)

	if generatedPrivKey || generatedAesKey {
		os.MkdirAll(*keysDir, 0755)

		if generatedPrivKey {
			// Save private key
			privKeyPath := *keysDir + "/private.pem"
			err = os.WriteFile(privKeyPath, privateKeyPEM, 0600)
			if err != nil {
				fmt.Printf("Failed to save private key: %v\n", err)
				os.Exit(1)
			}

			// Save public key
			pubKeyPEM, err := generator.GetPublicKeyPEM()
			if err != nil {
				fmt.Printf("Failed to get public key PEM: %v\n", err)
				os.Exit(1)
			}
			pubKeyPath := *keysDir + "/public.pem"
			err = os.WriteFile(pubKeyPath, pubKeyPEM, 0644)
			if err != nil {
				fmt.Printf("Failed to save public key: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("New RSA key pair saved to %s and %s\n", privKeyPath, pubKeyPath)
		}

		if generatedAesKey {
			aesKeyPath := *keysDir + "/aes.key"
			encodedKey := crypto.EncodeBase64(aesKeyBytes)
			err = os.WriteFile(aesKeyPath, []byte(encodedKey), 0600)
			if err != nil {
				fmt.Printf("Failed to save AES key: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("New AES key saved to %s\n", aesKeyPath)
		}
	}

	fmt.Printf("License ID: %s\n", lic.ID)
	fmt.Printf("Expires at: %s\n", lic.ExpiresAt.Format("2006-01-02 15:04:05"))
}

func handleVerify() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: lkctl verify <license-file>")
		os.Exit(1)
	}

	licenseFile := os.Args[2]

	// Create verifier
	verifier, err := license.NewVerifierFromFiles("keys/public.pem", "keys/aes.key")
	if err != nil {
		fmt.Printf("Failed to create verifier: %v\n", err)
		fmt.Println("Please make sure the key files exist: keys/public.pem, keys/aes.key")
		os.Exit(1)
	}

	// Verify license
	result, err := verifier.VerifyFile(licenseFile)
	if err != nil {
		fmt.Printf("Verification failed: %v\n", err)
		os.Exit(1)
	}

	// Output result
	if result.Valid {
		fmt.Println("✓ License verification passed")
		fmt.Printf("License ID: %s\n", result.License.ID)
		fmt.Printf("Product Name: %s\n", result.License.ProductName)
		fmt.Printf("Customer Name: %s\n", result.License.CustomerName)
		fmt.Printf("Expires at: %s\n", result.License.ExpiresAt.Format("2006-01-02 15:04:05"))

		days := result.ExpiresIn / (24 * 3600)
		fmt.Printf("Remaining days: %d days\n", days)
	} else {
		fmt.Println("✗ License verification failed")
		fmt.Printf("Error: %s\n", result.Error)
	}

	// Output machine info match status
	if result.MachineInfo.Matched {
		fmt.Println("✓ Machine information matched")
	} else {
		fmt.Println("✗ Machine information does not match")
		fmt.Printf("Current MAC: %s\n", result.MachineInfo.MAC)
		fmt.Printf("Current UUID: %s\n", result.MachineInfo.UUID)
		fmt.Printf("Current CPUID: %s\n", result.MachineInfo.CPUID)
	}
}

func handleInfo() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: lkctl info <license-file>")
		os.Exit(1)
	}

	licenseFile := os.Args[2]

	// Create verifier
	verifier, err := license.NewVerifierFromFiles("keys/public.pem", "keys/aes.key")
	if err != nil {
		fmt.Printf("Failed to create verifier: %v\n", err)
		fmt.Println("Please make sure the key files exist: keys/public.pem, keys/aes.key")
		os.Exit(1)
	}

	// Get license info
	lic, err := verifier.GetLicenseInfo(licenseFile)
	if err != nil {
		fmt.Printf("Failed to get license info: %v\n", err)
		os.Exit(1)
	}

	// Output info
	data, err := json.MarshalIndent(lic, "", "  ")
	if err != nil {
		fmt.Printf("Failed to serialize license info: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(data))
}

func handleKeys() {
	fs := flag.NewFlagSet("keys", flag.ExitOnError)
	output := fs.String("output", ".", "Output directory")
	fs.Parse(os.Args[2:])

	// Create generator
	generator, err := license.NewGenerator()
	if err != nil {
		fmt.Printf("Failed to create generator: %v\n", err)
		os.Exit(1)
	}

	// Ensure output directory exists
	err = os.MkdirAll(*output, 0755)
	if err != nil {
		fmt.Printf("Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	// Save keys
	privateKeyPath := *output + "/private.pem"
	publicKeyPath := *output + "/public.pem"
	aesKeyPath := *output + "/aes.key"

	err = generator.SaveKeys(privateKeyPath, publicKeyPath, aesKeyPath)
	if err != nil {
		fmt.Printf("Failed to save keys: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Keys generated:\n")
	fmt.Printf("  Private key: %s\n", privateKeyPath)
	fmt.Printf("  Public key: %s\n", publicKeyPath)
	fmt.Printf("  AES key: %s\n", aesKeyPath)
}
