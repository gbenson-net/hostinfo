package hostinfo

import "encoding/json"

// gatherInterfaces gathers the content of `ip address`.
func gatherInterfaces(gi *gatherInvoker, r *HostInfo) error {
	s, err := gi.Invoke("ip", "--json", "address", "show")
	if err != nil {
		return err
	}

	var devices []map[string]any
	if err = json.Unmarshal([]byte(s), &devices); err != nil {
		return err
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
