package common

import (
	"log"
	"os"
)

var (
	debugLogEnabled, dumpPayloads bool
)

// EnableDebugLogging enables debug logging.
func EnableDebugLogging() {
	debugLogEnabled = true
	_, dumpPayloads = os.LookupEnv("PROXY_DUMP_PAYLOADS")
}

// Debug prints a debug log with the given parameters if debug logging is
// enabled.
func Debug(msg ...interface{}) {
	if debugLogEnabled {
		log.Println("DEBUG:", msg)
	}
}

// Debugf prints a debug log with the given parameters and format if debug
// logging is enabled.
func Debugf(format string, args ...interface{}) {
	if debugLogEnabled {
		format = "DEBUG: " + format
		log.Printf(format, args...)
	}
}

// DumpPayload dumps a payload if both debug logging is enabled and the
// PROXY_DUMP_PAYLOADS environment variable is set.
func DumpPayload(label string, pl interface{}) {
	if dumpPayloads {
		Debugf("%s: %v", label, pl)
	}
}
