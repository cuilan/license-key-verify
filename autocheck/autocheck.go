// Package autocheck 提供「import 即校验」的零代码集成方式：
//
//	import _ "github.com/cuilan/license-key-verify/v2/autocheck"
//
// 包初始化时自动定位许可证文件并使用内置公钥验签：
//   - 解析顺序：环境变量 LICENSE_FILE → 可执行文件同目录 license.lic → /etc/<程序名>/license.lic
//   - 校验失败：输出明确错误并以退出码 1 终止进程
//
// 同时自动支持 `--fingerprint` 参数：带该参数启动时打印本机指纹后退出，
// 供客户把指纹发回给你签发许可证。
//
// 公钥来源（Go 初始化顺序要求）：
//
//	Go 先初始化被导入包、后初始化 main 包，而本校验发生在 autocheck 的 init 里，
//	因此公钥必须来自「被导入的包」：
//	  - fork 本仓库：lkctl embed 覆盖 pkg/license/key.go（同包，天然满足）；
//	  - 直接引用模块：把 embed 生成的 key.go 放进应用的一个子包
//	    （如 internal/licensekey/），并在 autocheck 之前 import 它：
//	      import _ "yourapp/internal/licensekey"
//	      import _ "github.com/cuilan/license-key-verify/v2/autocheck"
//	    放在 package main 里的 key.go 注册时机太晚，校验会报 no public key available。
package autocheck

import (
	"fmt"
	"os"

	"github.com/cuilan/license-key-verify/v2/pkg/license"
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
