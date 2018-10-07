package gh

import (
	"context"
	"fmt"

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
	client     *github.Client
	ctx        context.Context
	logger     log.Logger
}

// Githubber provides funcs to retrieve Github repositories
type Githubber interface {
	Repos() ([]*github.Repository, error)
}

// NewClient returns a new initialized GH client, context and error
func NewClient(ctx context.Context, config map[string]string) (*github.Client, context.Context, error) {

	APIKey := config["GithubAPIKey"]
	if ctx == nil {
		ctx = context.Background()
	}

	var err error

	// cannot call github w/o api key
	if APIKey == "" {
		return nil, nil, errors.Wrap(errors.New("missing github api key"), "environment variable not set")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: APIKey},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return client, ctx, err

}

// New returns a configured GH struct, it uses NewClient if no *github.Client was passed
func New(ctx context.Context, config map[string]string, logger log.Logger, client *github.Client) (Githubber, error) {

	var err error

	if client == nil {
		client, ctx, err = NewClient(ctx, config)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create github client")
		}
	}

	return &GH{
		APIKey:     config["GithubAPIKey"],
		OrgMatch:   config["GithubOrgMatch"],
		TopicMatch: config["GithubTopicMatch"],
		logger:     logger,
		client:     client,
		ctx:        ctx,
	}, nil
}

// Repos implements Githubber Github repositories that we'd like to create GoCD config repos for
func (gh *GH) Repos() ([]*github.Repository, error) {

	// make sure foundRepos is not nil
	var foundRepos = make([]*github.Repository, 0)
	var repos []*github.Repository
	var err error
	var resp *github.Response

	// protect against nil panic
	if gh.client.Repositories == nil {
		return nil, errors.Wrap(fmt.Errorf("nil pointer"), "unable to parse response from github")
	}

	// get all repos
	if gh.OrgMatch != "" {
		repos, resp, err = gh.client.Repositories.ListByOrg(gh.ctx, gh.OrgMatch, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to get repos (ListByOrg): %v", resp.Response.Status)
		}
	} else {
		repos, resp, err = gh.client.Repositories.List(gh.ctx, "", nil)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to get repos (List): %v", resp.Response.Status)
		}
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
					level.Debug(gh.logger).Log("msg", "found repo: "+*rr.FullName)
				}
			}
		}

	}

	// return the repos we care about
	return foundRepos, nil
}
