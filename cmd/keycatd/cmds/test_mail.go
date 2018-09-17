package cmds

import (
	"client"
	"fmt"
	"util"

	"github.com/spf13/cobra"
)

func TestMailCmd(cmd *cobra.Command, args []string) error {
	wc, err := client.OpenWorkingCopy(false)
	if err != nil {
		return util.NewErrorf("Cannot open project: %s", err)
	}
	flags := cmd.Flags()
	checkValid, _ := flags.GetBool("check_valid")
	data := getProjectInfo(wc, checkValid)
	if client.GetOutputFormat() == "json" {
		return printJson(data)
	}
	if b, _ := flags.GetBool("root_path"); b {
		fmt.Printf("root_path: %s\n", data.RootPath)
	}
	if b, _ := flags.GetBool("project_url"); b {
		fmt.Printf("project_url: %s\n", data.ProjectUrl)
	}
	if b, _ := flags.GetBool("check_valid"); b {
		fmt.Printf("check_valid: %s\n", data.Valid)
	}
	return nil
}
