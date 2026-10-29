#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e
# Print commands and their arguments as they are executed.
set -x

# Get the directory of this script, which is the 'tests' directory
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)
PROJECT_ROOT="$SCRIPT_DIR/.."

BIN_DIR="$PROJECT_ROOT/bin"
LKCTL="$BIN_DIR/lkctl"
TEST_DIR="$SCRIPT_DIR/test_run"
PRIVATE_KEY="$PROJECT_ROOT/testdata/private.pem" # 演示私钥，与 pkg/license/key.go 内嵌公钥配对

# 1. Setup test environment
echo "---- 1. Setting up test environment ----"
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"

# 2. Test 'lkctl fingerprint'
echo "---- 2. Testing 'lkctl fingerprint' ----"
FP=$($LKCTL fingerprint)
echo "fingerprint: $FP"
if [ ${#FP} -ne 64 ]; then
    echo "ERROR: fingerprint should be 64 hex chars, got: $FP"
    exit 1
fi

# 3. Sign and verify a license with the embedded public key (zero key files)
echo "---- 3.1. Signing a license ----"
LICENSE_FILE="$TEST_DIR/license.lic"
$LKCTL sign \
    --private-key "$PRIVATE_KEY" \
    --fingerprint "$FP" \
    --customer "E2E客户" \
    --product "demo" \
    --features "full,api" \
    --duration 30 \
    --out "$LICENSE_FILE"

echo "---- 3.2. Verifying with lkctl verify (embedded public key, no key files) ----"
$LKCTL verify "$LICENSE_FILE"

echo "---- 3.3. JSON output ----"
$LKCTL verify "$LICENSE_FILE" --json > /dev/null

echo "---- 3.4. lkctl info ----"
$LKCTL info "$LICENSE_FILE" > /dev/null

# 4. Negative cases
echo "---- 4.1. Tampered license must be rejected ----"
set +e
sed 's/"customer": *"[^"]*"/"customer": "attacker"/' "$LICENSE_FILE" > "$TEST_DIR/tampered.lic"
$LKCTL "$TEST_DIR/tampered.lic" --quiet
if [ $? -eq 0 ]; then
    echo "ERROR: tampered license passed verification"
    exit 1
fi
set -e

echo "---- 4.2. Foreign fingerprint must be rejected ----"
set +e
$LKCTL sign \
    --private-key "$PRIVATE_KEY" \
    --fingerprint "$(printf 'a%.0s' $(seq 1 64))" \
    --out "$TEST_DIR/foreign.lic"
$LKCTL "$TEST_DIR/foreign.lic" --quiet
if [ $? -eq 0 ]; then
    echo "ERROR: license with foreign fingerprint passed verification"
    exit 1
fi
set -e

echo "---- 4.3. Sign without fingerprint must fail ----"
set +e
$LKCTL sign --private-key "$PRIVATE_KEY" --out "$TEST_DIR/nofp.lic" 2> /dev/null
if [ $? -eq 0 ]; then
    echo "ERROR: sign without fingerprint should fail"
    exit 1
fi
set -e

echo "---- All tests passed successfully! ----"
