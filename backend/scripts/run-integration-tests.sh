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
COMPOSE_FILE="$PROJECT_ROOT/docker-compose.yml"
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

TEST_SUITE:
    all         Run all integration tests (default)
    workspace   Run workspace integration tests only
    user        Run user integration tests only

OPTIONS:
    -h, --help              Show this help message
    -v, --verbose           Enable verbose test output
    -k, --keep-running      Keep containers running after tests
    -s, --skip-build        Skip rebuilding Docker images
    -t, --timeout DURATION  Test timeout (default: 120s)

EXAMPLES:
    $(basename "$0")                    # Run all tests
    $(basename "$0") workspace          # Run workspace tests only
    $(basename "$0") -v user            # Run user tests with verbose output
    $(basename "$0") -k all             # Run all tests and keep containers running

EOF
    exit 0
}

# Function to cleanup
cleanup() {
    if [ "$KEEP_RUNNING" != "true" ]; then
        print_info "Stopping test environment..."
        docker compose -f "$COMPOSE_FILE" down -v 2>/dev/null || true
        print_success "Test environment stopped"
    else
        print_warning "Containers kept running (use -k flag)"
        print_info "To stop manually: docker compose -f $COMPOSE_FILE down"
    fi
}

# Function to wait for service health
wait_for_service() {
    local service_name=$1
    local max_attempts=30
    local attempt=1

    print_info "Waiting for $service_name to be healthy..."

    while [ $attempt -le $max_attempts ]; do
        if docker compose -f "$COMPOSE_FILE" ps | grep "$service_name" | grep -q "healthy"; then
            print_success "$service_name is healthy"
            return 0
        fi

        echo -n "."
        sleep 2
        attempt=$((attempt + 1))
    done

    print_error "$service_name failed to become healthy after $max_attempts attempts"
    return 1
}

# Parse command line arguments
VERBOSE=""
KEEP_RUNNING="false"
SKIP_BUILD="false"
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
        -k|--keep-running)
            KEEP_RUNNING="true"
            shift
            ;;
        -s|--skip-build)
            SKIP_BUILD="true"
            shift
            ;;
        -t|--timeout)
            TEST_TIMEOUT="$2"
            shift 2
            ;;
        all|workspace|user)
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
        TEST_PATTERN="./integration_tests/workspace_test.go"
        TEST_FILES="./integration_tests/workspace_test.go ./integration_tests/setup_test.go ./integration_tests/helpers_test.go"
        TEST_NAME="Workspace integration tests"
        ;;
    user)
        TEST_PATTERN="./integration_tests/user_test.go"
        TEST_FILES="./integration_tests/user_test.go ./integration_tests/setup_test.go ./integration_tests/helpers_test.go"
        TEST_NAME="User integration tests"
        ;;
    *)
        print_error "Invalid test suite: $TEST_SUITE"
        usage
        ;;
esac

# Set trap to cleanup on exit
trap cleanup EXIT INT TERM

# Main execution
print_info "Starting integration tests: $TEST_NAME"
echo ""

# Check if docker compose is available
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed or not in PATH"
    exit 1
fi

# Navigate to project root
cd "$PROJECT_ROOT"

# Start test environment
print_info "Starting test environment..."
if [ "$SKIP_BUILD" = "true" ]; then
    docker compose -f "$COMPOSE_FILE" up -d
else
    docker compose -f "$COMPOSE_FILE" up -d --build
fi

# Wait for services to be healthy
wait_for_service "postgres"
wait_for_service "backend"

echo ""
print_success "Test environment is ready"
echo ""

# Run tests
print_info "Running $TEST_NAME..."
echo ""

if [ "$TEST_SUITE" = "all" ]; then
    # Run all tests using the package pattern
    if go test $TEST_PATTERN $VERBOSE -timeout "$TEST_TIMEOUT"; then
        EXIT_CODE=0
        echo ""
        print_success "All tests passed! ✓"
    else
        EXIT_CODE=$?
        echo ""
        print_error "Some tests failed! ✗"
    fi
else
    # Run specific test file with required dependencies
    if go test $TEST_FILES $VERBOSE -timeout "$TEST_TIMEOUT"; then
        EXIT_CODE=0
        echo ""
        print_success "Tests passed! ✓"
    else
        EXIT_CODE=$?
        echo ""
        print_error "Tests failed! ✗"
    fi
fi

echo ""

# Show logs on failure
if [ $EXIT_CODE -ne 0 ]; then
    print_warning "Showing backend logs:"
    echo ""
    docker compose -f "$COMPOSE_FILE" logs backend --tail=50
fi

exit $EXIT_CODE
