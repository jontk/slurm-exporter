#!/bin/bash

# Performance profiling script for SLURM exporter
set -e

# Configuration
PROFILE_TYPE="${PROFILE_TYPE:-cpu}"
PROFILE_DURATION="${PROFILE_DURATION:-30s}"
OUTPUT_DIR="${OUTPUT_DIR:-profile-results}"
EXPORTER_PID="${EXPORTER_PID:-}"
BENCHMARK_ARGS="${BENCHMARK_ARGS:--bench=. -benchmem}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}SLURM Exporter Performance Profiling${NC}"
echo "======================================"
echo -e "Profile type: ${BLUE}${PROFILE_TYPE}${NC}"
echo -e "Duration: ${BLUE}${PROFILE_DURATION}${NC}"
echo -e "Output directory: ${BLUE}${OUTPUT_DIR}${NC}"
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Check if required tools are available
check_tool() {
    if ! command -v "$1" > /dev/null 2>&1; then
        echo -e "${RED}Error: $1 is not installed${NC}"
        echo "Please install $1 to run profiling"
        exit 1
    fi
}

# Check for required tools
echo -e "${YELLOW}Checking for profiling tools...${NC}"
check_tool go

# Check for pprof
if ! go tool pprof --help > /dev/null 2>&1; then
    echo -e "${RED}Error: go tool pprof is not available${NC}"
    exit 1
fi
echo -e "${GREEN}✓ go tool pprof found${NC}"

# Check for graphviz (for generating visual profiles)
GRAPHVIZ_AVAILABLE=false
if command -v dot > /dev/null 2>&1; then
    GRAPHVIZ_AVAILABLE=true
    echo -e "${GREEN}✓ Graphviz found (visual profiles available)${NC}"
else
    echo -e "${YELLOW}! Graphviz not found (visual profiles will not be generated)${NC}"
fi

# Function to profile running process
profile_running_process() {
    local pid="$1"
    local profile_type="$2"
    local duration="$3"

    echo -e "${YELLOW}Profiling running process (PID: $pid)...${NC}"

    if ! kill -0 "$pid" 2>/dev/null; then
        echo -e "${RED}Error: Process $pid is not running${NC}"
        return 1
    fi

    local profile_url="http://localhost:6060/debug/pprof/${profile_type}"
    local output_file="$OUTPUT_DIR/runtime_${profile_type}.prof"

    echo "Collecting $profile_type profile for ${duration}..."

    if [ "$profile_type" = "profile" ]; then
        # CPU profiling
        curl -s "${profile_url}?seconds=${duration%s}" > "$output_file"
    else
        # Other profiles (heap, goroutine, etc.)
        curl -s "$profile_url" > "$output_file"
    fi

    if [ -s "$output_file" ]; then
        echo -e "${GREEN}✓ Profile saved: $output_file${NC}"
        analyze_profile "$output_file" "$profile_type"
    else
        echo -e "${RED}✗ Failed to collect profile${NC}"
        return 1
    fi
}

# Function to run benchmark profiling
profile_benchmarks() {
    local profile_type="$1"

    echo -e "${YELLOW}Running benchmark profiling...${NC}"

    local output_file="$OUTPUT_DIR/bench_${profile_type}.prof"
    local bench_output="$OUTPUT_DIR/benchmark_results.txt"

    case "$profile_type" in
        "cpu")
            echo "Running CPU profiling on benchmarks..."
            go test $BENCHMARK_ARGS -cpuprofile="$output_file" ./internal/performance/ > "$bench_output" 2>&1
            ;;
        "mem"|"memory")
            echo "Running memory profiling on benchmarks..."
            go test $BENCHMARK_ARGS -memprofile="$output_file" ./internal/performance/ > "$bench_output" 2>&1
            ;;
        "block")
            echo "Running blocking profiling on benchmarks..."
            go test $BENCHMARK_ARGS -blockprofile="$output_file" ./internal/performance/ > "$bench_output" 2>&1
            ;;
        "mutex")
            echo "Running mutex profiling on benchmarks..."
            go test $BENCHMARK_ARGS -mutexprofile="$output_file" ./internal/performance/ > "$bench_output" 2>&1
            ;;
        *)
            echo -e "${RED}Unknown profile type: $profile_type${NC}"
            return 1
            ;;
    esac

    if [ -s "$output_file" ]; then
        echo -e "${GREEN}✓ Benchmark profile saved: $output_file${NC}"
        echo -e "${GREEN}✓ Benchmark results saved: $bench_output${NC}"
        analyze_profile "$output_file" "$profile_type"
    else
        echo -e "${RED}✗ Failed to generate benchmark profile${NC}"
        return 1
    fi
}

