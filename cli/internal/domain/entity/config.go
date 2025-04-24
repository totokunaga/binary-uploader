package entity

const (
	DefaultServerURL          = "http://localhost:38080"
	DefaultChunkSize          = 1024 * 128 // 128 KiB
	DefaultRetries            = 3
	DefaultMaxConcurrency     = 5
	DefaultCompressionEnabled = false
)

// Define custom types for context keys
type ContextKey string

const (
	ConcurrencyKey        ContextKey = "concurrency"
	RetriesKey            ContextKey = "retries"
	CompressionEnabledKey ContextKey = "compressionEnabled"
)
