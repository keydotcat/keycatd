package main

import (
	"github.com/keydotcat/keycatd/cmd/keycatd/cmds"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "keycatd",
		Short: "Keycat service to provide identities management",
		Run:   cmds.RunCmd,
	}

	rootCmd.PersistentFlags().String("config", "", "Configuration file (default is ./keycatd.yaml)")
	var testMailCmd = &cobra.Command{
		Use:   "testmail",
		Short: "Send a test mail to verify email parameters",
		Run:   cmds.TestMailCmd,
	}
	testMailCmd.Flags().String("to", "", "Who to send the test mail to")
	rootCmd.AddCommand(testMailCmd)

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the keycatd version",
		Run:   cmds.VersionCmd,
	}
	rootCmd.AddCommand(versionCmd)
	rootCmd.Execute()
}
