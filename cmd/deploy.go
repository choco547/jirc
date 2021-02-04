package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/ikeberlein/jirc/utils"
	"github.com/spf13/cobra"
)

const testingStatus string = "Тестирование"

var (
	skipTransitionToTest   bool
	noAssignBackToReporter bool
)

var cmdDeploy = &cobra.Command{
	Use:   "deploy PROJECT BUILD_NUMBER [APPLICATION]",
	Args:  cobra.MinimumNArgs(2),
	Short: "Register deployment in Jira, update issues",
	Long: `
Sets project release status to "released", adds deploy date to release description.
Transition release issues to "Testing" status if available and assign back to reporters
if transition reached or succeeded.`,
	Run: doDeploy,
}

func init() {
	rootCmd.AddCommand(cmdDeploy)

	flags := cmdDeploy.Flags()
	flags.BoolVarP(
		&noAssignBackToReporter, "no-assign-back", "a", false, "do not assign issues back to reporters",
	)
	flags.BoolVarP(
		&skipTransitionToTest, "skip-transition", "s", false, "skip issues transition to "+testingStatus+" (implies -a)",
	)
}

var (
	projectKey  string
	application string
	release     string
)

func doDeploy(cmd *cobra.Command, args []string) {
	client, err := utils.NewJiraClientFromConfig()
	if err != nil {
		log.Fatal(err)
		return
	}

	deployDate := time.Now().UTC()
	projectKey = args[0]
	buildNumber = args[1]

	if len(args) > 2 {
		application += args[2]
		release = fmt.Sprintf(releaseAppFmt, application, buildNumber)
	} else {
		release = fmt.Sprintf(releaseDefaultFmt, buildNumber)
	}

	fmt.Println("Going to register deployment in jira")

	project, _, err := client.Project.Get(projectKey)
	if err != nil {
		log.Fatalln(err)
	}

	changed := false

	versions := project.FindUnreleasedVersionsUpto(release)
	if len(versions) == 0 {
		fmt.Printf("Project %s has no suitable releases upto %s *******\n", projectKey, release)
	}

	for _, version := range versions {
		if !version.Released {
			fmt.Printf("%s found release %s\n", project.Key, version.Name)

			description := fmt.Sprintf("%s - deployed %s", version.Description, deployDate.Format("2006-01-02 15:04"))
			_, _, err = client.Version.Update(&jira.Version{
				ID:          version.ID,
				ReleaseDate: deployDate.Format("2006-01-02"),
				Description: description,
				Released:    true,
			})

			if err != nil {
				log.Fatalf("Error releasing %s v %s: %s\n", projectKey, release, err)
			}

			changed = true
		} else {
			fmt.Printf("%s v %s is already released: %s\n", project.Key, version.Name, version.Description)
		}

		if skipTransitionToTest {
			log.Println("Skipping issues transition to", testingStatus)
		} else {
			query := fmt.Sprintf("project = %s AND fixVersion = %s", projectKey, version.Name)
			issues, _, err := client.Issue.Search(query, nil)
			if err != nil {
				log.Fatalf("Error searching '%s' issues: %v\n", query, err)
			}

			for _, issue := range issues {
				changed = deployIssue(&issue) || changed
			}
		}
	}

	if changed {
		fmt.Println("Jira has been modified")
	}
}

func deployIssue(issue *utils.JiraIssue) bool {
	changed := false
	status := issue.Fields.Status.Name

	if status == testingStatus {
		fmt.Printf("%s status [%s]\n", issue.Key, status)
	} else {
		tr, _, err := issue.GetTransitionTo(testingStatus)
		if err != nil {
			log.Printf("Unable to transition issue %s to %s, skipping it\n", issue.Key, testingStatus)
			return false
		}

		if tr == nil {
			log.Printf(
				"%s has no transition from %s to %s\n", issue.Key, issue.Fields.Status.Name, testingStatus,
			)
			return false
		}

		_, err = issue.DoTransition(tr.ID)
		if err != nil {
			log.Printf("Error transiting %s to [%s] with [%s], skipping...\n", issue.Key, tr.To.Name, tr.Name)
			return false
		}
		fmt.Printf("%s transition [%s] -> [%s] with [%s]\n", issue.Key, status, tr.To.Name, tr.Name)
		changed = true
	}

	if noAssignBackToReporter {
		log.Printf("Skip assign issue %s back to reporter\n", issue.Key)
		return changed
	}

	assignee := issue.Fields.Assignee
	reporter := issue.Fields.Reporter

	if assignee != nil && assignee.Key == reporter.Key {
		return changed
	}

	_, err := issue.ReturnToReporter()
	if err != nil {
		log.Printf(
			"Error assigning issue %s back to reporter %s\n", issue.Key, reporter.DisplayName,
		)
		return changed
	}
	fmt.Printf("%s assigned back to reporter: %s\n", issue.Key, reporter.DisplayName)

	return changed
}
