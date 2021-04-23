package redis

import "regexp"

type device struct {
	LampHours       int    `json:"lamp-hours"`
	HardwareVersion string `json:"hardware-version"`
	DeviceType      string `json:"device-type"`
}

type query func(id string, dev device) bool

func defaultQueries() map[string]query {
	return map[string]query{
		"lampHours": func(id string, dev device) bool {
			reg := regexp.MustCompile("^(Panasonic).*((EZ770)|(EZ570))")
			return dev.DeviceType == "display" && dev.LampHours > 2850 && reg.MatchString(dev.HardwareVersion)
		},
	}
}
