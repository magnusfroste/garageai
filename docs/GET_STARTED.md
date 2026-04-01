# 🚀 Getting Started with Garage AI

Welcome to Garage AI! This guide helps you get up and running quickly, whether you're setting up a node or contributing to the project.

## 📋 Contents

- [Quick Start for Node Operators](#quick-start-for-node-operators)
- [Technical Installation](#technical-installation)
- [Community & Support](#community--support)
- [Contributing](#contributing)

---

## 🖥️ Quick Start for Node Operators

### What is Garage AI?
Garage AI lets you use your gaming PC or dedicated hardware to contribute to a sovereign European AI network when you're not using it. You earn tokens by lending your GPU compute to AI inference workloads.

### Step-by-Step Guide

#### 1. **Check Compatibility**
   - **GPU**: RTX 3060 or better (NVIDIA)
   - **RAM**: At least 16GB
   - **Storage**: 64GB free space
   - **Internet**: Stable 100 Mbps+ connection

#### 2. **One-Command Setup**
   ```bash
   # Complete installation (recommended for beginners)
   curl -fsSL https://garage.ai/install-all.sh | bash
   ```

#### 3. **Start Your Node**
   ```bash
   bash scripts/garage_start.sh
   ```

#### 4. **What Happens Automatically**
   - GPU drivers verified
   - Podman-in-Docker configured for inference isolation
   - Node identity generated (UUID-based)
   - Node registered with the Garage AI network
   - Hardware benchmark executed
   - Inference worker started

### What Happens Next?

- **Automatic operation**: Your machine contributes when idle
- **Token rewards**: Earn GAI tokens for each AI task completed
- **No lock-in**: Stop anytime, your hardware stays yours
- **Costs**: Only electricity (roughly 50–150W per node)

### Troubleshooting

#### "GPU not found"?
```bash
nvidia-smi
```

#### "No network connection"?
```bash
ping 8.8.8.8
```

#### Need help?
- **[GitHub Discussions](https://github.com/magnusfroste/garageai/discussions)** — Ask the community
- **Email**: powerup@garageai.eu

---

## 💻 Technical Installation

For developers and advanced users who want to modify or contribute.

### Prerequisites

- **Ubuntu 22.04+** or compatible Linux distribution
- **Python 3.9+**
- **NVIDIA GPU** with CUDA 11.8+
- **Git**

### Clone Repository

```bash
git clone https://github.com/magnusfroste/garageai.git
cd garageai
```

### Quick Install

```bash
# Install all dependencies
./scripts/setup.sh

# Configure GPU
./scripts/setup_gpu.sh

# Start development server
npm run dev
```

### Manual Installation

#### 1. Python Environment
```bash
python3 -m venv garage-env
source garage-env/bin/activate
pip install -r requirements.txt
```

#### 2. AI Libraries
```bash
# PyTorch with CUDA
pip install torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cu118

# AI inference stack
pip install transformers accelerate vllm ray[all] fastapi uvicorn
```

#### 3. GPU Setup
```bash
# Verify CUDA
nvidia-smi
python3 -c "import torch; print(f'CUDA: {torch.cuda.is_available()}')"

# Install vLLM
pip install vllm
```

### Test Installation

```bash
# Run diagnostics
python3 garage_node.py --diagnostics

# Register test node
python3 garage_node.py --register --api-url http://localhost:8000

# Start monitoring
python3 garage_node.py --monitor
```

### Docker Installation (Alternative)

```bash
docker build -t garage-ai .
docker run --gpus all -p 8000:8000 garage-ai
```

---

## 🌍 Community & Support

### Communication Channels

- **💬 [GitHub Discussions](https://github.com/magnusfroste/garageai/discussions)** — Real-time discussion, help, events
- **🐛 [GitHub Issues](https://github.com/magnusfroste/garageai/issues)** — Bug reports, feature requests
- **📧 [Email](mailto:powerup@garageai.eu)** — Direct contact

### Support Response Times

| Type | Response Time | Channel |
|------|--------------|---------|
| **General Questions** | 24h | GitHub Discussions |
| **Technical Support** | 12h | GitHub Issues |
| **Bug Reports** | 6h | GitHub Issues |
| **Security** | 2h | security@garageai.eu |

---

## 🤝 Contributing

### Why Contribute?

- **Shape the future**: Influence how decentralized AI works
- **GAI Token rewards**: Earn for code contributions
- **Community recognition**: Featured in releases
- **Learning**: Work with cutting-edge AI infrastructure

### Easy First Contributions

1. **🐛 Report Bugs**: Clearly describe what went wrong
2. **💡 Suggestions**: Share ideas for improvements
3. **📚 Documentation**: Improve guides and tutorials
4. **🧪 Testing**: Help test new features

### Technical Contributions

#### Code Standards
```python
# Follow PEP 8
# Use type hints
def process_inference(input_data: Dict[str, Any]) -> Dict[str, Any]:
    """Process AI inference request.

    Args:
        input_data: Input data dictionary

    Returns:
        Processed result dictionary
    """
    pass
```

#### Pull Request Process
1. **Fork** the repository
2. **Create branch**: `feature/your-feature` or `fix/issue-number`
3. **Commit messages**: "Add: New feature" or "Fix: Issue description"
4. **Test**: All tests must pass
5. **PR**: Include description, screenshots, test instructions

### Contribution Areas

#### 🔥 High Priority
- **Node Setup UX**: Simpler onboarding for non-technical users
- **GPU Compatibility**: Support for AMD and Intel GPUs
- **Network Optimization**: Better latency handling

#### 📈 Medium Priority
- **Web Dashboard**: Improved node management UI
- **Model Marketplace**: Dynamic model loading
- **Energy Monitoring**: Solar/battery integration hooks

#### 🚀 Future
- **Federated Learning**: Collaborative model training
- **Multi-Modal AI**: Image/text/video support
- **Edge Computing**: Offline capabilities

### Recognition

Contributors are credited in:
- **Changelog**: Release notes
- **Contributors file**: GitHub contributors list
- **Token Rewards**: GAI tokens for qualifying contributions

---

## 📚 Next Steps

### For Node Operators
1. **Try it yourself**: Run the setup script
2. **Join the community**: GitHub Discussions
3. **Share feedback**: Help improve the experience

### For Contributors
1. **Read CONTRIBUTING.md**: Detailed guidelines
2. **Browse Issues**: Find something to work on
3. **Start simple**: Documentation or small bug fixes
4. **Ask questions**: Community is friendly

### For the Technically Curious
1. **Study the architecture**: Read the technical docs
2. **Experiment**: Test proof-of-concept code
3. **Build on it**: Create new features or improvements

---

*Questions? Open a [GitHub Discussion](https://github.com/magnusfroste/garageai/discussions)*

*Built with ❤️ by the open source community · Inspired by [Nosana](https://nosana.com)'s proven decentralized GPU architecture*
