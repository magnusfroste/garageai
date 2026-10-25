## In short

- Liteit AB runs GarageAI and is responsible for the personal data described here.
- Prompts go from you to the EU gateway and straight on to the garage, also in the portal's chat;
  we do not store them or the answers. If you turn on "Save my chats to my account", your chat
  conversations are kept in your account so you can use them on other devices.
- We keep what we need to run accounts, bill per token, pay operators and keep the network
  healthy: account details, usage records (including the IP address a request came from, kept 30
  days), payments, and for operators what Bridge reports about their machine.
- We do not sell personal data and do not use it for advertising.

## 1. Who is responsible

Liteit AB, org.nr 556616-1658, Kaplansbacken 3, 112 24 Stockholm, Sweden, is the controller. Contact for privacy questions:
privacy@garageai.eu.

## 2. What we process, why, and on what basis

| What | Whose | Why | Legal basis | Kept |
|---|---|---|---|---|
| E-mail, name, password (hashed), sign-in times | Everyone | Your account | Contract | While the account exists, then 30 days |
| API keys and key names | Buyers | Access to the API; the portal keeps the key so you can use it in the chat | Contract | Until you delete the key or the account |
| Usage records: time, model, route, garage, tokens, cost, duration, status, the key used, and the IP address the request came from | Buyers, and operators using their own machine through My machine (recorded at a cost of 0) | Billing, the request log in the portal, abuse and fraud prevention, capacity planning | Contract; legitimate interest (security) | 30 days in the gateway; daily totals without IP for 7 years (bookkeeping) |
| Prompts and answers sent to the API | Buyers | Passed through to the operator's machine and back | Contract | **Not stored** by GarageAI |
| Conversations in the portal's chat, when you turn on "Save my chats to my account" | Buyers who choose it | Showing your chat history on other devices. Without it, the history stays in your browser only. Chat messages go from your browser straight to the gateway either way | Consent (the setting) | Until you delete the conversation, turn the setting off and choose Delete, or close the account |
| Web-search queries from the portal's chat | Buyers who turn on web search | The search query only (never the conversation) passes through the portal to Liteit's search server | Contract | Not stored |
| Payments: amount, currency, time, Stripe customer and payment ids, receipt details | Buyers | Credits, receipts, bookkeeping | Contract; legal obligation (bookkeeping) | 7 years (Swedish bookkeeping law) |
| Garage details: garage name, prices, offered models, terms acceptance | Operators | Running the marketplace | Contract | While the garage exists |
| Bridge's machine reports: OS and version, CPU and GPU model, memory, disk, power source and battery level, GPU temperature and power draw, energy used while serving (watt-hours and tokens per kWh, per model), whether the machine is in use or paused, runtime and models, request counts, Bridge version, a machine identifier, problems found by `garageai doctor` | Operators | Routing, health checks, alerts, support, updates | Contract; legitimate interest (a working network) | Only the latest report, replaced every 5 minutes; deleted with the garage. Per-minute health samples (up, queue, speed) 30 days |
| The public IP address and port the operator's tunnel connects from | Operators | Keeping the encrypted tunnel up | Contract | While the tunnel is up, in 30 days of operations data |
| Onboarding events: the steps of `garageai connect` and what went wrong | Operators | Helping operators get live, improving onboarding | Legitimate interest | 12 months |
| Earnings and, when payouts start, payout and tax details | Operators | Payouts; reporting to the tax authority where required | Contract; legal obligation | 7 years |
| Messages you send us | Everyone | Support | Legitimate interest | 2 years |

Bridge never reports prompts, answers, API keys, passwords or command lines.

## 3. Who else processes it

We use these providers; they process data on our behalf under data-processing agreements, or as
independent controllers where noted:

| Provider | What for | Where |
|---|---|---|
| Hetzner Online | The API gateway, usage records, operations data, encrypted backups | Helsinki, Finland (EU) |
| Lovable Cloud (Supabase on AWS) | Accounts, the portal's database, chat conversations you chose to save | Stockholm, Sweden (EU); Amazon and Supabase are US companies |
| Lovable | Hosting the portal (app.garageai.eu) | EU/US |
| Vercel | Hosting the website (garageai.eu) | EU/US |
| Stripe | Card payments (Stripe is an independent controller for card data; we never see card numbers) | EU/US |
| Lovable Cloud | Sign-in and service e-mails sent by the portal | EU/US |
| Cloudflare (Email Routing) | Receiving e-mail to @garageai.eu and forwarding it | EU/US |
| Google (Gmail) | The inbox where e-mail to @garageai.eu is read and answered | EU/US |
| GitHub | Bridge downloads (GitHub sees the downloading IP address) | US |

**Operators** receive your prompts in order to answer them. They are independent and have agreed
not to log, store, read, forward or analyse prompts or answers; they do not receive your name,
e-mail or IP address. Until verified garages with a data-processing agreement exist, do not send
personal data in prompts (see the terms).

Where a provider is outside the EU/EEA, transfers rely on the EU-US Data Privacy Framework or the
EU's standard contractual clauses.

## 4. Your rights

You can ask to see the personal data we have about you, to correct it, to have it deleted, to
restrict or object to processing based on legitimate interest, and to receive your data in a
machine-readable form. Write to privacy@garageai.eu. Deleting your account deletes your account
data and keys; records we must keep for bookkeeping stay for the time the law requires.

You can complain to the Swedish Authority for Privacy Protection (IMY), imy.se.

## 5. Security

Traffic is encrypted: TLS between you and the gateway, WireGuard between the gateway and the
garage. Passwords are stored hashed; API keys are kept in the portal's database, readable only by
you and the portal's servers. Backups are encrypted. Access to production systems is limited to
the people who run GarageAI.

## 6. Cookies

The portal uses only the cookies and local storage needed to keep you signed in and to remember
settings such as language. The website garageai.eu sets no cookies and has no tracking. No advertising
or cross-site tracking.

## 7. Changes

We publish changes here and tell account holders by e-mail before significant changes apply.
