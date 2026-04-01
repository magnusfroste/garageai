# 🧠 Garage AI Knowledge Base

**Version**: 2.1 | **Last Updated**: April 2026 | **Status**: Active

This knowledge base contains technical documentation, implementation details, and architectural overview for Garage AI — a distributed AI network built on idle hardware in European homes and garages.

---

## 📋 Table of Contents

- [Overview](#overview)
- [Core Concepts](#core-concepts)
- [Resource Utilization](#resource-utilization)
- [Swedish Infrastructure](#swedish-infrastructure)
- [Technical Implementation](#technical-implementation)
- [AI Inference](#ai-inference)
- [Inspiration: Nosana](#inspiration-nosana)
- [Community & Development](#community--development)
- [Future Vision](#future-vision)

---

## 🎯 Overview

### Vision
**Garage AI**: Democratize AI through decentralized infrastructure. Help build Europe's most sustainable AI network — powered by idle hardware and renewable energy from homes across the continent.

### Key Numbers
- **75M+ Garages** across the EU as potential AI nodes
- **1.7 Million Garages** in Sweden alone
- **98% Fiber Coverage** enabling distributed compute in Sweden
- **300,000+ Solar Installations** in Swedish homes (2024)
- **1.5M+ EVs** registered in Sweden — V2H buffer potential

### Unique Value Proposition
- **100% Local Control** — No data leaves the user's hardware
- **Resource Efficiency** — Reuse existing idle hardware
- **Environmentally Positive** — CO₂-neutral through renewable energy
- **Open Source** — MIT licensed, community-driven
- **Scalable** — From a single node to a national network

## 🔑 Core Concepts

### Resource Efficiency First
Garage AI focuses on smart reuse of hardware that would otherwise sit idle. Instead of gaming PCs collecting dust while you sleep or work, they can contribute to AI inference and generate value.

### Why This Is Different
- **No hardware waste**: Hundreds of thousands of gaming PCs in Sweden sit idle 16–20 hours/day
- **Fast technology cycles**: Use hardware while it's relevant, not after it's obsolete
- **Simple to join**: One computer becomes one node — easy and accessible

---

## 🔄 Resource Utilization

### Hardware Reused
```
Swedish Gaming PC Market:
├── Hundreds of thousands of gaming PCs available
├── 16–20 hours/day typically idle
├── Rapid GPU improvement cycles
└── Join as a ready-made cluster node — easy to start
```

### Technical Reality
- **RTX 4090**: ~120 tokens/sec for Llama-70B
- **Mac Mini M4 Pro (64GB)**: ~15 tokens/sec for Llama-70B — entirely locally
- **Idle draw**: 50–150W when not gaming
- **Repurpose**: Create AI value instead of waste
- **Scalability**: From one node to a network

---

## 🇸🇪 Swedish Infrastructure

### Fiber Network
```
PTS Statistics 2024:
├── 98% fiber coverage in Sweden
├── 123,000+ households with 1 Gbit/s+
├── 10 Gbit/s available on fiber networks
└── Operators: Telia, Telenor, Tele2, local ISPs
```

### Solar Panels & Energy
```
Swedish Solar Installations:
├── 300,000+ households have solar panels (2024)
├── 25% annual growth in solar installations
├── 10–12 peak sun hours/day (May–September)
└── 85% of solar energy produced by private homeowners
```

### Smart Energy Synergy
- **Daytime production**: Solar panels power AI nodes during the day
- **Surplus energy**: AI runs on clean solar power
- **Locally produced**: No grid load or transmission loss
- **CO₂-neutral**: 100% renewable energy for AI inference

### Energy Flow (Honest Assessment)
| Time | Primary Source | AI Node Status |
|------|---------------|----------------|
| Daytime (solar) | Solar direct or via Powerwall | ✅ Runs on renewables |
| Evening (EV home) | EV V2H discharge | ✅ Buffer available |
| Night (base load) | Grid | ⚡ Grid electricity |
| Weekends/WFH | Full solar + EV | ✅ Best renewable ratio |

---

## 💻 Technical Implementation

### Getting Started
If you already have a gaming PC with an RTX 30/40/50 series GPU, you can start immediately. All software is free. The only cost is electricity.

### Simple Setup
```bash
# One-command installation
bash <(wget -qO- https://garage.ai/start.sh)
```

**What happens:**
1. GPU and Docker prerequisites checked
2. Podman-in-Docker configured (GPU passthrough)
3. Node identity generated (UUID-based)
4. Node registered with Garage AI API
5. Hardware benchmarked
6. vLLM inference worker started

### Hardware Requirements
```yaml
Minimum:
  GPU: RTX 3060 (12GB VRAM)
  RAM: 16GB
  Storage: 64GB free
  Network: 100 Mbps stable

Recommended:
  GPU: RTX 4070+ (24GB+ VRAM)
  RAM: 32GB
  Storage: 256GB SSD
  Network: 1 Gbps fiber
```

### Enterprise & Business
Companies can set up dedicated nodes for private AI inference — ideal for GDPR-sensitive processes such as document analysis, customer support, or internal AI assistants.

Key benefits for businesses:
- **Secure data handling**: Sensitive information stays within company control
- **GDPR compliance**: No data leaves local servers
- **IP protection**: Protect trade secrets and innovation
- **Internal development**: AI assistants for coding, design, analysis

---

## 🌍 AI Inference

### What is AI Inference?
AI inference is the process by which a trained AI model generates responses to questions or creates content. Instead of sending data to the cloud (OpenAI, Google), everything runs locally on your hardware.

### Security Comparison
```
Cloud AI (OpenAI, Google):
❌ Data sent to external servers
❌ Processed in distant data centers
❌ Logging for training/safety
❌ Cost per request

Garage AI:
✅ 100% local processing
✅ No data leaves your hardware
✅ No logging or analytics
✅ Only electricity cost
```

### The Agentic Demand Surge
Autonomous AI agents (like OpenClaw, FlowWink, Silicon Soap) make hundreds to thousands of inference calls per task. This creates exponential demand for compute that centralized cloud providers cannot serve sustainably or affordably. Local inference becomes the rational default.

---

## 🙏 Inspiration: Nosana

Garage AI's core architecture — **Podman-in-Docker GPU isolation for AI inference** — was directly inspired by [Nosana](https://nosana.com), a Dutch open-source GPU compute marketplace.

### What Nosana Proved
Nosana demonstrated at scale that:
- Consumer and prosumer GPUs can serve real AI workloads reliably
- Podman-in-Docker provides the right isolation layer: the host system cannot see or interfere with the inference workload
- A pull-based job model works behind NAT without port forwarding
- Decentralized GPU networks can process millions of AI jobs (826,000+ as of mid-2025; 2M+ deployments by August 2025)

### Nosana's Architecture (What We Cloned)
```bash
# Nosana's proven setup command (the template for garage_start.sh)
bash <(wget -qO- https://nosana.com/start.sh)

# Their Podman-in-Docker isolation pattern:
docker run -d --gpus=all --name podman nosana/podman:latest
docker exec podman podman run --device nvidia.com/gpu=all vllm/vllm-openai:latest
```

This pattern ensures the AI workload runs in a fully isolated Podman namespace — the gaming PC host remains untouched and invisible to the workload.

### Key Difference: Garage AI vs Nosana
| | Nosana | Garage AI |
|---|---|---|
| **Chain** | Solana blockchain | Community / open |
| **Token** | NOS (crypto) | GAI (community points) |
| **Focus** | Global GPU marketplace | European sovereign AI |
| **Energy** | Any source | Solar + EV priority |
| **Governance** | Crypto DAO | Community-owned |
| **Geography** | Global | Sweden first → EU |

Nosana is a great project and a genuine pioneer. Garage AI takes the same technical foundation and applies it to European data sovereignty, renewable energy integration, and community governance — not crypto speculation.

---

## 👥 Community & Development

### Open Source & Community-Driven
- **MIT Licensed**: Free to use and modify
- **GitHub Community**: Discussions, issues, contributions
- **Partners**: Autoversio & Liteit
- **Inspired by**: Nosana's proven architecture

### Getting Started for Contributors
1. **Clone repo**: `git clone https://github.com/magnusfroste/garageai`
2. **Install dependencies**: `npm install`
3. **Start development**: `npm run dev`
4. **Contribute**: Documentation, code, tests

### Future Development
- **More GPU models**: Support for AMD, Intel
- **Edge computing**: Offline AI capabilities
- **Federated learning**: Collaboration between nodes
- **API development**: Local AI services for businesses

---

## 🎯 Future Vision

### Next Steps
- **Prototype expansion**: More nodes, geographic clustering
- **Community building**: Better onboarding and documentation
- **Partner integrations**: Deeper work with Autoversio & Liteit
- **Scaling**: From proof-of-concept to national network

### Vision 2026
- **Pilot nodes** running across Sweden
- **Community growing** across Scandinavia
- **B2B offering** live for early enterprise customers
- **CO₂-neutral AI** through solar integration demonstrated

### Vision 2027
- **European expansion**: Norway, Finland, Germany, Netherlands
- **10,000+ active nodes** across Europe
- **Sovereign EU AI infrastructure** as a reference architecture

---

## 📚 References

### Official Sources
- **SCB**: Statistics on Swedish households and housing
- **PTS**: Fiber infrastructure and coverage
- **Energimyndigheten**: Solar energy and renewables in Sweden
- **Eurostat**: EU housing and garage statistics
- **IEA / ACEA**: EV adoption data

### Technical Documentation
- **Nosana**: [nosana.com](https://nosana.com) — Original Podman-in-Docker GPU isolation pattern
- **vLLM**: [docs.vllm.ai](https://docs.vllm.ai) — Inference engine
- **LiteLLM**: Load balancing and token tracking
- **Podman**: Container isolation runtime

---

## 🚀 Summary

Garage AI is a practical answer to AI's centralization problem. By reusing idle hardware, we build a distributed AI network that:

- **Protects privacy**: 100% local data processing
- **Optimizes resources**: No hardware waste
- **Advances sovereignty**: Swedish and European AI infrastructure
- **Respects the environment**: CO₂-neutral through renewables
- **Builds community**: Open source, accessible to all

**Ready to start?** → [GET_STARTED.md](GET_STARTED.md)

---

## 🙏 Credits

**Kilo Code** — AI inference powered by Kilo Code, produced in our own garage on solar energy! ☀️🏠

*Decentralized AI, for the people, by the people.*
