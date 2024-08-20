package environment

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/slashdoom/aruba_exporter/rpc"
	"github.com/slashdoom/aruba_exporter/util"
	"regexp"
	"strings"
)

func (c *environmentCollector) ParseTemp(ostype string, output string) (map[string]Environment, error) {
	log.Debugf("OS: %s\n", ostype)
	switch ostype {
	case rpc.ArubaSwitch:
		return c.ParseArubaSwitchTemp(output)
	case rpc.ArubaCXSwitch:
		return c.ParseArubaSwitchTemp(output)
	default:
		return nil, errors.New("'show environment' is not implemented for " + ostype)
	}
}

func (c *environmentCollector) ParsePower(ostype string, output string) (map[string]Environment, error) {
	log.Debugf("OS: %s\n", ostype)
	switch ostype {
	case rpc.ArubaSwitch:
		return c.ParseArubaSwitchPower(output)
	case rpc.ArubaCXSwitch:
		return c.ParseArubaSwitchPower(output)
	default:
		return nil, errors.New("'show environment power-supply' is not implemented for " + ostype)
	}
}

func (c *environmentCollector) ParseFan(ostype string, output string) (map[string]Environment, error) {
	log.Debugf("OS: %s\n", ostype)
	switch ostype {
	case rpc.ArubaSwitch:
		return c.ParseArubaSwitchFan(output)
	case rpc.ArubaCXSwitch:
		return c.ParseArubaSwitchFan(output)
	default:
		return nil, errors.New("'show environment fan' is not implemented for " + ostype)
	}
}

func (c *environmentCollector) ParseArubaSwitchTemp(output string) (map[string]Environment, error) {
	environments := make(map[string]Environment)

	sensorRegex := regexp.MustCompile(`^\S+`)
	moduleTypeRegex := regexp.MustCompile(`\s{2,}([a-zA-Z-]+)\s{2,}`)
	temperatureRegex := regexp.MustCompile(`\d+\.\d{2}\sC`)
	statusRegex := regexp.MustCompile(`\s{2,}(\w+)$`)

	lines := strings.Split(output, "\n")
	for _, line := range lines {

		if strings.Contains(line, "Temperature") ||
			strings.Contains(line, "Slot") ||
			strings.Contains(line, "-----") ||
			strings.Contains(line, "show") ||
			strings.Contains(line, "#") ||
			strings.Contains(line, "Current") ||
			line == "" || line == "\n" {
			continue
		}
		slotSensor := sensorRegex.FindString(line)
		moduleTypeMatch := moduleTypeRegex.FindStringSubmatch(line)
		temperature := temperatureRegex.FindString(line)
		statusMatch := statusRegex.FindStringSubmatch(line)

		temperature = strings.TrimSuffix(temperature, " C")
		status := statusMatch[1]
		moduleType := moduleTypeMatch[1]

		environments[slotSensor] = Environment{
			TemperatureSlotSensor: slotSensor,
			Temperature:           util.Str2float64(temperature),
			TemperatureStatus:     status,
			TemperatureModuleType: moduleType,
		}
	}

	return environments, nil
}

func (c *environmentCollector) ParseArubaSwitchPower(output string) (map[string]Environment, error) {
	environments := make(map[string]Environment)

	slotRegex := regexp.MustCompile(`^\S+`)
	statusRegex := regexp.MustCompile(`\s{2,}([a-zA-Z-]+)\s{2,}`)                   // [1]
	productNumberRegex := regexp.MustCompile(`(?m)^\s*\d+/\d+\s+(\w+|N/A)\s+`)      // [1]
	serialNumberRegex := regexp.MustCompile(`(?m)^\s*\d+/\d+\s+\w+\s+(\w+|N/A)\s+`) // [1]

	lines := strings.Split(output, "\n")
	for _, line := range lines {

		if strings.Contains(line, "-----") ||
			strings.Contains(line, "PSU") ||
			strings.Contains(line, "Status") ||
			strings.Contains(line, "show") ||
			strings.Contains(line, "#") ||
			line == "" || line == "\n" {
			continue
		}
		slot := slotRegex.FindString(line)
		statusMatch := statusRegex.FindStringSubmatch(line)
		productNumberMatch := productNumberRegex.FindStringSubmatch(line)
		serialNumberMatch := serialNumberRegex.FindStringSubmatch(line)
		serialNumber := "N/A"
		status := statusMatch[1]
		productNumber := productNumberMatch[1]
		if serialNumberMatch != nil {
			serialNumber = serialNumberMatch[1]
		}

		environments[slot] = Environment{
			PowerSupplySlot:          slot,
			PowerSupplyStatus:        status,
			PowerSupplyProductNumber: productNumber,
			PowerSupplySerialNumber:  serialNumber,
		}
	}

	return environments, nil
}

func (c *environmentCollector) ParseArubaSwitchFan(output string) (map[string]Environment, error) {
	environments := make(map[string]Environment)

	fanInfoRegex := regexp.MustCompile(`(?s)Fan information.*?\n-{78}\n(.*?)(?:\n\n|\z)`)
	slotRegex := regexp.MustCompile(`^\S+`)
	speedRegex := regexp.MustCompile(`\s+(slow|fast|normal)\s+`)
	directionRegex := regexp.MustCompile(`\w+-to-\w+`)
	statusRegex := regexp.MustCompile(`\s+(ok|fail)\s+`)
	rpmRegex := regexp.MustCompile(`\s+(\d+)\s*$`)
	var lines []string
	fanInfoMatch := fanInfoRegex.FindStringSubmatch(output)

	if len(fanInfoMatch) > 1 {
		table := fanInfoMatch[1]
		lines = strings.Split(table, "\n")
	} else {
		return nil, errors.New("no fan information found")
	}

	for _, line := range lines {

		if strings.Contains(line, "-----") ||
			strings.Contains(line, "Name") ||
			strings.Contains(line, "Status") ||
			strings.Contains(line, "show") ||
			strings.Contains(line, "#") ||
			line == "" || line == "\n" {
			continue
		}
		slot := slotRegex.FindString(line)
		speedMatch := speedRegex.FindStringSubmatch(line)
		statusMatch := statusRegex.FindStringSubmatch(line)
		rpmMatch := rpmRegex.FindStringSubmatch(line)
		direction := directionRegex.FindString(line)
		speed := speedMatch[1]
		status := statusMatch[1]
		rpm := rpmMatch[1]

		log.Debugf("slot: %s", slot)
		log.Debugf("speed: %s", speed)
		log.Debugf("status: %s", status)
		log.Debugf("rpm: %s", rpm)
		log.Debugf("direction: %s", direction)

		environments[slot] = Environment{
			FanSlot:      slot,
			FanSpeed:     speed,
			FanStatus:    status,
			FanRPM:       util.Str2float64(rpm),
			FanDirection: direction,
		}
	}

	return environments, nil
}
