package cmd

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/andygrunwald/go-jira"

	"github.com/ikeberlein/jirc/utils"
)

var cmdBuild = &cobra.Command{
	Use:   "build -a | BUILD_NUMBER TASK [TASK]...",
	Args:  orChainArgs(hasFlag("show-apps-map"), cobra.MinimumNArgs(2)),
	Short: "Register build in Jira",
	Long: `
For each task adds label in form Server-full-x.x.x and creates release
Server-x.x.x or Server-app-x.x.x when build is for distinct server project.
`,
	Run: doBuild,
}

var (
	showAppsMap bool
)

func hasFlag(flagName string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		flag := cmd.Flag(flagName)
		if flag.Changed {
			return nil
		}
		return fmt.Errorf("requires --%s flag", flag.Name)
	}
}

func orChainArgs(checkers ...cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		var result error

		errList := make([]string, 0)
		for _, checker := range checkers {
			result = checker(cmd, args)
			if result == nil {
				return nil
			}
			errList = append(errList, result.Error())
		}

		return errors.New(strings.Join(errList, "\nor "))
	}
}

func init() {
	rootCmd.AddCommand(cmdBuild)

	flags := cmdBuild.Flags()

	flags.BoolVarP(
		&showAppsMap, "show-apps-map", "a", false, "show applicaitons map and exit",
	)
}

type updateCmd map[string]interface{}
type fieldOps map[string][]updateCmd

const (
	applicationField  = "customfield_10400"
	releaseDefaultFmt = "Server-%s"
	releaseAppFmt     = "Server-%s-%s"
	labelFmt          = "Server-full-%s"
)

var (
	buildNumber string
	tasks       []string
)

var appSuffix = map[string]string{
	"Video Line":       "videoline",
	"Knout":            "knout",
	"Social Discovery": "socialdisc",
}

var selectedComponents = []string{
	"PHP", "WEB",
}

func doBuild(cmd *cobra.Command, args []string) {
	if showAppsMap {
		fmt.Println("Applications suffix map:")
		for key, value := range appSuffix {
			fmt.Printf("  %s:%s\n", key, value)
		}
		return
	}

	options := utils.JiraClientOptions{
		EndPoint: viper.GetString("jira.url"),
		Username: viper.GetString("jira.user"),
		Password: viper.GetString("jira.pass"),
	}

	fmt.Printf(
		"Jira base URL: %s, Username: %s\n",
		options.EndPoint,
		options.Username,
	)
	fmt.Println("Going to compose jira issues in release")

	client, err := utils.NewJiraClient(options)
	if err != nil {
		log.Fatal(err)
		return
	}

	buildDate := time.Now().UTC()
	buildNumber = args[0]
	tasks = args[1:]

	for _, key := range tasks {
		release := fmt.Sprintf(releaseDefaultFmt, buildNumber)

		issue, resp, err := client.Issue.Get(key, nil)
		if err != nil {
			log.Printf("Unable to get issue %s, server response: %s\n", key, resp.Status)
			continue
		}

		if !issue.HasAnyComponent(selectedComponents) {
			log.Printf("%s: has none of selected components: [%s]\n", key, strings.Join(selectedComponents, ", "))
			continue
		}

		customFields, _, err := issue.CustomFields()
		application := customFields[applicationField]
		if application != "" {
			if suffix, ok := appSuffix[application]; ok {
				release = fmt.Sprintf(releaseAppFmt, suffix, buildNumber)
			}
		}

		updates := fieldOps{}

		if !issue.HasVersion(release) {
			project, _, err := issue.Project()
			if err != nil {
				log.Printf("Error fetching project: %s\n", err)
				continue
			}

			version := project.FindVersion(release)
			if version == nil {
				fmt.Printf("%s: create release %s\n", project.Key, release)

				version, _, err = project.CreateVersion(&jira.Version{
					Name:        release,
					Description: fmt.Sprintf("Built %s", buildDate.Format("2006-01-02 15:04")),
				})

				if err != nil {
					log.Printf("%s: unable to add release: %v\n", project.Key, err)
					continue
				}
			}

			updates["fixVersions"] = []updateCmd{
				{"add": version},
			}

			fmt.Printf("%s: adding to release %s\n", issue.Key, version.Name)
		}

		label := fmt.Sprintf(labelFmt, buildNumber)

		if !issue.HasLabel(label) {
			updates["labels"] = []updateCmd{
				{"add": label},
			}
			fmt.Printf("%s: adding label %s\n", issue.Key, label)
		}

		if len(updates) == 0 {
			continue
		}

		updateData := map[string]interface{}{
			"update": updates,
		}

		resp, err = issue.Update(updateData)
		if err != nil {
			log.Println(err)
			var buf strings.Builder
			io.Copy(&buf, resp.Body)
			log.Print(buf.String())
		}
	}
}
