package cmd

import (
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/cmd/actuator"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/cmd/central-server"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/cmd/local-server"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/cmd/sensor"
	"github.com/saman2000hoseini/iot-platform/internal/app/iot-platform/config"
	"github.com/spf13/cobra"
)

// NewRootCommand creates a new iot-platform root command.
func NewRootCommand() *cobra.Command {
	var root = &cobra.Command{
		Use: "iot-platform",
	}

	cfg := config.Init()

	central_server.Register(root, cfg)
	local_server.Register(root, cfg)
	sensor.Register(root, cfg)
	actuator.Register(root, cfg)

	return root
}
