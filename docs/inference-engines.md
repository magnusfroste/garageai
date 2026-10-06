# Local / self-hosted LLM inference engines — GarageAI support matrix

Research date: 2026-10-05. Every engine claim carries a source tag `[Sn]` that points to the list at the end.
Version numbers and dates come from each project's GitHub releases API on 2026-10-05 [S0] unless another source is given.
Where something could not be checked, the report says "unverified" and does not guess.

**What the gateway allows today:** TCP 11434, 1234, 8080 and 8000 into garages.
**What `garageai-connect.sh` knows today:** runtime ids `ollama|lmstudio|llamacpp|vllm|paddock|unsloth|other`. Only the first four have default ports. The probe is fixed to `GET http://<ip>:<port>/v1/models` and reads `.data[].id`. An optional `RUNTIME_API_KEY` is sent as a Bearer token.

---

## 1. Support matrix

Legend: Y = yes, N = no, ? = unverified. "Bind flag" is the exact way to listen on the mesh IP or on 0.0.0.0. A port in **bold** is not open on the gateway today.

| Engine (latest seen) | OpenAI API (/v1/models + chat SSE) | Bind flag | Default port | API key | Usage in stream | Concurrency | Linux / macOS / Win | License | GarageAI status |
|---|---|---|---|---|---|---|---|---|---|
| **Ollama** v0.40.0-rc3 | Y [S1] | `OLLAMA_HOST=<ip>:11434` (default 127.0.0.1) [S2] | 11434 [S2] | N on the local server [S3] | Y, `stream_options.include_usage` [S1] | `OLLAMA_NUM_PARALLEL` (default 1), `OLLAMA_MAX_QUEUE` 512 [S2] | Y/Y/Y | MIT [S0] | Supported now |
| **LM Studio** 0.4.x / llmster | Y [S4] | GUI "Serve on Local Network", or `lms server start --bind 0.0.0.0 --port 1234` [S5][S6] | 1234 [S5] | Y, "Require Authentication" toggle [S6] | Y since 0.3.18 [S7] | Continuous batching, "Max Concurrent Predictions" default 4 (llama.cpp engine ≥2.0) [S8] | Y/Y/Y; headless `llmster` / `lms daemon up` [S9] | Proprietary, free | Supported now (hint tweak) |
| **llama.cpp** `llama-server` b11401 | Y [S10] | `--host 0.0.0.0` (default 127.0.0.1) [S10] | 8080 [S10] | Y, `--api-key` / `LLAMA_API_KEY` [S10] | Y [S10] | `-np N` slots + continuous batching; router mode for several models [S10] | Y/Y/Y | MIT [S0] | Supported now (hint tweak) |
| **vLLM** v0.30.0 | Y [S11] | `--host 0.0.0.0` [S11] | 8000 [S11] | Y, `--api-key` (protects /v1 only) [S11] | Y [S12] | Continuous batching, `--tensor-parallel-size` [S11] | Y / via vllm-metal plugin [S13] / N | Apache-2.0 [S0] | Supported now |
| **Paddock** (Truespar) v0.1.13 | Y [S14][S15] | runner binds **0.0.0.0 by default**; `--host` / `PADDOCK_HOST` [S16] | **11540** (runner); manager/Studio 11500 on localhost [S16][S17] | Y, `--api-key` / `PADDOCK_API_KEY`; a key is **auto-generated and required** on a network bind [S16] | Y, always sends a final usage chunk [S18] | Continuous batching, `--max-batch`; single GPU only (no tensor parallelism yet) [S14] | Y/Y (pre-release, macOS 26+)/Y [S14] | MIT or Apache-2.0 [S14] | Supported with small change |
| **Unsloth Studio / `unsloth run`** | Y (built on llama-server) [S19][S20] | `-H 0.0.0.0 -p 8888` [S19][S21] | **8888** [S21] | Y, `sk-unsloth-…` keys created in the UI [S19] | ? (llama-server backend, not checked) | Comes from llama-server | Y/Y/Y [S19] | Core Apache-2.0, Studio AGPL-3.0 [S22] | Supported with small change |
| **SGLang** v0.5.21 | Y [S23] | `--host 0.0.0.0` (default 127.0.0.1) [S23] | **30000** [S23] | Y, `--api-key` [S23] | Y, `include_usage` + `continuous_usage_stats` [S24] | Continuous batching, `--tp N` [S23] | Y/N/N | Apache-2.0 [S0] | Supported with small change |
| **TensorRT-LLM** `trtllm-serve` | Y [S25] | `--host` [S25] | 8000 in the examples [S25] | ? | ? | Y, tp/pp/ep sizes [S25] | Y/N/N | Apache-2.0 | Supported now as `other` (expert users) |
| **NVIDIA NIM** (LLM) | Y [S26] | container, `-p 8000:8000` [S26] | 8000 (`NIM_SERVER_PORT`) [S26] | ? | ? | Y (vLLM/TRT-LLM inside) [S26] | Y/N/N | Developer Program: dev/test only, ≤16 GPUs; production needs NVIDIA AI Enterprise [S27] | Not suitable (license) |
| **HF TGI** v3.3.7 | Y | — | — | — | — | — | Linux | Apache-2.0 | Not suitable: in maintenance mode, repo now **archived** [S0][S28] |
| **LocalAI** v4.11.0 | Y [S29] | `LOCALAI_ADDRESS` [S29] | 8080 [S29] | Y, `LOCALAI_API_KEY` [S29] | ? | Backend-dependent; `LOCALAI_PARALLEL_REQUESTS`, `LLAMACPP_PARALLEL` [S30] | Y/Y/Y (Docker first) [S29] | MIT [S0] | Supported now as `other` |
| **Jan** v0.8.4 | Y [S31] | Host setting `0.0.0.0` in the GUI [S31] | **1337** [S31] | Y, optional [S31] | ? | ? | Y/Y/Y [S31] | ? (GitHub: NOASSERTION) [S0] | Supported with small change (low priority, desktop app) |
| **KoboldCpp** v1.122.1 | Y [S32] | `--host` [S32] | **5001** [S32] | Y, `--password` [S32] | Usage in non-stream replies; stream unverified [S33] | `--multiuser` queue (on by default) [S32] | Y/Y/Y | AGPL-3.0 [S0] | Supported with small change (low priority) |
| **text-generation-webui** v4.9 | Y [S34] | `--api --listen` [S34] | **5000** [S34] | Y, `--api-key` [S34] | ? | llama.cpp loader needs `--parallel N`; ExLlamaV3 native [S34] | Y/Y/Y | AGPL-3.0 [S0] | Waitlist (hobby UI, needs several flags) |
| **mlx-lm** `mlx_lm.server` v0.31.3 | Y [S35][S36] | `--host` (default 127.0.0.1) [S36] | 8080 [S36] | N (no key option in the server) [S36] | Y, `include_usage` [S36] | BatchGenerator batching [S36] | N/Y/N | MIT [S0] | Supported with small change (but "not recommended for production" [S35]) |
| **vllm-mlx** / **vllm-metal** | Y [S13][S37] | ? | ? | ? | ? | Continuous batching [S37] | N/Y/N | ? | Waitlist (check and add later) |
| **exo** v1.0.71 | chat Y; `/v1/models` unverified (README lists `/models`) [S38] | ? | **52415** [S38] | ? | ? | Model sharded across devices [S38] | Linux CPU only / macOS GPU [S38] | Apache-2.0 [S0] | Waitlist |
| **llamafile** 0.10.6 | Y [S39][S40] | `--server --host 0.0.0.0` [S40] | 8080 [S40] | ? (llama.cpp based) | ? | ? | Y/Y/Y (≤4 GB per file on Windows) [S39] | Apache-2.0 + MIT [S39] | Supported now as `llamacpp`/`other` |
| **Aphrodite → "Sonar"** v0.24.1 | Y [S41] | ? | **2242** (127.0.0.1) [S41] | ? | ? | vLLM-based [S41] | Y/Y (Metal)/N [S41] | AGPL-3.0 [S41] | Supported with small change (low priority) |
| **LMDeploy** v0.18.0 | Y [S42] | (`--server-name`, unverified) | **23333** [S42] | ? | ? | `--tp` [S42] | Y/N/? | Apache-2.0 [S0] | Supported with small change (low priority) |
| **Xinference** v3.5.0 | Y [S43] | `-H 0.0.0.0` [S43] | **9997** [S43] | N by default [S43] | ? | Backend-dependent (vLLM/SGLang/llama.cpp/MLX) [S43] | Y/Y/WSL [S43] | Apache-2.0 [S0] | Supported with small change |
| **GPUStack** v2.3.0rc2 | Y, under `/v1-openai` (older) / `/v1` (README) [S44][S45] | Docker, `-p 80:80` [S45] | **80** [S45] | Y, required [S44][S45] | ? | Cluster manager over vLLM/SGLang/TRT-LLM [S45] | Workers: Linux only [S45] | Apache-2.0 [S0] | Waitlist (it is a cluster manager) |
| **Lemonade** (AMD) v2026.40.0 | Y (`/api/v1` and `/v1`) [S46][S47] | `lemond --host 0.0.0.0`, `lemonade config set host=0.0.0.0`, `LEMONADE_HOST` [S47] | **13305** [S46][S47] | Y, `LEMONADE_API_KEY` [S47] | ? | `max_loaded_models` [S47] | Y/Y/Y; AMD NPU on Win+Linux [S46] | Apache-2.0 [S0] | Supported with small change |
| **Docker Model Runner** | Y, under `/engines/v1` [S48] | TCP `12434` [S48] | **12434** [S48] | ? | ? | llama.cpp or vLLM inside [S48] | Y/Y/Y | — | Waitlist (path is not `/v1`) |
| **Foundry Local** (Microsoft) | Y (optional web server) [S49] | ? | random by default; `foundry service set --port` [S49] | ? | ? | ? | N/Y/Y [S49] | — | Not suitable now (Windows/macOS only, random port) |

