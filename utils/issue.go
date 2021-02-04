package utils

import "github.com/andygrunwald/go-jira"

// JiraIssueService is wrapper for jira.IssueService
type JiraIssueService struct {
	*jira.IssueService
	client *JiraClient
}

// JiraIssue is wrapper for jira.Issue
type JiraIssue struct {
	*jira.Issue
	client *JiraClient
}

// Get returns a full representation of the issue for the given issue key.
func (s *JiraIssueService) Get(issueID string, options *jira.GetQueryOptions) (*JiraIssue, *jira.Response, error) {
	issue, resp, err := s.IssueService.Get(issueID, options)
	if err != nil {
		return nil, resp, err
	}

	result := &JiraIssue{
		Issue:  issue,
		client: s.client,
	}

	return result, resp, err
}

// Search will search for tickets according to the jql
func (s *JiraIssueService) Search(jql string, options *jira.SearchOptions) ([]JiraIssue, *jira.Response, error) {
	issues, resp, err := s.IssueService.Search(jql, options)
	if err != nil {
		return nil, resp, err
	}

	result := make([]JiraIssue, len(issues))

	for i := range issues {
		result[i] = JiraIssue{
			Issue:  &issues[i],
			client: s.client,
		}
	}

	return result, resp, nil
}

// HasVersion checks whether issue has version
func (issue *JiraIssue) HasVersion(version string) bool {
	for _, v := range issue.Fields.FixVersions {
		if v.Name == version {
			return true
		}
	}
	return false
}

// HasLabel checks whether issue has label
func (issue *JiraIssue) HasLabel(label string) bool {
	for _, l := range issue.Fields.Labels {
		if l == label {
			return true
		}
	}
	return false
}

// HasComponent check whether issue has named component
func (issue *JiraIssue) HasComponent(name string) bool {
	for _, c := range issue.Fields.Components {
		if c.Name == name {
			return true
		}
	}
	return false
}

// HasAnyComponent check whether issue has any named components
func (issue *JiraIssue) HasAnyComponent(names []string) bool {
	for _, name := range names {
		if issue.HasComponent(name) {
			return true
		}
	}
	return false
}

// Update updates an issue from a JSON representation.
func (issue *JiraIssue) Update(data map[string]interface{}) (*jira.Response, error) {
	return issue.client.Issue.UpdateIssue(issue.Key, data)
}

// CustomFields returns a map of customfield_* keys with string values
func (issue *JiraIssue) CustomFields() (jira.CustomFields, *jira.Response, error) {
	return issue.client.Issue.GetCustomFields(issue.ID)
}

// Project returns a full representation of the project for the given issue.
func (issue *JiraIssue) Project() (*JiraProject, *jira.Response, error) {
	return issue.client.Project.Get(issue.Fields.Project.ID)
}

// GetTransitions gets a list of the transitions possible for this issue by the current user,
// along with fields that are required and their types.
func (issue *JiraIssue) GetTransitions() ([]jira.Transition, *jira.Response, error) {
	return issue.client.Issue.GetTransitions(issue.ID)
}

// GetTransitionTo gets transition to named state if any
func (issue *JiraIssue) GetTransitionTo(state string) (*jira.Transition, *jira.Response, error) {
	transitions, resp, err := issue.GetTransitions()
	if err != nil {
		return nil, resp, err
	}

	for _, tr := range transitions {
		if tr.To.Name == state {
			return &tr, nil, nil
		}
	}

	return nil, nil, nil
}

// DoTransition performs a transition on an issue.
func (issue *JiraIssue) DoTransition(transitionID string) (*jira.Response, error) {
	return issue.client.Issue.DoTransition(issue.ID, transitionID)
}

// UpdateAssignee updates the user assigned to work on the given issue
func (issue *JiraIssue) UpdateAssignee(assignee *jira.User) (*jira.Response, error) {
	return issue.client.Issue.UpdateAssignee(issue.ID, assignee)
}

// ReturnToReporter sets the user assigned to work on the issue to reporter
func (issue *JiraIssue) ReturnToReporter() (*jira.Response, error) {
	reporter := issue.Fields.Reporter
	return issue.UpdateAssignee(reporter)
}
