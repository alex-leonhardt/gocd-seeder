package gh_test

import (
	"context"
	"os"
	"testing"

	"github.com/alex-leonhardt/gocd-seeder/gh"
	"github.com/go-kit/kit/log"
	"github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"
)

// TestNew tests we can create a new Githubber
func TestNewWithoutEnv(t *testing.T) {
	g, err := gh.New(
		nil,
		map[string]string{},
		log.NewLogfmtLogger(os.Stderr),
		nil,
	)

	assert.NotNil(t, err)
	assert.Nil(t, g)

}

func TestNewWithEnv(t *testing.T) {
	config := map[string]string{
		"GithubAPIKey": "aabbcc",
	}
	g, err := gh.New(
		nil,
		config,
		log.NewLogfmtLogger(os.Stderr),
		nil,
	)

	assert.Nil(t, err)
	assert.IsType(t, &gh.GH{}, g)

}

func TestRepos(t *testing.T) {

	if os.Getenv("GITHUB_API_KEY") == "" {
		t.Log("skipping... no GITHUB_API_KEY env var specified")
		return
	}

	ctx := context.Background()

	c, err := gh.New(
		ctx,
		map[string]string{
			"GithubAPIKey":     os.Getenv("GITHUB_API_KEY"),
			"GithubOrgMatch":   os.Getenv("GITHUB_ORG"),
			"GithubTopicMatch": "ci-gocd",
		},
		log.NewLogfmtLogger(os.Stderr),
		nil,
	)

	assert.Nil(t, err)
	assert.NotNil(t, c)

	repos, err := c.Repos()
	if err != nil {
		t.Fatal(err)
	}
	assert.IsType(t, []*github.Repository{}, repos)

}
