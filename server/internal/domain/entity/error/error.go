package error

type CustomError interface {
	Error() string      // returns the error message
	ErrorCode() string  // returns the error code which identifies error category
	ErrorObject() error // returns the error object
	StatusCode() int    // returns the HTTP status code
}
