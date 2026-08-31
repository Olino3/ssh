package cmd

import (
	"github.com/spf13/cobra"

	"sshutil/internal/sshkeys"
	"sshutil/internal/system"
	"sshutil/internal/tailscale"
	"sshutil/internal/ui"
	"sshutil/internal/wizard"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Regenerate tailnet ssh Host entries from tailscale status",
	RunE: func(cmd *cobra.Command, args []string) error {
		r := system.NewExecRunner()
		peers, err := tailscale.Peers(r)
		if err != nil {
			ui.ErrorBox("tailscale status failed", err.Error())
			return err
		}
		m, err := sshkeys.NewDefaultManager(r)
		if err != nil {
			return err
		}
		username := wizard.CurrentUser()
		blocks := make([]string, 0, len(peers))
		for _, p := range peers {
			blocks = append(blocks, sshkeys.TailnetPeerBlock(p.Alias(), p.MagicDNS(), username))
		}
		if err := sshkeys.WriteTailnetConfig(m.SSHDir, blocks); err != nil {
			return err
		}
		ui.Success("Wrote %d tailnet host entries", len(peers))
		return nil
	},
}
