package hostinfo

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"gbenson.net/go/strcase"
)

type keyValuePairParser struct {
	Name string
}

type kvpp = keyValuePairParser

var strtobool = map[string]bool{
	"false": false,
	"no":    false,
	"true":  true,
	"yes":   true,
}

var floatyIntegerRx = regexp.MustCompile(`^(\d+)\.0+$`)

// ParseLine splits a colon-separated line from /proc/cpuinfo or
// /proc/meminfo or a key-value pair.  The returned key will be
// converted to snake case as necessary; the returned value will be
// converted to type "int" or "bool" if possible, returned as type
// "string otherwise.  Values comprising an integer followed by
// whitespace and a string are interpreted as dimensioned values and
// returned as type "int", with the string part (the unit) appended
// to the returned key.
func (p *kvpp) ParseLine(line string) (key string, value any, err error) {
	key, v, found := strings.Cut(line, ":")
	if !found {
		return "", nil, p.Error(line)
	}
	if key = strcase.ToSnake(key); key == "" {
		return "", nil, p.Error(line)
	}

	v = strings.TrimSpace(v)

	// Transform floats that are whole numbers into integers.
	v = floatyIntegerRx.ReplaceAllString(v, "${1}")

	// Parse dimensionless integers.
	if n, err := strconv.ParseInt(v, 0, 64); err == nil {
		return key, int(n), nil
	}

	// Parse boolean values.
	if b, found := strtobool[strings.ToLower(v)]; found {
		return key, b, nil
	}

	// Parse integers with units.
	fields := strings.Fields(v)
	if len(fields) != 2 {
		return key, v, nil
	}
	unit := strings.ToLower(fields[1])
	if unit != "kb" {
		// XXX allow others?
		return key, v, nil
	}
	n, err := strconv.Atoi(fields[0])
	if err != nil {
		return key, v, nil // not a number
	}
	err = nil // unnecessary, but...
	key = fmt.Sprintf("%s_%s", key, unit)
	return key, n, nil
}

func (p *kvpp) Error(line string) error {
	return &InvalidLineError{p.Name, line}
}
