package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cuilan/license-key-verify/pkg/license"
)

func handleVerify() {
	// 手动解析：文件名与 flag 顺序随意（flag.FlagSet 遇到位置参数会停止解析）
	var (
		publicKey  string
		jsonOutput bool
		quiet      bool
	)
	rest := os.Args[2:]
	var args []string
	for i := 0; i < len(rest); i++ {
		arg := rest[i]
		switch arg {
		case "--json":
			jsonOutput = true
		case "--quiet":
			quiet = true
		case "--public-key":
			if i+1 >= len(rest) {
				fmt.Println("--public-key requires a file path")
				os.Exit(2)
			}
			i++
			publicKey = rest[i]
		default:
			if strings.HasPrefix(arg, "-") {
				fmt.Printf("Unknown option: %s\n", arg)
				os.Exit(2)
			}
			args = append(args, arg)
		}
	}
	if len(args) == 0 {
		fmt.Println("Usage: lkctl verify <license-file> [--public-key <pem>] [--json] [--quiet]")
		os.Exit(2)
	}
	if len(args) > 1 {
		fmt.Println("Only one license file can be specified")
		os.Exit(2)
	}

	// 通用工具：无内嵌公钥时 --public-key 必需（fork 后内嵌了公钥可省略）
	if publicKey == "" && !license.HasEmbeddedKey() {
		fatal("no public key: pass --public-key <pem> to verify\n" +
			"  (lkctl built from a fork with an embedded key does not need this flag)")
	}

	applyPublicKeyOverride(publicKey)
	result := license.VerifyFile(args[0])

	if jsonOutput {
		printResultJSON(result)
	} else if !quiet {
		printResult(result)
	}
	if !result.Valid {
		os.Exit(1)
	}
}
