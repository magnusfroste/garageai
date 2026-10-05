# 🚀 Garage AI Quick Start

> **⚠️ Previous architecture.** This document describes GarageAI's earlier design (a custom OS / boot image, vLLM-only nodes, Podman-in-Docker isolation and an overlay network). That approach has been dropped and this page is kept for history only.
> GarageAI now uses a NetBird (WireGuard) mesh and a LiteLLM gateway: operators keep their own macOS or Linux machine and any supported runtime (Ollama, LM Studio, llama.cpp, vLLM).
> To get started, use the portal at https://app.garageai.eu (operators: https://app.garageai.eu/auth?intent=operator). Technical setup: [infra/gateway/README.md](https://github.com/magnusfroste/garageai/blob/main/infra/gateway/README.md) and [scripts/garageai-connect.sh](https://github.com/magnusfroste/garageai/blob/main/scripts/garageai-connect.sh).

**Garage AI - Podman-in-Docker Setup**

**📚 Quick Navigation:**
- **[🏠 README](../README.md)** - Overview & learning path
- **[🚀 START HERE](GARAGE_AI_START_HERE.md)** - Vision & innovation
- **[🏗️ BLUEPRINT](GARAGE_AI_IMPLEMENTATION_BLUEPRINT.md)** - Technical deep dive

---

## 🎯 One Command Setup

### Option A: Complete Installation (Recommended for Beginners)
```bash
# Installs everything automatically: GPU drivers, Docker, Python AI environment
curl -fsSL https://garage.ai/install-all.sh | bash
```

**What this installs:**
- ✅ NVIDIA GPU drivers (reboot required)
- ✅ Docker + NVIDIA Container Toolkit
- ✅ Python AI environment (PyTorch, vLLM, etc.)
- ✅ Garage AI repository

**After installation, run:**
```bash
# Start your Garage AI node
bash scripts/garage_start.sh
```

### Option B: Garage AI Only (For Advanced Users)
```bash
# Assumes you have GPU drivers, Docker, and Python AI environment already set up
bash <(wget -qO- https://garage.ai/start.sh)
```

**What this does:**
1. ✅ Checks Ubuntu + GPU prerequisites
2. ✅ Sets up Podman-in-Docker (GPU passthrough)
3. ✅ Generates node identity (UUID-based)
4. ✅ Registers with Garage AI API
5. ✅ Benchmarks hardware
6. ✅ Starts inference worker

**Result:** Your gaming PC becomes a Garage AI node in minutes!

---

## 🔧 Manual Setup (Advanced Users Only)

**⚠️ Warning:** Manual setup is complex and error-prone. Use the automated script above instead!

If you must do manual setup, follow these steps in order:

### Prerequisites Check
```bash
# Ubuntu 20.04+ required
lsb_release -a

# Internet connectivity
curl -s https://garage.ai > /dev/null && echo "✅ Connected"
```

### GPU Drivers (Reboot Required)
```bash
# Install NVIDIA drivers
sudo apt update
sudo ubuntu-drivers autoinstall

# REBOOT required after this step
sudo reboot

# After reboot, verify:
nvidia-smi
```

### Docker Installation
```bash
# Remove old versions
sudo apt-get remove docker docker-engine docker.io containerd runc

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Enable non-root usage
sudo usermod -aG docker $USER
sudo systemctl enable docker

# Logout/login or run: newgrp docker
```

### NVIDIA Container Toolkit
```bash
# Add NVIDIA repository
curl -fsSL https://nvidia.github.io/libnvidia-container/gpgkey | sudo gpg --dearmor -o /usr/share/keyrings/nvidia-container-toolkit-keyring.gpg
curl -s -L https://nvidia.github.io/libnvidia-container/stable/deb/nvidia-container-toolkit.list | sed 's#deb https://#deb [signed-by=/usr/share/keyrings/nvidia-container-toolkit-keyring.gpg] https://#g' | sudo tee /etc/apt/sources.list.d/nvidia-container-toolkit.list

# Install toolkit
sudo apt-get update
sudo apt-get install -y nvidia-container-toolkit

# Configure Docker
sudo nvidia-ctk runtime configure --runtime=docker
sudo systemctl restart docker
```

### Step 5: Podman-in-Docker

```bash
# Create volumes for persistence
docker volume create podman-socket
docker volume create podman-cache

# Run Podman daemon in Docker
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

# Verify Podman is running
docker exec podman curl -s http://localhost:8080/v4.5.0/libpod/info
```

### Step 6: GPU Test

```bash
docker exec -it podman bash
podman run --rm --device nvidia.com/gpu=all --security-opt=label=disable ubuntu nvidia-smi -L
```

---

## 🧪 Testing Your Setup

### Basic Inference Test

```bash
# Test vLLM in Podman container
docker exec podman podman run --rm --device nvidia.com/gpu=all \
  -p 8000:8000 \
  vllm/vllm-openai:latest \
  --model microsoft/DialoGPT-small \
  --tensor-parallel-size 1 \
  --port 8000

# Test API
curl http://localhost:8000/v1/models
```

### Performance Benchmark

```bash
# Test inference performance
curl -X POST http://localhost:8000/v1/completions \
  -H "Content-Type: application/json" \
  -d '{"model": "microsoft/DialoGPT-small", "prompt": "Hello world", "max_tokens": 10}'
```

---

## 🚀 Scaling to Multiple Nodes

### Current Status
- **Single Node**: ✅ Working via Podman-in-Docker
- **Multi-Node**: 🔄 Next phase (API coordination)

### Adding More Nodes
```bash
# On each gaming PC:
bash <(wget -qO- https://garage.ai/start.sh)
# Nodes auto-register and coordinate via API
```

---

## 📊 Performance Expectations

### RTX 4090 Single Node
- **Llama-7B**: 150-300 tokens/sec
- **Setup Time**: 5-10 minutes
- **Memory**: ~14GB VRAM usage

### Network Scaling
- **Current**: Independent nodes
- **Future**: API-coordinated cluster
- **Goal**: Distributed Llama-70B support

---

## 🐛 Troubleshooting

### GPU Not Working
```bash
# Check NVIDIA setup
nvidia-smi
docker run --rm --gpus=all ubuntu nvidia-smi

# Reinstall toolkit
sudo apt-get install -y nvidia-container-toolkit
sudo nvidia-ctk runtime configure --runtime=docker
sudo systemctl restart docker
```

### Podman Issues
```bash
# Restart Podman
docker restart podman
docker logs podman

# Test connection
docker exec podman podman info
```

### Network Issues
```bash
# Test API connectivity
curl -I https://api.garage.ai/health
```

---

## 📚 Resources

- **Podman GPU**: NVIDIA container toolkit docs
- **vLLM**: https://docs.vllm.ai/
- **Docker GPU**: Docker documentation

---

## 🎉 Success!

When you see:
```
✅ nvidia-smi works in containers
✅ Podman daemon running
✅ vLLM serving models
✅ API responding
✅ Node registered
```

**Your gaming PC is now part of the Swedish AI network!** 🚀

---

*Focus: Build Swedish AI infrastructure*
*Status: Podman-in-Docker core ready*
*Next: API backend + worker images*
