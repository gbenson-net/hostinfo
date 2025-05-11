package hostinfo

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
)

// gatherInterfaces gathers the content of `ip address`.
func gatherInterfaces(gi *gatherInvoker, r *HostInfo) error {
	var devices []map[string]any

	if s, err1 := gi.Invoke("ip", "--json", "address", "show"); err1 == nil {
		if err1 = json.Unmarshal([]byte(s), &devices); err1 != nil {
			return err1
		}
	} else if s, err2 := gi.Invoke("ip", "address", "show"); err2 == nil {
		if devices, err2 = unmarshalInterfaces(s); err2 != nil {
			return err2
		}
	} else {
		return errors.Join(err1, err2)
	}

	for _, device := range devices {
		name, _ := device["ifname"].(string)
		if name == "" {
			gi.Logger().Warn().Msg("Skipping unnamed interface")
			continue
		}
		if _, found := r.Interfaces[name]; found {
			gi.Logger().Warn().
				Str("name", name).
				Msg("Skipping duplicate interface")
			continue
		}
		if r.Interfaces == nil {
			r.Interfaces = make(map[string]map[string]any)
		}
		r.Interfaces[name] = device
	}

	return nil
}

var ifHdrRx = regexp.MustCompile(`^(\d+): (\S+):`)
var ifLinkRx = regexp.MustCompile(`^\s+link/(\S+) (\S+)`)
var ifAddrRx = regexp.MustCompile(`^\s+(inet6?) (\S+)/(\d+)`)

// unmarshalInterfaces parses the non-JSON output of `ip address show`.
// This is very minimal at the moment, only a very small amount of the
// provided information is gathered.
func unmarshalInterfaces(s string) (devices []map[string]any, err error) {
	var device map[string]any
	scanner := bufio.NewScanner(bufio.NewReader(bytes.NewBufferString(s)))
	for scanner.Scan() {
		line := scanner.Text()

		if m := ifHdrRx.FindStringSubmatch(line); m != nil {
			if device, err = unmarshalInterfaceHeader(m); err != nil {
				return nil, err
			}
			devices = append(devices, device)
		} else if m := ifLinkRx.FindStringSubmatch(line); m != nil {
			if err = unmarshalInterfaceLink(device, m); err != nil {
				return nil, err
			}
		} else if m := ifAddrRx.FindStringSubmatch(line); m != nil {
			if err = unmarshalInterfaceAddress(device, m); err != nil {
				return nil, err
			}
		}
	}
	return
}

// unmarshalInterfaceHeader parses lines like
// "5: eth0: <UP,LOWER_UP> mtu 1500 qdisc noqueue state UP qlen 1000".
func unmarshalInterfaceHeader(m []string) (map[string]any, error) {
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return nil, err // shouldn't be possible
	}

	device := make(map[string]any)
	device["ifindex"] = n
	device["ifname"] = m[2]

	return device, nil
}

// unmarshalInterfaceLink parses lines like:
// "    link/ether 00:16:3e:ba:ab:67 brd ff:ff:ff:ff:ff:ff link-netnsid 0".
func unmarshalInterfaceLink(device map[string]any, m []string) error {
	device["link_type"] = m[1]
	device["address"] = m[2] // MAC

	return nil
}

// unmarshalInterfaceAddress parses lines like:
// "    inet 100.115.92.201/28 brd 100.115.92.207 scope global eth0"
// and "    inet6 fe80::216:3eff:feba:ab67/64 scope link".
func unmarshalInterfaceAddress(device map[string]any, m []string) error {
	n, err := strconv.Atoi(m[3])
	if err != nil {
		return err // shouldn't be possible
	}

	addr := make(map[string]any)
	addr["family"] = m[1]
	addr["local"] = m[2]
	addr["prefixlen"] = n

	addrs, _ := device["addr_info"].([]map[string]any)
	device["addr_info"] = append(addrs, addr)

	return nil
}
