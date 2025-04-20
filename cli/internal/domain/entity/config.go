package entity

const (
	DefaultServerURL      = "http://localhost:18080"
	DefaultChunkSize      = 1024 * 8 // 8 KiB
	DefaultRetries        = 3
	DefaultMaxConcurrency = 5
)

type ServiceConfig struct {
	ServerURL      string
	ChunkSize      int64
	Retries        int
	MaxConcurrency int
}

func NewServiceConfig(serverOrigin string, chunkSize int64, retries int, maxConcurrency int) *ServiceConfig {
	return &ServiceConfig{
		ServerURL:      serverOrigin,
		ChunkSize:      chunkSize,
		Retries:        retries,
		MaxConcurrency: maxConcurrency,
	}
}
