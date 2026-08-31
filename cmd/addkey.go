package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"sshutil/internal/sshkeys"
	"sshutil/internal/state"
	"sshutil/internal/system"
	"sshutil/internal/ui"
	"sshutil/internal/wizard"
)

var addKeyCmd = &cobra.Command{
	Use:   "add-key",
	Short: "Add a per-service SSH key",
	RunE: func(cmd *cobra.Command, args []string) error {
		r := system.NewExecRunner()
		m, err := sshkeys.NewDefaultManager(r)
		if err != nil {
			return err
		}
		name, err := wizard.AskServiceName()
		if err != nil {
			return err
		}
		keyName := sshkeys.ServiceKeyName(name)
		if m.Exists(keyName) {
			ui.Warn("Key %s already exists — nothing to do", keyName)
			return nil
		}
		pw, err := wizard.AskPassphrase(keyName)
		if err != nil {
			return err
		}
		host, _ := os.Hostname()
		comment := fmt.Sprintf("%s@%s:%s", wizard.CurrentUser(), host, name)
		if err := m.Generate(keyName, comment, pw); err != nil {
			return err
		}
		if err := sshkeys.WriteServiceConfig(m.SSHDir, name); err != nil {
			return err
		}
		st, err := state.Load()
		if err != nil {
			return err
		}
		if st == nil {
			st = &state.State{}
		}
		st.UpsertKey(state.Key{Name: keyName, Path: m.KeyPath(keyName), Comment: comment})
		st.AddService(name)
		if err := state.Save(st); err != nil {
			return err
		}
		pub, err := m.PublicKey(keyName)
		if err != nil {
			return err
		}
		fmt.Println(ui.KeyPanel(keyName, pub))
		return nil
	},
}
