package cmds

import (
	"log"

	"github.com/keydotcat/keycatd/util"
	"github.com/spf13/cobra"
)

func VersionCmd(cmd *cobra.Command, args []string) {
	log.Printf("Keycat server version is %s (web %s)", util.GetServerVersion(), util.GetWebVersion())
}
