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
				fmt.Fprintf(w, `[
{
	"id": "000",
	"node_id": "???",
	"name": "000",
	"fullname": "000/000",
	"description": "...no desc...",
	"git_url": "git://..."
}
]`)
			}))
	defer hs.Close()

	ghurl, err := url.Parse(hs.URL + "/")
	t.Log(*ghurl)

	// c := &Client{client: httpClient, BaseURL: baseURL, UserAgent: userAgent, UploadURL: uploadURL}

	ghc := &github.Client{
		BaseURL:   ghurl,
		UserAgent: "go-github",
		UploadURL: ghurl,
	}

	// copy from github.go
	ghc.Activity = &github.ActivityService{}
	ghc.Admin = &github.AdminService{}
	ghc.Apps = &github.AppsService{}
	ghc.Authorizations = &github.AuthorizationsService{}
	ghc.Checks = &github.ChecksService{}
	ghc.Gists = &github.GistsService{}
	ghc.Git = &github.GitService{}
	ghc.Gitignores = &github.GitignoresService{}
	ghc.Issues = &github.IssuesService{}
	ghc.Licenses = &github.LicensesService{}
	ghc.Marketplace = &github.MarketplaceService{Stubbed: true}
	ghc.Migrations = &github.MigrationService{}
	ghc.Organizations = &github.OrganizationsService{}
	ghc.Projects = &github.ProjectsService{}
	ghc.PullRequests = &github.PullRequestsService{}
	ghc.Reactions = &github.ReactionsService{}
	ghc.Repositories = &github.RepositoriesService{}
	ghc.Search = &github.SearchService{}
	ghc.Teams = &github.TeamsService{}
	ghc.Users = &github.UsersService{}
	// end
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

	t.Log("ghc:", ghc, "c:", c, "ctx:", ctx)

	assert.Nil(t, err)
	assert.NotNil(t, c)

	repos, err := c.Repos()
	if err != nil {
		t.Fatal(err)
	}
	assert.IsType(t, []*github.Repository{}, repos)

}
