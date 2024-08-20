package environment

type Environment struct {
	TemperatureSlotSensor string
	Temperature           float64
	TemperatureStatus     string
	TemperatureModuleType string

	PowerSupplySlot          string
	PowerSupplyStatus        string
	PowerSupplyProductNumber string
	PowerSupplySerialNumber  string

	FanSlot      string
	FanSpeed     string
	FanStatus    string
	FanRPM       float64
	FanDirection string
}
