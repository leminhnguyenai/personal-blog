package apis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Commit struct {
	URL         string `json:"url"`
	Sha         string `json:"sha"`
	NodeID      string `json:"node_id"`
	HTMLURL     string `json:"html_url"`
	CommentsURL string `json:"comments_url"`
	Commit      struct {
		URL    string `json:"url"`
		Author struct {
			Name  string    `json:"name"`
			Email string    `json:"email"`
			Date  time.Time `json:"date"`
		} `json:"author"`
		Committer struct {
			Name  string    `json:"name"`
			Email string    `json:"email"`
			Date  time.Time `json:"date"`
		} `json:"committer"`
		Message string `json:"message"`
		Tree    struct {
			URL string `json:"url"`
			Sha string `json:"sha"`
		} `json:"tree"`
		CommentCount int `json:"comment_count"`
		Verification struct {
			Verified   bool        `json:"verified"`
			Reason     string      `json:"reason"`
			Signature  interface{} `json:"signature"`
			Payload    interface{} `json:"payload"`
			VerifiedAt interface{} `json:"verified_at"`
		} `json:"verification"`
	} `json:"commit"`
	Author struct {
		Login             string `json:"login"`
		ID                int    `json:"id"`
		NodeID            string `json:"node_id"`
		AvatarURL         string `json:"avatar_url"`
		GravatarID        string `json:"gravatar_id"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		OrganizationsURL  string `json:"organizations_url"`
		ReposURL          string `json:"repos_url"`
		EventsURL         string `json:"events_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		Type              string `json:"type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"author"`
	Committer struct {
		Login             string `json:"login"`
		ID                int    `json:"id"`
		NodeID            string `json:"node_id"`
		AvatarURL         string `json:"avatar_url"`
		GravatarID        string `json:"gravatar_id"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		OrganizationsURL  string `json:"organizations_url"`
		ReposURL          string `json:"repos_url"`
		EventsURL         string `json:"events_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		Type              string `json:"type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"committer"`
	Parents []struct {
		URL string `json:"url"`
		Sha string `json:"sha"`
	} `json:"parents"`
}

type Commits []Commit

type Branch struct {
	Name   string `json:"name"`
	Commit struct {
		Sha string `json:"sha"`
		URL string `json:"url"`
	} `json:"commit"`
	Protected  bool `json:"protected"`
	Protection struct {
		RequiredStatusChecks struct {
			EnforcementLevel string   `json:"enforcement_level"`
			Contexts         []string `json:"contexts"`
		} `json:"required_status_checks"`
	} `json:"protection"`
	ProtectionURL string `json:"protection_url"`
}

type Branches []Branch

func GetGithubBranches(ctx context.Context, owner string, repo string) ([]Branch, error) {
	apiKey := os.Getenv("GITHUB_API_KEY")
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/branches", owner, repo)

	type Response struct {
		branches []Branch
		error    error
	}

	respChan := make(chan Response)

	go func() {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			respChan <- Response{error: err}
		}

		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			respChan <- Response{error: err}
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			respChan <- Response{error: err}
		}

		var branches []Branch

		err = json.Unmarshal(body, &branches)
		if err != nil {
			respChan <- Response{error: err}
		}

		respChan <- Response{branches, nil}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("Unexpected error occured\n")
		case res := <-respChan:
			return res.branches, res.error
		}
	}
}

func GetGithubLatestCommit(ctx context.Context, owner, repo string) (Commit, error) {
	apiKey := os.Getenv("GITHUB_API_KEY")
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits?per_page=1", owner, repo)

	type Response struct {
		commit Commit
		error  error
	}

	respChan := make(chan Response)

	go func() {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			respChan <- Response{error: err}
		}

		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			respChan <- Response{error: err}
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			respChan <- Response{error: err}
		}

		var commits []Commit

		err = json.Unmarshal(body, &commits)
		if err != nil {
			respChan <- Response{error: err}
		}

		if len(commits) == 0 {
			respChan <- Response{error: fmt.Errorf("Failed to retrieve commits from https://api.github.com\n")}
		}

		respChan <- Response{commits[0], nil}
	}()

	for {
		select {
		case <-ctx.Done():
			return Commit{}, fmt.Errorf("Unexpected error occured\n")
		case res := <-respChan:
			return res.commit, res.error
		}
	}
}
