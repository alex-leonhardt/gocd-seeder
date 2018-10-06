package gh_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
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

	hs := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, `{}`)
			}))
	defer hs.Close()

	ghurl, err := url.ParseRequestURI(hs.URL)

	ghc := &github.Client{
		BaseURL: ghurl,
	}
	// ghc := github.NewClient(hc)
	ctx := context.Background()

	c, err := gh.New(
		ctx,
		map[string]string{
			"GithubAPIKey": "aabbcc",
		},
		log.NewLogfmtLogger(os.Stderr),
		ghc,
	)

	assert.Nil(t, err)
	assert.NotNil(t, c)

	repos, err := c.Repos()
	if err != nil {
		t.Fatal(err)
	}
	assert.IsType(t, []*github.Repository{}, repos)

}
