package entity

const (
	DefaultServerURL      = "http://localhost:38080"
	DefaultChunkSize      = 1024 * 32 // 32 KiB
	DefaultRetries        = 3
	DefaultMaxConcurrency = 5
)

type Config struct {
	ServerURL      string
	ChunkSize      int64
	Retries        int
	MaxConcurrency int
}

func NewServiceConfig() *Config {
	return &Config{
		ServerURL:      DefaultServerURL,
		ChunkSize:      DefaultChunkSize,
		Retries:        DefaultRetries,
		MaxConcurrency: DefaultMaxConcurrency,
	}
}
