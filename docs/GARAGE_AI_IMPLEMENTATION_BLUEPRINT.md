# 🚀 Garage AI: Implementation Blueprint

**Version**: 2.1 | **Status**: Active | **Focus**: Chat + Batch Workloads

This document describes the technical architecture for Garage AI's distributed AI inference cluster, including the Nosana-inspired Podman-in-Docker isolation model.

**📚 Quick Navigation:**
- **[🏠 README](../README.md)** — Overview & learning path
- **[🚀 START HERE](GARAGE_AI_START_HERE.md)** — Vision & innovation
- **[⚡ QUICK START](GARAGE_AI_QUICK_START.md)** — Set up your node

---

## 🙏 Architecture Credit: Nosana

Garage AI's core isolation pattern is directly inspired by **[Nosana](https://nosana.com)** — a Dutch open-source decentralized GPU compute marketplace. Nosana proved that:

- Podman-in-Docker provides correct GPU isolation for AI inference
- A pull-based node model works reliably behind NAT
- Consumer-grade GPUs can power real production AI workloads at scale

Nosana's node setup command was the template for `garage_start.sh`:
```bash
# Nosana's original pattern (what we studied and adapted)
bash <(wget -qO- https://nosana.com/start.sh)
```

Their architecture is proven: 826,000+ AI jobs, 2M+ deployments across 2,000+ nodes globally.

Garage AI adapts this architecture for **European data sovereignty**, **renewable energy integration**, and **community governance** — without blockchain/crypto dependency.

---

## 🎯 Phase 1: Garage-Scale Prototype

### Goal: Multiple rigs accessible via LiteLLM proxy

#### Architecture
```
Your Network + LiteLLM → Multiple Garage Rigs via Load Balancing
├── Rig 1 (2× RTX 4090) — Direct access for chat
├── Rig 2 (2× RTX 4090) — Batch jobs for large models
├── Rig 3 (1× RTX 4080) — Backup capacity
└── All coordinated through central API
```

#### LiteLLM Integration
- **API Keys**: Per-user for token tracking
- **Usage Insights**: Token consumption and cost
- **Load Balancing**: Automatic distribution across GPUs
- **Fallback**: Chat → batch when needed

#### Token System
- **Simple Tracking**: Track tokens via LiteLLM
- **Compensation**: Community points for contributed GPU hours
- **Blockchain-ready**: Can be upgraded to full transparency later

---

## 🏗️ Dual-Path Implementation

### Path A: Real-Time Chat (<500ms)
```bash
# Direct to GPUs via LiteLLM
curl -X POST http://localhost:8000/v1/chat/completions \
  -H "Authorization: Bearer $API_KEY" \
  -d '{"model": "garage-chat", "messages": [...]}'

# LiteLLM routes to the nearest GPU automatically
# Result back within milliseconds
```

### Path B: Batch Jobs (5sec+)
```bash
# For large models, data analysis, etc.
garage job submit --model Qwen/Qwen3-80B --input data.json

# Queued and processed in distributed fashion
# Result fetched later via API
```

---

## 🔒 Podman-in-Docker: Security Deep Dive

### The Problem with Direct Docker vLLM
```bash
# Direct Docker approach — what NOT to do
docker run -p 8000:8000 vllm/vllm-openai:latest
# ❌ Port 8000 exposed directly on host
# ❌ Host system has access to AI workloads
# ❌ No workload isolation between users
# ❌ Security risks if model or prompt is adversarial
```

### Our Podman-in-Docker Solution (from Nosana)
```bash
# Podman creates an isolated network layer
docker run -d --name podman garageai/podman:v1.0.0
docker exec podman podman run -p 8000:8000 vllm/vllm-openai:latest

# ✅ AI workloads isolated in Podman network
# ✅ Host system cannot access vLLM directly
# ✅ Multiple workloads run independently
# ✅ Enterprise-grade isolation on consumer hardware
```

### Network Isolation Layers
```
Host System (Gaming PC)
├── Docker Layer (manages Podman)
│   └── Podman Daemon (isolated runtime)
│       ├── Inference Network (virtual LAN)
│       │   ├── vLLM Container (port 8000 — internal only)
│       │   ├── Worker Container (port 3000 — internal only)
│       │   └── Model Cache (isolated storage)
│       └── ← No direct host access
└── Host Network (completely separated)
```

### Security Benefits

**1. Workload Isolation**
- Each AI workload runs in a separate Podman container
- No interference between models or users
- Resource limits prevent abuse

**2. Network Security**
- No inbound ports required — NAT-friendly
- Internal networking only accessible via API proxy
- Encrypted communication channels

**3. Host Protection**
- Gaming PC remains untouched by AI workloads
- Easy to stop/remove all AI services
- No persistent changes to host system

