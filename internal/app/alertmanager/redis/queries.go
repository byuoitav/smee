package redis

import (
	"regexp"
	"time"
)

type device struct {
	DeviceID          string    `json:"deviceID"`
	RoomID            string    `json:"room"`
	LampHours         int       `json:"lamp-hours"`
	HardwareVersion   string    `json:"hardware-version"`
	DeviceType        string    `json:"device-type"`
	Temperature       int       `json:"temperature"`
	LastStateReceived time.Time `json:"last-state-received"`
	LastHeartbeat     time.Time `json:"last-heartbeat"`
	WebsocketCount    int       `json:"websocket-count"`
	BatteryType       string    `json:"battery-type"`

	FieldStateReceived struct {
		WebsocketCount time.Time `json:"websocket-count"`
	} `json:"field-state-received"`
}

type query func(id string, dev device) bool

// TODO have something on the redis client to name the type of the alerts based on these
// OR just register the queries in the main function
func defaultQueries() map[string]query {
	return map[string]query{
		"display-temperature": func(id string, dev device) bool {
			return dev.DeviceType == "display" && dev.Temperature > 109
		},
		"lamp-hours": func(id string, dev device) bool {
			reg := regexp.MustCompile("^(Panasonic).*((EZ770)|(EZ570))")
			return dev.DeviceType == "display" && dev.LampHours > 2850 && reg.MatchString(dev.HardwareVersion)
		},
		"no-state-updates": func(id string, dev device) bool {
			reg := regexp.MustCompile("-(LA|DMPS|CP)[0-9]*$")
			return time.Since(dev.LastStateReceived) > 10*time.Minute && reg.MatchString(dev.DeviceID)
		},
		"sys-offline": func(id string, dev device) bool {
			reg := regexp.MustCompile("-(LA|DMPS|CP|AGW|DS|TC|SP)[0-9]*$")
			return time.Since(dev.LastHeartbeat) > 6*time.Minute && reg.MatchString(dev.DeviceID)
		},
		"sys-offline-custom": func(id string, dev device) bool {
			reg := regexp.MustCompile("-(TECLITE1|CUSTOM1|TECSD1)$")
			return time.Since(dev.LastHeartbeat) > 6*time.Minute && reg.MatchString(dev.DeviceID)
		},
		"websocket": func(id string, dev device) bool {
			return (dev.DeviceType == "control-processor" || dev.DeviceType == "scheduling-panel") && dev.WebsocketCount == 0 && time.Since(dev.FieldStateReceived.WebsocketCount) > 3*time.Minute
		},
		"mic-battery-type": func(id string, dev device) bool {
			return dev.DeviceType == "microphone" && dev.BatteryType == "ALKA"
		},
	}
}