### Other notes per engine

| Engine | Model id in /v1/models = id clients send? | Tool calling | Reasoning output | Multi-GPU | Apple Silicon |
|---|---|---|---|---|---|
| Ollama | Y (tag such as `qwen3:8b`) [S1] | Y [S1] | Y ("Reasoning Output") [S1] | Y | Y |
| LM Studio | Y (JIT load on request) [S6] | Y | ? | ? | Y (MLX + GGUF) |
| llama.cpp | Y (loaded model) [S10] | Y, needs `--jinja` [S10] | `reasoning_content`, `--reasoning-format` [S10] | Y | Y (Metal) |
| vLLM | Y; `--served-model-name`, else the model path [S11] | Y, `--enable-auto-tool-choice` + parser [S11] | `reasoning` (≥0.9); `reasoning_content` still accepted [S12] | Y (tp/pp) [S11] | via vllm-metal [S13] |
| SGLang | Y | Y, `--tool-call-parser` [S23] | `reasoning_content` (`separate_reasoning`) [S24] | Y `--tp` [S23] | N |
| Paddock | Y | Y [S14] | `reasoning_content` deltas [S18] | N yet [S14] | Y (pre-release) [S14] |
| Unsloth | Y (loaded models) [S19] | Y ("self-healing tool calling") [S20] | ? | from llama-server | Y |
| mlx-lm | Lists models cached on the machine [S35] | Y | `reasoning` field [S36] | N | Y |
| Xinference | model **UID** (defaults to the model name) [S43] | ? | ? | Y | Y (MLX) |

