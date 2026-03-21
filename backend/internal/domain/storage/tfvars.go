package storage

import (
	"fmt"
	"sort"
	"strings"
)

// FormatTFVars converts key/value maps into HCL terraform.tfvars content.
// Keys are sorted for deterministic output.
func FormatTFVars(vars map[string]string) []byte {
	if len(vars) == 0 {
		return nil
	}

	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&b, "%s = %q\n", k, vars[k])
	}

	return []byte(b.String())
}
