package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ahur-system/sysmedic/pkg/cli"
)

var (
	version = "1.0.0"
	commit  = "dev"
	date    = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "sysmedic",
		Short: "Cross-platform Linux server monitoring CLI tool",
		Long: `SysMedic is a comprehensive server monitoring tool that tracks system
and user resource usage spikes with daemon capabilities and intelligent alerting.`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
		Run: func(cmd *cobra.Command, args []string) {
			// Default command shows dashboard
			cli.ShowDashboard()
		},
	}

	// Daemon commands
	daemonCmd := &cobra.Command{
		Use:   "daemon",
		Short: "Manage the SysMedic monitoring daemon",
		Long:  "Start, stop, or check the status of the SysMedic background monitoring daemon",
	}

	daemonStartCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the monitoring daemon",
		Run: func(cmd *cobra.Command, args []string) {
			cli.StartDaemon()
		},
	}

	daemonStopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the monitoring daemon",
		Run: func(cmd *cobra.Command, args []string) {
			cli.StopDaemon()
		},
	}

	daemonStatusCmd := &cobra.Command{
		Use:   "status",
		Short: "Check daemon status",
		Run: func(cmd *cobra.Command, args []string) {
			cli.DaemonStatus()
		},
	}

	daemonCmd.AddCommand(daemonStartCmd, daemonStopCmd, daemonStatusCmd)

	// Config commands
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage SysMedic configuration",
		Long:  "View and modify SysMedic configuration settings",
	}

	configShowCmd := &cobra.Command{
		Use:   "show",
		Short: "Display current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			cli.ShowConfig()
		},
	}

	configSetCmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set configuration value",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			cli.SetConfig(args[0], args[1])
		},
	}

	configSetUserCmd := &cobra.Command{
		Use:   "set-user [username] [key] [value]",
		Short: "Set user-specific configuration",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			cli.SetUserConfig(args[0], args[1], args[2])
		},
	}

	configCmd.AddCommand(configShowCmd, configSetCmd, configSetUserCmd)

	// Reports commands
	reportsCmd := &cobra.Command{
		Use:   "reports",
		Short: "View system and user activity reports",
		Long:  "Generate reports on system alerts, user activity, and resource usage patterns",
		Run: func(cmd *cobra.Command, args []string) {
			period, _ := cmd.Flags().GetString("period")
			cli.ShowReports(period)
		},
	}

	reportsCmd.Flags().StringP("period", "p", "hourly", "Report period (hourly, daily, weekly)")

	reportsUsersCmd := &cobra.Command{
		Use:   "users",
		Short: "Show detailed user activity reports",
		Run: func(cmd *cobra.Command, args []string) {
			top, _ := cmd.Flags().GetInt("top")
			user, _ := cmd.Flags().GetString("user")
			period, _ := cmd.Flags().GetString("period")
			cli.ShowUserReports(top, user, period)
		},
	}

	reportsUsersCmd.Flags().IntP("top", "t", 0, "Show top N users")
	reportsUsersCmd.Flags().StringP("user", "u", "", "Show specific user")
	reportsUsersCmd.Flags().StringP("period", "p", "hourly", "Report period")

	reportsCmd.AddCommand(reportsUsersCmd)

	// Add all commands to root
	rootCmd.AddCommand(daemonCmd, configCmd, reportsCmd)

	// Execute the CLI
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
