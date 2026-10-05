---
section: "faq"
title: "Frequently Asked Questions"
description: "Practical answers about using GarageAI and offering your GPU."
order: 11
faqs:
  - question: "What is GarageAI?"
    answer: "A two-sided marketplace for local AI inference. Garage owners (operators) connect a GPU they already own and offer open models on it. Buyers call those models through one OpenAI-compatible API at https://llm.garageai.eu/v1, or use the chat in the portal, and pay per token. The operator earns the token price for the requests their garage serves."
  - question: "What hardware do I need to offer my GPU?"
    answer: "A machine that can run an open model with acceptable speed: typically a Mac with Apple silicon or a Linux machine with a capable GPU. There is no fixed minimum spec. Every model has to pass an automated acceptance test through the gateway, which measures time-to-first-token and tokens per second, before it can be sold."
  - question: "Which operating systems and runtimes are supported?"
    answer: "macOS and Linux. Windows is not supported yet. Supported runtimes are Ollama, LM Studio, llama.cpp, vLLM, SGLang and Paddock (beta). Unsloth, MLX on Apple Silicon and Lemonade on AMD also work. You keep your own OS and runtime; the onboarding wizard shows how to prepare the runtime, and one command installs the NetBird client, joins the encrypted mesh and registers your models."
  - question: "Do I need to open ports on my router?"
    answer: "No. Your machine joins a WireGuard mesh (self-hosted NetBird) with an outgoing connection. You open no inbound ports, garages cannot reach each other, and only the GarageAI gateway can reach your runtime."
  - question: "Is my data safe?"
    answer: "Traffic is encrypted end to end in transit over the WireGuard mesh, and the gateway runs on an EU VPS (Hetzner, Helsinki). But prompts are processed on the operator's machine, and operators are not yet formally verified. Verified garages and data-processing agreements are planned before we sell to buyers with sensitive data. Until then, don't send sensitive or personal data through the marketplace."
  - question: "How do operators get paid?"
    answer: "You earn the token price buyers pay for requests your garage serves, and there is no platform fee during launch (a platform fee will come later). Earnings are tracked per token. The wallet and payouts are being built and are coming next."
  - question: "What does it cost to use?"
    answer: "You pay per token with prepaid credits, topped up by card via Stripe. There is no subscription. Buying from the Pool is cheaper than buying from a specific garage."
  - question: "What is the difference between Pool and Specific garage?"
    answer: "Pool spreads your requests across every garage offering that model, with load balancing and automatic failover. Specific garage sends requests to a named garage you choose, for example for location. It is shared capacity, not exclusive; booked or reserved capacity is planned."
  - question: "Which models are available?"
    answer: "Open models that operators choose to serve and that pass the acceptance test. The current list is in the model catalogue in the portal at app.garageai.eu, and it changes as garages come and go."
  - question: "Is it OpenAI-compatible?"
    answer: "Yes. The API at https://llm.garageai.eu/v1 is OpenAI-compatible, so the OpenAI SDKs, agents and coding tools work by setting the base URL and your GarageAI API key."
  - question: "Can I be both a buyer and an operator?"
    answer: "Yes. One account can buy inference and offer GPUs, and you can register several garages under the same account."
  - question: "Is GarageAI open source?"
    answer: "The building blocks are open source (NetBird and LiteLLM), and GarageAI's gateway setup (infra/gateway/README.md) and connect script (scripts/garageai-connect.sh) are in the public GitHub repository under the MIT license."
  - question: "Can my company run private garages?"
    answer: "Yes, talk to us. Companies can use the API, offer their own GPUs on the marketplace, or run private garages. Contact powerup@garageai.eu about a pilot."
  - question: "Where do solar panels and EV batteries come in?"
    answer: "That is the long-term vision, not today's product. The idea is that garage nodes can eventually run largely on locally produced energy (rooftop solar, home batteries and EV batteries). Today any machine on the grid can join."
---

## FAQ Items

The questions and answers above are defined in this file's frontmatter (`faqs`)
and rendered by the FAQ section. Edit them there to change the content.
