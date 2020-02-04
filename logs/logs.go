package logs

// Writer writes logs
type Writer interface {
	LogTracef(txid string, format string, args ...interface{})
	LogWarningf(txid string, format string, args ...interface{})
	LogErrorf(txid string, format string, args ...interface{})
}
