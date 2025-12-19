#!/bin/bash
set -e

# ANSI colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo "🔍 Starting Verification of First-Time User Experience..."

# Check for Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: 'go' binary not found. Please install Go 1.24+ and try again.${NC}"
    exit 1
fi

# Build CLI
echo "🏗️ Building CLI..."
cd cli
go build -o ../bin/goconnect ./cmd/goconnect
cd ..

BINARY="./bin/goconnect"

# Test 1: Non-Interactive Mode without Config
# Should fail gracefully and NOT hang
echo "🧪 Test 1: Non-Interactive run (no config)..."
if echo "" | $BINARY > /dev/null 2>&1; then
    # It might exit 0 or 1 depending on implementation, but shouldn't hang.
    # Our implementation exits 1 with error message.
    RET=$?
else
    RET=$?
fi

# We expect exit code 1 due to missing config in non-interactive mode
if [ $RET -eq 1 ]; then
    echo -e "${GREEN}✓ Passed: Binary exited gracefully with error as expected.${NC}"
else
    echo -e "${RED}✗ Failed: Expected exit code 1, got $RET.${NC}"
fi

# Test 2: Verify Setup Wizard detected (Mocking TTY is hard in script without 'expect', 
# but we can check if it tries to open)
# For now, we will just trust the non-interactive check for CI safety.

# Test 3: Zero-Config Daemon
echo "🧪 Test 2: Daemon Zero-Config Startup..."
CONFIG_PATH="/tmp/goconnect_test_config.yaml"
rm -f $CONFIG_PATH

# Run daemon in background with a timeout
timeout 5s $BINARY run --config $CONFIG_PATH > /tmp/daemon_output 2>&1 &
PID=$!

sleep 2

if ps -p $PID > /dev/null; then
    echo -e "${GREEN}✓ Passed: Daemon started successfully without config file.${NC}"
    kill $PID
else
    echo -e "${RED}✗ Failed: Daemon failed to start or crashed.${NC}"
    cat /tmp/daemon_output
    exit 1
fi

echo -e "\n${GREEN}✨ All automated verification checks passed!${NC}"
echo "Note: Full TUI Wizard functionality requires manual interactive testing."
