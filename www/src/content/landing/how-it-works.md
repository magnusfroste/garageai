---
section: "how-it-works"
order: 2
title: "How It Works"
subtitle: "Two sides. One marketplace."
description: "Buyers get AI from GPUs in European garages through one API. Garage owners turn hardware they already have into a paid inference node. One account can do both, and can own several garages."
tracks:
  - id: "buyers"
    icon: "⚡"
    title: "Use AI"
    tagline: "For developers, companies and anyone who wants local AI"
    color: "var(--color-accent)"
    steps:
      - title: "Sign up and add credits"
        text: "Create an account at app.garageai.eu and top up prepaid credits by card (Stripe). You pay per token, with no subscription."
      - title: "Create an API key"
        text: "Generate a key in the portal. The API at https://api.garageai.eu/v1 is OpenAI-compatible, so the OpenAI SDKs, agents and coding tools work by changing the base URL."
      - title: "Pick a model and how to buy it"
        text: "Choose a Specific garage (you pay the price it has set, for example when location matters) or the Pool (one model name; each request goes to one of the garages offering the model and is billed at that garage's price, with automatic failover). Or just use the chat in the portal."
      - title: "Call it"
        text: "Point any OpenAI client at https://api.garageai.eu/v1. garageai/auto answers with a working model for you; any model name from the catalogue works too."
        commands:
          - label: "curl"
            command: |-
              curl https://api.garageai.eu/v1/chat/completions \
                -H "Authorization: Bearer $GARAGEAI_API_KEY" \
                -H "Content-Type: application/json" \
                -d '{"model": "garageai/auto", "messages": [{"role": "user", "content": "Hello from a European garage"}]}'
          - label: "Python (OpenAI SDK)"
            command: |-
              from openai import OpenAI

              client = OpenAI(base_url="https://api.garageai.eu/v1", api_key="YOUR_GARAGEAI_KEY")
              reply = client.chat.completions.create(
                  model="garageai/auto",
                  messages=[{"role": "user", "content": "Hello from a European garage"}],
              )
              print(reply.choices[0].message.content)
            note: "Agents and coding tools work the same way: set the base URL and your GarageAI key."
    ctaText: "Use AI"
    ctaUrl: "https://app.garageai.eu/auth?intent=buyer"
  - id: "operators"
    icon: "🏠"
    title: "Offer Your GPU"
    tagline: "For anyone with a capable GPU in a garage, office or server room"
    color: "var(--color-primary)"
    requirements: "A Mac (Apple silicon), a Linux machine or a Windows PC with a GPU, and Ollama, LM Studio, vLLM, llama.cpp or Xyntetik Runner running with a model loaded. No runtime yet? garageai runtime checks what fits your machine and installs Xyntetik Runner with a model."
    steps:
      - title: "Create an account"
        text: "Sign up at app.garageai.eu. Connect your GPU takes you to the page that connects your computer."
      - title: "Run one line"
        text: "Paste the line for your system on the machine with the GPU. It installs GarageAI Bridge, a small open-source program, which looks at your GPU, memory, runtime and models. It asks for your password once."
        commands:
          - label: "macOS and Linux, in the terminal"
            command: "curl -fsSL https://raw.githubusercontent.com/magnusfroste/garageai/main/cli/install.sh | sh -s -- connect"
          - label: "Windows, in PowerShell"
            command: "irm https://raw.githubusercontent.com/magnusfroste/garageai/main/cli/install.ps1 | iex; garageai connect"
            note: "Nothing secret in either line: no token or key. On Windows, Bridge asks for Administrator once."
      - title: "Approve your computer in the browser"
        text: "Bridge opens app.garageai.eu with a code. Check that it is your computer, pick a name for your garage and approve with one click. Bridge then opens its encrypted tunnel to the gateway by itself. You open no ports, and your runtime can stay on localhost."
      - title: "Go live"
        text: "Your models pass an acceptance test with a real streamed request, and the page shows \"Your garage is live\" with tokens per second for each model. It takes a few minutes. Set your price per million tokens in the portal; you earn it for every token you serve, also through the pool. No platform fee during launch."
    helpText: "No runtime yet, or not sure which model fits your GPU? Pick your setup ↓"
    helpUrl: "#pick-your-setup"
    ctaText: "Connect your GPU"
    ctaUrl: "https://app.garageai.eu/connect"
    noteDetail: "Curious what Bridge does? It is open source, one binary, and garageai doctor shows what is wrong and how to fix it."
    noteLinkText: "Bridge on GitHub"
    noteLinkUrl: "https://github.com/magnusfroste/garageai/blob/main/cli/README.md"
whyTitle: "Why GarageAI"
whyDescription: "The principle hasn't changed. AI capacity can be built from hardware Europe already has, close to the people who use it."
whyItems:
  - icon: "🇪🇺"
    title: "Sovereign"
    text: "Inference runs on hardware owned by people and companies in Europe, routed through a gateway in the EU, instead of only in hyperscaler data centres."
    color: "var(--color-accent)"
  - icon: "📍"
    title: "Local"
    text: "Garages are named. When location matters, buyers can choose a specific garage instead of the pool."
    color: "var(--color-primary)"
  - icon: "🔓"
    title: "Open source"
    text: "Built on WireGuard and LiteLLM. The gateway setup and GarageAI Bridge, the program that connects a garage, are in the GitHub repository for anyone to read."
    color: "var(--color-warning)"
  - icon: "💶"
    title: "Get paid for your hardware"
    text: "A GPU that sits idle earns nothing. On GarageAI it earns the price you set for every token it serves."
    color: "var(--color-tertiary)"
---

Content for this section is defined in the frontmatter above.
