package utils

import (
	"github.com/andygrunwald/go-jira"
	"github.com/spf13/viper"
)

// JiraClientOptions are options to create new Jira client
type JiraClientOptions struct {
	EndPoint string
	Username string
	Password string
}

// JiraClient is wrapper for jira.Client
type JiraClient struct {
	*jira.Client
	Issue   *JiraIssueService
	Project *JiraProjectService
}

// NewJiraClientFromConfig create jira client with options from config
func NewJiraClientFromConfig() (*JiraClient, error) {
	return NewJiraClient(JiraClientOptions{
		EndPoint: viper.GetString("jira.url"),
		Username: viper.GetString("jira.user"),
		Password: viper.GetString("jira.pass"),
	})
}

// NewJiraClient creates jira client
func NewJiraClient(options JiraClientOptions) (*JiraClient, error) {
	tp := jira.BasicAuthTransport{
		Username: options.Username,
		Password: options.Password,
	}

	client, err := jira.NewClient(tp.Client(), options.EndPoint)
	if err != nil {
		return nil, err
	}

	result := &JiraClient{
		Client: client,
		Issue: &JiraIssueService{
			IssueService: client.Issue,
		},
		Project: &JiraProjectService{
			ProjectService: client.Project,
		},
	}

	result.Issue.client = result
	result.Project.client = result

	return result, nil
}
