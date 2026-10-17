package profile

import "testing"

func TestPowercfgAC(t *testing.T) {
	standby := `Power Setting GUID: 29f6c1db-86da-48c5-9fdb-f2b67b1f44da  (Sleep after)
      Minimum Possible Setting: 0x00000000
      Maximum Possible Setting: 0xffffffff
      Possible Settings increment: 0x00000001
      Possible Settings units: Seconds
    Current AC Power Setting Index: 0x00000708
    Current DC Power Setting Index: 0x00000384`
	if v, ok := powercfgAC(standby); !ok || v != 1800 {
		t.Fatalf("standby AC = %d,%v want 1800", v, ok)
	}
	// Swedish Windows: other words, same hex values.
	lid := `Energiinställningens GUID: 5ca83367-6e45-459f-a27b-476b1d01c936  (Stängningsåtgärd för locket)
      Möjligt inställningsindex: 000
      Möjligt inställningsindex: 001
    Index för aktuell inställning för nätström: 0x00000001
    Index för aktuell inställning för likström: 0x00000001`
	if v, ok := powercfgAC(lid); !ok || v != 1 {
		t.Fatalf("lid AC = %d,%v want 1", v, ok)
	}
	if _, ok := powercfgAC(""); ok {
		t.Fatal("empty output must be unknown")
	}
}

func TestOllamaRunnerByBlobName(t *testing.T) {
	if !ollamaRunner("", []Model{{ID: "sha256-7f4030143c1c477224c5434f8272c662a8b042079a0a584f0a27a1684fe2e1fa"}}) {
		t.Fatal("Ollama's model process (blob name, unknown process) must be hidden")
	}
	if ollamaRunner("llama-server", []Model{{ID: "qwen3-8b-q4_k_m.gguf"}}) {
		t.Fatal("an operator's llama-server must stay")
	}
}
