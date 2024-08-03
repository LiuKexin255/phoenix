package env

import "os"

// GetUptraceDSN 获取 uptrace 的 DSN 
func GetUptraceDSN() string {
	return os.Getenv("UPTRACE_DSN")
}

func GetUptraceEndpoint () string {
	return os.Getenv("UPTRACE_ENDPOINT")
}