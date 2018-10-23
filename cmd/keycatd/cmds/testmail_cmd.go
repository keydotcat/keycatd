package cmds

import (
	"log"

	"github.com/keydotcat/keycatd/api"
	"github.com/spf13/cobra"
)

func TestMailCmd(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	cfgfile, err := flags.GetString("config")
	if err != nil {
		log.Fatalf("Could not get config file: %s", err)
		return
	}
	to, err := flags.GetString("to")
	switch {
	case err != nil:
		log.Fatalf("Could not get destination recipient: %s", err)
		return
	case len(to) == 0:
		log.Fatalf("Who to send the test email to?")
		return
	}
	c := processConf(cfgfile)
	if err != nil {
		log.Fatalf("Could not parse configuration: %s", err)
		return
	}
	if err = api.SendTestEmail(c, to); err != nil {
		log.Fatalf("Could not send test email: %s", err)
	} else {
		log.Println("Mail sent")
	}
}
