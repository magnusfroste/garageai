#!/bin/bash

# Garage AI Complete Installation Script
# One-command setup for junior Linux system engineers
# Installs all prerequisites: GPU drivers, Docker, NVIDIA toolkit, Python environment

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
        log_error "Do not run as root. Use sudo when prompted."
        exit 1
    fi
}

# Detect system
detect_system() {
    log_step "Detecting System..."

    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        OS=$ID
        VERSION=$VERSION_ID
        log_info "Detected: $OS $VERSION"
    else
        log_error "Unsupported operating system"
        exit 1
    fi

    ARCH=$(uname -m)
    log_info "Architecture: $ARCH"

    # Check if Ubuntu/Debian-based
    if [[ "$OS" != "ubuntu" && "$OS" != "debian" && "$OS" != "pop" ]]; then
        log_error "This script requires Ubuntu/Debian-based system"
        exit 1
    fi

    # Check Ubuntu version
    UBUNTU_VERSION=$(lsb_release -rs 2>/dev/null || echo "0")
    if (( $(echo "$UBUNTU_VERSION < 20.04" | bc -l 2>/dev/null || echo 1) )); then
        log_warning "Ubuntu $UBUNTU_VERSION detected. Ubuntu 20.04+ recommended."
    fi
}

# Install NVIDIA GPU drivers
install_nvidia_drivers() {
    log_step "Installing NVIDIA GPU Drivers..."

    # Update package lists
    log_info "Updating package lists..."
    sudo apt-get update

    # Install Ubuntu drivers
    log_info "Installing NVIDIA drivers via ubuntu-drivers..."
    sudo ubuntu-drivers autoinstall

    # Alternative: Install specific driver version
    # sudo apt-get install -y nvidia-driver-470

    log_success "NVIDIA drivers installed"

    # Prompt for reboot
    log_warning "GPU drivers installed. A system reboot is required."
    echo "After reboot, run this script again to continue."
    read -p "Press Enter to reboot now, or Ctrl+C to reboot later: "
    sudo reboot
}

# Verify GPU installation
verify_gpu() {
    log_step "Verifying GPU Installation..."

    if ! command -v nvidia-smi &> /dev/null; then
        log_error "nvidia-smi not found. GPU drivers may not be installed correctly."
        return 1
    fi

    # Get GPU info
    GPU_INFO=$(nvidia-smi --query-gpu=name,memory.total --format=csv,noheader,nounits | head -1)
    GPU_NAME=$(echo "$GPU_INFO" | cut -d',' -f1 | xargs)
    GPU_MEMORY=$(echo "$GPU_INFO" | cut -d',' -f2 | xargs)

    log_success "GPU detected: $GPU_NAME with ${GPU_MEMORY}MB VRAM"

    # Test basic GPU functionality
    if nvidia-smi &> /dev/null; then
        log_success "GPU drivers working correctly"
    else
        log_error "GPU drivers not working"
        return 1
    fi
}

# Install Docker
install_docker() {
    log_step "Installing Docker..."

    # Remove any old Docker installations
    sudo apt-get remove -y docker docker-engine docker.io containerd runc 2>/dev/null || true

    # Update package index
    sudo apt-get update

    # Install prerequisites
    sudo apt-get install -y ca-certificates curl gnupg lsb-release

    # Add Docker's official GPG key
    sudo mkdir -p /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg

    # Set up the repository
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

    # Install Docker Engine
    sudo apt-get update
    sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

    # Start and enable Docker
    sudo systemctl start docker
    sudo systemctl enable docker

    # Add user to docker group (requires logout/login to take effect)
    sudo usermod -aG docker $USER

    log_success "Docker installed"

    # Test Docker installation
    if sudo docker run --rm hello-world &> /dev/null; then
        log_success "Docker working correctly"
    else
        log_error "Docker test failed"
        return 1
    fi

    log_warning "You may need to log out and back in for Docker group changes to take effect"
}

