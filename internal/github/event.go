package github

import (
	"encoding/json"
	"fmt"
	"os"
)

// PullRequest identifies the Pull Request an event refers to.
type PullRequest struct {
	Owner  string
	Repo   string
	Number int
}

// event is the minimal shape we need from the GitHub webhook payload.
type event struct {
	Number      int `json:"number"`
	PullRequest struct {
		Number int `json:"number"`
	} `json:"pull_request"`
	Repository struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
}

// ParseEvent extracts the Pull Request coordinates from the GitHub event
// payload located at path.
func ParseEvent(path string) (PullRequest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return PullRequest{}, fmt.Errorf("reading event file: %w", err)
	}

	var e event
	if err := json.Unmarshal(data, &e); err != nil {
		return PullRequest{}, fmt.Errorf("unmarshaling event JSON: %w", err)
	}

	number := e.Number
	if number == 0 {
		number = e.PullRequest.Number
	}
	if number == 0 {
		return PullRequest{}, fmt.Errorf("could not determine pull request number from event payload")
	}
	if e.Repository.Owner.Login == "" || e.Repository.Name == "" {
		return PullRequest{}, fmt.Errorf("could not determine repository owner/name from event payload")
	}

	return PullRequest{
		Owner:  e.Repository.Owner.Login,
		Repo:   e.Repository.Name,
		Number: number,
	}, nil
}
