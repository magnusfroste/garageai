# GarageAI gateway: NetBird mesh + LiteLLM

This is how garage nodes get connected to the GarageAI gateway without exposing
anything to the public internet, and without US-run tunnel services.

```
 Buyers ──HTTPS──▶ [Hetzner VM 1: Easypanel]                [Hetzner VM 2]
                    Traefik :443 → LiteLLM ─┐                NetBird server
                    NetBird client (host)   │                (management, signal,
                                            │                 relay, dashboard)
                     private WireGuard mesh │                        ▲
              ┌─────────────┬───────────────┘        coordination    │
          garage-lund   garage-malmo   garage-…   ◀───────────────────┘
          (Ollama /     (llama.cpp /   (Paddock /
           LM Studio)    vLLM)          Unsloth)
```

- **Data plane** (prompts and tokens): buyer → LiteLLM on Hetzner → WireGuard mesh → garage.
  Encrypted end to end. Garages open no inbound ports.
- **Control plane**: NetBird server, self-hosted on Hetzner. NetBird is fully open
  source (BSD-3 clients, AGPLv3 server) and made by NetBird GmbH, Berlin.

Two VMs, because Easypanel's Traefik already owns ports 80/443 on VM 1 and the NetBird
server needs 80/443 plus UDP 3478 for itself. A CX22 (~€4/month) is plenty for VM 2.

## 1. NetBird server (VM 2)

1. Point a DNS record at VM 2, e.g. `netbird.garageai.eu`.
2. Open TCP 80, TCP 443 and UDP 3478 in the Hetzner firewall.
3. Run the official installer, which generates the compose file, config and the
   embedded identity provider:

   ```bash
   curl -fsSL https://github.com/netbirdio/netbird/releases/latest/download/getting-started.sh | bash
   ```

4. Log in to the dashboard at `https://netbird.garageai.eu`.

## 2. Groups, access policy and setup keys (NetBird dashboard)

1. **Groups:** create `gateway` and `garages`.
2. **Access control — important:** delete the *Default* policy. It lets every peer
   reach every other peer, so garages could reach each other. Add one policy instead:

   | Source    | Destination | Protocol | Ports |
   |-----------|-------------|----------|-------|
   | `gateway` | `garages`   | TCP      | 11434, 1234, 8080, 8000 (+ any custom runtime port) |

   Garages can now only be reached by the gateway, and only on runtime ports.
3. **Setup keys:**
   - one *one-off* key with auto-group `gateway`, for VM 1;
   - one key per operator with auto-group `garages`. Make it one-off or give it a
     short expiry, so a leaked key cannot enrol extra machines. Revoke it in the
     dashboard to cut an operator off.

## 3. Join the Easypanel host to the mesh (VM 1)

Install the NetBird client on the **host**, not as an Easypanel app. The host gets a
route into the mesh, and containers, including LiteLLM, reach garages through it.

```bash
curl -fsSL https://pkgs.netbird.io/install.sh | sh
sudo env NB_SETUP_KEY=<gateway-setup-key> netbird up --management-url https://netbird.garageai.eu
netbird status        # note the mesh IP; the peer should be in group "gateway"
```

Check that the LiteLLM container can reach a garage once one is connected:

```bash
docker ps --format '{{.ID}} {{.Names}}' | grep -i litellm
docker exec <container-id> python -c \
  "import urllib.request; print(urllib.request.urlopen('http://<garage-mesh-ip>:11434/v1/models', timeout=5).read()[:200])"
```

If that times out, but the same request works from the host itself, check that the
host has a mesh route (`ip route | grep wt0`) and that the access policy in step 2
includes the port.

## 4. LiteLLM settings (Easypanel → your LiteLLM app → Environment)

```
STORE_MODEL_IN_DB=True          # required: lets register-node.sh add/remove models at runtime
DATABASE_URL=postgresql://...   # required by STORE_MODEL_IN_DB
LITELLM_MASTER_KEY=sk-...       # admin key; never give it to garage operators
LITELLM_SALT_KEY=sk-...         # encrypts stored credentials; cannot be rotated later
```

Residential nodes drop out more often than datacenter ones, so let the router retry
and cool down failing garages quickly (in LiteLLM's `config.yaml`):

```yaml
router_settings:
  routing_strategy: latency-based-routing   # prefer the garage that answers fastest
  num_retries: 2                            # retry on another garage in the pool
  allowed_fails: 2                          # failures before a garage is cooled down
  cooldown_time: 60                         # seconds out of rotation
```

Streaming (SSE) passes through Easypanel's Traefik without extra configuration.

## 5. Onboard a garage

**Operator** (in their garage), after starting their runtime:

```bash
curl -fsSLO https://raw.githubusercontent.com/magnusfroste/garageai/main/scripts/garageai-connect.sh
bash garageai-connect.sh --setup-key <their-key> --management-url https://netbird.garageai.eu --runtime ollama
```

The script joins the mesh, checks that the runtime answers on the mesh IP, explains
how to rebind it if it only listens on localhost, and prints the node details.

**You** (anywhere that can reach LiteLLM's admin API):

```bash
export LITELLM_URL=https://llm.garageai.eu LITELLM_MASTER_KEY=sk-...
infra/gateway/register-node.sh add garage-lund 100.92.1.7 11434 qwen3:32b gemma4:27b
infra/gateway/register-node.sh list
infra/gateway/register-node.sh remove garage-lund qwen3:32b     # when they stop sharing
```

## 6. The two marketplace tiers

`register-node.sh` registers every model twice:

| Model name                     | Routes to                          | Sold as |
|--------------------------------|------------------------------------|---------|
| `garage/garage-lund/qwen3:32b` | that one garage only               | Booked, single-tenant access to a named garage |
| `qwen3:32b`                    | any garage offering the model      | Cheaper pool access, load-balanced |

Give a buyer a LiteLLM virtual key limited to what they bought:

```bash
curl -X POST "$LITELLM_URL/key/generate" \
  -H "Authorization: Bearer $LITELLM_MASTER_KEY" -H "Content-Type: application/json" \
  -d '{"models": ["garage/garage-lund/qwen3:32b"], "max_budget": 25, "duration": "30d",
       "metadata": {"buyer": "acme-ab", "booking": "b_123"}}'
```

LiteLLM's spend tracking per key and per model is the basis for operator payouts.

## Security notes

- The LiteLLM master key stays on the gateway side. Garage nodes never see it;
  `garageai-connect.sh` does not need it.
- Prompts are processed on the operator's machine. Mesh encryption protects them in
  transit, not on the garage itself. Verified operators and a data processing
  agreement are needed before selling to buyers with sensitive data.
- Next step: replace the manual `register-node.sh` call with a registration endpoint
  (it keeps the master key server-side) and pass its URL and a per-node token to
  `garageai-connect.sh --register-url ... --register-token ...`.
