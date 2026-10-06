---
section: "token-economy"
title: "Pricing"
subtitle: "Every garage sets its own price."
description: "GarageAI has no price list of its own. Each garage sets a price per million tokens, input and output, for the models it offers. Buyers prepay credits and pay per token: the price of the garage that served the request. No subscriptions and no reserved-capacity contracts."
order: 5
buyOptions:
  - icon: "📍"
    title: "Specific garage"
    tag: "One named garage"
    color: "var(--color-accent)"
    points:
      - "You pay the price that garage has set"
      - "Choose it when it matters where the request runs"
      - "Shared capacity, not exclusive to you"
  - icon: "🌊"
    title: "Pool"
    tag: "One model name, many garages"
    color: "var(--color-primary)"
    points:
      - "Each request goes to one garage offering the model"
      - "You pay the price of the garage that served it"
      - "If a garage drops out, the request moves to another"
buyerNotes:
  - "The model catalogue shows each model's prices and the garages that offer it"
  - "Prepaid credits, topped up by card via Stripe"
  - "Usage is metered per API key and per model"
operatorPoints:
  - "You set your price per million tokens, input and output, when you create your garage"
  - "The same price applies when your garage serves requests through the pool"
  - "You earn your price for every token your garage serves. No platform fee during launch; a platform fee will be introduced later."
  - "Your earnings are tracked per token. Wallet and payouts are being built."
poolTitle: "How the pool works"
poolDescription: "The pool lets buyers use a model without picking a garage. Nobody sets a pool price: each request is billed at the price of the garage that answered it."
poolSteps:
  - title: "Only tested garages take part"
    text: "A garage joins the pool for a model once that model has passed the acceptance test through the gateway. Regular checks keep it there, and a garage that fails is taken out of rotation for a while."
  - title: "One garage answers each request"
    text: "The gateway sends each request to one of the healthy garages offering the model. If that garage fails mid-way, the request is retried on another one."
  - title: "Price and reliability decide the share"
    text: "Requests are spread across the healthy garages rather than all going to one. Garages with a good price and a solid track record get a larger share, but every healthy garage gets requests."
  - title: "You pay the garage that answered"
    text: "Because garages set their own prices, the cost of a pool request depends on which garage served it. The catalogue shows the range before you start."
poolNote: "Rolling out: price- and reliability-weighted sharing is being introduced. Until then, the pool favours the garage that responds fastest."
marketTitle: "The Bigger Picture"
marketDescription: "Why this matters beyond today's marketplace. Third-party market estimates, not GarageAI figures."
marketStats:
  - num: 106
    prefix: "$"
    suffix: "B"
    label: "AI inference market 2025¹"
    sub: "estimate"
    color: "var(--color-warning)"
  - num: 255
    prefix: "$"
    suffix: "B"
    label: "projected by 2030¹"
    sub: "estimate"
    color: "var(--color-primary)"
  - num: 19.2
    prefix: ""
    suffix: "%"
    label: "annual growth rate¹"
    sub: "CAGR 2025–2030, estimate"
    color: "var(--color-accent)"
whyAgents:
  - "Gartner expects 40% of enterprise apps to have embedded AI agents by end of 2026, up from 5% in Sept 2025²"
  - "Inference, not training, accounts for most AI compute usage (estimated 80–90%)³"
  - "Autonomous agents make many model calls per task, so demand per user compounds"
footnotes:
  - "¹ MarketsandMarkets: AI Inference Market Report 2025–2030"
  - "² Gartner, cited in Landbase: 39 Agentic AI Statistics 2026"
  - "³ MIT Technology Review, May 2025: \"Inferencing drives 80–90% of all AI compute\""
---

Content for this section is defined in the frontmatter above.
