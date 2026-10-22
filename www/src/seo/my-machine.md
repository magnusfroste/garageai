---
path: "/my-machine"
title: "Use Ollama or LM Studio From Your Phone — GarageAI My Machine"
description: "Reach the AI on your home computer from your phone or any computer, privately and free. No open ports, no VPN app: GarageAI Bridge opens the tunnel."
kicker: "My machine · your own AI, from anywhere"
breadcrumb: "Your AI from your phone"
h1: "Use the AI on your home computer from your phone, anywhere"
lead: "Ollama, LM Studio and the other local runtimes are great at the desk. My machine makes the same computer reachable from your phone or any other computer, without port forwarding, a VPN app or a public URL. It is private, it is free, and it works whether or not you sell anything on GarageAI."
commands:
  - label: "Connect the computer: macOS and Linux"
    command: "curl -fsSL https://raw.githubusercontent.com/magnusfroste/garageai/main/cli/install.sh | sh -s -- connect"
  - label: "Connect the computer: Windows, in PowerShell"
    command: "irm https://raw.githubusercontent.com/magnusfroste/garageai/main/cli/install.ps1 | iex; garageai connect"
  - label: "Then call it from anywhere (Python, OpenAI SDK)"
    command: |-
      from openai import OpenAI

      client = OpenAI(base_url="https://api.garageai.eu/v1", api_key="YOUR_GARAGEAI_KEY")
      reply = client.chat.completions.create(
          model="mine/<your-garage>/<model>",
          messages=[{"role": "user", "content": "Hi from my phone"}],
      )
      print(reply.choices[0].message.content)
    note: "Your machines' model names are listed in the portal and in /v1/models, for you only."
ctas:
  - text: "Connect your computer"
    url: "https://app.garageai.eu/connect"
  - text: "Is it safe?"
    url: "/security"
faqs:
  - question: "Does it cost anything to use my own machine?"
    answer: "No. Requests to your own garage are free, even with no credits on your account, and they never count as earnings."
  - question: "Can other people see or use my machine?"
    answer: "Not through My machine. Your mine/ models are hidden from everyone else and the gateway refuses anyone else's key. If you choose Me and the market in My garages, buyers can also use the models you offer for sale, at your price; choose Only me to keep the machine off the market entirely."
  - question: "What if my computer is asleep or switched off?"
    answer: "The chat shows the machine's state next to its models: awake, asleep since a given time, or not answering. When it wakes up and Bridge's tunnel is back, it answers again."
  - question: "Do I need Tailscale, a VPN or port forwarding?"
    answer: "No. GarageAI Bridge opens one outgoing, encrypted WireGuard tunnel to GarageAI's gateway; you open no ports and install no VPN app. If you already use Tailscale, Bridge runs next to it."
  - question: "Which apps can I use?"
    answer: "The GarageAI portal's chat in your phone's browser, or any app or tool that speaks the OpenAI API: point it at https://api.garageai.eu/v1 with your GarageAI key and use your mine/<garage>/<model> name."
related:
  - text: "Earn with your GPU while you are not using it"
    url: "/bridge"
  - text: "What Bridge opens on your computer, and what it does not"
    url: "/security"
  - text: "The OpenAI-compatible API"
    url: "/api"
---

## What you get

- **In the chat:** your garage's models come first in the GarageAI portal's model picker, under *Your machines*, with the machine's state: *awake*, *asleep since 14:02* or *not answering*. Open the portal in your phone's browser and talk to your own GPU.
- **In the API:** the same models answer as `mine/<garage>/<model>` with any of your keys, from any OpenAI-compatible app, script or agent.
- **Private and free:** nobody else can see or call your `mine/` models, and using them costs nothing.

## How to set it up

1. Have a runtime running with a model: Ollama, LM Studio, llama.cpp, vLLM or Xyntetik Runner. It can stay on localhost.
2. Run the connect line for your system and approve the computer in the browser.
3. In *My garages*, choose who can use the machine: **Only me** (it stays off the market) or **Me and the market** (it also earns from buyers).

## Your machine first

If you also sell on GarageAI, you still come first: while you use your own machine, new requests from buyers go to other garages, and come back a few minutes after your last message. A buyer's long request never keeps you waiting at your own computer.

## Compared with doing it yourself

| | What it means |
|---|---|
| Port forwarding on your router | Your runtime is open to the whole internet; Ollama has no password |
| A public tunnel URL (ngrok and similar) | Anyone with the link can call it; free addresses change |
| A personal VPN (Tailscale and similar) | Works well for your own devices, but every phone and computer needs the app and an account |
| GarageAI My machine | One outgoing tunnel, no public address, only you can call your models, from any browser or OpenAI-compatible app |
