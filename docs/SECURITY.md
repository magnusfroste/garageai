# Security Policy

## Reporting a vulnerability

Please **do not** open a public issue for security problems.

Report it privately through GitHub: go to the repository's **Security** tab and choose
**[Report a vulnerability](https://github.com/magnusfroste/garageai/security/advisories/new)**.
Include what you found, how to reproduce it, and the impact you expect.

You will get an answer as soon as possible, and credit when the fix ships unless you prefer otherwise.

## Scope

- The connect script ([`scripts/garageai-connect.sh`](../scripts/garageai-connect.sh)) and the heartbeat it installs
- The gateway setup in [`infra/gateway/`](../infra/gateway/README.md)
- The site in [`src/`](../src)

## Design notes

- Garages never expose their runtime to the public internet; only the gateway reaches them, over the WireGuard mesh.
- The connect script never needs the gateway's LiteLLM master key.
- Operators run the [security checklist](../infra/gateway/README.md#security-checklist) on the gateway after every upgrade, after adding a provider, and before handing out keys to external testers.