**4. Enterprise Features**
- Rootless containers (no privileged access needed)
- SELinux/AppArmor integration
- Audit logging of all container activities

---

## 🎮 LiteLLM + Garage GPUs Setup

### Current LiteLLM (Already Perfect for This)
- ✅ User management built in
- ✅ Token tracking built in
- ✅ Load balancing across multiple endpoints
- ✅ OpenAI-compatible API

### Garage GPUs Config
```yaml
# LiteLLM config for garage nodes
model_list:
  - model_name: garage-chat
    litellm_params:
      model: openai/gpt-4
      api_base: http://rig1:8000
      api_key: garage-key
  - model_name: garage-chat
    litellm_params:
      model: openai/gpt-4
      api_base: http://rig2:8001
      api_key: garage-key
```

### For Chat
- **AnythingLLM** → **LiteLLM** → **Garage GPUs** (fast!)
- Load balancing across all available GPUs
- Token tracking and rate limiting

### For Batch
- Large models requiring multi-GPU
- Background processing
- Results via API when ready

---

## 📊 Performance & Scaling

### Chat Performance (Target)
- **Latency**: <500ms per response
- **Throughput**: 100+ requests/minute
- **Models**: GPT-4 level via GPU cluster

### Batch Performance (Target)
- **Qwen3-80B**: 4-GPU distributed (2+2 rigs)
- **Processing**: Background jobs
- **Results**: Async retrieval

### Token Economics
- **Usage tracking**: Via LiteLLM
- **Community rewards**: GPU hours = GAI tokens
- **Scalable**: From simple tracking to on-chain later

---

## 🛠️ Implementation Roadmap

### Phase 1: Core Setup ✅
- ✅ LiteLLM config for garage GPUs
- ✅ Chat test via LiteLLM proxy
- ✅ GPU load balancing
- ✅ Podman-in-Docker GPU isolation

### Phase 2: Multi-Rig
- 🔄 `garage.ai/start.sh` for each rig
- 🔄 API coordination between rigs
- 🔄 Batch job system

### Phase 3: Polish & Scale
- 🔄 Token system integration
- 🔄 Performance optimization
- 🔄 Community features

---

## 🔧 Technical Setup Per Node

### On Each Rig (Ubuntu)
```bash
# 1. Install prerequisites
sudo apt install nvidia-driver docker.io nvidia-container-toolkit

# 2. Run garage setup
bash <(wget -qO- https://garage.ai/start.sh)

# 3. Result: vLLM server on port 8000 (inside Podman)
```

### LiteLLM Configuration
```yaml
# config.yaml
model_list:
  - model_name: garage-chat
    litellm_params:
      model: openai/gpt-4
      api_base: http://rig1-ip:8000
  - model_name: garage-chat
    litellm_params:
      model: openai/gpt-4
      api_base: http://rig2-ip:8000
```

### AnythingLLM Config
```bash
# Point to LiteLLM proxy
OpenAI Base URL: http://localhost:8000
API Key: your-litellm-key
```

---

## 💡 Architecture Insights

### Nosana's Pull-Model (What We Adapted)
- Nodes poll for jobs (works behind NAT)
- Clients post jobs via API
- Perfect for batch, acceptable for chat

### LiteLLM's Proxy-Model (What We Added)
- Load balancing for real-time chat
- Token tracking built in
- OpenAI-compatible

### Our Hybrid
- **Chat**: LiteLLM proxy → Direct GPU access
- **Batch**: Job polling → Distributed processing
- **Best of both worlds**

---

## 🎉 Vision: Real-Time Cluster for Europe

With this setup:
```
Chat in AnythingLLM / any OpenAI-compatible app:
✅ Real-time responses from garage GPUs
✅ Load balancing across nodes
✅ Token tracking & community rewards

Large models:
✅ Qwen3-80B on 4 GPUs distributed
✅ Batch processing in background
✅ Community-powered AI
```

**From garage prototype to European AI infrastructure.** 🇪🇺🤖

---

## 🙏 Credits

- **[Nosana](https://nosana.com)** — Podman-in-Docker GPU isolation pattern, pull-based NAT-friendly node model. Proven architecture that made this possible.
- **[vLLM](https://docs.vllm.ai)** — High-throughput inference engine
- **[LiteLLM](https://litellm.ai)** — Load balancing and token tracking proxy
- **[Podman](https://podman.io)** — Rootless container isolation
- **[Ollama](https://ollama.com)** — Apple Silicon / CPU inference path

---

*Updated: April 2026*
*Focus: Real-Time Chat + Batch Jobs + European Sovereignty*
*Status: Prototype Running — Sweden Pilot Phase*
