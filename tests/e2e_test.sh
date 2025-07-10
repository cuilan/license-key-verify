#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e
# Print commands and their arguments as they are executed.
set -x

# Get the directory of this script, which is the 'tests' directory
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)
PROJECT_ROOT="$SCRIPT_DIR/.."

# Define paths relative to the project root
BIN_DIR="$PROJECT_ROOT/bin"
LKCTL="$BIN_DIR/lkctl"
LKVERIFY="$BIN_DIR/lkverify"
TEST_DIR="$SCRIPT_DIR/test_run" # Run tests inside the tests/ directory
KEYS_DIR="$TEST_DIR/keys"
LICENSE_FILE="$TEST_DIR/license.lic"

# 1. Setup test environment
echo "---- 1. Setting up test environment ----"
rm -rf $TEST_DIR
mkdir -p $KEYS_DIR

# 2. Test 'lkctl get'
echo "---- 2. Testing 'lkctl get all' ----"
$LKCTL get all

# 3. Test 'lkctl gen' and successful verification
echo "---- 3.1. Generating a new license and keys ----"
$LKCTL gen --keys-dir $KEYS_DIR $LICENSE_FILE

echo "---- 3.2. Verifying the license with the correct keys ----"
$LKVERIFY $LICENSE_FILE --public-key $KEYS_DIR/public.pem --aes-key $KEYS_DIR/aes.key

echo "---- 3.3. Generating a second license reusing existing keys ----"
LICENSE_FILE_2="$TEST_DIR/license2.lic"
$LKCTL gen --private-key $KEYS_DIR/private.pem --aes-key $KEYS_DIR/aes.key $LICENSE_FILE_2

echo "---- 3.4. Verifying the second license ----"
$LKVERIFY $LICENSE_FILE_2 --public-key $KEYS_DIR/public.pem --aes-key $KEYS_DIR/aes.key

# 4. Test 'lkverify' with incorrect keys
echo "---- 4.1. Generating a different set of keys ----"
KEYS_DIR_2="$TEST_DIR/wrong_keys"
mkdir -p $KEYS_DIR_2
$LKCTL gen --keys-dir $KEYS_DIR_2 "$TEST_DIR/dummy.lic"

echo "---- 4.2. Attempting to verify the original license with wrong keys (expected to fail) ----"
# We expect this command to fail, so we temporarily disable 'exit on error'
set +e
$LKVERIFY $LICENSE_FILE --public-key $KEYS_DIR_2/public.pem --aes-key $KEYS_DIR_2/aes.key
# Check that the command failed as expected (exit code 1)
if [ $? -ne 1 ]; then
    echo "ERROR: Verification with wrong keys succeeded, but it should have failed."
    exit 1
fi
set -e

echo "---- All tests passed successfully! ----" 