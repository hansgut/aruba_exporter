package environment

import (
	log "github.com/sirupsen/logrus"
	"github.com/slashdoom/aruba_exporter/collector"
	"github.com/slashdoom/aruba_exporter/rpc"

	"github.com/prometheus/client_golang/prometheus"
)

const prefix string = "aruba_environment_"

var (
	TemperatureDesc       *prometheus.Desc
	TemperatureStatusDesc *prometheus.Desc
	PowerSupplyStatusDesc *prometheus.Desc
	FanStatusDesc         *prometheus.Desc
	FanSpeedDesc          *prometheus.Desc
	FanDirectionDesc      *prometheus.Desc
	FanRPMDesc            *prometheus.Desc
)

func init() {
	lt := []string{"target", "slot_sensor", "module_type"}
	lp := []string{"target", "power_slot", "product_number", "product_serial_number"}
	lf := []string{"target", "fan_slot"}

	TemperatureDesc = prometheus.NewDesc(prefix+"temperature", "Temperature in Celsius", lt, nil)
	TemperatureStatusDesc = prometheus.NewDesc(prefix+"temperature_status", "Status of the sensor: normal = 1", lt, nil)

	PowerSupplyStatusDesc = prometheus.NewDesc(prefix+"power_supply_status", "Status of the power supply: ok = 1", lp, nil)

	FanStatusDesc = prometheus.NewDesc(prefix+"fan_status", "Status of the fan", lf, nil)
	FanSpeedDesc = prometheus.NewDesc(prefix+"fan_speed", "Speed of the fan: 0 = slow, 1 = fast", lf, nil)
	FanDirectionDesc = prometheus.NewDesc(prefix+"fan_direction", "Direction of the fan: 0 = front-to-back, 1 = back-to-front", lf, nil)
	FanRPMDesc = prometheus.NewDesc(prefix+"fan_rpm", "RPM of the fan", lf, nil)
}

type environmentCollector struct {
}

// NewCollector creates a new collector
func NewCollector() collector.RPCCollector {
	return &environmentCollector{}
}

func (*environmentCollector) Name() string {
	return "Environment"
}

func (c *environmentCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- TemperatureDesc
	ch <- TemperatureStatusDesc
}

func (c *environmentCollector) Collect(client *rpc.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	var (
		outTemp    string
		outPower   string
		outFan     string
		itemsTemp  map[string]Environment
		itemsPower map[string]Environment
		itemsFan   map[string]Environment
		err        error
	)

	switch client.OSType {
	//case "ArubaController":
	//	out, err = client.RunCommand([]string{"show interface", "show interface counters"})
	//	if err != nil {
	//		return err
	//	}
	//case "ArubaInstant":
	//	out, err = client.RunCommand([]string{"show interface counters"})
	//	if err != nil {
	//		return err
	//	}
	case "ArubaSwitch":
		outTemp, err = client.RunCommand([]string{"show environment temperature"})
		if err != nil {
			return err
		}

		outPower, err = client.RunCommand([]string{"show environment power-supply"})
		if err != nil {
			return err
		}
	default:
		outTemp, err = client.RunCommand([]string{"show environment temperature"})
		if err != nil {
			return err
		}

		outPower, err = client.RunCommand([]string{"show environment power-supply"})
		if err != nil {
			return err
		}

		outFan, err = client.RunCommand([]string{"show environment fan"})
		if err != nil {
			return err
		}
	}

	itemsTemp, err = c.ParseTemp(client.OSType, outTemp)
	if err != nil {
		log.Warnf("Parse environments failed for %s: %s\n", labelValues[0], err.Error())
		return nil
	}

	for envName, envData := range itemsTemp {
		l := append(labelValues, envName, envData.TemperatureModuleType)

		tempStatus := 0
		if envData.TemperatureStatus == "normal" {
			tempStatus = 1
		} else {
			tempStatus = 0
		}

		ch <- prometheus.MustNewConstMetric(TemperatureDesc, prometheus.GaugeValue, envData.Temperature, l...)
		ch <- prometheus.MustNewConstMetric(TemperatureStatusDesc, prometheus.GaugeValue, float64(tempStatus), l...)
	}

	itemsPower, err = c.ParsePower(client.OSType, outPower)
	if err != nil {
		log.Warnf("Parse environments failed for %s: %s\n", labelValues[0], err.Error())
		return nil
	}

	for envName, envData := range itemsPower {
		l := append(labelValues, envName, envData.PowerSupplyProductNumber, envData.PowerSupplySerialNumber)

		powerStatus := 0
		if envData.PowerSupplyStatus == "OK" {
			powerStatus = 1
		} else {
			powerStatus = 0
		}

		ch <- prometheus.MustNewConstMetric(PowerSupplyStatusDesc, prometheus.GaugeValue, float64(powerStatus), l...)
	}

	itemsFan, err = c.ParseFan(client.OSType, outFan)
	if err != nil {
		log.Warnf("Parse environments failed for %s: %s\n", labelValues[0], err.Error())
	}

	for envName, envData := range itemsFan {
		l := append(labelValues, envName)

		fanStatus := 0
		fanDirection := 0
		speed := 0
		if envData.FanStatus == "ok" {
			fanStatus = 1
		}

		if envData.FanDirection == "front-to-back" {
			fanDirection = 0
		} else {
			fanDirection = 1
		}

		if envData.FanSpeed == "slow" {
			speed = 0
		} else {
			speed = 1
		}

		ch <- prometheus.MustNewConstMetric(FanStatusDesc, prometheus.GaugeValue, float64(fanStatus), l...)
		ch <- prometheus.MustNewConstMetric(FanDirectionDesc, prometheus.GaugeValue, float64(fanDirection), l...)
		ch <- prometheus.MustNewConstMetric(FanSpeedDesc, prometheus.GaugeValue, float64(speed), l...)
		ch <- prometheus.MustNewConstMetric(FanRPMDesc, prometheus.GaugeValue, envData.FanRPM, l...)
	}

	return nil
}
