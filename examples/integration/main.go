// 本示例展示新 API 下的两种集成方式（公钥已编译进二进制，部署机器零密钥文件）。
//
// 方式 A（零代码）：import 即校验
//
//	import _ "github.com/cuilan/license-key-verify/autocheck"
//
//	func main() { /* 业务代码 */ }
//
// 方式 B（一行调用）：显式校验
//
//	func main() {
//	    license.MustCheck() // 失败输出原因并以退出码 1 退出
//	    /* 业务代码 */
//	}
//
// 生产环境：用 lkctl embed 生成新的 pkg/license/key.go（内嵌你自己的公钥）后重新编译。
package main

import (
	"fmt"

	"github.com/cuilan/license-key-verify/pkg/license"
)

func main() {
	// 一行集成：校验失败直接退出
	license.MustCheck()

	// 运行到这里说明许可证有效
	fmt.Println("license OK, application starts...")

	// 需要更细的控制时（例如按 feature 开关功能）：
	result := license.VerifyFile(license.DefaultLicensePath())
	if len(result.License.Features) > 0 {
		fmt.Printf("licensed features: %v\n", result.License.Features)
	}
	fmt.Printf("expires in %d days\n", result.ExpiresIn/(24*3600))
}