# Function to analyze profile
analyze_profile() {
    local profile_file="$1"
    local profile_type="$2"

    echo -e "${YELLOW}Analyzing $profile_type profile...${NC}"

    local base_name=$(basename "$profile_file" .prof)
    local text_output="$OUTPUT_DIR/${base_name}_analysis.txt"
    local top_output="$OUTPUT_DIR/${base_name}_top.txt"

    # Generate text analysis
    echo "Generating text analysis..."
    go tool pprof -text "$profile_file" > "$text_output" 2>&1

    # Generate top functions/allocations
    echo "Generating top report..."
    go tool pprof -top "$profile_file" > "$top_output" 2>&1

    # Generate visual outputs if graphviz is available
    if [ "$GRAPHVIZ_AVAILABLE" = true ]; then
        echo "Generating visual graphs..."

        # Generate SVG graph
        local svg_output="$OUTPUT_DIR/${base_name}_graph.svg"
        go tool pprof -svg "$profile_file" > "$svg_output" 2>&1

        # Generate flame graph (if supported)
        local flame_output="$OUTPUT_DIR/${base_name}_flame.svg"
        go tool pprof -http=:0 -svg "$profile_file" > "$flame_output" 2>&1 || true

        echo -e "${GREEN}✓ Visual graphs generated${NC}"
    fi

    # Show top results
    echo ""
    echo -e "${BLUE}Top 10 Results:${NC}"
    head -15 "$top_output" || echo "No top results available"

    echo -e "${GREEN}✓ Analysis complete${NC}"
    echo "Files generated:"
    echo "  - Text analysis: $text_output"
    echo "  - Top report: $top_output"
    if [ "$GRAPHVIZ_AVAILABLE" = true ]; then
        echo "  - SVG graph: $OUTPUT_DIR/${base_name}_graph.svg"
    fi
}

# Function to run memory analysis
analyze_memory() {
    echo -e "${YELLOW}Running comprehensive memory analysis...${NC}"

    # Run multiple memory-related benchmarks
    echo "Running memory benchmarks..."
    go test -bench=BenchmarkMemory -benchmem -memprofile="$OUTPUT_DIR/memory_detailed.prof" ./internal/performance/ > "$OUTPUT_DIR/memory_bench.txt" 2>&1

    # Analyze heap profile
    if [ -s "$OUTPUT_DIR/memory_detailed.prof" ]; then
        analyze_profile "$OUTPUT_DIR/memory_detailed.prof" "memory"

        # Generate additional memory reports
        echo "Generating detailed memory reports..."

        # Allocation analysis
        go tool pprof -alloc_space "$OUTPUT_DIR/memory_detailed.prof" > "$OUTPUT_DIR/memory_alloc_space.txt" 2>&1
        go tool pprof -alloc_objects "$OUTPUT_DIR/memory_detailed.prof" > "$OUTPUT_DIR/memory_alloc_objects.txt" 2>&1

        # In-use analysis
        go tool pprof -inuse_space "$OUTPUT_DIR/memory_detailed.prof" > "$OUTPUT_DIR/memory_inuse_space.txt" 2>&1
        go tool pprof -inuse_objects "$OUTPUT_DIR/memory_detailed.prof" > "$OUTPUT_DIR/memory_inuse_objects.txt" 2>&1

        echo -e "${GREEN}✓ Memory analysis complete${NC}"
    fi
}