# Install NVIDIA Container Toolkit
install_nvidia_toolkit() {
    log_step "Installing NVIDIA Container Toolkit..."

    # Add NVIDIA GPG key
    curl -fsSL https://nvidia.github.io/libnvidia-container/gpgkey | sudo gpg --dearmor -o /usr/share/keyrings/nvidia-container-toolkit-keyring.gpg

    # Add NVIDIA repository
    curl -s -L https://nvidia.github.io/libnvidia-container/stable/deb/nvidia-container-toolkit.list | sed 's#deb https://#deb [signed-by=/usr/share/keyrings/nvidia-container-toolkit-keyring.gpg] https://#g' | sudo tee /etc/apt/sources.list.d/nvidia-container-toolkit.list

    # Update and install
    sudo apt-get update
    sudo apt-get install -y nvidia-container-toolkit

    # Configure Docker to use NVIDIA runtime
    sudo nvidia-ctk runtime configure --runtime=docker
    sudo systemctl restart docker

    log_success "NVIDIA Container Toolkit installed"

    # Test GPU access in containers
    if sudo docker run --rm --gpus=all ubuntu nvidia-smi &> /dev/null; then
        log_success "GPU access in Docker containers working"
    else
        log_error "GPU access in Docker containers failed"
        return 1
    fi
}

# Install Python and development tools
install_python_dev() {
    log_step "Installing Python Development Environment..."

    # Install Python and pip
    sudo apt-get install -y python3 python3-pip python3-venv python3-dev

    # Install development tools
    sudo apt-get install -y build-essential git curl wget

    # Verify Python
    if command -v python3 &> /dev/null; then
        PYTHON_VERSION=$(python3 --version)
        log_success "Python installed: $PYTHON_VERSION"
    else
        log_error "Python installation failed"
        return 1
    fi

    # Create Garage AI virtual environment
    if [[ ! -d ~/garage-env ]]; then
        python3 -m venv ~/garage-env
        log_success "Python virtual environment created at ~/garage-env"
    else
        log_info "Python virtual environment already exists"
    fi
}

# Install basic AI/ML libraries
install_ai_libraries() {
    log_step "Installing AI/ML Libraries..."

    # Activate virtual environment
    source ~/garage-env/bin/activate

    # Install PyTorch with CUDA support
    log_info "Installing PyTorch with CUDA support..."
    pip install torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cu118

    # Install other AI libraries
    pip install transformers accelerate vllm fastapi uvicorn

    # Verify installations
    if python3 -c "import torch; print(f'PyTorch CUDA: {torch.cuda.is_available()}')"; then
        log_success "AI libraries installed successfully"
    else
        log_error "AI library installation failed"
        return 1
    fi
}

# Setup Git repository
setup_repository() {
    log_step "Setting up Garage AI Repository..."

    # Check if already in repository
    if [[ -d .git ]]; then
        log_info "Already in Git repository"
        return 0
    fi

    # Clone repository
    if [[ ! -d ~/garage-ai ]]; then
        git clone https://github.com/garage-ai/platform.git ~/garage-ai
        cd ~/garage-ai
        log_success "Repository cloned to ~/garage-ai"
    else
        log_info "Repository already exists at ~/garage-ai"
        cd ~/garage-ai
    fi
}

# Test complete installation
test_installation() {
    log_step "Testing Complete Installation..."

    # Test GPU
    if ! verify_gpu; then
        log_error "GPU test failed"
        return 1
    fi

    # Test Docker GPU
    if ! sudo docker run --rm --gpus=all ubuntu nvidia-smi &> /dev/null; then
        log_error "Docker GPU test failed"
        return 1
    fi

    # Test Python environment
    source ~/garage-env/bin/activate
    if ! python3 -c "import torch; print('PyTorch working')"; then
        log_error "Python environment test failed"
        return 1
    fi

    log_success "All installation tests passed!"
}

