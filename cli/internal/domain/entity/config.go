package entity

const (
	DefaultServerURL          = "http://localhost:38080"
	DefaultChunkSize          = 1024 * 256 // 256 KiB
	DefaultRetries            = 3
	DefaultMaxConcurrency     = 10
	DefaultCompressionEnabled = false
)

type ContextKey string

const (
	ConcurrencyKey        ContextKey = "concurrency"
	RetriesKey            ContextKey = "retries"
	CompressionEnabledKey ContextKey = "compressionEnabled"
)