# Function to run CPU analysis
analyze_cpu() {
    echo -e "${YELLOW}Running comprehensive CPU analysis...${NC}"

    # Run CPU-intensive benchmarks
    echo "Running CPU benchmarks..."
    go test -bench=BenchmarkJobsCollector -benchtime=10s -cpuprofile="$OUTPUT_DIR/cpu_detailed.prof" ./internal/performance/ > "$OUTPUT_DIR/cpu_bench.txt" 2>&1

    if [ -s "$OUTPUT_DIR/cpu_detailed.prof" ]; then
        analyze_profile "$OUTPUT_DIR/cpu_detailed.prof" "cpu"

        # Generate additional CPU reports
        echo "Generating detailed CPU reports..."

        # Call graph analysis
        go tool pprof -call_tree "$OUTPUT_DIR/cpu_detailed.prof" > "$OUTPUT_DIR/cpu_call_tree.txt" 2>&1 || true

        # Function list
        go tool pprof -list=".*" "$OUTPUT_DIR/cpu_detailed.prof" > "$OUTPUT_DIR/cpu_function_list.txt" 2>&1 || true

        echo -e "${GREEN}✓ CPU analysis complete${NC}"
    fi
}

# Function to run goroutine analysis
analyze_goroutines() {
    echo -e "${YELLOW}Running goroutine analysis...${NC}"

    if [ -n "$EXPORTER_PID" ]; then
        # Profile running process
        local goroutine_profile="$OUTPUT_DIR/goroutines.prof"
        curl -s "http://localhost:6060/debug/pprof/goroutine" > "$goroutine_profile"

        if [ -s "$goroutine_profile" ]; then
            echo "Analyzing goroutines..."
            go tool pprof -text "$goroutine_profile" > "$OUTPUT_DIR/goroutine_analysis.txt" 2>&1
            go tool pprof -traces "$goroutine_profile" > "$OUTPUT_DIR/goroutine_traces.txt" 2>&1

            echo -e "${GREEN}✓ Goroutine analysis complete${NC}"
        fi
    else
        echo -e "${YELLOW}No PID provided, skipping runtime goroutine analysis${NC}"
    fi
}