**Usage reporting for billing.** These engines document usage in streamed replies: Ollama, LM Studio, llama.cpp, vLLM, SGLang, mlx-lm and Paddock. Paddock always sends the usage chunk. For every other engine it is unverified. The gateway should **not** trust that streams carry usage. Two options: always send `stream_options.include_usage=true` from LiteLLM and reject or flag garages whose streams come back without `usage` (a smoke test in the connect script), or count tokens at the gateway as a fallback.

---

## 2. Status per engine (detail)

**Supported now** (today's script and ports):
- **Ollama.** It has no API key, so security comes only from the mesh ACL. `OLLAMA_NUM_PARALLEL` defaults to 1 [S2], so the hint should suggest raising it to 4 or more for marketplace load.
- **LM Studio.** Add the headless command to the hint: `lms server start --bind 0.0.0.0 --port 1234` [S5]. Also mention the optional auth toggle [S6] and Max Concurrent Predictions [S8].
- **llama.cpp.** Add `--api-key`, `-np 4` and `--jinja` to the hint [S10].
- **vLLM.** Add `--api-key` to the hint [S11].
- **llamafile.** Run it as `llamacpp` (same port 8080 and `--host`) [S40].
- **LocalAI.** Run it as `other --port 8080` [S29].
- **TensorRT-LLM.** Run it as `other --port 8000` [S25]. Expert users only.

**Supported with small change:**
- **Paddock.** Add default port **11540** and open it on the gateway. A network bind without `--api-key` auto-generates a key [S16], so the operator must pass `--runtime-api-key`, or the hint should tell them to set `--api-key` explicitly. Hint: `paddock-runner --model /path/model.gguf --host <mesh-ip> --port 11540 --api-key <key>`. Managed endpoints that the Paddock manager starts take ports upward from 11540 [S17]: 11541, 11542 and so on.
- **Unsloth.** Add default port **8888** [S21] and open it, or tell operators to use `-p 8000`. Hint: `unsloth run --model <hf-repo>:<quant> -H <mesh-ip> -p 8888` [S19], plus a key from Settings → API → Create [S19], plus `--disable-tools` [S21].
- **SGLang.** Add the runtime id `sglang` with port 30000, or tell operators to pass `--port 8000` (simplest). Hint: `python -m sglang.launch_server --model-path <m> --host <mesh-ip> --port 8000 --api-key <key>` [S23].
- **mlx-lm.** Add the runtime id `mlx` with port 8080. Hint: `mlx_lm.server --model <m> --host <mesh-ip> --port 8080` [S36]. Upstream warns it is not for production [S35].
- **Lemonade.** Port **13305** [S47]. It also serves `/v1/*` [S47], so the probe works. Hint: `lemond --host <mesh-ip> --port 8000` plus `LEMONADE_API_KEY`.
- **Xinference.** `xinference-local -H <mesh-ip> --port 8000` (port 9997 by default) [S43]. Buyers must use the model UID.
- **Jan** (1337), **KoboldCpp** (5001), **Aphrodite/Sonar** (2242), **LMDeploy** (23333). Each one only needs its port changed to 8000 or 8080 with `--port`, or the port opened on the gateway. All are low priority.

**Waitlist:**
- **GPUStack.** It is a cluster manager with user auth and its own metering [S45]. The OpenAI path has been `/v1-openai` [S44]. It needs base-path support in the probe and in the gateway's `api_base`. Better: point GarageAI at the worker's engine directly.
- **Docker Model Runner.** The path is `/engines/v1` on 12434 [S48], so it needs base-path support.
- **exo.** Port 52415. `/v1/models` is unverified. Model placement goes through its own `/instance` API [S38].
- **text-generation-webui.** Many flags are needed (`--api --listen --api-port --api-key`) [S34]. The hobby UI is not a good marketplace fit.
- **vllm-mlx / vllm-metal.** Promising for Mac garages [S13][S37]. Check them again in Q1 2027.

**Not suitable:**
- **NVIDIA NIM.** The Developer Program license is for dev/test only. Selling tokens needs NVIDIA AI Enterprise (≈$4,500/GPU/yr) [S27]. Operators who own an AIE license could run it as `other --port 8000`.
- **HF TGI.** In maintenance mode and the repo is now archived [S0][S28]. Send operators to vLLM or SGLang.
- **Foundry Local.** Windows and macOS only, and the port is random unless pinned [S49]. Windows is unsupported by our script.

---

## 3. Recommended lists for the operator wizard

**Officially supported (6):**
1. Ollama: the easiest path for consumers.
2. LM Studio (incl. headless llmster): GUI users on Mac and Linux.
3. llama.cpp (`llama-server`): power users with GGUF.
4. vLLM: NVIDIA GPU servers, best throughput and parity with our gateway.
5. SGLang: NVIDIA servers, especially multi-GPU.
6. Paddock: **beta.** Young (repo created 2026-09-03, v0.1.13, 122 stars [S0][S50]) and single-GPU only. It has the best default posture for us: binds 0.0.0.0, key required, usage always sent.

**Also accepted via "Other / Unsloth":** Unsloth (8888), mlx-lm, Lemonade, Xinference, LocalAI, llamafile, TensorRT-LLM.

**Waitlist:** GPUStack, Docker Model Runner, exo, text-generation-webui, vllm-mlx/vllm-metal, Jan, KoboldCpp, Aphrodite/Sonar, LMDeploy.

**Do not offer:** NVIDIA NIM (license), TGI (archived), Foundry Local (Windows, random port).

---

## 4. Concrete changes

### 4a. `scripts/garageai-connect.sh` (proposal only; the file is being edited on another branch)

```bash
default_port() {
  case "$1" in
    ollama)    echo 11434 ;;
    lmstudio)  echo 1234 ;;
    llamacpp)  echo 8080 ;;
    vllm)      echo 8000 ;;
    sglang)    echo 30000 ;;   # or recommend --port 8000 in the hint and use 8000 here
    paddock)   echo 11540 ;;   # paddock-runner default
    unsloth)   echo 8888 ;;    # unsloth studio / unsloth run default
    mlx)       echo 8080 ;;
    lemonade)  echo 13305 ;;
    *)         echo "" ;;
  esac
}
# case: ollama|lmstudio|llamacpp|vllm|sglang|paddock|unsloth|mlx|lemonade|other
```

Runtime hints, with `${bind}` / `${PORT}` as today:
- ollama: keep as is, and add `Environment="OLLAMA_NUM_PARALLEL=4"`. Add a note that Ollama has no API key, so the mesh ACL is the only protection.
- lmstudio: add `Headless: lms server start --bind 0.0.0.0 --port ${PORT}`. Add "optional: Settings → Require Authentication, then pass --runtime-api-key".
- llamacpp: `llama-server -m model.gguf --host ${bind} --port ${PORT} -np 4 --jinja [--api-key KEY]`.
- vllm: `vllm serve <model> --host ${bind} --port ${PORT} [--api-key KEY] [--served-model-name <name>]`.
- sglang: `python -m sglang.launch_server --model-path <model> --host ${bind} --port ${PORT} [--api-key KEY]`.
- paddock: `paddock-runner --model /path/model.gguf --host ${bind} --port ${PORT} --api-key KEY`. Add: "Paddock requires a key on network binds; pass the same key with --runtime-api-key."
- unsloth: `unsloth run --model <repo>:<quant> -H ${bind} -p ${PORT} --disable-tools`. Add: "create an API key in Settings → API and pass it with --runtime-api-key", and the alternative "export to GGUF and use --runtime llamacpp/ollama".
- mlx: `mlx_lm.server --model <model> --host ${bind} --port ${PORT}`. Add "(no API key support)".
- lemonade: `lemond --host ${bind} --port ${PORT}` plus `LEMONADE_API_KEY=KEY`.

Other script items:
- After `/v1/models`, add a smoke test: a 1-token `POST /v1/chat/completions` with `stream:true, stream_options:{include_usage:true}`. Warn if no `usage` chunk arrives, because buyers pay per token.
- Optional: a `--base-path` flag, default `/v1`, for GPUStack (`/v1-openai`) and Docker Model Runner (`/engines/v1`) later.

### 4b. Gateway port policy

Today: 11434, 1234, 8080, 8000.
- **Add now:** 11540 (Paddock), 8888 (Unsloth), 30000 (SGLang).
- **Add if we list them:** 13305 (Lemonade). 11541–11549 if Paddock-manager-spawned endpoints should work.
- **Better long-term:** have the connect script register the garage's chosen port and open only that `peer:port` in the NetBird policy for that garage. Then there is no global allowlist to maintain, and any engine with a `/v1` API works with `--port`.
- **Until then:** the hints for low-priority engines (Jan, KoboldCpp, Aphrodite, LMDeploy, Xinference, text-generation-webui) should tell operators to run them on 8000 or 8080.

---

## 5. Unsloth

- **What it is:** Unsloth (Daniel and Michael Han) is mainly a fine-tuning library. Since 2026 it also ships **Unsloth Studio**, a local web UI for running, training and exporting models, launched 2026-03-17 [S51][S22].
- **Serving:** yes, since about May 2026 [S20]. `unsloth run --model <repo>:<quant>` (or `unsloth studio`) exposes OpenAI `/v1/chat/completions`, `/v1/responses` and `/v1/models`, plus Anthropic `/v1/messages`, on one port. Streaming, tools and vision are supported [S19][S20]. It is a wrapper around **llama-server** (llama.cpp) [S19].
- **Ports and binding:** the default port is 8888 [S21]. It binds to localhost by default, `-H 0.0.0.0` exposes it, and `-p` sets the port [S19][S21].
- **API key:** API keys (`sk-unsloth-…`) are made in the UI, stored hashed, and sent as Bearer [S19]. The LAN doc warns that traffic is plain HTTP and that server-side tools (web search, Python, terminal) run as the user. Use `--disable-tools` when exposing it [S21].
- **License:** the core is Apache-2.0 and Studio is AGPL-3.0 [S22]. AGPL matters only if the operator modifies Studio.
- **Recommendation:** Unsloth's own docs list GGUF/llama.cpp, Ollama, vLLM, SGLang and llama-server as deployment paths [S52]. For a garage, the most robust path is to **export the fine-tuned model** and serve it with an officially supported engine:
  - GGUF goes to llama.cpp, Ollama or LM Studio.
  - Merged 16-bit or FP8 weights go to vLLM or SGLang.
- Keep the `unsloth` runtime id for operators who want to serve straight from Studio. Set its port to 8888 and require `--runtime-api-key`.

## 6. Jens Nylander and Paddock

- **Who:** Jens Nylander is a Swedish developer and serial founder. He founded Jens of Sweden (MP3 players), JAYS (headphones) and Automile (acquired 2020). He is now a co-founder of The Intelligence Company (tic.io) [S53][S54].
- **Project:** **Paddock** is listed on his site as his project ("A high-throughput inference engine for open models on NVIDIA GPUs") [S53]. It is published by **Truespar**, "a research arm of The Intelligence Company AB (publ)" [S15]. The repo is github.com/truespar/paddock [S14].
  - The vendor page and README do not name individual developers. The link to Jens comes from his own site [S53].
  - His site also lists Sentio (an email API for agents) and Traverse (a graph DB) [S53].
- **What it is:**
  - A native Rust engine with its own CUDA and Metal kernels. "NOT A WRAPPER" [S14].
  - Two binaries: `paddock-runner` (one model, one port) and `paddock` (manager plus the web Studio) [S14].
  - Platforms: NVIDIA on Windows and Linux x64, DGX Spark arm64, and Apple Silicon (pre-release, macOS 26+). There is no ROCm or Vulkan backend [S14].
  - Model formats: GGUF and safetensors, with FP8, NVFP4, MXFP4 and k-quants [S14].
  - Serving: continuous batching, paged KV and prefix caching [S14]. Single GPU only for now [S14].
  - Version: v0.1.13, released 2026-10-03. The repo was created 2026-09-03 [S0].
  - License: MIT or Apache-2.0 [S14].
- **Vendor performance claim (not independently checked):** it says it beats llama.cpp, vLLM and SGLang on RTX PRO 6000 and B200 [S14][S15].
- **OpenAI compatibility: yes.**
  - Endpoints: chat completions, completions, Responses, embeddings and `/v1/models`, plus the Anthropic Messages API [S14].
  - Streaming emits `reasoning_content` deltas, and every stream ends with a server-counted usage chunk whether or not `include_usage` is set [S18].
- **Network defaults** [S16]:
  - The runner binds **0.0.0.0:11540** by default (`--host` / `--port`, or `PADDOCK_HOST` / `PADDOCK_PORT`).
  - A network bind without `--api-key` / `PADDOCK_API_KEY` **auto-generates and requires** a key. `--no-auth` opts out.
  - The manager/Studio listens on 127.0.0.1:11500 and allocates runner ports upward from 11540 [S17].
- **How to support it:** use the existing `paddock` runtime id.
  - Set `default_port` to 11540 and open 11540 on the gateway.
  - Hint: `paddock-runner --model … --host <mesh-ip> --port 11540 --api-key <key>`.
  - Make `--runtime-api-key` mandatory in the paddock flow.
  - Then LiteLLM calls it as `openai/<id>` with `api_key` set.
- **Caveats:** the project is about one month old and single-GPU. NVIDIA consumer cards before Ampere are not in the release pack, and Ada is marked "UNVALIDATED" [S14]. List it as **beta**.

---

## 7. Sources

- [S0] GitHub REST API, `GET /repos/{owner}/{repo}` and `/releases`, queried 2026-10-05. Gives latest tags, licenses and archived status. Repos: vllm-project/vllm, sgl-project/sglang, ollama/ollama, ggml-org/llama.cpp, mudler/LocalAI, janhq/jan, LostRuins/koboldcpp, oobabooga/text-generation-webui, ml-explore/mlx-lm, exo-explore/exo, mozilla-ai/llamafile, aphrodite-engine/aphrodite-engine, InternLM/lmdeploy, xorbitsai/inference, gpustack/gpustack, lemonade-sdk/lemonade, huggingface/text-generation-inference, truespar/paddock. Example: https://api.github.com/repos/huggingface/text-generation-inference
- [S1] Ollama OpenAI compatibility: https://docs.ollama.com/api/openai-compatibility
- [S2] Ollama FAQ (OLLAMA_HOST, NUM_PARALLEL): https://docs.ollama.com/faq
- [S3] Ollama authentication: https://docs.ollama.com/api/authentication
- [S4] LM Studio server: https://lmstudio.ai/docs/developer/core/server
- [S5] LM Studio `lms server start`: https://lmstudio.ai/docs/cli/server-start
- [S6] LM Studio server settings: https://lmstudio.ai/docs/developer/core/server/settings
- [S7] LM Studio API changelog (stream_options, 0.3.18): https://www.lmstudio.ai/docs/developer/api-changelog
- [S8] LM Studio 0.4.0 (parallel requests): https://lmstudio.ai/blog/0.4.0 and https://lmstudio.ai/docs/app/advanced/parallel-requests
- [S9] LM Studio headless / llmster: https://lmstudio.ai/docs/developer/core/headless
- [S10] llama.cpp server README: https://github.com/ggml-org/llama.cpp/blob/master/tools/server/README.md
- [S11] vLLM `vllm serve` CLI: https://docs.vllm.ai/en/stable/cli/serve/
- [S12] vLLM reasoning outputs: https://docs.vllm.ai/en/stable/features/reasoning_outputs/ (include_usage: https://www.mintlify.com/vllm-project/vllm/serving/cli-usage)
- [S13] Docker blog, vllm-metal on macOS: https://www.docker.com/blog/docker-model-runner-vllm-metal-macos/ ; https://modelfit.io/blog/vllm-metal-apple-silicon-concurrent-serving/
- [S14] Paddock README: https://github.com/truespar/paddock
- [S15] Truespar Paddock page: https://truespar.com/paddock
- [S16] Paddock runner config/CLI: https://github.com/truespar/paddock/blob/main/crates/paddock-runner/src/config.rs and https://github.com/truespar/paddock/blob/main/crates/paddock-runner/src/startup.rs
- [S17] Paddock manager config: https://github.com/truespar/paddock/blob/main/crates/paddock-manager/src/config.rs
- [S18] Paddock chat endpoint: https://github.com/truespar/paddock/blob/main/crates/paddock-runner/src/chat.rs
- [S19] Unsloth API endpoint docs: https://unsloth.ai/docs/basics/api
- [S20] Unsloth API endpoint announcement (May 2026): https://github.com/unslothai/unsloth/discussions/5285
- [S21] Unsloth LAN docs: https://unsloth.ai/docs/basics/lan ; https://unsloth.ai/docs/basics/how-to-serve-local-llms-anywhere-secure-remote-access-with-cloudflare-and-unsloth
- [S22] Unsloth Studio intro / licensing: https://www.unsloth.ai/docs/new ; https://themenonlab.blog/blog/unsloth-studio-fine-tune-llms-no-code
- [S23] SGLang server arguments: https://docs.sglang.io/advanced_features/server_arguments.html
- [S24] SGLang OpenAI protocol (include_usage, reasoning_content): https://github.com/sgl-project/sglang/blob/main/python/sglang/srt/entrypoints/openai/protocol.py
- [S25] trtllm-serve: https://nvidia.github.io/TensorRT-LLM/commands/trtllm-serve/trtllm-serve.html
- [S26] NIM for LLMs configuration: https://docs.nvidia.com/nim/large-language-models/1.12.0/configuration.html ; https://docs.nvidia.com/nim/large-language-models/2.0.0/reference/architecture.html
- [S27] NIM licensing: https://forums.developer.nvidia.com/t/nvidia-nim-faq/300317 ; https://www.nvidia.com/en-gb/data-center/products/ai-enterprise/get-started/
- [S28] TGI maintenance mode: https://aiwiki.ai/wiki/huggingface_tgi ; https://dev.to/runcai/text-generation-inference-tgi-what-it-is-how-it-works-and-when-to-use-it-1hm5
- [S29] LocalAI getting started: https://localai.io/docs/basics/getting_started/
- [S30] LocalAI parallel requests: https://github.com/mudler/LocalAI/discussions/1670
- [S31] Jan API server: https://www.jan.ai/docs/desktop/api-server
- [S32] KoboldCpp wiki: https://github.com/LostRuins/koboldcpp/wiki
- [S33] KoboldCpp source (usage fields): https://github.com/LostRuins/koboldcpp/blob/concedo/koboldcpp.py
- [S34] text-generation-webui OpenAI API: https://github.com/oobabooga/text-generation-webui/wiki/12-%E2%80%90-OpenAI-API
- [S35] mlx-lm SERVER.md: https://github.com/ml-explore/mlx-lm/blob/main/mlx_lm/SERVER.md
- [S36] mlx-lm server.py: https://github.com/ml-explore/mlx-lm/blob/main/mlx_lm/server.py
- [S37] vllm-mlx continuous batching: https://vllm-mlx.is-a.dev/guides/continuous-batching/
- [S38] exo README: https://github.com/exo-explore/exo
- [S39] llamafile README: https://github.com/mozilla-ai/llamafile
- [S40] llamafile server usage: https://github.com/Mozilla-Ocho/llamafile
- [S41] Aphrodite / Sonar README: https://github.com/aphrodite-engine/aphrodite-engine
- [S42] LMDeploy api_server: https://lmdeploy.readthedocs.io/en/latest/llm/api_server.html
- [S43] Xinference: https://inference.readthedocs.io/en/latest/getting_started/using_xinference.html
- [S44] GPUStack (v1-openai path, API keys): https://pypi.org/project/gpustack/0.3.1
- [S45] GPUStack README: https://github.com/gpustack/gpustack
- [S46] AMD Lemonade playbook: https://developer.amd.com/playbooks/lemonade-getting-started
- [S47] Lemonade configuration: https://lemonade-server.ai/docs/guide/configuration/
- [S48] Docker Model Runner: https://www.docker.com/blog/run-llms-locally/
- [S49] Foundry Local GA: https://devblogs.microsoft.com/foundry/foundry-local-ga/ ; https://ravichaganti.com/blog/local-model-serving-using-foundry-local/
- [S50] Paddock repo metadata: https://api.github.com/repos/truespar/paddock
- [S51] Unsloth Studio launch: https://aiautomationglobal.com/blog/unsloth-studio-no-code-local-llm-finetuning-2026
- [S52] Unsloth inference & deployment: https://unsloth.ai/docs/basics/inference-and-deployment
- [S53] Jens Nylander personal site: https://jensnylander.com/
- [S54] Jens Nylander LinkedIn (public): https://www.linkedin.com/in/nylanderjens/
