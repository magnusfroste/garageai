#!/bin/bash

# Garage AI LiteLLM Integration Script
# Dynamic setup - skips what's already installed
# Focus: Podman layer + LiteLLM connection for existing vLLM rigs

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

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
        log_error "Do not run as root"
        exit 1
    fi
}

# Detect system
detect_system() {
    log_step "Detecting System..."

    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        OS=$ID
        log_info "System: $OS"
    else
        log_error "Unsupported system"
        exit 1
    fi
}

# Check existing installations
check_existing() {
    log_step "Checking Existing Setup..."

    # Check Docker
    if command -v docker &> /dev/null && docker info &> /dev/null 2>&1; then
        log_success "Docker is already installed and running"
        DOCKER_INSTALLED=true
    else
        log_warning "Docker not found or not running"
        DOCKER_INSTALLED=false
    fi

    # Check Podman
    if docker ps --filter "name=podman" --format "{{.Names}}" | grep -q "^podman$"; then
        log_success "Podman container is already running"
        PODMAN_RUNNING=true
    else
        log_warning "Podman container not found"
        PODMAN_RUNNING=false
    fi

    # Check NVIDIA GPU
    if command -v nvidia-smi &> /dev/null; then
        GPU_INFO=$(nvidia-smi --query-gpu=name,memory.total --format=csv,noheader,nounits | head -1)
        GPU_NAME=$(echo "$GPU_INFO" | cut -d',' -f1 | xargs)
        GPU_MEMORY=$(echo "$GPU_INFO" | cut -d',' -f2 | xargs)
        log_success "GPU detected: $GPU_NAME with ${GPU_MEMORY}MB VRAM"
        GPU_PRESENT=true
    else
        log_warning "NVIDIA GPU not detected"
        GPU_PRESENT=false
    fi

    # Check vLLM containers
    VLLM_CONTAINERS=$(docker ps --filter "ancestor=vllm/vllm-openai" --format "{{.Names}}" | wc -l)
    if [[ $VLLM_CONTAINERS -gt 0 ]]; then
        log_success "Found $VLLM_CONTAINERS vLLM container(s) running"
        VLLM_EXISTS=true
    else
        log_warning "No vLLM containers found"
        VLLM_EXISTS=false
    fi

    # Check Python environment
    if [[ -d ~/garage-env ]] && source ~/garage-env/bin/activate && python3 -c "import litellm" &> /dev/null; then
        log_success "LiteLLM is already installed"
        LITELLM_INSTALLED=true
    else
        log_warning "LiteLLM not found"
        LITELLM_INSTALLED=false
    fi
}

# Setup Podman if needed
setup_podman() {
    if [[ "$PODMAN_RUNNING" == true ]]; then
        log_info "Podman already running, skipping setup"
        return 0
    fi

    log_step "Setting up Podman Container..."

    # Create volumes
    docker volume create podman-socket 2>/dev/null || true
    docker volume create podman-cache 2>/dev/null || true

    # Stop any existing podman container
    docker stop podman 2>/dev/null || true
    docker rm podman 2>/dev/null || true

    # Start Podman daemon
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
            PODMAN_RUNNING=true
            return 0
        fi
        sleep 2
        ((retries--))
    done

    log_error "Podman daemon failed to start"
    exit 1
}

# Setup Python environment if needed
setup_python() {
    if [[ "$LITELLM_INSTALLED" == true ]]; then
        log_info "Python environment already set up, skipping"
        return 0
    fi

    log_step "Setting up Python Environment..."

    # Create virtual environment if it doesn't exist
    if [[ ! -d ~/garage-env ]]; then
        python3 -m venv ~/garage-env
        log_success "Created Python virtual environment"
    fi

    # Activate and install packages
    source ~/garage-env/bin/activate

    log_info "Installing LiteLLM and dependencies..."
    pip install litellm fastapi uvicorn python-dotenv

    # Verify installation
    if python3 -c "import litellm; print('LiteLLM version:', litellm.__version__)" 2>/dev/null; then
        log_success "LiteLLM installed successfully"
        LITELLM_INSTALLED=true
    else
        log_error "LiteLLM installation failed"
        exit 1
    fi
}

