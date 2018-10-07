package gocd_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/alex-leonhardt/gocd-seeder/gocd"
	"github.com/go-kit/kit/log"
	"github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"
)

func TestGetConfigReposEmpty(t *testing.T) {

	ctx := context.Background()
	hc := &http.Client{}

	hs := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{"msg": "Hello World."}`)
		}))

	testGoCD := gocd.New(
		ctx,
		map[string]string{
			"GoCDURL":      hs.URL,
			"GoCDUser":     os.Getenv("GOCD_USER"),
			"GoCDPassword": os.Getenv("GOCD_PASSWORD"),
		},
		hc,
		log.NewNopLogger(),
	)

	configRepos, err := testGoCD.GetConfigRepos()

	assert.Nil(t, err)
	assert.IsType(t, []gocd.ConfigRepo{}, configRepos)
	assert.Len(t, configRepos, 0)
}

func TestGetConfigRepos(t *testing.T) {

	ctx := context.Background()
	hc := &http.Client{}

	hs := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{
				"_links": {
					"self": {
						"href": "https://ci.example.com/go/api/admin/config_repos"
					}
				},
				"_embedded": {
					"config_repos": [
						{
							"_links": {
								"self": {
									"href": "https://ci.example.com/go/api/admin/config_repos/repo1"
								},
								"doc": {
									"href": "https://api.gocd.org/#config-repos"
								},
								"find": {
									"href": "https://ci.example.com/go/api/admin/config_repos/:id"
								}
							},
							"id": "repo1",
							"plugin_id": "json.config.plugin",
							"material": {
								"type": "git",
								"attributes": {
									"url": "https://github.com/config-repo/gocd-json-config-example.git",
									"name": null,
									"branch": "master",
									"auto_update": true
								}
							},
							"configuration": [
							]
						}
					]
				}
			}`)
		}))
	defer hs.Close()

	testGoCD := gocd.New(
		ctx,
		map[string]string{
			"GoCDURL":      hs.URL,
			"GoCDUser":     os.Getenv("GOCD_USER"),
			"GoCDPassword": os.Getenv("GOCD_PASSWORD"),
		},
		hc,
		log.NewNopLogger(),
	)

	configRepos, err := testGoCD.GetConfigRepos()

	assert.Nil(t, err)
	assert.IsType(t, []gocd.ConfigRepo{}, configRepos)
	assert.Len(t, configRepos, 1)
}

func TestGetConfigRepoExists(t *testing.T) {
	ctx := context.Background()
	hc := &http.Client{}

	hs := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{
				"_links": {
					"self": {
						"href": "https://ci.example.com/go/api/admin/config_repos/repo1"
					},
					"doc": {
						"href": "https://api.gocd.org/#config-repos"
					},
					"find": {
						"href": "https://ci.example.com/go/api/admin/config_repos/:id"
					}
				},
				"id": "myprefix-repo1",
				"plugin_id": "json.config.plugin",
				"material": {
					"type": "git",
					"attributes": {
						"url": "https://github.com/config-repo/gocd-json-config-example.git",
						"name": "null",
						"branch": "master",
						"auto_update": true
					}
				},
				"configuration": [
			
				]
			}`)
		}))
	defer hs.Close()

	testGoCD := gocd.New(
		ctx,
		map[string]string{
			"GoCDURL":      hs.URL,
			"GoCDUser":     os.Getenv("GOCD_USER"),
			"GoCDPassword": os.Getenv("GOCD_PASSWORD"),
		},
		hc,
		log.NewNopLogger(),
	)

	exampleGithubRepo := &github.Repository{
		ID:   github.Int64(1234567890),
		Name: github.String("null"),
	}

	configRepo, err := testGoCD.GetConfigRepo(exampleGithubRepo, "myprefix")
	assert.Nil(t, err)
	assert.IsType(t, gocd.ConfigRepo{}, configRepo)

}

func TestGetConfigRepoNotExists(t *testing.T) {
	ctx := context.Background()
	hc := &http.Client{}

	hs := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(404)
		}))
	defer hs.Close()

	testGoCD := gocd.New(
		ctx,
		map[string]string{
			"GoCDURL":      hs.URL,
			"GoCDUser":     os.Getenv("GOCD_USER"),
			"GoCDPassword": os.Getenv("GOCD_PASSWORD"),
		},
		hc,
		log.NewNopLogger(),
	)

	exampleGithubRepo := &github.Repository{
		ID:   github.Int64(1234567890),
		Name: github.String("null"),
	}

	configRepo, err := testGoCD.GetConfigRepo(exampleGithubRepo, "myprefix")
	assert.NotNil(t, err)
	assert.IsType(t, gocd.ConfigRepo{}, configRepo)

}

func TestCreateConfigRepo(t *testing.T) {
	t.Log("not implemented")
	t.Fail()
}

func TestDeleteConfigRepo(t *testing.T) {
	t.Log("not implemented")
	t.Fail()
}

func TestReconcile(t *testing.T) {
	t.Log("not implemented")
	t.Fail()
}
