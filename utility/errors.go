package utility

// ResponseError groups status code, error, and message together
type ResponseError struct {
	StatusCode int
	Error      error
	Message    string
}
