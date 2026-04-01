# 🚀 Garage AI: Start Here

**Version 2.1** | **April 2026** | **Status: Prototype Running**

**📚 Quick Navigation:**
- **[🏠 README](../README.md)** — GitHub overview & getting started
- **[⚡ QUICK START](GARAGE_AI_QUICK_START.md)** — Set up your node
- **[🏗️ BLUEPRINT](GARAGE_AI_IMPLEMENTATION_BLUEPRINT.md)** — Technical deep dive

---

## 🎯 Vision: Democratize AI in Europe

We're building **Europe's first distributed sovereign AI infrastructure** — from gaming PCs, spare hardware, and garage nodes. Instead of depending on Big Tech cloud services, we create a community-owned AI network where anyone in Europe can:

- **Contribute**: Lend GPU compute from a gaming rig or dedicated node
- **Use**: Access powerful private AI without cloud costs
- **Own**: Data stays on your hardware, in your country

**Result**: GPT-4 level AI performance from European garages, with 100% local data sovereignty.

---

## 🌊 The Four Waves

Every major technology shift has followed the same pattern: power concentrates at the frontier, then decentralizes back to individuals.

1. **Personal Computing (1980s)** — Compute left the mainframe. A PC on every desk.
2. **The Internet (1990s–2000s)** — Information left the library. Anyone could publish and connect.
3. **Generative AI (2022–2024)** — Creation was democratized — but sovereignty moved back to the cloud.
4. **Autonomous Agents (2025→)** — Action is being democratized. Agents make thousands of inference calls per task. The intelligence that left the mainframe in the 1980s is now leaving the cloud — and coming back to your garage.

Garage AI is infrastructure for Wave Four.

---

## 🏗️ Architecture

### Core Stack
```
User Apps → LiteLLM Proxy → Garage AI Network
    ↓              ↓              ↓
AnythingLLM    Token Tracking   Podman + vLLM
Custom Apps    Load Balancing   Multi-GPU Tensor Parallelism
Batch Jobs     Rate Limiting    Enterprise Isolation
```

### What Makes It Different

| Feature | Traditional AI | Garage AI |
|---------|---------------|-----------|
| **Hardware** | Data centers ($M) | Gaming PCs & garage nodes |
| **Latency** | 100–1000ms | <500ms chat |
| **Privacy** | Cloud logging | 100% local |
| **Cost** | $/hour | Electricity only |
| **Ownership** | Big tech | Community-owned |
| **Data** | Leaves your country | Stays in Europe |

---

## 🔒 Security: Podman-in-Docker Isolation

The core security innovation — borrowed and adapted from [Nosana](https://nosana.com)'s proven architecture — is **Podman-in-Docker**:

```
Host System (Gaming PC / Garage Node)
├── Docker Layer (manages Podman daemon)
│   └── Podman Daemon (isolated runtime)
│       ├── Inference Network (virtual LAN)
│       │   ├── vLLM Container (port 8000 — internal only)
│       │   ├── Worker Container
│       │   └── Model Cache (isolated storage)
│       └── ← Host system cannot see this layer
└── Host Network (completely separated)
```

**Why this matters**: AI workloads run in a fully isolated Podman namespace. The host gaming PC is invisible to the inference workload, and the workload is invisible to the host. Multiple tenants can share a node without seeing each other's data.

This architecture was proven at scale by Nosana, which processed 826,000+ AI jobs across 2,000+ nodes globally before we built Garage AI.

---

## 📊 Performance & Economics

### Node Performance
- **RTX 4090**: 150–300 tokens/sec (Llama-7B)
- **RTX 4090**: ~30 tokens/sec (Llama-70B quantized)
- **Mac Mini M4 Pro 64GB**: ~15 tokens/sec (Llama-70B local)
- **Cost per 1K tokens**: ~$0.001 (electricity only)

### The Inference Market
- **$106B** AI inference market (2025)
- **$255B** projected by 2030 (CAGR 19.2%)
- **1,000×** demand multiplier from agentic AI

The market has been proven by Anthropic ($380B valuation, $19B annualized revenue) and others. Garage AI is the community alternative to centralized inference monopolies.

---

## 🚀 Roadmap

### Phase 1: Sweden Pilot (2025–2026) ✅ Underway
- ✅ Podman-in-Docker GPU isolation working
- ✅ LiteLLM integration for chat + tokens
- ✅ Dual-path architecture (chat + batch)
- 🔄 Node operator community growing
- 🔄 Solar + EV energy integration

### Phase 2: Scandinavian Expansion (2026–2027)
- 🔄 Multi-node geographic clustering
- 🔄 Community token system
- 🔄 B2B private inference offering
- 🔄 Dashboard and monitoring

### Phase 3: European Scale (2027+)
- 🔄 Germany, Netherlands, Finland, Norway expansion
- 🔄 10,000+ active nodes
- 🔄 Enterprise SLA tier
- 🔄 EU policy engagement

---

## 🎮 Getting Started

### Prerequisites
- RTX 30/40 series GPU (8GB+ VRAM) or Apple Silicon (32GB+ unified memory)
- Ubuntu 20.04+ Linux (for NVIDIA nodes) or macOS (for Apple Silicon via Ollama)
- Docker installed

### One-Command Setup
```bash
bash <(wget -qO- https://garage.ai/start.sh)

# What happens:
# 1. Podman-in-Docker with GPU passthrough
# 2. Node registration in the network
# 3. Hardware benchmark
# 4. vLLM server startup
```

### Integration with LiteLLM
```yaml
# Add to litellm/config.yaml
model_list:
  - model_name: garage-chat
    litellm_params:
      model: openai/gpt-4
      api_base: http://your-node-ip:8000
```

---

## 🛡️ Privacy & Sovereignty

- **100% Local**: No data leaves your hardware
- **Container Isolation**: Each workload runs separately
- **Host Protection**: Your PC is untouched by AI workloads
- **NAT-Friendly**: Works behind home routers without port forwarding
- **Swedish First**: GDPR-compliant by architecture

---

## 🙏 Standing on the Shoulders of Giants

Garage AI did not invent decentralized GPU inference. We built on what works:

- **[Nosana](https://nosana.com)** (Amsterdam) — Pioneered the Podman-in-Docker pattern for decentralized GPU AI inference. Proven at scale: 2M+ deployments. Their architecture is the technical foundation for Garage AI's node setup.
- **[vLLM](https://docs.vllm.ai)** — State-of-the-art inference engine
- **[LiteLLM](https://litellm.ai)** — Load balancing and token tracking
- **[Ollama](https://ollama.com)** — Local model serving (Apple Silicon path)
- **Open Source AI Community** — For making models free and accessible

---

*Garage AI: From European garages to sovereign AI infrastructure* 🇪🇺🤖

**April 2026 — Sweden Pilot Running**
