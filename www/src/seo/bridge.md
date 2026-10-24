---
path: "/bridge"
title: "Earn Money With Your GPU — GarageAI Bridge"
description: "Connect a gaming PC, Mac or Linux machine to GarageAI with one line, approve it in the browser and earn your price for every token it serves. Open source."
kicker: "GarageAI Bridge · for GPU owners"
breadcrumb: "Earn with your GPU"
h1: "Earn money with the GPU you already own"
lead: "GarageAI Bridge connects your computer to GarageAI's marketplace for AI inference. Run one line, approve the computer in the browser, and in a few minutes your GPU earns the price you set for every token it serves. Bridge is open source, opens no ports, and your machine stays yours first."
commands:
  - label: "macOS and Linux, in the terminal"
    command: "curl -fsSL https://raw.githubusercontent.com/magnusfroste/garageai/main/cli/install.sh | sh -s -- connect"
  - label: "Windows, in PowerShell"
    command: "irm https://raw.githubusercontent.com/magnusfroste/garageai/main/cli/install.ps1 | iex; garageai connect"
    note: "Nothing secret in either line. On Windows, Bridge asks for Administrator once."
ctas:
  - text: "Connect your GPU"
    url: "https://app.garageai.eu/connect"
  - text: "Bridge on GitHub"
    url: "https://github.com/magnusfroste/garageai/blob/main/cli/README.md"
faqs:
  - question: "How much can I earn with my GPU?"
    answer: "It depends on demand for the models you serve and on the price you set per million tokens; GarageAI does not promise a figure. You earn your price for every token your garage serves, also through the pool, and there is no platform fee during launch. A GPU that sits idle earns nothing, so popular models, a fair price and a machine that stays on earn the most."
  - question: "Which GPU do I need?"
    answer: "Any machine that runs an open model at a useful speed: a Mac with Apple silicon, a gaming PC with an NVIDIA RTX card, a Linux workstation or an AI box such as the DGX Spark. There is no fixed minimum. Every model passes an acceptance test with a real request, which measures time to first token and tokens per second, before it is sold. The site's Pick your setup section suggests a first model per amount of GPU memory."
  - question: "Can I still use my computer for games or work?"
    answer: "Yes. garageai pause takes your garage off the market for a while and garageai resume brings it back; when you use your own machine through GarageAI, buyers are sent to other garages until a few minutes after your last request. Pausing automatically when a game or a render takes the GPU is coming to Bridge."
  - question: "Does a laptop work?"
    answer: "Yes, while it is plugged in. Set it not to sleep on the power adapter (garageai doctor tells you how); on battery the garage is shown as away, not down, and comes back when you plug in."
  - question: "How do I stop?"
    answer: "garageai pause stops selling for a while. garageai uninstall removes Bridge, its two services and its settings. Your runtime and your models stay as they were."
related:
  - text: "Is it safe to share from my computer?"
    url: "/security"
  - text: "Use your own machine from your phone, privately and free"
    url: "/my-machine"
  - text: "The OpenAI-compatible API buyers use"
    url: "/api"
  - text: "Audit Bridge yourself, or let your AI do it (AUDIT.md)"
    url: "https://github.com/magnusfroste/garageai/blob/main/AUDIT.md"
---

## How it works

1. **Create an account** at app.garageai.eu and open *Connect your GPU*.
2. **Run one line** on the machine with the GPU. It installs GarageAI Bridge, a small open-source program (`garageai`), which looks at your GPU, memory, runtime and models.
3. **Approve your computer in the browser.** Bridge opens app.garageai.eu with a code; check that it is your computer, name your garage and approve. No token is copied anywhere.
4. **Go live.** Your models pass an acceptance test with a real request, the page says *Your garage is live*, and you set your price per million tokens in the portal.

## What you need

- **macOS** (Apple silicon), **Linux** or **Windows**, with a GPU.
- A runtime with a model loaded: **Ollama, LM Studio, llama.cpp, vLLM, SGLang** or **Xyntetik Runner**. No runtime yet? `garageai runtime` checks whether Xyntetik Runner, a one-file engine made in Sweden, fits your machine, and `garageai runtime install` sets it up with a model.
- A normal home or office internet connection. You open no ports on your router.

## How you get paid

You set the price per million tokens, input and output. You earn it for every token your garage serves, whether a buyer picked your garage or the request came through the pool, and there is no platform fee during launch.

- **Now:** every token your garage serves is counted at your price and shown in *My garages*.
- **Next:** use your earnings as GarageAI credits on any model, for example larger models than your own machine runs. Everything earned from launch carries over.
- **Later:** payouts to your bank account through Stripe.

## What Bridge does on your computer

- **Its own encrypted tunnel.** WireGuard, built into Bridge. Your machine dials out to the EU gateway, so it works behind any router and through NAT. No VPN app, no network interface, no ports opened, and it runs fine next to Tailscale.
- **Your runtime stays on localhost.** Bridge forwards one port into the tunnel, and only inference requests pass: downloading or deleting models and every other management call is refused. Nothing else on your computer or your home network can be reached.
- **No secrets leave the machine.** The tunnel key is made on your computer; only its public half is sent. Bridge never reports prompts, keys or command lines.
- **It keeps itself up to date.** A heartbeat reports your models and their context windows every 5 minutes, and Bridge updates itself to the current stable release, checked against its checksum.
- **It tells you what is wrong.** `garageai doctor` shows what runs on the machine, what is wrong and exactly how to fix it.
- **GarageAI in your menu bar.** On macOS (menu bar) and Windows (notification area), a small GarageAI icon shows at a glance whether your garage is live, which models it serves and how many requests it handled in the last hour, with *Open My garage* and *Check this machine*. It is read-only and optional; `garageai connect` offers it at the end.
- **It measures energy.** On NVIDIA GPUs and Apple silicon, Bridge samples power every 5 seconds and reports the energy your garage used while serving: watt-hours and tokens per kWh, per model. The menu shows today's figure, for example *Energy today: 420 Wh · 1.2M tokens/kWh*. Only the model name and the token counts are kept; prompts and answers are not stored.

## Your machine stays yours first

Your own garage is available to you from your phone or any other computer, privately and free, whether or not it is on the market: see [My machine](/my-machine). While you use it, buyers are sent to other garages. And other people's prompts are confidential: the operator terms forbid logging, storing or reading them.
