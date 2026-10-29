// Package autocheck 提供「import 即校验」的零代码集成方式：
//
//	import _ "github.com/cuilan/license-key-verify/autocheck"
//
// 包初始化时自动定位许可证文件并使用内置公钥验签：
//   - 解析顺序：环境变量 LICENSE_FILE → 可执行文件同目录 license.lic → /etc/<程序名>/license.lic
//   - 校验失败：输出明确错误并以退出码 1 终止进程
//
// 同时自动支持 `--fingerprint` 参数：带该参数启动时打印本机指纹后退出，
// 供客户把指纹发回给你签发许可证。
//
// 注意：autocheck 使用编译进 pkg/license/key.go 的内置公钥。
// 直接引用本模块（未 fork 替换 key.go）的项目请改用 license.UsePublicKeyPEM + license.MustCheck。
package autocheck

import (
	"fmt"
	"os"

	"github.com/cuilan/license-key-verify/pkg/license"
)

func init() {
	// 应用带 --fingerprint 启动时：打印本机指纹并退出，跳过校验
	for _, arg := range os.Args[1:] {
		if arg == "--fingerprint" || arg == "-fingerprint" {
			license.PrintFingerprint()
			os.Exit(0)
		}
	}

	if err := license.Check(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
