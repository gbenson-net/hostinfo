package hostinfo

import (
	"bufio"
	"bytes"
	"regexp"
	"strconv"
	"strings"
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

	attrs := make(map[string]string)
	if r.Disks == nil {
		r.Disks = make(map[string]map[string]string)
	}
	r.Disks[device] = attrs

	attr, rest, found := strings.Cut(rest, "=")
	attr = strings.TrimSpace(attr)
	for {
		if !found {
			if attr == "" {
				return nil
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

		attrs[attr] = value

		attr = nextattr
	}
}
