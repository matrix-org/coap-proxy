// Copyright 2019 New Vector Ltd
//
// This file is part of coap-proxy.
//
// coap-proxy is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// coap-proxy is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with coap-proxy.  If not, see <https://www.gnu.org/licenses/>.

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
