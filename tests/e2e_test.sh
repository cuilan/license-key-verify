#!/bin/bash

# E2E test for license-key-verify.
# Self-contained: generates its own key pair (the repo ships no demo keys).
# Run via: make test-e2e  (or ./tests/e2e_test.sh after `make build`)
set -e
set -x

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)
PROJECT_ROOT="$SCRIPT_DIR/.."

BIN_DIR="$PROJECT_ROOT/bin"
# Pick the binary matching the current shell environment:
# Git Bash/MSYS expects lkctl.exe, Linux/WSL expects lkctl (see the "build" target)
case "$(uname -s)" in
    MINGW*|MSYS*|CYGWIN*)
        [ -f "$BIN_DIR/lkctl.exe" ] && LKCTL="$BIN_DIR/lkctl.exe" || LKCTL="$BIN_DIR/lkctl"
        ;;
    *)
        [ -f "$BIN_DIR/lkctl" ] && LKCTL="$BIN_DIR/lkctl" || LKCTL="$BIN_DIR/lkctl.exe"
        ;;
esac
TEST_DIR="$SCRIPT_DIR/test_run"

# 1. Setup test environment
echo "---- 1. Setting up test environment ----"
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"

# 2. Generate a fresh key pair (users must do this too; no demo keys exist)
echo "---- 2. lkctl keygen ----"
$LKCTL keygen --out "$TEST_DIR/keys"
PRIVATE_KEY="$TEST_DIR/keys/private.pem"
PUBLIC_KEY="$TEST_DIR/keys/public.pem"

# 3. Test 'lkctl fingerprint'
echo "---- 3. Testing 'lkctl fingerprint' ----"
FP=$($LKCTL fingerprint)
echo "fingerprint: $FP"
if [ ${#FP} -ne 64 ]; then
    echo "ERROR: fingerprint should be 64 hex chars, got: $FP"
    exit 1
fi

# 4. Core flows: sign works standalone (lkctl is a generic tool, no embedded key needed)
echo "---- 4.1. sign works without any embedded key ----"
LICENSE_FILE="$TEST_DIR/license.lic"
$LKCTL sign \
    --private-key "$PRIVATE_KEY" \
    --fingerprint "$FP" \
    --customer "E2E客户" \
    --product "demo" \
    --features "full,api" \
    --duration 30 \
    --out "$LICENSE_FILE"

echo "---- 4.2. verify without embedded key must ask for --public-key ----"
set +e
$LKCTL verify "$LICENSE_FILE" 2> /dev/null
if [ $? -eq 0 ]; then
    echo "ERROR: verify should fail without --public-key"
    exit 1
fi
set -e

echo "---- 4.3. verify with --public-key ----"
$LKCTL verify "$LICENSE_FILE" --public-key "$PUBLIC_KEY"

echo "---- 4.4. JSON output ----"
$LKCTL verify "$LICENSE_FILE" --public-key "$PUBLIC_KEY" --json > /dev/null

echo "---- 4.5. lkctl info ----"
$LKCTL info "$LICENSE_FILE" --public-key "$PUBLIC_KEY" > /dev/null

# 5. Negative cases
echo "---- 5.1. Tampered license must be rejected ----"
set +e
sed 's/"customer": *"[^"]*"/"customer": "attacker"/' "$LICENSE_FILE" > "$TEST_DIR/tampered.lic"
$LKCTL verify "$TEST_DIR/tampered.lic" --public-key "$PUBLIC_KEY" --quiet
if [ $? -eq 0 ]; then
    echo "ERROR: tampered license passed verification"
    exit 1
fi
set -e

echo "---- 5.2. Foreign fingerprint must be rejected ----"
set +e
$LKCTL sign \
    --private-key "$PRIVATE_KEY" \
    --fingerprint "$(printf 'a%.0s' $(seq 1 64))" \
    --out "$TEST_DIR/foreign.lic"
$LKCTL verify "$TEST_DIR/foreign.lic" --public-key "$PUBLIC_KEY" --quiet
if [ $? -eq 0 ]; then
    echo "ERROR: license with foreign fingerprint passed verification"
    exit 1
fi
set -e

echo "---- 5.3. Sign without fingerprint must fail ----"
set +e
$LKCTL sign --private-key "$PRIVATE_KEY" --out "$TEST_DIR/nofp.lic" 2> /dev/null
if [ $? -eq 0 ]; then
    echo "ERROR: sign without fingerprint should fail"
    exit 1
fi
set -e

# 6. embed generates a working key.go: compile and verify with it
echo "---- 6. lkctl embed + compile check ----"
mkdir -p "$TEST_DIR/embedcheck"
$LKCTL embed --public-key "$PUBLIC_KEY" --out "$TEST_DIR/embedcheck/key_embedded.go" --package main

cat > "$TEST_DIR/embedcheck/main.go" <<'EOF'
package main

import (
	"fmt"
	"os"

	"github.com/cuilan/license-key-verify/v2/pkg/license"
)

func main() {
	if !license.HasEmbeddedKey() {
		fmt.Println("FAIL: no embedded key")
		os.Exit(1)
	}
	// 这里只验证"内嵌公钥与签发私钥配对"（签名通过即 License 非空）。
	// 指纹可能因签发与验证运行在不同环境而不同，不作为本检查的判据。
	r := license.VerifyFile(os.Args[1])
	if r.License == nil {
		fmt.Println("FAIL: signature:", r.Error)
		os.Exit(1)
	}
	fmt.Println("embedded-key verify OK")
}
EOF

# Convert a path for consumption by native Windows executables (go.exe / check.exe)
to_native_path() {
    if [ -n "$WSL_DISTRO_NAME" ] && command -v wslpath > /dev/null 2>&1; then
        wslpath -w "$1" | sed 's|\\|/|g'
    elif [ -n "$MSYSTEM" ] && command -v cygpath > /dev/null 2>&1; then
        cygpath -m "$1"
    else
        echo "$1"
    fi
}

# go.mod with a local replace so no network access is needed
GO_BIN=$(command -v go || command -v go.exe || true)
if [ -z "$GO_BIN" ]; then
    echo "WARN: go not found, skipping embed compile check"
else
    ROOT_WIN=$(to_native_path "$PROJECT_ROOT")
    cat > "$TEST_DIR/embedcheck/go.mod" <<EOF
module embedcheck

go 1.23

require github.com/cuilan/license-key-verify v0.0.0

replace github.com/cuilan/license-key-verify => $ROOT_WIN
EOF

    CHECK_ARG=$(to_native_path "$LICENSE_FILE")
    (cd "$TEST_DIR/embedcheck" && "$GO_BIN" mod tidy && "$GO_BIN" build -o check.exe . && ./check.exe "$CHECK_ARG")
    $TEST_DIR/embedcheck/check.exe "$CHECK_ARG" | grep -q "embedded-key verify OK" || {
        echo "ERROR: embedded key compile/verify check failed"
        exit 1
    }
fi

echo "---- All tests passed successfully! ----"