# Show next steps
show_next_steps() {
    echo
    echo -e "${CYAN}🎉 Installation Complete!${NC}"
    echo
    echo -e "${CYAN}Next Steps:${NC}"
    echo "1. Log out and back in (or reboot) for Docker group changes"
    echo "2. Run: source ~/garage-env/bin/activate  # Activate Python environment"
    echo "3. Run: bash scripts/garage_start.sh      # Start your Garage AI node"
    echo
    echo -e "${CYAN}Useful Commands:${NC}"
    echo "• Check GPU: nvidia-smi"
    echo "• Test Docker: docker run --rm --gpus all ubuntu nvidia-smi"
    echo "• Activate Python: source ~/garage-env/bin/activate"
    echo
    echo -e "${CYAN}Need Help?${NC}"
    echo "• Documentation: https://github.com/garage-ai/platform#readme"
    echo "• Discord: https://discord.gg/garage-ai"
    echo "• Issues: https://github.com/garage-ai/platform/issues"
    echo
    log_success "Welcome to Garage AI! 🚀"
}

# Main installation function
main() {
    echo -e "${CYAN}"
    echo " ██████   █████  ██████   ██████   ██████  ███████"
    echo "██       ██   ██ ██   ██ ██    ██ ██       ██"
    echo "██   ███ ███████ ██████  ██    ██ ██   ███ █████"
    echo "██    ██ ██   ██ ██   ██ ██    ██ ██    ██ ██"
    echo " ██████  ██   ██ ██   ██  ██████   ██████  ███████"
    echo "${NC}"
    echo -e "${CYAN}Garage AI Complete Installation${NC}"
    echo "=================================="
    echo

    check_root
    detect_system

    # Check what needs to be installed
    local needs_gpu_drivers=true
    local needs_docker=true
    local needs_nvidia_toolkit=true
    local needs_python=true

    # Check current state
    if command -v nvidia-smi &> /dev/null && nvidia-smi &> /dev/null; then
        log_info "NVIDIA drivers already installed"
        needs_gpu_drivers=false
    fi

    if command -v docker &> /dev/null && docker info &> /dev/null 2>&1; then
        log_info "Docker already installed"
        needs_docker=false
    fi

    if docker run --rm --gpus=all ubuntu nvidia-smi &> /dev/null 2>&1; then
        log_info "NVIDIA Container Toolkit already configured"
        needs_nvidia_toolkit=false
    fi

    if [[ -d ~/garage-env ]] && source ~/garage-env/bin/activate && python3 -c "import torch" &> /dev/null; then
        log_info "Python AI environment already set up"
        needs_python=false
    fi

    # Install components
    if [[ "$needs_gpu_drivers" == true ]]; then
        install_nvidia_drivers
        # Note: This will reboot, so script will restart
    fi

    if [[ "$needs_docker" == true ]]; then
        install_docker
    fi

    if [[ "$needs_nvidia_toolkit" == true ]]; then
        install_nvidia_toolkit
    fi

    if [[ "$needs_python" == true ]]; then
        install_python_dev
        install_ai_libraries
    fi

    # Setup repository
    setup_repository

    # Test everything
    test_installation

    # Show next steps
    show_next_steps
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "Garage AI Complete Installation Script"
        echo
        echo "This script installs all prerequisites for running Garage AI:"
        echo "• NVIDIA GPU drivers"
        echo "• Docker + NVIDIA Container Toolkit"
        echo "• Python AI environment (PyTorch, vLLM, etc.)"
        echo "• Garage AI repository"
        echo
        echo "Usage: $0 [OPTIONS]"
        echo
        echo "Options:"
        echo "  --help, -h          Show this help message"
        echo "  --test              Test current installation"
        echo "  --skip-gpu          Skip GPU driver installation"
        echo "  --reinstall         Force reinstallation of all components"
        echo
        exit 0
        ;;
    --test)
        check_root
        detect_system
        test_installation
        exit $?
        ;;
    --skip-gpu)
        SKIP_GPU=true
        ;;
    --reinstall)
        FORCE_REINSTALL=true
        ;;
    *)
        main "$@"
        ;;
esac
