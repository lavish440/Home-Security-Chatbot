package models

type Config struct {
	Port             string
	GeminiAPIKey     string
	Origin           string
	ReverseProxyIP   string
	EnforceHTTPS     bool
	EnableMonitoring bool
	EnableDebug      bool
	BasicAuthUser    string
	BasicAuthPass    string
}
