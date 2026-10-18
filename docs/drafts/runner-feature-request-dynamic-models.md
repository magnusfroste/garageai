**Title:** Serve mode: add and remove models while running (dynamic model set), plus a download-only fetch

Hi Joakim! We run Runner as a managed engine in GarageAI Bridge (`garageai runtime install`,
https://github.com/magnusfroste/garageai/blob/main/docs/bridge.md). Bridge installs a pinned
release, verifies SHA256SUMS and runs `runner -hf MODEL --serve --port 8080 --ttl 0` as a
system service on Linux, macOS and Windows. The garage owner never opens a terminal after that.

**What already works for us:** swap mode with up to 16 models, lazy load on the first request,
`--ttl`, per-request `keep_alive`, `POST /unload`, `/v1/models` with residency, `-hf` with
SHA-256 verification, `runner_` metrics. Thank you, that is most of what we need.

**What we are missing:** the model set is fixed when the server starts. In GarageAI the owner
(or our "Smart" mode, from buyer demand) picks models in a web portal, and Bridge fetches and
offers them, the way it does with Ollama's `/api/pull`. With Runner, adding a model today means
restarting the server, which drops requests in flight and the prefix cache.

**Proposal** (any shape that gives the same result is fine):

1. **Register and unregister models at runtime:** for example `POST /v1/runner/models`
   `{"name": "qwen3-4b", "hf": "owner/repo:TAG"}` or `{"name": ..., "path": ...}`, and
   `DELETE /v1/runner/models/{name}`. The model joins the swap set and loads on its first request,
   as the startup set does today. Alternatively, `--models-dir DIR` rescanned on a signal or an
   endpoint.
2. **Fetch without serving:** `runner --pull owner/repo:TAG` (download into the HF cache, verify,
   exit), or an async `POST /v1/runner/pull` with progress in a GET. Bridge reports download
   progress to the portal ("2.1 of 4.7 GB").
3. **Optional, later:** keep more than one model resident when they fit (LRU eviction under a
   memory budget), so two buyers on two small models do not take turns loading.

**Smaller asks:**

- A `linux-arm64` release binary (DGX Spark / GB10 and ARM servers). Bridge already serves those
  machines with other engines, and would offer Runner there too.
- A stable channel or tag that Bridge can follow, next to the frequent releases.

Happy to test pre-releases on our garages: we have an M1, an M2, an M4, a DGX Spark, a Windows
VM and a Windows laptop on the network, and a nightly chain test that exercises fetch → serve →
request → remove end to end.

— Magnus Froste, GarageAI
