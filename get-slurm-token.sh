#!/bin/bash
# Utility script to get SLURM JWT token

SLURM_HOST="${1:-rocky9.ar.jontk.com}"

echo "Acquiring JWT token from ${SLURM_HOST}..."

# Get token and parse it
TOKEN_OUTPUT=$(ssh "root@${SLURM_HOST}" "scontrol token" 2>&1)

if [ $? -ne 0 ]; then
    echo "Error: Failed to connect to ${SLURM_HOST}"
    echo "$TOKEN_OUTPUT"
    exit 1
fi

# Extract token value
TOKEN=$(echo "$TOKEN_OUTPUT" | grep -oP 'token=\K.*' || true)

if [ -z "$TOKEN" ]; then
    echo "Error: Could not extract token from response"
    echo "Response: $TOKEN_OUTPUT"
    exit 1
fi

echo ""
echo "Successfully acquired token!"
echo ""
echo "Token value (full):"
echo "$TOKEN"
echo ""
echo "Token value (export for use):"
echo "export SLURM_AUTH_TOKEN=\"$TOKEN\""
echo ""
echo "Token preview:"
TOKEN_PREVIEW="${TOKEN:0:30}...${TOKEN: -30}"
echo "$TOKEN_PREVIEW"
echo ""
echo "To use this token in tests:"
echo "  export SLURM_AUTH_TOKEN=\"$TOKEN\""
echo "  go test -v ./test/integration"
