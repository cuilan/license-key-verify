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

### 1. Build

```bash
make build   # produces bin/lkctl (bin/lkctl.exe on Windows)
```

### 2. Generate a key pair (on your machine)

```bash
lkctl keygen --out keys/
# keys/private.pem  signing key, keep it secret
# keys/public.pem   verification key, used for embedding
```

### 3. Get the fingerprint from the customer machine

```bash
# Option 1: your app embeds autocheck — just run with the flag
./your-app --fingerprint
# output: 090499363c0b2202d7f025cd7dc2dedce65421c0a25eaccb55b890e64bf73625

# Option 2: run lkctl on the customer machine
lkctl fingerprint
```

### 4. Sign the license (on your machine)

`sign` only needs the private key — lkctl is a generic tool, no embedding or rebuild required:

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

### 5. Deliver and verify

Send `license.lic` to the customer. Drop it at the app's default location (or point `LICENSE_FILE` to it) and the app validates on startup. Manual verification (a generic lkctl needs the public key):

```bash
lkctl verify license.lic --public-key keys/public.pem    # exit code 0=valid 1=invalid
lkctl verify license.lic --public-key keys/public.pem --json
lkctl info license.lic   --public-key keys/public.pem    # show license content
```

## Application Integration

### Mode A: import-side-effect check (zero code changes)

```go
package main

import _ "github.com/cuilan/license-key-verify/autocheck"

func main() {
    // validation happens in init(); the process exits on failure
}
```

### Mode B: explicit one-line call

```go
package main

import "github.com/cuilan/license-key-verify/pkg/license"

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
lkctl embed --public-key keys/public.pem --out pkg/license/key.go
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
2. **The repo ships no keys at all**: there is no demo key pair; signing and verification are impossible until `lkctl embed` has been run, which prevents "forgot to swap the key" incidents
3. **Tamper protection comes from the signature**: licenses are readable JSON; nobody can modify content without the private key
4. **No rollback/revocation**: offline verification does not detect clock rollback and has no revocation list; extend it yourself if needed

## Docker

```bash
docker build -t license-key-verify .
docker run --rm license-key-verify help
docker run --rm -v $(pwd):/workspace license-key-verify fingerprint
```

## Development

```bash
make test          # unit tests
make test-e2e      # end-to-end tests (generates its own keys)
make build-all     # cross-platform build
```

## FAQ

**Q: Can the license file be copied to another machine?**
No. The fingerprint mismatch fails verification.

**Q: What if I forget to embed the public key?**
`sign` is unaffected (it only needs the private key). `lkctl verify` tells you to pass `--public-key <pem>`; your app without an embedded key reports `no public key available` at check time — run `lkctl embed` and put key.go into your app, then rebuild.

**Q: Are Windows / macOS / Linux supported?**
Yes. Fingerprint collection and builds cover all three platforms.

## License

MIT. See the LICENSE file.
