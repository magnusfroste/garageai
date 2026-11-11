// Package profile builds the garage profile for GarageAI Bridge: what runs on this machine and what is wrong, as
// JSON (schema 1) for the portal and the Operations Center. Discovery is deterministic: it
// lists the TCP ports that listen, asks each candidate whether it is an OpenAI-compatible API, and
// reads a few start flags. Keys and whole command lines are never part of a profile.
package profile

// Profile is schema 1. Field names and meaning are the contract with the portal and the
// Operations Center: change them only together with those.
type Profile struct {
	Schema        int       `json:"schema"`
	ScriptVersion string    `json:"script_version"`
	Tool          string    `json:"tool"`
	GeneratedAt   string    `json:"generated_at"`
	Machine       Machine   `json:"machine"`
	GPUs          []GPU     `json:"gpus"`
	Runtimes      []Runtime `json:"runtimes"`
	Firewall      Firewall  `json:"firewall"`
	Heartbeat     Heartbeat `json:"heartbeat"`
	Tunnel        *Tunnel   `json:"tunnel,omitempty"`
	GarageAI      GarageAI  `json:"garageai"`
	Problems      []Problem `json:"problems"`
	OK            bool      `json:"ok"`

	// Inputs to the problem rules that are not part of the profile itself.
	darwin         bool
	heartbeatLast  string
	sleepMinutes   int
	laptop         bool // a Mac with a battery: its "prevent sleeping" setting only holds on the power adapter
	winSleepACMin  int  // Windows: minutes until sleep on the power adapter (0 = never, -1 = unknown)
	winLidSleepsAC bool // Windows laptop: closing the lid on the power adapter sleeps or hibernates
	displayMinutes int
	diskFreeGB     *int // free space on the system disk (models, Bridge's updates)
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
	Port          int           `json:"port"`
	Kind          string        `json:"kind"`
	Process       *string       `json:"process"`
	PID           int           `json:"-"`
	API           string        `json:"api"`
	Binds         []string      `json:"binds"`
	Network       bool          `json:"network"` // listens on all interfaces (not needed: the tunnel reaches localhost)
	Models        []Model       `json:"models"`
	Version       *string       `json:"version"`
	OllamaLoaded  []LoadedModel `json:"ollama_loaded"`
	RunnerContext *int          `json:"runner_context,omitempty"` // Xyntetik Runner: the window it runs with (/v1/capabilities)
	Flags         Flags         `json:"flags"`
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

// Tunnel is Bridge's own tunnel (docs/bridge-tunnel.md), when this machine is linked.
type Tunnel struct {
	IP        string `json:"ip"`
	Installed bool   `json:"installed"`
	Active    bool   `json:"active"`
}

type GarageAI struct {
	ConfiguredRuntime *string `json:"configured_runtime"`
	ConfiguredPort    *int    `json:"configured_port"`
	Transport         string  `json:"transport,omitempty"` // "tunnel" when this machine is linked (the only transport)
	TunnelIP          string  `json:"tunnel_ip,omitempty"`
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
