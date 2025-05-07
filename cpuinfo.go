package hostinfo

import (
	"bufio"
	"bytes"
	"slices"
	"strconv"
	"strings"

	"gbenson.net/go/strcase"
)

// gatherCPUInfo gathers the content of `/proc/cpuinfo`.
func gatherCPUInfo(gi *gatherInvoker, r *HostInfo) error {
	s, err := gi.ReadFile("/proc/cpuinfo")
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(bufio.NewReader(bytes.NewBufferString(s)))
	for scanner.Scan() {
		line := scanner.Text()

		key, value, err := cpuinfoParseLine(line)
		if err != nil {
			return err
		}

		if key != "processor" {
			if _, exists := r.CPUInfo[key]; exists {
				return cpuinfoError(line)
			}
			if r.CPUInfo == nil {
				r.CPUInfo = make(map[string]any)
			}
			r.CPUInfo[key] = value
			continue
		}

		if n, ok := value.(int); !ok {
			return cpuinfoError(line)
		} else if n != len(r.CPUs) {
			return cpuinfoError(line)
		}

		p, err := unmarshalProcessor(scanner)
		if err != nil {
			return err
		}

		r.CPUs = append(r.CPUs, p)
	}

	return nil
}

func unmarshalProcessor(scanner *bufio.Scanner) (map[string]any, error) {
	p := make(map[string]any)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}

		key, value, err := cpuinfoParseLine(line)
		if err != nil {
			return nil, err
		} else if _, exists := p[key]; exists {
			return nil, cpuinfoError(line)
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

// cpuinfoParseLine splits a colon-separated line from /proc/cpuinfo
// into a key-value pair.  The returned key will be converted to snake
// case as necessary.  The returned value will be converted to type
// "int" if [strconv.ParseInt] with base 0 handles it without error,
// otherwise the returned value will be of type "string".
func cpuinfoParseLine(line string) (key string, value any, err error) {
	var val string
	var found bool
	if key, val, found = strings.Cut(line, ":"); !found {
		return "", nil, cpuinfoError(line)
	}
	if key = strcase.ToSnake(key); key == "" {
		return "", nil, cpuinfoError(line)
	}
	val = strings.TrimSpace(val)
	switch val {
	case "yes":
		return key, true, nil
	}

	if n, err := strconv.ParseInt(val, 0, 64); err == nil {
		return key, int(n), nil
	}
	return key, val, nil
}

func cpuinfoError(line string) error {
	return &InvalidLineError{"cpuinfo", line}
}
