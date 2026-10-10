// Package profile builds the garage profile: what runs on this machine and what is wrong, in the
// same JSON (schema 1) as `garageai-connect.sh --doctor --json`. Discovery is deterministic: it
// lists the TCP ports that listen, asks each candidate whether it is an OpenAI-compatible API, and
// reads a few start flags. Keys and whole command lines are never part of a profile.
package profile

// Profile is schema 1. Field names and meaning are the contract with the portal and the
// Operations Center: change them only together with garageai-connect.sh.
type Profile struct {
	Schema        int       `json:"schema"`
	ScriptVersion string    `json:"script_version"`
	Tool          string    `json:"tool"`
	GeneratedAt   string    `json:"generated_at"`
	Machine       Machine   `json:"machine"`
	GPUs          []GPU     `json:"gpus"`
	NetBird       NetBird   `json:"netbird"`
	Runtimes      []Runtime `json:"runtimes"`
	Firewall      Firewall  `json:"firewall"`
	Heartbeat     Heartbeat `json:"heartbeat"`
	GarageAI      GarageAI  `json:"garageai"`
	Problems      []Problem `json:"problems"`
	OK            bool      `json:"ok"`
}

type Machine struct {
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	MemoryGB *int   `json:"memory_gb"`
}

type GPU struct {
	Vendor        string `json:"vendor"`
	Name          string `json:"name"`
	MemoryMB      *int   `json:"memory_mb"`
	Driver        string `json:"driver,omitempty"`
	UnifiedMemory bool   `json:"unified_memory,omitempty"`
}

type NetBird struct {
	Installed bool    `json:"installed"`
	Version   *string `json:"version"`
	Connected bool    `json:"connected"`
	MeshIP    *string `json:"mesh_ip"`
}

type Model struct {
	ID      string  `json:"id"`
	Context *int    `json:"context"`
	OwnedBy *string `json:"owned_by"`
}

type LoadedModel struct {
	Name    string `json:"name"`
	Context *int   `json:"context"`
}

// Flags are the few start flags of vLLM/SGLang worth knowing. Never the command line itself.
type Flags struct {
	Host                *string `json:"host,omitempty"`
	Port                *int    `json:"port,omitempty"`
	MaxModelLen         *int    `json:"max_model_len,omitempty"`
	ContextLength       *int    `json:"context_length,omitempty"`
	APIKeySet           *bool   `json:"api_key_set,omitempty"`
	PromptTokensDetails *bool   `json:"prompt_tokens_details,omitempty"`
}

type Runtime struct {
	Port         int           `json:"port"`
	Kind         string        `json:"kind"`
	Process      *string       `json:"process"`
	API          string        `json:"api"`
	Binds        []string      `json:"binds"`
	Network      bool          `json:"network"`
	Models       []Model       `json:"models"`
	Version      *string       `json:"version"`
	OllamaLoaded []LoadedModel `json:"ollama_loaded"`
	Flags        Flags         `json:"flags"`
}

type Firewall struct {
	Tool                  string `json:"tool"`
	Active                bool   `json:"active"`
	GarageAIRuleInstalled bool   `json:"garageai_rule_installed"`
}

type Heartbeat struct {
	Installed bool `json:"installed"`
	Active    bool `json:"active"`
}

type GarageAI struct {
	ConfiguredRuntime *string `json:"configured_runtime"`
	ConfiguredPort    *int    `json:"configured_port"`
}

type Problem struct {
	Severity string `json:"severity"` // error, warning, info
	Code     string `json:"code"`
	Message  string `json:"message"`
	Fix      string `json:"fix"`
}

func strp(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func intp(n int) *int { return &n }
