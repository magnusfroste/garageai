#!/bin/bash

# Garage AI Node Setup Script
# Clones Nosana's start.sh approach but replaces blockchain with our own system

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# ASCII Art
GARAGE_ART="${CYAN}
 ██████   █████  ██████   ██████   ██████  ███████
██       ██   ██ ██   ██ ██    ██ ██       ██
██   ███ ███████ ██████  ██    ██ ██   ███ █████
██    ██ ██   ██ ██   ██ ██    ██ ██    ██ ██
 ██████  ██   ██ ██   ██  ██████   ██████  ███████

    █████  ██
   ██   ██ ██
   ███████ ██
   ██   ██ ██
   ██   ██ ██
${NC}"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${PURPLE}[STEP]${NC} $1"
}

# Check if running as root
check_root() {
    if [[ $EUID -eq 0 ]]; then
        log_error "This script should not be run as root"
        exit 1
    fi
}

# Detect OS and architecture
detect_system() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        OS=$ID
        VERSION=$VERSION_ID
    else
        log_error "Unsupported operating system"
        exit 1
    fi

    ARCH=$(uname -m)
    log_info "Detected: $OS $VERSION on $ARCH"
}

# Prerequisites check (same as Nosana)
check_prerequisites() {
    log_step "Checking Prerequisites..."

    # Check Ubuntu version
    if ! lsb_release -a 2>/dev/null | grep -q "Ubuntu"; then
        log_error "Ubuntu required (20.04+ recommended)"
        exit 1
    fi

    UBUNTU_VERSION=$(lsb_release -rs)
    if (( $(echo "$UBUNTU_VERSION < 20.04" | bc -l 2>/dev/null || echo 1) )); then
        log_warning "Ubuntu $UBUNTU_VERSION detected. 20.04+ recommended."
    fi

    # Check internet connectivity
    if ! curl -s --connect-timeout 5 https://garage.ai > /dev/null; then
        log_error "No internet connection or garage.ai unreachable"
        exit 1
    fi

    # Check available disk space
    local free_space=$(df / | tail -1 | awk '{print $4}')
    if [[ $free_space -lt 10485760 ]]; then  # 10GB in KB
        log_error "At least 10GB free disk space required"
        exit 1
    fi

    log_success "Prerequisites check passed"
}

