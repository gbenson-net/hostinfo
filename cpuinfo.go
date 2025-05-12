package hostinfo

import (
	"bufio"
	"bytes"
	"iter"
	"maps"
	"slices"
	"strings"

	"github.com/google/go-cmp/cmp"
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

	compactCPUInfo(&r.CPUInfo, r.CPUs)
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

// compactCPUInfo moves values common to all CPUs into a global template.
func compactCPUInfo(template *map[string]any, cpus []map[string]any) {
	result := *template

	for k, v := range maps.Collect(sharedKVPs(cpus)) {
		tv, found := result[k]
		if found {
			if !cmp.Equal(tv, v) {
				continue // template has a different value
			}
		} else {
			if result == nil {
				result = make(map[string]any)
			}
			result[k] = v
		}

		for _, cpu := range cpus {
			delete(cpu, k)
		}
	}

	*template = result
}

// sharedKVPs returns an iterator over all key-value pairs which are
// identical across all mm.
func sharedKVPs[Map ~map[K]V, K comparable, V any](mm []Map) iter.Seq2[K, V] {
	if len(mm) < 2 {
		return func(_ func(K, V) bool) {}
	}
	ref := mm[0]
	mm = mm[1:]

	return func(yield func(K, V) bool) {
	keys:
		for k, rv := range ref {
			for _, m := range mm {
				mv, found := m[k]
				if !found {
					continue keys
				}
				if !cmp.Equal(mv, rv) {
					continue keys
				}
			}
			if !yield(k, rv) {
				return
			}
		}
	}
}
