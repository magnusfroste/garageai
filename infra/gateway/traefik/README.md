# Traefik: the gateway's reverse proxy

`docker-compose.yml` runs the one Traefik on the VPS (container `garageai-traefik`). It holds
:80 and :443, gets and renews the Let's Encrypt certificates (TLS challenge), and routes
api.garageai.eu, llm.garageai.eu and ops.garageai.eu to the containers that publish themselves
with Docker labels on the network `garageai_proxy` (LiteLLM and the Operations Center). The
certificates and the ACME account live in the volume `garageai_traefik_letsencrypt`.

Deployed at `/opt/garageai/traefik/docker-compose.yml`. Upgrade: change the image tag, then
`cd /opt/garageai/traefik && sudo docker compose pull && sudo docker compose up -d` at a quiet
time (a few seconds without HTTPS).

## Cut-over from the NetBird installer's Traefik (one time)

Until now the proxy was `netbird-traefik` from the NetBird installer (`/opt/garageai/netbird`,
network `netbird_netbird`, volume `netbird_netbird_traefik_letsencrypt`). This moves it to this
stack and then switches NetBird off. Every garage already runs on the Bridge tunnel hub
(`garageai-hub`, `wg-hub`), which this does not touch.

**Interruption:** from step 5 to step 7 buyers get no answers (about 1 to 2 minutes). Do it in a
quiet window (the Live page shows one) and add a note first:
`sudo garageai-ops-collect --note "Maintenance: Traefik moves out of the NetBird stack"`.

Run as a user with sudo on the VPS. Every command is meant to be run as written, in order.

### 1. Prepare

The gateway checkout must contain this change (merged to main and pulled):

```bash
REPO=/home/garageai/garageai
git -C "$REPO" pull --ff-only
test -f "$REPO/infra/gateway/traefik/docker-compose.yml" && echo ok
sudo systemctl start garageai-backup && sudo journalctl -u garageai-backup -n 3 --no-pager   # "backup ok"
```

Keep the files the cut-over replaces, for a rollback:

```bash
sudo cp -p /opt/garageai/litellm/docker-compose.yml /opt/garageai/litellm/docker-compose.yml.pre-traefik
sudo cp -p /opt/garageai/litellm/.env /opt/garageai/litellm/.env.pre-traefik
sudo cp -p /opt/garageai/ops/docker-compose.yml /opt/garageai/ops/docker-compose.yml.pre-traefik
```

### 2. Create the new stack without starting it

`up --no-start` creates the network `garageai_proxy`, the volume `garageai_traefik_letsencrypt`
and the container, but binds no port, so it does not collide with the old Traefik on :80/:443.

```bash
sudo install -d -m 0755 /opt/garageai/traefik
sudo install -m 0644 "$REPO/infra/gateway/traefik/docker-compose.yml" /opt/garageai/traefik/docker-compose.yml
cd /opt/garageai/traefik && sudo docker compose up --no-start
sudo docker network ls | grep garageai_proxy
sudo docker volume ls | grep garageai_traefik_letsencrypt
```

### 3. Copy the certificates

Traefik refuses an `acme.json` that is not mode 0600, so the copy keeps the mode and sets it
again:

```bash
sudo docker run --rm \
  -v netbird_netbird_traefik_letsencrypt:/from:ro \
  -v garageai_traefik_letsencrypt:/to \
  alpine:3 sh -c 'cp -p /from/acme.json /to/acme.json && chmod 600 /to/acme.json && ls -l /to/acme.json'
```

Expect `-rw-------` and a size of some kilobytes (not 0).

### 4. Install the new compose files for LiteLLM and the Operations Center

Nothing restarts yet.

```bash
sudo install -m 0644 "$REPO/infra/gateway/litellm/docker-compose.yml" /opt/garageai/litellm/docker-compose.yml
sudo sed -i 's/^NETBIRD_NETWORK=.*/PROXY_NETWORK=garageai_proxy/' /opt/garageai/litellm/.env
sudo grep -n 'NETWORK' /opt/garageai/litellm/.env          # exactly: PROXY_NETWORK=garageai_proxy
sudo install -m 0644 "$REPO/infra/gateway/ops/docker-compose.yml" /opt/garageai/ops/docker-compose.yml
cd /opt/garageai/litellm && sudo docker compose config --quiet && echo litellm-ok
cd /opt/garageai/ops && sudo docker compose config --quiet && echo ops-ok
```

### 5. Move LiteLLM and Ops to the new network (the interruption starts)

`up -d` recreates the two containers on `garageai_proxy`. The old Traefik can no longer reach
them from here on. Postgres is not recreated.

```bash
cd /opt/garageai/litellm && sudo docker compose up -d
cd /opt/garageai/ops && sudo docker compose up -d
sudo docker network inspect garageai_proxy --format '{{range .Containers}}{{.Name}} {{end}}'
# expect garageai-litellm-litellm-1 and garageai-ops-ops-1 (and garageai-traefik, created)
```

### 6. Stop the NetBird stack

This stops the NetBird server, dashboard and the old Traefik, and frees :80/:443. Its volumes
stay (`down` without `-v`).

```bash
cd /opt/garageai/netbird && sudo docker compose down
sudo ss -ltnp | grep -E ':(80|443) ' || echo "80/443 free"
```

### 7. Start the new Traefik (the interruption ends)

