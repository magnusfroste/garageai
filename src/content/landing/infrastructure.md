---
section: "infrastructure"
title: "The Garage Stack"
subtitle: "How Your Garage Becomes an AI Data Center"
description: "Five layers. All already installed in millions of Swedish homes. Together they form the sovereign, decentralized AI infrastructure Europe needs."
order: 4
---

## Stack Layers

### Layer 1: Solar Panels
- **Step**: 01
- **Icon**: ☀️
- **Title**: Solar Panels
- **Subtitle**: Energy Production
- **Color**: var(--color-warning)
- **Details**: Roof-mounted solar (8–12 kW) generates 40–60 kWh on a good summer day in Sweden. This is the primary energy source — powering the home, charging the battery, and running the AI node directly during daylight hours.
- **Footnote**: Energimyndigheten: Swedish solar irradiation data 2024

### Layer 2: Home Battery Storage
- **Step**: 02
- **Icon**: 🔋
- **Title**: Home Battery Storage
- **Subtitle**: The Critical Link — Powers the Node 24/7
- **Color**: var(--color-accent)
- **Details**: Home batteries (Tesla Powerwall: 13.5 kWh, BYD: 10–15 kWh) are the backbone of the energy stack. They absorb solar surplus during the day and release it at night — decoupling the AI node from solar timing entirely. The node runs around the clock regardless of weather or time of day.
- **Footnote**: Tesla Powerwall 3 specs; BYD Battery Box specs

### Layer 3: EV Battery
- **Step**: 03
- **Icon**: 🚗
- **Title**: EV Battery — Evening & Weekend Buffer
- **Subtitle**: V2H (Vehicle-to-Home), not primarily V2G
- **Color**: rgba(255,59,48,0.9)
- **Details**: The honest picture: cars are away during solar peak hours on weekdays. But they're home during evenings and weekends — exactly when V2H (Vehicle-to-Home) adds value. A 60–100 kWh EV battery parked at home from 17:00 can supply the house and AI node overnight. Studies show cars are parked ~92% of the time; roughly half of that is at home, concentrated in evenings and weekends.
- **Footnote**: INRIX / IEA: average car parked 92–95% of time. IEA Global EV Outlook 2024 on V2H deployment.

### Layer 4: AI Inference Node
- **Step**: 04
- **Icon**: 🖥️
- **Title**: AI Inference Node
- **Subtitle**: The Computing Layer
- **Color**: var(--color-primary)
- **Details**: A Mac Mini M4 Pro, RTX 4090 rig, or purpose-built inference server in your garage. Runs local models (Llama, Mistral, Qwen) at 15–120 tokens/sec. Consumes 300–600W.

### Layer 5: Fiber Connection
- **Step**: 05
- **Icon**: 🌐
- **Title**: Fiber Connection
- **Subtitle**: Distribution
- **Color**: rgba(0,122,255,0.9)
- **Details**: Sweden's 1–10 Gbit/s fiber network connects your node to neighbours, local businesses, and the broader Garage AI mesh. Inference stays local — bandwidth is just for coordination.

## CTA

- **Text**: "The stack is already sitting in your driveway. The only question is whether Europe uses it."
- **Button**: "See the Setup Guide" → https://github.com/magnusfroste/garageai/blob/main/docs/GET_STARTED.md
