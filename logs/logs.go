package logs

import "fmt"

// Writer writes logs
type Writer interface {
	LogTracef(txid string, format string, args ...interface{})
	LogWarningf(txid string, format string, args ...interface{})
	LogErrorf(txid string, format string, args ...interface{})
}

// StdOutLogger logs to stdout
type StdOutLogger struct{}

// LogTracef ...
func (s StdOutLogger) LogTracef(txid string, format string, args ...interface{}) {
	fmt.Printf("TRACE - TXID=%s, MSG=%s\n", txid, fmt.Sprintf(format, args...))
}

// LogWarningf ...
func (s StdOutLogger) LogWarningf(txid string, format string, args ...interface{}) {
	fmt.Printf("WARNING - TXID=%s, MSG=%s\n", txid, fmt.Sprintf(format, args...))
}

// LogErrorf ...
func (s StdOutLogger) LogErrorf(txid string, format string, args ...interface{}) {
	fmt.Printf("ERROR - TXID=%s, MSG=%s\n", txid, fmt.Sprintf(format, args...))
}