# Create LiteLLM configuration
create_litellm_config() {
    log_step "Creating LiteLLM Configuration..."

    # Ask for rig information
    echo
    log_info "LiteLLM Configuration Setup"
    echo "============================"

    read -p "Enter Rig 1 IP address [192.168.1.100]: " RIG1_IP
    RIG1_IP=${RIG1_IP:-192.168.1.100}

    read -p "Enter Rig 1 vLLM port [8000]: " RIG1_PORT
    RIG1_PORT=${RIG1_PORT:-8000}

    read -p "Enter Rig 2 IP address [192.168.1.101]: " RIG2_IP
    RIG2_IP=${RIG2_IP:-192.168.1.101}

    read -p "Enter Rig 2 vLLM port [8001]: " RIG2_PORT
    RIG2_PORT=${RIG2_PORT:-8001}

    # Create config directory
    mkdir -p ~/litellm-config

    # Create config.yaml
    cat > ~/litellm-config/config.yaml << EOF
model_list:
  - model_name: garage-gpt4
    litellm_params:
      model: openai/gpt-4
      api_base: http://$RIG1_IP:$RIG1_PORT
      api_key: garage-key-1
  - model_name: garage-gpt4
    litellm_params:
      model: openai/gpt-4
      api_base: http://$RIG2_IP:$RIG2_PORT
      api_key: garage-key-2

general_settings:
  master_key: sk-garage-ai-master
  database_url: sqlite:///./litellm.db

# Rate limiting
router_settings:
  routing_strategy: usage-based-routing
  model_group_retry_policy: retry_different_provider
EOF

    log_success "Created LiteLLM config at ~/litellm-config/config.yaml"
}

# Test vLLM connections
test_vllm_connections() {
    log_step "Testing vLLM Connections..."

    source ~/litellm-config/config.yaml 2>/dev/null || true

    # Extract rig IPs from config
    RIG1_IP=$(grep "api_base:" ~/litellm-config/config.yaml | head -1 | sed 's/.*http:\/\/\([^:]*\).*/\1/')
    RIG1_PORT=$(grep "api_base:" ~/litellm-config/config.yaml | head -1 | sed 's/.*:\([0-9]*\).*/\1/')
    RIG2_IP=$(grep "api_base:" ~/litellm-config/config.yaml | tail -1 | sed 's/.*http:\/\/\([^:]*\).*/\1/')
    RIG2_PORT=$(grep "api_base:" ~/litellm-config/config.yaml | tail -1 | sed 's/.*:\([0-9]*\).*/\1/')

    # Test Rig 1
    if curl -s "http://$RIG1_IP:$RIG1_PORT/v1/models" > /dev/null 2>&1; then
        log_success "Rig 1 ($RIG1_IP:$RIG1_PORT) is accessible"
    else
        log_warning "Rig 1 ($RIG1_IP:$RIG1_PORT) is not accessible"
    fi

    # Test Rig 2
    if curl -s "http://$RIG2_IP:$RIG2_PORT/v1/models" > /dev/null 2>&1; then
        log_success "Rig 2 ($RIG2_IP:$RIG2_PORT) is accessible"
    else
        log_warning "Rig 2 ($RIG2_IP:$RIG2_PORT) is not accessible"
    fi
}

# Create systemd service for LiteLLM
create_litellm_service() {
    log_step "Creating LiteLLM Service..."

    sudo tee /etc/systemd/system/garage-litellm.service > /dev/null << EOF
[Unit]
Description=Garage AI LiteLLM Proxy
After=network.target

[Service]
Type=simple
User=$USER
Environment=PATH=/home/$USER/garage-env/bin
WorkingDirectory=/home/$USER
ExecStart=/home/$USER/garage-env/bin/litellm --config /home/$USER/litellm-config/config.yaml --port 8000
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

    sudo systemctl daemon-reload
    sudo systemctl enable garage-litellm.service

    log_success "LiteLLM service created"
}

