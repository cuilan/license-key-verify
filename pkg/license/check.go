package license

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cuilan/license-key-verify/pkg/machine"
)

// DefaultLicensePath 返回默认的许可证文件路径，解析顺序：
//  1. 环境变量 LICENSE_FILE
//  2. 可执行文件同目录下的 license.lic
//  3. Unix: /etc/<程序名>/license.lic（Windows 无此层级）
func DefaultLicensePath() string {
	if p := os.Getenv("LICENSE_FILE"); p != "" {
		return p
	}
	exe, err := os.Executable()
	if err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "license.lic")
		if _, statErr := os.Stat(candidate); statErr == nil {
			return candidate
		}
		if runtime.GOOS != "windows" {
			return "/etc/" + appName() + "/license.lic"
		}
	}
	return "license.lic"
}

// Check 校验默认路径下的许可证，成功返回 nil，失败返回包含原因的错误。
func Check() error {
	path := DefaultLicensePath()
	result := VerifyFile(path)
	if result.Valid {
		return nil
	}
	return fmt.Errorf("license check failed (%s): %s", path, result.Error)
}

// MustCheck 等同于 Check，失败时输出错误到 stderr 并以退出码 1 终止进程。
func MustCheck() {
	if err := Check(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// PrintFingerprint 打印当前机器指纹（供应用支持 --fingerprint 参数）。
func PrintFingerprint() {
	fp, err := machine.Fingerprint()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to compute fingerprint: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(fp)
}

// appName 返回可执行文件名（去掉路径与 .exe 后缀），用于默认许可证路径。
func appName() string {
	if len(os.Args) == 0 {
		return "app"
	}
	name := filepath.Base(os.Args[0])
	return strings.TrimSuffix(name, ".exe")
}
