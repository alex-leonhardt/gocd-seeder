package gh_test

import (
	_ "net/http/httptest"
	"os"
	"testing"

	"github.com/alex-leonhardt/gocd-seeder/gh"
	"github.com/go-kit/kit/log"
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

// TODO: need to break off creation of client + auth from actual retrieval of repos
// TODO: change gh struct to contain github client, so we can replace it when we're testing ..
// TODO: create fake github client,