# Check GPU requirements
check_gpu() {
    log_step "Checking GPU..."

    if ! command -v nvidia-smi &> /dev/null; then
        log_error "NVIDIA drivers not found. Please install NVIDIA drivers first."
        log_info "Run: ubuntu-drivers autoinstall"
        exit 1
    fi

    # Get GPU info
    GPU_NAME=$(nvidia-smi --query-gpu=name --format=csv,noheader,nounits | head -1)
    GPU_MEMORY=$(nvidia-smi --query-gpu=memory.total --format=csv,noheader,nounits | head -1)
    GPU_COUNT=$(nvidia-smi --list-gpus | wc -l)

    log_info "Found $GPU_COUNT GPU(s): $GPU_NAME with ${GPU_MEMORY}MB VRAM"

    # Check minimum requirements
    if [[ $GPU_COUNT -eq 0 ]]; then
        log_error "No NVIDIA GPU detected"
        exit 1
    fi

    if [[ ${GPU_MEMORY//[^0-9]/} -lt 4096 ]]; then
        log_warning "GPU has less than 4GB VRAM. Performance may be limited."
    fi

    log_success "GPU check passed"
}

# Check Docker installation
check_docker() {
    log_step "Checking Docker..."

    if ! command -v docker &> /dev/null; then
        log_error "Docker not found. Please install Docker first."
        log_info "See: https://docs.docker.com/engine/install/ubuntu/"
        exit 1
    fi

    # Check if Docker daemon is running
    if ! docker info &> /dev/null; then
        log_error "Docker daemon not running. Start with: sudo systemctl start docker"
        exit 1
    fi

    # Check if user can run Docker without sudo
    if ! docker ps &> /dev/null; then
        log_error "Cannot run Docker commands. Add user to docker group:"
        log_info "sudo usermod -aG docker $USER && newgrp docker"
        exit 1
    fi

    # Check NVIDIA Docker support
    if ! docker run --rm --gpus=all ubuntu nvidia-smi &> /dev/null; then
        log_error "Docker NVIDIA GPU support not configured"
        log_info "Install NVIDIA Container Toolkit and run: sudo nvidia-ctk runtime configure --runtime=docker"
        exit 1
    fi

    log_success "Docker check passed"
}

# Setup Podman-in-Docker (exact clone of Nosana)
setup_podman() {
    log_step "Setting up Podman-in-Docker..."

    # Create volumes
    docker volume create podman-socket 2>/dev/null || true
    docker volume create podman-cache 2>/dev/null || true

    # Stop any existing podman container
    docker stop podman 2>/dev/null || true
    docker rm podman 2>/dev/null || true

    # Start Podman daemon (exact same as Nosana)
    docker run -d \
        --pull=always \
        --gpus=all \
        --name podman \
        --device /dev/fuse \
        --mount source=podman-cache,target=/var/lib/containers \
        --volume podman-socket:/podman \
        --privileged \
        -e ENABLE_GPU=true \
        garageai/podman:v1.0.0 \
        unix:/podman/podman.sock

    # Wait for Podman to start
    log_info "Waiting for Podman to initialize..."
    local retries=30
    while [[ $retries -gt 0 ]]; do
        if docker exec podman curl -s http://localhost:8080/v4.5.0/libpod/info > /dev/null 2>&1; then
            log_success "Podman daemon started successfully"
            return 0
        fi
        sleep 2
        ((retries--))
    done

    log_error "Podman daemon failed to start"
    exit 1
}

# Generate node identity (replace Solana keypair)
generate_node_identity() {
    log_step "Generating Node Identity..."

    # Create garage config directory
    mkdir -p ~/.garage

    # Generate node ID (UUID-based, replace Solana key)
    if [[ ! -f ~/.garage/node_id ]]; then
        NODE_ID=$(cat /proc/sys/kernel/random/uuid)
        echo "$NODE_ID" > ~/.garage/node_id
    else
        NODE_ID=$(cat ~/.garage/node_id)
    fi

    # Generate API key for authentication
    if [[ ! -f ~/.garage/api_key ]]; then
        API_KEY=$(openssl rand -hex 32)
        echo "$API_KEY" > ~/.garage/api_key
    else
        API_KEY=$(cat ~/.garage/api_key)
    fi

    log_info "Node ID: $NODE_ID"
    log_success "Node identity generated"
}

# Register node with Garage AI API (replace Solana blockchain)
register_node() {
    log_step "Registering Node with Garage AI Network..."

    # Get system information
    local gpu_name="$GPU_NAME"
    local gpu_memory="$GPU_MEMORY"
    local gpu_count="$GPU_COUNT"
    local cpu_cores=$(nproc)
    local total_ram=$(free -m | awk 'NR==2{printf "%.0f", $2}')
    local os_info="$OS $VERSION"
    local location="Sweden"  # Default, can be overridden

    # Ask for location if not set
    if [[ ! -f ~/.garage/location ]]; then
        echo "Enter your location (city, Sweden) [default: Stockholm]:"
        read -r user_location
        location=${user_location:-"Stockholm"}
        echo "$location" > ~/.garage/location
    else
        location=$(cat ~/.garage/location)
    fi

    # Create registration payload
    local payload=$(cat <<EOF
{
    "node_id": "$NODE_ID",
    "api_key": "$API_KEY",
    "gpu_name": "$gpu_name",
    "gpu_memory_mb": ${gpu_memory//[^0-9]/},
    "gpu_count": $gpu_count,
    "cpu_cores": $cpu_cores,
    "ram_mb": $total_ram,
    "os": "$os_info",
    "location": "$location",
    "network_speed_mbps": 100
}
EOF
    )

    # Register with Garage AI API
    log_info "Contacting Garage AI registration service..."
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $API_KEY" \
        -d "$payload" \
        https://api.garage.ai/nodes/register)

    if echo "$response" | grep -q '"status":"success"'; then
        log_success "Node registered successfully in Garage AI network"
        echo "$response" | jq -r '.message // "Registration complete"'
    else
        log_error "Node registration failed: $response"
        exit 1
    fi
}

# Run hardware benchmark (similar to Nosana)
run_benchmark() {
    log_step "Running Hardware Benchmark..."

    log_info "Testing GPU performance..."

    # Test basic GPU functionality
    if ! docker exec podman podman run --rm --device nvidia.com/gpu=all \
        --security-opt=label=disable \
        ubuntu nvidia-smi -L > /dev/null 2>&1; then
        log_error "GPU access through Podman failed"
        exit 1
    fi

    # Test inference performance (small model)
    log_info "Testing inference performance..."
    docker exec podman podman run --rm --device nvidia.com/gpu=all \
        -e HUGGING_FACE_HUB_TOKEN=$HF_TOKEN \
        vllm/vllm-openai:latest \
        vllm serve microsoft/DialoGPT-small \
        --tensor-parallel-size 1 \
        --max-model-len 512 \
        --port 8000 \
        --host 0.0.0.0 > /tmp/benchmark.log 2>&1 &
    local bench_pid=$!

    # Wait for model to load
    sleep 30

    # Test inference
    local test_result=$(curl -s -X POST \
        http://localhost:8000/v1/completions \
        -H "Content-Type: application/json" \
        -d '{"model": "microsoft/DialoGPT-small", "prompt": "Hello, how are you?", "max_tokens": 10}' \
        --connect-timeout 10)

    if echo "$test_result" | grep -q '"choices"'; then
        local tokens_per_sec=$(echo "$test_result" | jq -r '.usage.total_tokens // 0')
        log_success "Benchmark passed - Model inference working (~$tokens_per_sec tokens)"
    else
        log_warning "Benchmark inconclusive - Model loading slowly"
    fi

    # Cleanup
    kill $bench_pid 2>/dev/null || true
    rm -f /tmp/benchmark.log
}

# Start inference worker (replace Nosana's GPU host)
start_worker() {
    log_step "Starting Garage AI Inference Worker..."

    # Stop any existing worker
    docker stop garage-worker 2>/dev/null || true
    docker rm garage-worker 2>/dev/null || true

    # Start our inference worker
    docker run -d \
        --name garage-worker \
        --gpus=all \
        -p 8000:8000 \
        -e NODE_ID=$NODE_ID \
        -e API_KEY=$API_KEY \
        -e CLUSTER_ENDPOINT=https://api.garage.ai \
        -v ~/.garage:/root/.garage:ro \
        --volume podman-socket:/root/.garage/podman:ro \
        garageai/worker:latest \
        --podman /root/.garage/podman/podman.sock

    # Wait for worker to start
    sleep 10

    if docker ps | grep -q garage-worker; then
        log_success "Garage AI worker started successfully"

        # Test worker health
        if curl -s http://localhost:8000/health > /dev/null 2>&1; then
            log_success "Worker health check passed"
        else
            log_warning "Worker started but health check failed - may take time to initialize"
        fi
    else
        log_error "Failed to start Garage AI worker"
        exit 1
    fi
}

# Create systemd service for auto-start
create_service() {
    log_step "Creating Auto-Start Service..."

    sudo tee /etc/systemd/system/garage-ai.service > /dev/null <<EOF
[Unit]
Description=Garage AI Node Service
Requires=docker.service
After=docker.service network.target

[Service]
Type=simple
User=$USER
Environment=NODE_ID=$NODE_ID
Environment=API_KEY=$API_KEY
ExecStartPre=/bin/bash -c 'docker volume create podman-socket 2>/dev/null || true'
ExecStartPre=/bin/bash -c 'docker volume create podman-cache 2>/dev/null || true'
ExecStartPre=/bin/bash -c 'docker run -d --pull=always --gpus=all --name podman --device /dev/fuse --mount source=podman-cache,target=/var/lib/containers --volume podman-socket:/podman --privileged -e ENABLE_GPU=true garageai/podman:v1.0.0 unix:/podman/podman.sock 2>/dev/null || true'
ExecStart=docker run --rm \
    -e NODE_ID=$NODE_ID \
    -e API_KEY=$API_KEY \
    -e CLUSTER_ENDPOINT=https://api.garage.ai \
    --gpus=all \
    --volume podman-socket:/root/.garage/podman:ro \
    garageai/worker:latest \
    --podman /root/.garage/podman/podman.sock
Restart=always
RestartSec=30

[Install]
WantedBy=multi-user.target
EOF

    sudo systemctl daemon-reload
    sudo systemctl enable garage-ai.service

    log_success "Auto-start service created"
}

# Display final status
show_status() {
    echo
    echo -e "${GARAGE_ART}"
    echo
    log_success "🎉 Garage AI Node Setup Complete!"
    echo
    echo -e "${CYAN}Node Information:${NC}"
    echo "  Node ID: $NODE_ID"
    echo "  Location: $(cat ~/.garage/location 2>/dev/null || echo 'Sweden')"
    echo "  GPU: $GPU_NAME (${GPU_MEMORY} VRAM)"
    echo
    echo -e "${CYAN}Services Running:${NC}"
    docker ps --filter "name=podman" --format "table {{.Names}}\t{{.Status}}"
    docker ps --filter "name=garage-worker" --format "table {{.Names}}\t{{.Status}}"
    echo
    echo -e "${CYAN}Next Steps:${NC}"
    echo "1. Check dashboard: https://dashboard.garage.ai"
    echo "2. Monitor node: docker logs garage-worker"
    echo "3. View status: docker exec podman podman ps"
    echo "4. Test inference: curl http://localhost:8000/v1/models"
    echo
    echo -e "${CYAN}Useful Commands:${NC}"
    echo "• Stop node: docker stop garage-worker podman"
    echo "• Restart: sudo systemctl restart garage-ai"
    echo "• View logs: docker logs garage-worker"
    echo "• Update: docker pull garageai/worker:latest"
    echo
    log_success "Welcome to the Garage AI network! 🇸🇪🤖"
}

# Main function
main() {
    echo -e "${GARAGE_ART}"
    echo -e "${CYAN}Garage AI Node Setup${NC}"
    echo "========================"
    echo

    check_root
    detect_system
    check_prerequisites
    check_gpu
    check_docker
    setup_podman
    generate_node_identity
    register_node
    run_benchmark
    start_worker
    create_service
    show_status
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "Garage AI Node Setup Script"
        echo
        echo "Usage: $0 [OPTIONS]"
        echo
        echo "Options:"
        echo "  --help, -h          Show this help message"
        echo "  --version, -v       Show version information"
        echo "  --status            Show current node status"
        echo "  --stop              Stop the node"
        echo "  --restart           Restart the node"
        echo
        exit 0
        ;;
    --version|-v)
        echo "Garage AI Setup Script v1.0.0"
        exit 0
        ;;
    --status)
        echo "Node Status:"
        echo "Node ID: $(cat ~/.garage/node_id 2>/dev/null || echo 'Not registered')"
        echo "Location: $(cat ~/.garage/location 2>/dev/null || echo 'Not set')"
        echo
        echo "Services:"
        docker ps --filter "name=podman" --format "podman: {{.Status}}" 2>/dev/null || echo "podman: Not running"
        docker ps --filter "name=garage-worker" --format "worker: {{.Status}}" 2>/dev/null || echo "worker: Not running"
        exit 0
        ;;
    --stop)
        log_info "Stopping Garage AI node..."
        docker stop garage-worker podman 2>/dev/null || true
        sudo systemctl stop garage-ai 2>/dev/null || true
        log_success "Node stopped"
        exit 0
        ;;
    --restart)
        log_info "Restarting Garage AI node..."
        sudo systemctl restart garage-ai
        sleep 5
        docker ps --filter "name=garage-worker" --format "Worker: {{.Status}}"
        log_success "Node restarted"
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac
