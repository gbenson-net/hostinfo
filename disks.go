package hostinfo

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"gbenson.net/go/strcase"
)

// gatherDiskAttrs gathers the output of `/sbin/blkid`.
func gatherDiskAttrs(gi *gatherInvoker, r *HostInfo) error {
	s, err := gi.Invoke("/sbin/blkid")
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(bufio.NewReader(bytes.NewBufferString(s)))
	for scanner.Scan() {
		if err := gatherDiskAttr(gi, r, scanner.Text()); err != nil {
			return err
		}
	}
	return nil
}

var blkidAttrRx = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

func blkidError(line string) error {
	return &InvalidLineError{"blkid", line}
}

func gatherDiskAttr(gi *gatherInvoker, r *HostInfo, line string) error {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	device, rest, found := strings.Cut(line, ":")
	if !found {
		return blkidError(line)
	}

	log := gi.Logger()
	log.Trace().
		Str("device", device).
		Msg("Got")

	attrs := make(map[string]any)
	if r.Disks == nil {
		r.Disks = make(map[string]map[string]any)
	}
	r.Disks[device] = attrs

	attr, rest, found := strings.Cut(rest, "=")
	attr = strings.TrimSpace(attr)
	for {
		if !found {
			if attr == "" {
				break
			}
			return blkidError(line)
		}

		if blkidAttrRx.FindString(attr) != attr {
			return blkidError(line)
		}
		attr = strings.ToLower(attr)

		var value, nextattr string
		value, rest, found = strings.Cut(rest, "=")
		if found {
			// `value` includes the name of the next attribute.
			index := strings.LastIndex(value, " ")
			if index < 0 {
				return blkidError(line)
			}
			nextattr = value[index+1:]
			value = value[:index]
		}

		if len(value) < 2 || value[0] != '"' {
			return blkidError(line)
		}
		value, err := strconv.Unquote(value)
		if err != nil {
			return err
		}

		log.Trace().
			Str("attr", attr).
			Str("value", value).
			Msg("Got")

		untypedValue := any(value)
		if attr == "block_size" {
			if n, err := strconv.Atoi(value); err == nil {
				untypedValue = n
			}
		}

		attrs[attr] = untypedValue

		attr = nextattr
	}

	if attrs["type"] == "crypto_LUKS" {
		luks, err := gatherLUKSInfo(gi, device)
		if err == nil {
			attrs["luks"] = luks
		} else {
			log.Warn().
				Str("item", "LUKSInfo").
				Str("device", device).
				AnErr("reason", err).
				Msg("Gather failed")
		}
	}

	return nil
}

func luksDumpError(line string) error {
	return &InvalidLineError{"luksDump", line}
}

// gatherLUKSInfo gathers the output of `cryptsetup luksDump`.
func gatherLUKSInfo(gi *gatherInvoker, device string) (map[string]any, error) {
	s, err := gi.InvokeRetrySudo("cryptsetup", "luksDump", device)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bufio.NewReader(bytes.NewBufferString(s)))
	result, line, err := gatherLUKSInfoMapping(gi, scanner, "")
	if err != nil {
		return nil, err
	} else if line != "" {
		return nil, luksDumpError(line) // shouldn't happen
	}

	return result, nil
}

// gatherLUKSInfoMapping gathers a YAML-style mapping from the output
// of `cryptsetup luksDump`.
func gatherLUKSInfoMapping(
	gi *gatherInvoker,
	scanner *bufio.Scanner,
	indent string,
) (result map[string]any, line string, err error) {
	var pushedBack, lastKey string

	for {
		if pushedBack == "" {
			if !scanner.Scan() {
				return result, "", err // end of input
			}
			line = scanner.Text()
		} else {
			line = pushedBack
			pushedBack = ""
		}

		if line == "" {
			continue
		} else if dedented, ok := strings.CutPrefix(line, indent); ok {
			line = dedented
		} else {
			return // dedent => end of section
		}

		firstc, _ := utf8.DecodeRuneInString(line)
		if unicode.IsSpace(firstc) {
			// more indention => continuation
			lastValue, ok := result[lastKey].(string)
			if !ok {
				return nil, "", luksDumpError(line)
			}
			result[lastKey] = fmt.Sprintf("%s %s", lastValue, strings.TrimSpace(line))
			continue
		}

		key, value, err := luksParseKeyValuePair(line)
		if err != nil {
			return nil, "", err
		} else if key == "" {
			continue
		} else if _, found := result[key]; found {
			return nil, "", luksDumpError(line)
		} else if indent == "" && value == "" {
			value, pushedBack, err = gatherLUKSInfoSlice(gi, scanner, indent)
			if err != nil {
				return nil, "", err
			}
		} else if sv, ok := value.(string); ok {
			if sv == fmt.Sprintf("(no %s)", key) {
				continue
			} else if key == "length" && sv == "(whole device)" {
				continue
			}
		}

		gi.Logger().Trace().
			Str("key", key).
			Str("value", fmt.Sprintf("%v", value)).
			Msg("Got")

		if result == nil {
			result = make(map[string]any)
		}
		result[key] = value
		lastKey = key
	}
}

// gatherLUKSInfoSlice gathers a YAML-style list from the output of
// `cryptsetup luksDump`.
func gatherLUKSInfoSlice(
	gi *gatherInvoker,
	scanner *bufio.Scanner,
	indent string,
) (result []map[string]any, line string, err error) {
	var pushedBack string

	for {
		if pushedBack == "" {
			if !scanner.Scan() {
				return result, "", err // end of input
			}
			line = scanner.Text()
		} else {
			line = pushedBack
			pushedBack = ""
		}

		if line == "" {
			continue
		} else if dedented, ok := strings.CutPrefix(line, indent); ok {
			line = dedented
		} else {
			return // dedent => end of slice
		}

		if firstc, _ := utf8.DecodeRuneInString(line); !unicode.IsSpace(firstc) {
			return // no indentation at all => end of slice
		}

		key, value, err := luksParseKeyValuePair(line)
		if err != nil {
			return nil, "", err
		}
		index, err := strconv.Atoi(key)
		if err != nil {
			return nil, "", err
		} else if index != len(result) {
			return nil, "", luksDumpError(line) // index out of sequence
		}

		const itemValueKey = "type"
		var item map[string]any
		item, pushedBack, err = gatherLUKSInfoMapping(gi, scanner, indent+"\t")
		if err != nil {
			return nil, "", err
		} else if _, found := item[itemValueKey]; found {
			return nil, "", luksDumpError(line)
		} else if item == nil {
			return nil, "", luksDumpError(line)
		}

		item[itemValueKey] = value
		result = append(result, item)
	}
}

func luksParseKeyValuePair(line string) (key string, value any, err error) {
	key, v, found := strings.Cut(line, ":")
	if !found {
		if line == "LUKS header information" {
			return "", "", nil
		}
		return "", "", luksDumpError(line)
	}
	if key = strcase.ToSnake(key); key == "" {
		return "", "", luksDumpError(line)
	}

	v = strings.TrimSpace(v)
	if v == "" {
		return key, "", nil
	}

	if s, found := strings.CutSuffix(v, " [bytes]"); found {
		key += "_bytes"
		v = s
	} else if s, found := strings.CutSuffix(v, " bits"); found {
		key += "_bits"
		v = s
	}

	if n, err := strconv.Atoi(v); err == nil {
		return key, n, nil // integer
	}

	return key, v, nil // string
}
