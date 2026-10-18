# GarageAI privacy policy (DRAFT for legal review, 2026-10-17)

> Draft by Claude for Magnus Froste. Not legal advice. Facts checked against the running system on
> 2026-10-17 (gateway configuration and database, Bridge's reports); bracketed items are open.
> Plain language, one page, as the terms.

## In short

- Liteit AB runs GarageAI and is responsible for the personal data described here.
- We do not store the content of your prompts or the answers.
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
| E-mail, name, password (hashed), sign-in times | Everyone | Your account | Contract | While the account exists, then [30 days] |
| API keys (stored hashed), key names | Buyers | Access to the API | Contract | Until you delete the key or the account |
| Usage records: time, model, route, garage, tokens, cost, duration, status, the key used, and the IP address the request came from | Buyers | Billing, the request log in the portal, abuse and fraud prevention, capacity planning | Contract; legitimate interest (security) | 30 days in the gateway; daily totals without IP for [the accounting period] |
| Prompts and answers | Buyers | Passed through to the operator's machine and back | Contract | **Not stored** by GarageAI |
| Payments: amount, currency, time, Stripe customer and payment ids, receipt details | Buyers | Credits, receipts, bookkeeping | Contract; legal obligation (bookkeeping) | 7 years (Swedish bookkeeping law) |
| Garage details: garage name, prices, offered models, terms acceptance | Operators | Running the marketplace | Contract | While the garage exists |
| Bridge's machine reports: OS and version, CPU and GPU model, memory, disk, power source and battery level, whether the machine is in use, runtime and models, request counts, Bridge version, a machine identifier, problems found by `garageai doctor` | Operators | Routing, health checks, alerts, support, updates | Contract; legitimate interest (a working network) | Only the latest report, replaced every 5 minutes; deleted with the garage. Per-minute health samples (up, queue, speed) 30 days |
| The public IP address and port the operator's tunnel connects from | Operators | Keeping the encrypted tunnel up | Contract | While the tunnel is up, in [30 days] of operations data |
| Onboarding events: the steps of `garageai connect` and what went wrong | Operators | Helping operators get live, improving onboarding | Legitimate interest | 12 months |
| Earnings and, when payouts start, payout and tax details | Operators | Payouts; reporting to the tax authority where required | Contract; legal obligation | 7 years |
| Messages you send us | Everyone | Support | Legitimate interest | [2 years] |

Bridge never reports prompts, answers, API keys, passwords or command lines.

## 3. Who else processes it

We use these providers; they process data on our behalf under data-processing agreements, or as
independent controllers where noted:

| Provider | What for | Where |
|---|---|---|
| Hetzner Online | The API gateway, usage records, operations data, encrypted backups | Helsinki, Finland (EU) |
| Lovable Cloud (Supabase) | Accounts, the portal's database | [region] |
| Lovable | Hosting the portal (app.garageai.eu) | [region] |
| Vercel | Hosting the website (garageai.eu) | [EU/US] |
| Stripe | Card payments (Stripe is an independent controller for card data; we never see card numbers) | EU/US |
| Lovable Cloud | Sign-in and service e-mails sent by the portal | [region] |
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
garage. Keys are stored hashed. Backups are encrypted. Access to production systems is limited to
the people who run GarageAI.

## 6. Cookies

The portal uses only the cookies and local storage needed to keep you signed in and to remember
settings such as language. The website garageai.eu sets no cookies and has no tracking. No advertising
or cross-site tracking.

## 7. Changes

We publish changes here and tell account holders by e-mail before significant changes apply.

---

## Notes for the review (not part of the policy)

1. **Buyer IP in usage records.** LiteLLM stores `requester_ip_address` on every request (checked:
   all rows), kept 30 days (`maximum_spend_logs_retention_period`). Keep (abuse prevention) and
   disclose, or switch off if not needed.
2. **Prompts are not stored:** checked on 2026-10-17, `messages` and `response` are empty in every
   usage row. A test should keep it so (the gateway's configuration must not turn on prompt
   storage).
3. **Operators' machine reports** include power source, battery and "user active": personal data
   about a private person's home computer. Only the latest report is kept (garages.agent,
   overwritten every heartbeat); is legitimate interest the right basis?
4. **Roles for prompts.** GarageAI as controller for account and usage data; for prompts, GarageAI
   (and operators) only transmit and compute. If a buyer sends personal data anyway: processor
   (and operators sub-processors), or not? This decides the shape of the future DPA.
5. **Onboarding events are kept forever today** (263 rows since 2026-10-10, 47 with a full machine
   profile; no clean-up job). Add a scheduled delete after 12 months before publishing.
6. **Open facts to fill in:** Supabase/Lovable region, e-mail provider, whether the website sets
   cookies (Vercel analytics is not enabled in the site's code).
