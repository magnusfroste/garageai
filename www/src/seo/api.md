---
path: "/api"
title: "OpenAI-Compatible API for Open Models in Europe — GarageAI"
description: "Call open models on GPUs in European garages through one OpenAI-compatible API. Change the base URL, pay per token, no subscription. EU gateway."
kicker: "GarageAI API · for developers"
breadcrumb: "OpenAI-compatible API"
h1: "One OpenAI-compatible API for open models, served from European garages"
lead: "GarageAI's API at https://api.garageai.eu/v1 speaks the OpenAI format, so the OpenAI SDKs, agents and coding tools work by changing the base URL. The models run on GPUs that garage owners connect; you pay per token with prepaid credits, at the price the garage that answered has set."
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
  - label: "JavaScript / TypeScript (OpenAI SDK)"
    command: |-
      import OpenAI from "openai";

      const client = new OpenAI({ baseURL: "https://api.garageai.eu/v1", apiKey: process.env.GARAGEAI_API_KEY });
      const reply = await client.chat.completions.create({
        model: "garageai/auto",
        messages: [{ role: "user", content: "Hello from a European garage" }],
      });
      console.log(reply.choices[0].message.content);
    note: "Create a key at app.garageai.eu. Streaming works the same way as with OpenAI."
ctas:
  - text: "Get an API key"
    url: "https://app.garageai.eu/auth?intent=buyer"
  - text: "See the live models and prices"
    url: "https://app.garageai.eu/models"
faqs:
  - question: "Is the GarageAI API compatible with the OpenAI API?"
    answer: "Yes. It accepts OpenAI-format chat completions at https://api.garageai.eu/v1, including streaming, so the OpenAI SDKs for Python and JavaScript, agents and coding tools work by setting the base URL and your GarageAI key."
  - question: "Which models can I use?"
    answer: "Open models that garage owners serve and that passed the acceptance test; the catalogue in the portal lists them with their prices and the garages that offer them. garageai/auto answers with a working model and names it in the response."
  - question: "Where is my data processed?"
    answer: "Your request goes over TLS to GarageAI's gateway on an EU server (Hetzner, Helsinki), which stores no prompts or answers, and then over an encrypted WireGuard tunnel to the garage that serves the model, where it is processed. Garage owners agree not to log or read prompts, but garages are not yet formally verified, so don't send sensitive or personal data until verified garages and data-processing agreements are in place."
  - question: "How does billing work?"
    answer: "Prepaid credits, topped up by card through Stripe, with no subscription. Each request is billed per token at the price of the garage that served it; cached input is billed at that garage's lower cache-read price. Usage is shown per key and per model in the portal."
  - question: "What happens when a garage goes offline?"
    answer: "In a pool, the request is retried on another garage. If the only garage serving a model is away, the API answers 503 with Retry-After, which SDKs and agents retry, rather than model not found."
related:
  - text: "The full API reference (in the portal)"
    url: "https://app.garageai.eu/api"
  - text: "The live model catalogue, and the garages behind it"
    url: "https://app.garageai.eu/models"
  - text: "Pricing: every garage sets its own price"
    url: "/#pricing"
  - text: "Earn with your own GPU instead"
    url: "/bridge"
  - text: "Privacy policy"
    url: "/privacy"
---

## Models and routing

- **`garageai/auto`** answers with a working model for you, and the response names the model that answered.
- **Pool:** one model name across every garage that offers the model. Each request goes to one healthy garage and is billed at that garage's price; if it fails, the request is retried on another one. Requests are spread across the healthy garages, not only to the fastest.
- **Specific garage:** a named garage you choose, for example when location matters. You pay the price it has set.

Every model passes an acceptance test with a real streamed request before it is sold, and regular checks keep it in rotation.

## Pricing

GarageAI has no price list of its own. Each garage sets a price per million tokens, input and output, plus a lower cache-read price for cached input, and you pay the price of the garage that served the request. Credits are prepaid by card; there is no subscription and no reserved-capacity contract. The model catalogue in the portal shows each model's prices before you start.

## Data and privacy

| | |
|---|---|
| In transit | TLS from you to the gateway; WireGuard from the gateway to the garage |
| Gateway | An EU server (Hetzner, Helsinki); stores no prompts or answers |
| Processing | On the garage owner's machine, under terms that forbid logging or reading prompts |
| Not yet | Formally verified garages and data-processing agreements |

The [privacy policy](/privacy) lists everything GarageAI keeps. Until verified garages and data-processing agreements are in place, don't send sensitive or personal data.

## Reliability

Requests in a pool fail over to another garage. A garage that fails health checks is taken out of rotation for a while. When the only garage serving a model is away, the API answers `503` with `Retry-After` instead of an error that makes agents give up.
