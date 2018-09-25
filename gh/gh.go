package gh

import (
	"context"
	"errors"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// GH is GitHub
type GH struct {
	APIKey     string
	OrgMatch   string
	TopicMatch string
}

// New returns a configured GH struct
func New(config map[string]string) *GH {
	return &GH{
		APIKey:     config["GithubAPIKey"],
		OrgMatch:   config["GithubOrgMatch"],
		TopicMatch: config["GithubTopicMatch"],
	}
}

// Repos retruns Github repositories that we'd like to create GoCD config repos for
func (gh *GH) Repos() ([]*github.Repository, error) {

	var client *github.Client
	var foundRepos []*github.Repository
	var repos []*github.Repository
	var err error

	ctx := context.Background()

	// cannot call github w/o api key
	if gh.APIKey == "" {
		return nil, errors.New("missing github api Key")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gh.APIKey},
	)
	tc := oauth2.NewClient(ctx, ts)
	client = github.NewClient(tc)

	// get all repos
	if gh.OrgMatch != "" {
		repos, _, err = client.Repositories.ListByOrg(ctx, gh.OrgMatch, nil)
	} else {
		repos, _, err = client.Repositories.List(ctx, "", nil)
	}

	// return specific error when we hit the rate limit
	if _, ok := err.(*github.RateLimitError); ok {
		return nil, errors.New("hit rate limit")
	}

	// else just return the error
	if err != nil {
		return nil, err
	}

	// filter out only the repos we're interested in and return the slice
	for _, rr := range repos {

		// dont bother if there are no topics on the repo
		if len(rr.Topics) > 0 {

			// if we have > 0 topics, iterate over them until we have a match and add to the foundRepos slice
			for _, topic := range rr.Topics {
				if topic == gh.TopicMatch {
					foundRepos = append(foundRepos, rr)
				}
			}
		}

	}

	// return the repos we care about
	return foundRepos, nil
}
