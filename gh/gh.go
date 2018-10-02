package gh

import (
	"context"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

// GH is GitHub
type GH struct {
	APIKey     string
	OrgMatch   string
	TopicMatch string
	logger     log.Logger
}

// New returns a configured GH struct
func New(config map[string]string, logger log.Logger) *GH {
	return &GH{
		APIKey:     config["GithubAPIKey"],
		OrgMatch:   config["GithubOrgMatch"],
		TopicMatch: config["GithubTopicMatch"],
		logger:     logger,
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
		return nil, errors.Wrap(errors.New("missing github api key"), "environment variable not set")
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
		return nil, errors.Wrap(err, "github rate limit hit")
	}

	// else just return the error
	if err != nil {
		return nil, errors.Wrap(err, "error listing github repos")
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

	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		for _, repo := range foundRepos {
			level.Debug(gh.logger).Log("gh_repo", repo.FullName)
		}
	}

	// return the repos we care about
	return foundRepos, nil
}
