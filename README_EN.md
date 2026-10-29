# License Key Verify Tool

[![CI](https://github.com/cuilan/license-key-verify/actions/workflows/ci.yml/badge.svg)](https://github.com/cuilan/license-key-verify/actions/workflows/ci.yml)
[![Release](https://github.com/cuilan/license-key-verify/actions/workflows/release.yml/badge.svg)](https://github.com/cuilan/license-key-verify/actions/workflows/release.yml)
[![Docker](https://github.com/cuilan/license-key-verify/actions/workflows/docker.yml/badge.svg)](https://github.com/cuilan/license-key-verify/actions/workflows/docker.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

License signing and verification for software licensing: **one-line integration, zero key files on the deployment machine, pure RSA signing**.

[中文文档](README.md) | **English**

## Key Features

- **One-line integration**: `import _ ".../autocheck"` or `license.MustCheck()`
- **Zero key files**: public key is compiled into the binary (`lkctl embed`); deployment machines hold only the binary + license file
- **Pure RSA signing**: RSA-2048 + SHA256, no symmetric crypto, smaller attack surface
- **Machine binding**: single fingerprint (platform UUID + MAC via SHA-256), locked at signing time
- **Offline verification**: signature, validity window and fingerprint are all checked locally
- **Standard library only**: zero third-party dependencies

## Quick Start

### 1. Download lkctl

Grab the archive for your platform from the [Releases](https://github.com/cuilan/license-key-verify/releases/latest) page and extract it — you get the `lkctl` binary (`lkctl.exe` on Windows). No installation needed; optionally add it to your `PATH`.

### 2. Generate a key pair (on your machine)

```bash
lkctl keygen --out keys/
# keys/private.pem  signing key, keep it secret
# keys/public.pem   verification key, used for embedding
```

### 3. Integrate into your app (one-time, during development)

Generate the public key as a Go source file into a sub-package of your app (package name defaults to the directory name):

```bash
lkctl embed --public-key keys/public.pem --out internal/licensekey/key.go
```

Wire the check into your app with one line (fork scenario can use `import _ autocheck`, see [Application Integration](#application-integration)):

```go
import (
    _ "yourapp/internal/licensekey" // register the public key
    "github.com/cuilan/license-key-verify/v2/pkg/license"
)

func main() {
    license.MustCheck() // prints the reason and exits with code 1 on failure
    // ... your business code
}
```

Rebuild your app — it can now validate licenses, and deployment machines hold zero key files.

### 4. Get the fingerprint from the customer machine

```bash
# Option 1: your app has the check built in — just run with the flag
./your-app --fingerprint
# output: 090499363c0b2202d7f025cd7dc2dedce65421c0a25eaccb55b890e64bf73625

# Option 2: run lkctl on the customer machine
lkctl fingerprint
```

### 5. Sign the license (on your machine)

`sign` only needs the private key — lkctl is a generic tool, no need to rebuild lkctl:

```bash
lkctl sign \
  --private-key keys/private.pem \
  --fingerprint 090499363c0b... \
  --customer "Acme Corp" \
  --product "leosap" \
  --features "full" \
  --duration 365 \
  --out license.lic
```

### 6. Deliver and verify

Send `license.lic` to the customer. Drop it at the app's default location (or point `LICENSE_FILE` to it) and the app validates on startup. Manual verification (a generic lkctl needs the public key):

```bash
lkctl verify license.lic --public-key keys/public.pem    # exit code 0=valid 1=invalid
lkctl verify license.lic --public-key keys/public.pem --json
lkctl info license.lic   --public-key keys/public.pem    # show license content
```

## Application Integration

Fully runnable examples (both autocheck and SDK styles) live in [lkctl-examples](https://github.com/cuilan/lkctl-examples).

> **Where to put key.go?** Put it in a **sub-package** of your app (e.g. `internal/licensekey/`):
> `lkctl embed --public-key keys/public.pem --out internal/licensekey/key.go`
> (the package name defaults to the directory name). Do NOT put it in package main —
> Go initializes imported packages before the main package; a key.go in package main
> registers the key too late.

### Mode A: import-side-effect check (zero code changes, fork scenario)

Fork this repo and overwrite `pkg/license/key.go` with `lkctl embed` (key.go then lives in an imported package, initialized before autocheck):

```go
package main

import _ "github.com/cuilan/license-key-verify/v2/autocheck"

func main() {
    // validation happens in init(); the process exits on failure
}
```

### Mode B: explicit one-line call (direct module usage)

```go
package main

import (
    _ "yourapp/internal/licensekey" // register the public key
    "github.com/cuilan/license-key-verify/v2/pkg/license"
)

func main() {
    license.MustCheck() // prints the reason and exits with code 1

    // or handle it yourself:
    // if err := license.Check(); err != nil { ... }
}
```

License path resolution order:

1. `LICENSE_FILE` environment variable
2. `license.lic` next to the executable
3. Unix: `/etc/<program-name>/license.lic`

### Embedding the public key (application-side integration step)

`lkctl` itself is a generic tool: `sign` only needs the private key and `verify` takes `--public-key`; neither depends on embedding. `lkctl embed` is used when **you develop the customer-facing app** — it generates a Go source file containing your public key with light XOR obfuscation (so `strings` cannot dump it directly); it compiles into your app. Deployment machines hold no `.pem` / `.key` files at all:

```bash
# Fork scenario: overwrite pkg/license/key.go (pairs with Mode A)
lkctl embed --public-key keys/public.pem --out pkg/license/key.go

# Module scenario: generate into a sub-package of your app (package name = directory name, pairs with Mode B)
lkctl embed --public-key keys/public.pem --out internal/licensekey/key.go
```

- If you consume this project as a Go module and cannot ship a `key.go`, inject the key at runtime instead:
  call `license.UsePublicKeyPEM(publicKeyPEM)` before `license.MustCheck()`.

## Command Reference

### lkctl (vendor side)

```bash
lkctl keygen --out keys/                      Generate an RSA key pair
lkctl embed --public-key <pem> --out <file>   Emit the public key as a Go source file
lkctl fingerprint                             Print this machine's fingerprint
lkctl sign --private-key <pem> --fingerprint <fp> --out license.lic
        [--duration 365] [--customer ...] [--product ...] [--features a,b]
lkctl verify <license-file> [--public-key <pem>] [--json] [--quiet]
                                              Verify a license (exit codes: 0=valid 1=invalid 2=arg error)
lkctl info <license-file> [--public-key <pem>]     Show license content
```

## License File Format

Plain JSON + RSA signature (the signature covers every field except `signature` itself):

```json
{
  "license_id": "2366ee07-1c29-979a-6106-e253ed47301c",
  "customer": "Acme Corp",
  "product": "leosap",
  "fingerprint": "090499363c0b2202d7f025cd7dc2dedce65421c0a25eaccb55b890e64bf73625",
  "issued_at": "2026-10-27T00:00:00Z",
  "expires_at": "2027-10-27T00:00:00Z",
  "features": ["full"],
  "signature": "<base64 RSA-SHA256 signature>"
}
```

## SDK

```go
// Signing (vendor side)
lic, err := license.Sign(&license.SignOptions{
    Customer:    "Acme Corp",
    Product:     "leosap",
    Fingerprint: fp,
    Features:    []string{"full"},
    Duration:    365 * 24 * time.Hour,
}, privateKeyPEM)
err = lic.Save("license.lic")

// Verification (application side)
result := license.VerifyFile("license.lic")
if result.Valid {
    fmt.Println("days remaining:", result.ExpiresIn/(24*3600))
}
```

## Machine Fingerprint

fingerprint = `SHA-256(version | platform stable id | sorted MAC list)`

- Platform stable id: Linux prefers `product_uuid` (falls back to `machine-id` when DMI is unavailable), macOS uses `IOPlatformUUID`, Windows uses the system UUID
- No random fallback values; the same machine always computes the same fingerprint
- OS reinstall / hardware change / NIC change alters the fingerprint and requires re-signing

## Security Notes

1. **The private key never leaves you**: leaking it means anyone can forge licenses
2. **The repo ships no keys at all**: there is no demo key pair; generate your own via `lkctl keygen`
3. **Tamper protection comes from the signature**: licenses are readable JSON; nobody can modify content without the private key
4. **No rollback/revocation**: offline verification does not detect clock rollback and has no revocation list; extend it yourself if needed

## License

MIT. See the LICENSE file.
