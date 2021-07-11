package nodetype

const (
	TemperatureSensor = iota
	LightSensor
	Cooler
	LightBulb
)

func AllSensors() []int {
	return []int{TemperatureSensor, LightSensor}
}

func RelatedActuator(sensor int) int {
	return ActuatorMappedSensor()[sensor]
}

func ActuatorMappedSensor() map[int]int {
	return map[int]int{
		LightSensor:       LightBulb,
		TemperatureSensor: Cooler,
	}
}

func IsSensor(t int) bool {
	for _, node := range AllSensors() {
		if node == t {
			return true
		}
	}

	return false
}
