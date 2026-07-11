package multicall

import "strings"

// Add0xPrefix adds 0x hex prefix to a string, if needed.
func Add0xPrefix(s string) string {
	if strings.HasPrefix(s, "0x") {
		return s
	}
	return "0x" + s
}
