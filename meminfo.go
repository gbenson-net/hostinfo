package hostinfo

import (
	"bufio"
	"bytes"
)

// gatherMemInfo gathers the content of `/proc/meminfo`.
func gatherMemInfo(gi *gatherInvoker, r *HostInfo) error {
	s, err := gi.ReadFile("/proc/meminfo")
	if err != nil {
		return err
	}
	r.Memory = make(map[string]any)

	scanner := bufio.NewScanner(bufio.NewReader(bytes.NewBufferString(s)))
	parser := &keyValuePairParser{"meminfo"}
	for scanner.Scan() {
		line := scanner.Text()

		key, value, err := parser.ParseLine(line)
		if err != nil {
			return err
		} else if _, exists := r.Memory[key]; exists {
			return parser.Error(line)
		}

		r.Memory[key] = value
	}

	return nil
}
