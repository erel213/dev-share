#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TEST_TIMEOUT="${TEST_TIMEOUT:-120s}"

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to print usage
usage() {
    cat << EOF
Usage: $(basename "$0") [OPTIONS] [TEST_SUITE]

Run integration tests for the dev-share backend.
The backend server starts in-process (no Docker required).

TEST_SUITE:
    all         Run all integration tests (default)
    workspace   Run workspace integration tests only
    user        Run user integration tests only
    admin       Run admin initialization integration tests only
    template    Run template integration tests only

OPTIONS:
    -h, --help              Show this help message
    -v, --verbose           Enable verbose test output
    -t, --timeout DURATION  Test timeout (default: 120s)

EXAMPLES:
    $(basename "$0")                    # Run all tests
    $(basename "$0") workspace          # Run workspace tests only
    $(basename "$0") -v user            # Run user tests with verbose output

EOF
    exit 0
}

# Parse command line arguments
VERBOSE=""
TEST_SUITE="all"

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            ;;
        -v|--verbose)
            VERBOSE="-v"
            shift
            ;;
        -t|--timeout)
            TEST_TIMEOUT="$2"
            shift 2
            ;;
        all|workspace|user|admin|template)
            TEST_SUITE="$1"
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            usage
            ;;
    esac
done

# Validate test suite
case $TEST_SUITE in
    all)
        TEST_PATTERN="./integration_tests/..."
        TEST_NAME="All integration tests"
        ;;
    workspace)
        TEST_FILES="./integration_tests/workspace_test.go ./integration_tests/setup_test.go ./integration_tests/helpers_test.go"
        TEST_NAME="Workspace integration tests"
        ;;
    user)
        TEST_FILES="./integration_tests/user_test.go ./integration_tests/setup_test.go ./integration_tests/helpers_test.go"
        TEST_NAME="User integration tests"
        ;;
    admin)
        TEST_FILES="./integration_tests/admin_init_test.go ./integration_tests/setup_test.go ./integration_tests/helpers_test.go"
        TEST_NAME="Admin initialization integration tests"
        ;;
    template)
        TEST_PATTERN="./integration_tests/template_test.go"
        TEST_FILES="./integration_tests/template_test.go ./integration_tests/setup_test.go ./integration_tests/helpers_test.go"
        TEST_NAME="Template integration tests"
        ;;
    *)
        print_error "Invalid test suite: $TEST_SUITE"
        usage
        ;;
esac

# Main execution
print_info "Starting integration tests: $TEST_NAME"
echo ""

# Navigate to project root
cd "$PROJECT_ROOT"

# Run tests
print_info "Running $TEST_NAME..."
echo ""

EXIT_CODE=0

if [ "$TEST_SUITE" = "all" ]; then
    if go test $TEST_PATTERN $VERBOSE -timeout "$TEST_TIMEOUT" -count=1; then
        echo ""
        print_success "All tests passed! ✓"
    else
        EXIT_CODE=$?
        echo ""
        print_error "Some tests failed! ✗"
    fi
else
    if go test $TEST_FILES $VERBOSE -timeout "$TEST_TIMEOUT" -count=1; then
        echo ""
        print_success "Tests passed! ✓"
    else
        EXIT_CODE=$?
        echo ""
        print_error "Tests failed! ✗"
    fi
fi

exit $EXIT_CODE
