package utils

import (
	"strconv"
	"strings"

	"github.com/andygrunwald/go-jira"
	semver "github.com/hashicorp/go-version"
)

// JiraProjectService is wrapper for jira.ProjectService
type JiraProjectService struct {
	*jira.ProjectService
	client *JiraClient
}

// JiraProject is wrapper for jira.Project
type JiraProject struct {
	*jira.Project
	client *JiraClient
}

// Get returns a full representation of the project for the given prokect id.
func (s *JiraProjectService) Get(projectID string) (*JiraProject, *jira.Response, error) {
	project, resp, err := s.ProjectService.Get(projectID)
	if err != nil {
		return nil, resp, err
	}

	result := &JiraProject{
		Project: project,
		client:  s.client,
	}

	return result, resp, err
}

// FindVersion looks for version in project
func (project *JiraProject) FindVersion(version string) *jira.Version {
	for _, v := range project.Versions {
		if v.Name == version {
			return &v
		}
	}
	return nil
}

// FindUnreleasedVersionsUpto looks for unreleased versions in project
func (project *JiraProject) FindUnreleasedVersionsUpto(version string) []*jira.Version {
	var result = make([]*jira.Version, 0)

	prefix, reqNumber := splitVersion(version)

	reqVersion, err := semver.NewVersion(reqNumber)
	if err != nil {
		return result
	}

	for i, v := range project.Versions {
		if v.Name == version {
			result = append(result, &project.Versions[i])
			continue
		}

		if v.Released {
			continue
		}

		curPrefix, curNumber := splitVersion(v.Name)
		if curPrefix != prefix {
			continue
		}

		curVersion, err := semver.NewVersion(curNumber)
		if err != nil {
			continue
		}

		if curVersion.LessThan(reqVersion) {
			result = append(result, &project.Versions[i])
		}
	}

	return result
}

func splitVersion(version string) (string, string) {
	sepIdx := strings.LastIndex(version, "-")
	if sepIdx < 0 {
		return "", ""
	}
	prefix := version[0:sepIdx]
	number := version[sepIdx+1:]
	return prefix, number
}

// CreateVersion creates version (aka release) in project
func (project *JiraProject) CreateVersion(version *jira.Version) (*jira.Version, *jira.Response, error) {
	version.ProjectID, _ = strconv.Atoi(project.ID)
	return project.client.Version.Create(version)
}
