package cmd

import (
	"io"
	"log"
	"strings"

	"github.com/ikeberlein/jirc/utils"
	"github.com/spf13/cobra"
)

var cmdPingJira = &cobra.Command{
	Use:   "ping",
	Args:  cobra.MaximumNArgs(0),
	Short: "Ping Jira",
	Long:  `Ping jira server`,
	Run:   doPingJira,
}

func init() {
	rootCmd.AddCommand(cmdPingJira)
}

func doPingJira(cmd *cobra.Command, args []string) {
	client, err := utils.NewJiraClientFromConfig()
	if err != nil {
		log.Fatal(err)
		return
	}

	user, resp, err := client.User.GetSelf()
	if err != nil {
		var buf strings.Builder
		io.Copy(&buf, resp.Response.Body)
		log.Println(buf.String())
		log.Fatalln("Unable to login to jira server:", resp.Status)
	}

	log.Printf("Logged in as %s <%s>\n", user.DisplayName, user.EmailAddress)
}