# Function to generate comprehensive report
generate_performance_report() {
    echo -e "${YELLOW}Generating performance report...${NC}"

    local report_file="$OUTPUT_DIR/performance_report.md"

    {
        echo "# SLURM Exporter Performance Analysis Report"
        echo ""
        echo "**Generated:** $(date)"
        echo "**Profile Type:** $PROFILE_TYPE"
        echo "**Duration:** $PROFILE_DURATION"
        echo ""

        echo "## System Information"
        echo "\`\`\`"
        echo "OS: $(uname -a)"
        echo "Go Version: $(go version)"
        echo "CPU: $(lscpu | grep "Model name" | cut -d: -f2 | xargs)"
        echo "Memory: $(free -h | grep Mem | awk '{print $2}')"
        echo "\`\`\`"
        echo ""

        echo "## Benchmark Results"
        if [ -f "$OUTPUT_DIR/benchmark_results.txt" ]; then
            echo "\`\`\`"
            cat "$OUTPUT_DIR/benchmark_results.txt"
            echo "\`\`\`"
        else
            echo "No benchmark results available"
        fi
        echo ""

        echo "## Profile Analysis"

        # Include top results from different profiles
        for analysis_file in "$OUTPUT_DIR"/*_top.txt; do
            if [ -f "$analysis_file" ]; then
                local profile_name=$(basename "$analysis_file" _top.txt)
                echo ""
                echo "### $profile_name Profile"
                echo "\`\`\`"
                head -20 "$analysis_file"
                echo "\`\`\`"
            fi
        done

        echo ""
        echo "## Memory Analysis"
        if [ -f "$OUTPUT_DIR/memory_bench.txt" ]; then
            echo "\`\`\`"
            grep -E "(Benchmark|PASS|FAIL)" "$OUTPUT_DIR/memory_bench.txt" | head -20
            echo "\`\`\`"
        fi

        echo ""
        echo "## Files Generated"
        echo ""
        for file in "$OUTPUT_DIR"/*; do
            if [ -f "$file" ]; then
                local filename=$(basename "$file")
                local filesize=$(wc -c < "$file")
                echo "- $filename (${filesize} bytes)"
            fi
        done

    } > "$report_file"

    echo -e "${GREEN}✓ Performance report generated: $report_file${NC}"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -t, --type TYPE       Profile type: cpu, memory, goroutine, block, mutex (default: cpu)"
    echo "  -d, --duration TIME   Profile duration (default: 30s)"
    echo "  -p, --pid PID         PID of running exporter process"
    echo "  -o, --output DIR      Output directory (default: profile-results)"
    echo "  -b, --bench ARGS      Benchmark arguments (default: -bench=. -benchmem)"
    echo "  -a, --all             Run all profile types"
    echo "  -h, --help            Show this help"
    echo ""
    echo "Examples:"
    echo "  $0 -t cpu -d 60s                    # CPU profile for 60 seconds"
    echo "  $0 -t memory -p 1234                # Memory profile of process 1234"
    echo "  $0 -a                               # Run all profile types"
    echo "  $0 -t cpu -b '-bench=BenchmarkJobs' # Profile specific benchmark"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--type)
            PROFILE_TYPE="$2"
            shift 2
            ;;
        -d|--duration)
            PROFILE_DURATION="$2"
            shift 2
            ;;
        -p|--pid)
            EXPORTER_PID="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -b|--bench)
            BENCHMARK_ARGS="$2"
            shift 2
            ;;
        -a|--all)
            PROFILE_ALL=true
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Main execution
main() {
    if [ "$PROFILE_ALL" = true ]; then
        echo -e "${YELLOW}Running all profile types...${NC}"

        for profile_type in cpu memory block mutex; do
            echo ""
            echo -e "${BLUE}=== Running $profile_type profiling ===${NC}"

            if [ -n "$EXPORTER_PID" ]; then
                profile_running_process "$EXPORTER_PID" "$profile_type" "$PROFILE_DURATION"
            else
                profile_benchmarks "$profile_type"
            fi
        done

        # Run specialized analyses
        analyze_memory
        analyze_cpu
        analyze_goroutines

    else
        case "$PROFILE_TYPE" in
            "cpu"|"profile")
                if [ -n "$EXPORTER_PID" ]; then
                    profile_running_process "$EXPORTER_PID" "profile" "$PROFILE_DURATION"
                else
                    profile_benchmarks "cpu"
                fi
                analyze_cpu
                ;;
            "memory"|"mem"|"heap")
                if [ -n "$EXPORTER_PID" ]; then
                    profile_running_process "$EXPORTER_PID" "heap" "$PROFILE_DURATION"
                else
                    profile_benchmarks "memory"
                fi
                analyze_memory
                ;;
            "goroutine"|"goroutines")
                analyze_goroutines
                ;;
            "block"|"blocking")
                if [ -n "$EXPORTER_PID" ]; then
                    profile_running_process "$EXPORTER_PID" "block" "$PROFILE_DURATION"
                else
                    profile_benchmarks "block"
                fi
                ;;
            "mutex")
                if [ -n "$EXPORTER_PID" ]; then
                    profile_running_process "$EXPORTER_PID" "mutex" "$PROFILE_DURATION"
                else
                    profile_benchmarks "mutex"
                fi
                ;;
            *)
                echo -e "${RED}Unknown profile type: $PROFILE_TYPE${NC}"
                echo "Supported types: cpu, memory, goroutine, block, mutex"
                exit 1
                ;;
        esac
    fi

    # Generate final report
    generate_performance_report

    echo ""
    echo -e "${GREEN}Profiling completed!${NC}"
    echo -e "Results available in: ${BLUE}$OUTPUT_DIR/${NC}"

    if [ "$GRAPHVIZ_AVAILABLE" = true ]; then
        echo -e "Visual profiles: ${BLUE}$OUTPUT_DIR/*.svg${NC}"
    fi
}

# Run main function
main