# Start LiteLLM proxy
start_litellm() {
    log_step "Starting LiteLLM Proxy..."

    source ~/garage-env/bin/activate

    # Start LiteLLM in background
    nohup litellm --config ~/litellm-config/config.yaml --port 8000 > ~/litellm.log 2>&1 &

    # Wait for startup
    sleep 5

    # Test if running
    if curl -s http://localhost:8000/v1/models > /dev/null 2>&1; then
        log_success "LiteLLM proxy started successfully on port 8000"
    else
        log_error "LiteLLM proxy failed to start"
        log_info "Check logs: tail -f ~/litellm.log"
        exit 1
    fi
}

# Show final status
show_final_status() {
    echo
    echo -e "${CYAN}🎉 Garage AI LiteLLM Integration Complete!${NC}"
    echo
    echo -e "${CYAN}Active Services:${NC}"
    echo "• Podman Container: $(docker ps --filter "name=podman" --format "{{.Status}}" || echo "Not running")"
    echo "• LiteLLM Proxy: $(curl -s http://localhost:8000/health >/dev/null 2>&1 && echo "Running" || echo "Not responding")"
    echo
    echo -e "${CYAN}Your Rigs:${NC}"
    echo "• Rig 1: http://$RIG1_IP:$RIG1_PORT"
    echo "• Rig 2: http://$RIG2_IP:$RIG2_PORT"
    echo
    echo -e "${CYAN}API Endpoint:${NC}"
    echo "• LiteLLM: http://localhost:8000"
    echo
    echo -e "${CYAN}Test Commands:${NC}"
    echo "• curl http://localhost:8000/v1/models"
    echo "• curl -X POST http://localhost:8000/v1/chat/completions \\"
    echo "    -H 'Authorization: Bearer sk-garage-ai-master' \\"
    echo "    -d '{\"model\": \"garage-gpt4\", \"messages\": [{\"role\": \"user\", \"content\": \"Hello!\"}]}'"
    echo
    log_success "Ready to serve distributed AI! 🚀"
}

# Main function
main() {
    echo -e "${CYAN}"
    echo " ██████   █████  ██████   ██████   ██████  ███████"
    echo "██       ██   ██ ██   ██ ██    ██ ██       ██"
    echo "██   ███ ███████ ██████  ██    ██ ██   ███ █████"
    echo "██    ██ ██   ██ ██   ██ ██    ██ ██    ██ ██"
    echo " ██████  ██   ██ ██   ██  ██████   ██████  ███████"
    echo "${NC}"
    echo -e "${CYAN}Garage AI LiteLLM Integration${NC}"
    echo "==============================="
    echo

    check_root
    detect_system
    check_existing
    setup_podman
    setup_python
    create_litellm_config
    test_vllm_connections
    create_litellm_service
    start_litellm
    show_final_status
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "Garage AI LiteLLM Integration Script"
        echo
        echo "This script dynamically sets up LiteLLM integration with existing vLLM rigs."
        echo "It skips components that are already installed."
        echo
        echo "Usage: $0 [OPTIONS]"
        echo
        echo "Options:"
        echo "  --help, -h          Show this help message"
        echo "  --test              Test current setup"
        echo "  --status            Show current status"
        echo "  --restart           Restart LiteLLM service"
        echo
        exit 0
        ;;
    --test)
        check_root
        detect_system
        check_existing
        test_vllm_connections
        exit $?
        ;;
    --status)
        check_root
        detect_system
        check_existing
        echo
        echo "LiteLLM Status:"
        if pgrep -f litellm > /dev/null; then
            echo "• LiteLLM: Running"
        else
            echo "• LiteLLM: Not running"
        fi
