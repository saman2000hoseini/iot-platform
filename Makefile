export APP=iot-platform
export LDFLAGS="-w -s"

central-server:
	go run -ldflags $(LDFLAGS) ./cmd/iot-platform central-server

local-server:
	go run -ldflags $(LDFLAGS) ./cmd/iot-platform local-server

actuator:
	go run -ldflags $(LDFLAGS) ./cmd/iot-platform actuator

sensor:
	go run -ldflags $(LDFLAGS) ./cmd/iot-platform sensor

build:
	go build -ldflags $(LDFLAGS) ./cmd/iot-platform

install:
	go install -ldflags $(LDFLAGS) ./cmd/iot-platform
