package hostinfo

import (
	"bufio"
	"bytes"
	"slices"
	"strings"
)

// gatherCPUInfo gathers the content of `/proc/cpuinfo`.
func gatherCPUInfo(gi *gatherInvoker, r *HostInfo) error {
	s, err := gi.ReadFile("/proc/cpuinfo")
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(bufio.NewReader(bytes.NewBufferString(s)))
	parser := &keyValuePairParser{"cpuinfo"}
	for scanner.Scan() {
		line := scanner.Text()

		key, value, err := parser.ParseLine(line)
		if err != nil {
			return err
		}

		if key != "processor" {
			if _, exists := r.CPUInfo[key]; exists {
				return parser.Error(line)
			}
			if r.CPUInfo == nil {
				r.CPUInfo = make(map[string]any)
			}
			r.CPUInfo[key] = value
			continue
		}

		if n, ok := value.(int); !ok {
			return parser.Error(line)
		} else if n != len(r.CPUs) {
			return parser.Error(line)
		}

		p, err := unmarshalProcessor(scanner, parser)
		if err != nil {
			return err
		}

		r.CPUs = append(r.CPUs, p)
	}

	return nil
}

func unmarshalProcessor(
	scanner *bufio.Scanner,
	parser *keyValuePairParser,
) (map[string]any, error) {
	p := make(map[string]any)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}

		key, value, err := parser.ParseLine(line)
		if err != nil {
			return nil, err
		} else if _, exists := p[key]; exists {
			return nil, parser.Error(line)
		}

		if isFlags(key) {
			if s, ok := value.(string); ok {
				p[key] = parseFlags(s)
				continue
			}
		}

		p[key] = value
	}

	return p, nil
}

func isFlags(key string) bool {
	switch key {
	default:
		return false
	case "bugs":
	case "features":
	case "flags":
	case "vmx_flags":
	}
	return true
}

func parseFlags(s string) []string {
	result := strings.Fields(s)
	slices.Sort(result)
	return result
}
