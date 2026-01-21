#!/bin/bash

# Test coverage script for SLURM exporter
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
COVERAGE_THRESHOLD=20
COVERAGE_FILE="coverage.out"
COVERAGE_HTML="coverage.html"
COVERAGE_DIR="coverage"

echo -e "${GREEN}Running test coverage analysis...${NC}"

# Create coverage directory
mkdir -p $COVERAGE_DIR

# Run tests with coverage
echo -e "${YELLOW}Running tests with coverage...${NC}"
go test -v -race -coverprofile=$COVERAGE_DIR/$COVERAGE_FILE ./...

# Check if coverage file was created
if [ ! -f "$COVERAGE_DIR/$COVERAGE_FILE" ]; then
    echo -e "${RED}Error: Coverage file not generated${NC}"
    exit 1
fi

# Generate coverage report
echo -e "${YELLOW}Generating coverage report...${NC}"
go tool cover -html=$COVERAGE_DIR/$COVERAGE_FILE -o $COVERAGE_DIR/$COVERAGE_HTML

# Calculate total coverage percentage
COVERAGE_PERCENT=$(go tool cover -func=$COVERAGE_DIR/$COVERAGE_FILE | grep total | awk '{print $3}' | sed 's/%//')

echo -e "${YELLOW}Coverage Summary:${NC}"
go tool cover -func=$COVERAGE_DIR/$COVERAGE_FILE | tail -10

# Check coverage threshold
if (( $(echo "$COVERAGE_PERCENT >= $COVERAGE_THRESHOLD" | bc -l) )); then
    echo -e "${GREEN}✅ Coverage: ${COVERAGE_PERCENT}% (threshold: ${COVERAGE_THRESHOLD}%)${NC}"
    echo -e "${GREEN}Coverage report generated: $COVERAGE_DIR/$COVERAGE_HTML${NC}"
else
    echo -e "${RED}❌ Coverage: ${COVERAGE_PERCENT}% (below threshold: ${COVERAGE_THRESHOLD}%)${NC}"
    echo -e "${RED}Coverage report generated: $COVERAGE_DIR/$COVERAGE_HTML${NC}"
    exit 1
fi

# Generate coverage by package
echo -e "${YELLOW}Coverage by package:${NC}"
go tool cover -func=$COVERAGE_DIR/$COVERAGE_FILE | grep -v "total:" | awk '{
    package = $1
    gsub(/.*\//, "", package)
    gsub(/\.go.*/, "", package)
    if (packages[package] == "") {
        packages[package] = $3
        count[package] = 1
    } else {
        # Simple average - could be weighted by lines
        split(packages[package], a, "%")
        split($3, b, "%")
        packages[package] = ((a[1] * count[package] + b[1]) / (count[package] + 1)) "%"
        count[package]++
    }
}
END {
    for (pkg in packages) {
        printf "%-20s %s\n", pkg, packages[pkg]
    }
}' | sort

echo -e "${GREEN}Test coverage analysis complete!${NC}"

# Open coverage report in browser (optional) - don't fail if no browser available
if command -v open > /dev/null 2>&1; then
    echo -e "${YELLOW}Opening coverage report in browser...${NC}"
    open $COVERAGE_DIR/$COVERAGE_HTML 2>/dev/null || true
elif command -v xdg-open > /dev/null 2>&1; then
    echo -e "${YELLOW}Opening coverage report in browser...${NC}"
    xdg-open $COVERAGE_DIR/$COVERAGE_HTML 2>/dev/null || true
fi