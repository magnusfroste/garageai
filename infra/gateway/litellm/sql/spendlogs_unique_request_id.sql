-- LiteLLM_SpendLogs: make Ollama's response ids unique before they are inserted.
--
-- LiteLLM keys a spend-log row by the provider's response id (request_id is the primary key),
-- and Ollama's OpenAI-compatible API answers with `chatcmpl-<rand.Intn(999)>`
-- (ollama/middleware/openai.go): a thousand possible ids. LiteLLM inserts spend logs with
-- "skip duplicates", so every collision with a row from the last 30 days is silently dropped:
-- with 464 of the 1000 ids taken on 2026-10-13, 46 % of requests to Ollama garages left no
-- spend-log row (operators unpaid, Ops and quality figures short), rising by the day. vLLM
-- (chatcmpl-<uuid>) and OpenRouter (gen-<hex>) are not affected.
--
-- This trigger rewrites a short Ollama id to chatcmpl-<n>-<uuid> on insert, so every row lands.
-- Applied by hand on the gateway (psql in the postgres container); survives LiteLLM upgrades and
-- restarts (it lives in the database). Reported upstream to LiteLLM and Ollama.
create or replace function garageai_unique_spend_log_id() returns trigger as $$
begin
  if new.request_id ~ '^chatcmpl-[0-9]{1,4}$' then
    new.request_id := new.request_id || '-' || replace(gen_random_uuid()::text, '-', '');
  end if;
  return new;
end
$$ language plpgsql;

drop trigger if exists garageai_unique_spend_log_id on "LiteLLM_SpendLogs";
create trigger garageai_unique_spend_log_id
  before insert on "LiteLLM_SpendLogs"
  for each row execute function garageai_unique_spend_log_id();