```bash
cd /opt/garageai/traefik && sudo docker compose up -d
sleep 10
curl -sI https://api.garageai.eu/health/liveliness | head -1     # HTTP/2 200
curl -sI https://llm.garageai.eu/health/liveliness | head -1     # HTTP/2 200
curl -s -o /dev/null -w '%{http_code}\n' https://ops.garageai.eu/   # the sign-in page, not 404/502
for h in api llm ops; do echo "== $h"; echo | openssl s_client -connect $h.garageai.eu:443 -servername $h.garageai.eu 2>/dev/null | openssl x509 -noout -issuer -dates; done
sudo docker logs garageai-traefik 2>&1 | grep -iE 'acme|error' | tail -n 20
```

The issuer is Let's Encrypt and `notAfter` is the same date as before the cut-over: the
certificates were copied, not issued again. A browser that gets a self-signed "TRAEFIK DEFAULT
CERT" means `acme.json` was not read (step 3).

Then one real request through a garage, with a buyer key:

```bash
curl -s https://api.garageai.eu/v1/chat/completions -H "Authorization: Bearer $BUYER" -H "Content-Type: application/json" \
  -d '{"model":"garageai/auto","max_tokens":20,"messages":[{"role":"user","content":"hi"}]}' | head -c 300; echo
```

**Rollback** (only if step 7 fails and cannot be fixed at once):

```bash
cd /opt/garageai/traefik && sudo docker compose down
sudo cp -p /opt/garageai/litellm/docker-compose.yml.pre-traefik /opt/garageai/litellm/docker-compose.yml
sudo cp -p /opt/garageai/litellm/.env.pre-traefik /opt/garageai/litellm/.env
sudo cp -p /opt/garageai/ops/docker-compose.yml.pre-traefik /opt/garageai/ops/docker-compose.yml
cd /opt/garageai/netbird && sudo docker compose up -d
cd /opt/garageai/litellm && sudo docker compose up -d
cd /opt/garageai/ops && sudo docker compose up -d
```

### 8. Deploy the gateway services that know the new names

The guard reads `garageai-traefik`'s log, the backup saves the new volume, and the health check
and the Operations Center no longer ask NetBird.

```bash
sudo install -m 0755 "$REPO/infra/gateway/guard/garageai-guard.py" /usr/local/sbin/garageai-guard
sudo install -m 0755 "$REPO/infra/gateway/backup/garageai-backup" /usr/local/sbin/garageai-backup
sudo install -m 0755 "$REPO/infra/gateway/health/garageai-health.py" /usr/local/bin/garageai-health
sudo install -m 0644 "$REPO/infra/gateway/health/garageai-health.service" /etc/systemd/system/garageai-health.service
sudo install -m 0755 "$REPO/infra/gateway/ops/garageai-ops-collect.py" /usr/local/bin/garageai-ops-collect
sudo install -m 0644 "$REPO/infra/gateway/ops/index.html" /var/lib/garageai-ops/www/index.html
sudo sed -i '/^NETBIRD_API_/d' /etc/garageai/gateway-health.env
sudo systemctl daemon-reload
sudo garageai-guard && sudo garageai-health | tail -n 1 && sudo garageai-ops-collect
sudo systemctl start garageai-backup && sudo journalctl -u garageai-backup -n 3 --no-pager   # "backup ok"
```

The Operations Center's Server page should show no new alert, and the Updates page Traefik
from `garageai-traefik`.

### 9. Switch off the NetBird client on the host

```bash
sudo netbird down || true
sudo systemctl disable --now netbird
sudo apt remove -y netbird
ip link show wt0 2>/dev/null && echo "wt0 still there" || echo "wt0 gone"
ip route | grep wt0 || echo "no wt0 routes"
```

Nothing on the gateway may still expect `wt0` or a route through it. Find and remove any such
check (a cron entry, a unit or a script from the NetBird time):

```bash
sudo grep -rln 'wt0' /etc/cron* /etc/systemd/system /usr/local/bin /usr/local/sbin 2>/dev/null
```

Remove each file or line it lists that checks `wt0`, then `sudo systemctl daemon-reload` if a
unit changed. In the Hetzner Cloud Firewall, delete the inbound rules for UDP 3478 and UDP
51820; keep TCP 80/443 and UDP 51830 (the hub).

### 10. Archive the NetBird directory (keep 30 days)

```bash
cd /opt/garageai && sudo tar -czf /var/backups/garageai/netbird-$(date -u +%Y%m%d).tgz netbird && sudo chmod 600 /var/backups/garageai/netbird-*.tgz
ls -l /var/backups/garageai/netbird-*.tgz
sudo rm -rf /opt/garageai/netbird
```

The nightly backup's clean-up only deletes `garageai-*` archives, so this one stays until it is
removed by hand: delete it, and the old Docker volumes, 30 days after the cut-over:

```bash
sudo rm /var/backups/garageai/netbird-*.tgz
sudo docker volume ls -q | grep '^netbird_' | xargs -r sudo docker volume rm
sudo docker network rm netbird_netbird 2>/dev/null || true
sudo rm -f /opt/garageai/litellm/*.pre-traefik /opt/garageai/ops/*.pre-traefik
```

### 11. DNS

The record `netbird.garageai.eu` is removed separately, in the DNS provider, after the cut-over.
Nothing on the gateway uses it any more.

`sudo garageai-ops-collect --note "Traefik moved to /opt/garageai/traefik; NetBird switched off"`
closes the window in the event log.
