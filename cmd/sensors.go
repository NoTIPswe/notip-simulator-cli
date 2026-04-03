package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/NoTIPswe/notip-simulator-cli/internal/client"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var sensorsCmd = &cobra.Command{
	Use:   "sensors",
	Short: "Manage sensors attached to gateways",
}

// ── add ───────────────────────────────────────────────────────────────────────

var sensorsAddCmd = &cobra.Command{
	Use:   "add <gateway-int-id>",
	Short: "Add a sensor to a gateway (uses the numeric gateway ID, not the UUID)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		gatewayID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("gateway-id must be a numeric ID: %w", err)
		}

		req := client.AddSensorRequest{}
		req.Type, _ = cmd.Flags().GetString("type")
		req.MinRange, _ = cmd.Flags().GetFloat64("min")
		req.MaxRange, _ = cmd.Flags().GetFloat64("max")
		req.Algorithm, _ = cmd.Flags().GetString("algorithm")

		spinner, _ := pterm.DefaultSpinner.Start(
			fmt.Sprintf("Adding %s sensor to gateway %d...", req.Type, gatewayID),
		)
		sensor, err := client.New(simulatorURL).AddSensor(gatewayID, req)
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
	Use:   "list <gateway-int-id>",
	Short: "List all sensors for a gateway (uses the numeric gateway ID, not the UUID)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		gatewayID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("gateway-id must be a numeric ID: %w", err)
		}

		spinner, _ := pterm.DefaultSpinner.Start(
			fmt.Sprintf("Fetching sensors for gateway %d...", gatewayID),
		)
		sensors, err := client.New(simulatorURL).ListSensors(gatewayID)
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
	Use:   "delete <sensor-int-id>",
	Short: "Delete a sensor by its numeric ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sensorID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("sensor-id must be a numeric ID: %w", err)
		}

		spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Deleting sensor %d...", sensorID))
		if err := client.New(simulatorURL).DeleteSensor(sensorID); err != nil {
			spinner.Fail("Failed to delete sensor")
			return err
		}
		spinner.Success(fmt.Sprintf("Sensor %d deleted", sensorID))
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
			strconv.FormatInt(s.ID, 10),
			s.SensorID,
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
		if err := sensorsAddCmd.MarkFlagRequired(f); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
