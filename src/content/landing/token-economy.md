---
section: "token-economy"
title: "Pricing & Economics"
subtitle: "Pay per token. The garage earns it."
description: "No subscriptions and no reserved-capacity contracts. Buyers prepay credits and pay per token. The operator whose garage served the request earns that token price, and during launch GarageAI takes no platform fee."
order: 5
buyOptions:
  - icon: "🌊"
    title: "Pool"
    tag: "Cheaper"
    color: "var(--color-primary)"
    points:
      - "Load-balanced across every garage that offers the model"
      - "Automatic failover if a garage goes offline"
      - "The best default for most workloads"
  - icon: "📍"
    title: "Specific garage"
    tag: "Choose where it runs"
    color: "var(--color-accent)"
    points:
      - "Pick a named garage, for example for location"
      - "Shared capacity, not exclusive to you"
      - "Booked or reserved capacity is planned"
buyerNotes:
  - "Prepaid credits, topped up by card via Stripe"
  - "Usage is metered per API key and per model"
operatorPoints:
  - "You earn the token price buyers pay for requests your garage serves"
  - "No platform fee during launch. A platform fee will be introduced later."
  - "Your earnings are tracked per token"
  - "Wallet and payouts are being built and are coming next"
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
