package cmd

import (
	"fmt"

	"github.com/NoTIPswe/notip-simulator-cli/internal/client"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var sensorsCmd = &cobra.Command{
	Use:   "sensors",
	Short: "Manage sensors attached to gateways",
}

func resolveGatewayID(c *client.Client, input string) (string, error) {
	gw, err := c.GetGateway(input)
	if err != nil {
		return "", fmt.Errorf("gateway must be a valid gateway UUID: %w", err)
	}
	return gw.ID, nil
}

// ── add ───────────────────────────────────────────────────────────────────────

var sensorsAddCmd = &cobra.Command{
	Use:   "add <gateway-uuid>",
	Short: "Add a sensor to a gateway",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(simulatorURL).WithContext(cmd.Context())
		gatewayID, err := resolveGatewayID(c, args[0])
		if err != nil {
			return err
		}

		req := client.AddSensorRequest{}
		req.Type, _ = cmd.Flags().GetString("type")
		req.MinRange, _ = cmd.Flags().GetFloat64("min")
		req.MaxRange, _ = cmd.Flags().GetFloat64("max")
		req.Algorithm, _ = cmd.Flags().GetString("algorithm")

		spinner := startSpinner(
			fmt.Sprintf("Adding %s sensor to gateway %s...", req.Type, gatewayID),
		)
		sensor, err := c.AddSensor(gatewayID, req)
		if err != nil {
			spinner.Fail("Failed to add sensor")
			return err
		}
		spinner.Success("Sensor added")
		printSensorTable([]client.Sensor{*sensor})
		return nil
	},
}

// ── list ──────────────────────────────────────────────────────────────────────

var sensorsListCmd = &cobra.Command{
	Use:   "list <gateway-uuid>",
	Short: "List sensors for a gateway",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(simulatorURL).WithContext(cmd.Context())
		gatewayID, err := resolveGatewayID(c, args[0])
		if err != nil {
			return err
		}

		spinner := startSpinner(
			fmt.Sprintf("Fetching sensors for gateway %s...", gatewayID),
		)
		sensors, err := c.ListSensors(gatewayID)
		if err != nil {
			spinner.Fail("Failed to fetch sensors")
			return err
		}
		spinner.Success("Sensors retrieved")

		if len(sensors) == 0 {
			pterm.Info.Println("No sensors found for this gateway.")
			return nil
		}
		printSensorTable(sensors)
		return nil
	},
}

// ── delete ────────────────────────────────────────────────────────────────────

var sensorsDeleteCmd = &cobra.Command{
	Use:   "delete <sensor-uuid>",
	Short: "Delete a sensor by UUID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sensorID := args[0]

		spinner := startSpinner(fmt.Sprintf("Deleting sensor %s...", sensorID))
		if err := client.New(simulatorURL).WithContext(cmd.Context()).DeleteSensor(sensorID); err != nil {
			spinner.Fail("Failed to delete sensor")
			return err
		}
		spinner.Success(fmt.Sprintf("Sensor %s deleted", sensorID))
		return nil
	},
}

// ── helpers ───────────────────────────────────────────────────────────────────

func printSensorTable(sensors []client.Sensor) {
	if len(sensors) == 0 {
		return
	}
	tableData := pterm.TableData{{"ID", "UUID", "Type", "Min", "Max", "Algorithm"}}
	for _, s := range sensors {
		tableData = append(tableData, []string{
			s.ID,
			s.ID,
			s.Type,
			fmt.Sprintf("%.2f", s.MinRange),
			fmt.Sprintf("%.2f", s.MaxRange),
			s.Algorithm,
		})
	}
	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render() //nolint:errcheck
}

// ── init ──────────────────────────────────────────────────────────────────────

func init() {
	rootCmd.AddCommand(sensorsCmd)
	sensorsCmd.AddCommand(sensorsAddCmd, sensorsListCmd, sensorsDeleteCmd)

	sensorsAddCmd.Flags().String("type", "", "Sensor type: temperature|humidity|pressure|movement|biometric (required)")
	sensorsAddCmd.Flags().Float64("min", 0, "Minimum range value (required)")
	sensorsAddCmd.Flags().Float64("max", 100, "Maximum range value (required)")
	sensorsAddCmd.Flags().String("algorithm", "", "Generation algorithm: uniform_random|sine_wave|spike|constant (required)")
	for _, f := range []string{"type", "min", "max", "algorithm"} {
		mustMarkRequired(sensorsAddCmd, f)
	}
}
