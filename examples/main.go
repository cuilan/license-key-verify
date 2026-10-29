// 本示例演示公钥内嵌的完整流程（部署机器零密钥文件）。
//
// 复现本目录（key.go 的包名由 embed 自动检测为 main）：
//
//	lkctl keygen --out keys/
//	lkctl sign --private-key keys/private.pem \
//	           --fingerprint "$(lkctl fingerprint)" \
//	           --out license.lic
//	lkctl embed --public-key keys/public.pem --out key.go
//	go run .
//
// sign/verify 是通用能力：sign 只需 --private-key，无需任何内嵌公钥；
// embed 仅在应用侧集成 license SDK 时使用，与 lkctl 自身无关。
//
// 集成方式 A（零代码）：import 即校验
//
//	import _ "github.com/cuilan/license-key-verify/autocheck"
//
// 集成方式 B（一行调用）：显式校验
//
//	license.MustCheck() // 失败输出原因并以退出码 1 退出
package main

import (
	"fmt"

	_ "github.com/cuilan/license-key-verify/autocheck"
	// "github.com/cuilan/license-key-verify/pkg/license"
)

func main() {
	// 一行集成：校验失败直接退出
	// 	license.MustCheck()

	// 	// 运行到这里说明许可证有效
	// 	fmt.Println("license OK, application starts...")

	// // 需要更细的控制时（例如按 feature 开关功能）：
	// result := license.VerifyFile(license.DefaultLicensePath())
	//
	//	if len(result.License.Features) > 0 {
	//		fmt.Printf("licensed features: %v\n", result.License.Features)
	//	}
	//
	// fmt.Printf("expires in %d days\n", result.ExpiresIn/(24*3600))
	fmt.Println("成功启动")
}
