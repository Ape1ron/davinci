package pgsql

import (
	"fmt"
	"strings"
)

func PatchPgConfig(content, key, newValue string) string {
	lines := strings.Split(content, "\n")
	var adding = false
	var newline = fmt.Sprintf("%s = %s ", key, newValue)
	for i := 0; i < len(lines); i++ {
		line := strings.Trim(lines[i], " ")
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, key+" ") || strings.HasSuffix(line, key+"=") {
			lines[i] = newline
			adding = true
		}
	}
	if !adding {
		lines = append(lines, newline)
	}
	return strings.Join(lines, "\n")
}
