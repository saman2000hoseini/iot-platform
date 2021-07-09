export APP=iot-platform
export LDFLAGS="-w -s"

run:
	go run -ldflags $(LDFLAGS) ./cmd/iot-platform server

build:
	go build -ldflags $(LDFLAGS) ./cmd/iot-platform

install:
	go install -ldflags $(LDFLAGS) ./cmd/iot-platform
