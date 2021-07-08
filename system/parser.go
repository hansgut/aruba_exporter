package system

import (
	"errors"
	"regexp"
	"strings"

	"github.com/yankiwi/aruba_exporter/rpc"
	"github.com/yankiwi/aruba_exporter/util"
)

// ParseVersion parses cli output and tries to find the version number of the running OS
func (c *systemCollector) ParseVersion(ostype string, output string) (SystemVersion, error) {
	if ostype != rpc.IOSXE && ostype != rpc.NXOS && ostype != rpc.IOS {
		return SystemVersion{}, errors.New("'show version' is not implemented for " + ostype)
	}
	versionRegexp := make(map[string]*regexp.Regexp)
	versionRegexp[rpc.IOSXE], _ = regexp.Compile(`^.*, Version (.+) -.*$`)
	versionRegexp[rpc.IOS], _ = regexp.Compile(`^.*, Version (.+),.*$`)
	versionRegexp[rpc.NXOS], _ = regexp.Compile(`^\s+NXOS: version (.*)$`)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		matches := versionRegexp[ostype].FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		return SystemVersion{Version: ostype + "-" + matches[1]}, nil
	}
	return SystemVersion{}, errors.New("Version string not found")
}

// ParseMemory parses cli output and tries to find current memory usage
func (c *systemCollector) ParseMemory(ostype string, output string) ([]SystemMemory, error) {
	if ostype != rpc.IOSXE && ostype != rpc.IOS {
		return nil, errors.New("'show process memory' is not implemented for " + ostype)
	}
	memoryRegexp, _ := regexp.Compile(`^\s*(\S*) Pool Total:\s*(\d+) Used:\s*(\d+) Free:\s*(\d+)\s*$`)

	items := []SystemMemory{}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		matches := memoryRegexp.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		item := SystemMemory{
			Type:  matches[1],
			Total: util.Str2float64(matches[2]),
			Used:  util.Str2float64(matches[3]),
			Free:  util.Str2float64(matches[4]),
		}
		items = append(items, item)
	}
	return items, nil
}

// ParseCPU parses cli output and tries to find current CPU utilization
func (c *systemCollector) ParseCPU(ostype string, output string) (SystemCPU, error) {
	if ostype != rpc.IOSXE && ostype != rpc.IOS {
		return SystemCPU{}, errors.New("'show process cpu' is not implemented for " + ostype)
	}
	memoryRegexp, _ := regexp.Compile(`^\s*CPU utilization for five seconds: (\d+)%\/(\d+)%; one minute: (\d+)%; five minutes: (\d+)%.*$`)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		matches := memoryRegexp.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		return SystemCPU{
			FiveSeconds: util.Str2float64(matches[1]),
			Interrupts:  util.Str2float64(matches[2]),
			OneMinute:   util.Str2float64(matches[3]),
			FiveMinutes: util.Str2float64(matches[4]),
		}, nil
	}
	return SystemCPU{}, errors.New("Version string not found")
}
