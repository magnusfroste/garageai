# 🚀 Garage AI Quick Start

**Nosana Clone - Podman-in-Docker Setup**

---

## 🎯 One Command Setup (Like Nosana)

```bash
# Exactly like Nosana's approach:
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

## 🔧 Manual Setup (If One-Command Fails)

### Step 1: Prerequisites (Same as Nosana)

```bash
# Ubuntu 20.04+
lsb_release -a

# Internet connectivity
curl -s https://garage.ai > /dev/null && echo "Connected"
```

### Step 2: NVIDIA Setup (Clone Nosana)

```bash
# Install NVIDIA drivers (from their docs)
# Download from nvidia.com or:
ubuntu-drivers autoinstall

# Verify
nvidia-smi
```

### Step 3: NVIDIA Container Toolkit (Exact Clone)

```bash
curl -fsSL https://nvidia.github.io/libnvidia-container/gpgkey | sudo gpg --dearmor -o /usr/share/keyrings/nvidia-container-toolkit-keyring.gpg
curl -s -L https://nvidia.github.io/libnvidia-container/stable/deb/nvidia-container-toolkit.list | sed 's#deb https://#deb [signed-by=/usr/share/keyrings/nvidia-container-toolkit-keyring.gpg] https://#g' | sudo tee /etc/apt/sources.list.d/nvidia-container-toolkit.list
sudo apt-get update
sudo apt-get install -y nvidia-container-toolkit
sudo nvidia-ctk runtime configure --runtime=docker
sudo systemctl restart docker
```

### Step 4: Docker Setup (From Nosana)

```bash
# Follow: https://www.digitalocean.com/community/tutorials/how-to-install-and-use-docker-on-ubuntu-22-04
# Enable non-root usage
sudo usermod -aG docker $USER
newgrp docker
```

### Step 5: Podman-in-Docker (Exact Clone)

```bash
# Same command as Nosana docs
docker volume create podman-socket
docker volume create podman-cache

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

# Verify (same as Nosana)
docker exec podman curl -s http://localhost:8080/v4.5.0/libpod/info
```

### Step 6: GPU Test (From Nosana)

```bash
docker exec -it podman bash
podman run --rm --device nvidia.com/gpu=all --security-opt=label=disable ubuntu nvidia-smi -L
```

---

## 🧪 Testing Your Setup

### Basic Inference Test

```bash
# Test vLLM in Podman (like Nosana examples)
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
# Like Nosana's benchmark
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

- **Nosana Original**: https://docs.nosana.com/hosts/grid-ubuntu.html
- **Our Clone Analysis**: `GARAGE_AI_NOSANA_ANALYSIS.md`
- **Podman GPU**: NVIDIA container toolkit docs
- **vLLM**: https://docs.vllm.ai/

---

## 🎉 Success = Nosana Working!

When you see:
```
✅ nvidia-smi works in containers
✅ Podman daemon running
✅ vLLM serving models
✅ API responding
✅ Node registered
```

**Then you have successfully cloned Nosana's approach!** 🚀

---

*Focus: Clone what works, then add Swedish features*
*Status: Podman-in-Docker core ready*
*Next: API backend + worker images*
