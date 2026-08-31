package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"sshutil/internal/wizard"
)

var rootCmd = &cobra.Command{
	Use:          "sshutil",
	Short:        "SSH + Tailscale device setup utility",
	Long:         "sshutil sets up SSH keys, keychain/agent, secure ssh defaults,\nhost hardening, and Tailscale on Ubuntu/Debian devices.",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWizard()
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Run the full interactive setup wizard",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWizard()
	},
}

func runWizard() error { return wizard.Run() }

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd, addKeyCmd, syncCmd, statusCmd)
}
