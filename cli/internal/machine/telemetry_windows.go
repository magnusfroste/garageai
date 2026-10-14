package machine

// Windows: no telemetry yet beyond the GPU (nvidia-smi works there too); the Windows service is
// not finished either (docs/bridge.md).
func collectOS() Telemetry { return Telemetry{} }

func statfsFreeGB(string) int { return 0 }
