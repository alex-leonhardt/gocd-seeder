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
		hs.Client(),
		log.NewNopLogger(),
	)

	configRepos, err := testGoCD.GetConfigRepos()

	assert.Nil(t, err)
	assert.IsType(t, []gocd.ConfigRepo{}, configRepos)
	assert.Len(t, configRepos, 0)
}

func TestGetConfigRepos(t *testing.T) {

	ctx := context.Background()
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
		hs.Client(),
		log.NewNopLogger(),
	)

	configRepos, err := testGoCD.GetConfigRepos()

	assert.Nil(t, err)
	assert.IsType(t, []gocd.ConfigRepo{}, configRepos)
	assert.Len(t, configRepos, 1)
}

func TestGetConfigRepoExists(t *testing.T) {
	ctx := context.Background()
	hs := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{
				"_links": {
					"self": {
						"href": "https://ci.example.com/go/api/admin/config_repos/repo-2"
					},
					"doc": {
						"href": "https://api.gocd.org/#config-repos"
					},
					"find": {
						"href": "https://ci.example.com/go/api/admin/config_repos/:id"
					}
				},
				"id": "myprefix-one",
				"plugin_id": "json.config.plugin",
				"material": {
					"type": "git",
					"attributes": {
						"url": "https://github.com/config-repo/gocd-json-config-example2.git",
						"name": null,
						"branch": "master",
						"auto_update": true
					}
				},
				"configuration": [
					{
						"key": "pattern",
						"value": "*.myextension"
					}
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
		hs.Client(),
		log.NewNopLogger(),
	)

	exampleGithubRepo := &github.Repository{
		ID:       github.Int64(1234567890),
		Name:     github.String("one"),
		CloneURL: github.String(""),
		Topics:   []string{"ci-gocd"},
	}

	configRepo, err := testGoCD.GetConfigRepo(exampleGithubRepo, "myprefix")
	assert.Nil(t, err)
	assert.IsType(t, gocd.ConfigRepo{}, configRepo)

}

func TestGetConfigRepoNotExists(t *testing.T) {
	ctx := context.Background()
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
		hs.Client(),
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
	ctx := context.Background()
	hs := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{
				"_links": {
					"self": {
						"href": "https://ci.example.com/go/api/admin/config_repos/repo-2"
					},
					"doc": {
						"href": "https://api.gocd.org/#config-repos"
					},
					"find": {
						"href": "https://ci.example.com/go/api/admin/config_repos/:id"
					}
				},
				"id": "myprefix-one",
				"plugin_id": "json.config.plugin",
				"material": {
					"type": "git",
					"attributes": {
						"url": "https://github.com/config-repo/gocd-json-config-example2.git",
						"name": null,
						"branch": "master",
						"auto_update": true
					}
				},
				"configuration": [
					{
						"key": "pattern",
						"value": "*.myextension"
					}
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
		hs.Client(),
		log.NewNopLogger(),
	)

	exampleGithubRepo := &github.Repository{
		ID:       github.Int64(1234567890),
		Name:     github.String("one"),
		CloneURL: github.String("http://localhost/clone/repo/one"),
		Topics:   []string{"ci-gocd"},
	}

	configRepo, err := testGoCD.CreateConfigRepo(exampleGithubRepo, "myprefix")
	assert.Nil(t, err)
	assert.IsType(t, gocd.ConfigRepo{}, configRepo)
	assert.Equal(t, "myprefix-one", configRepo.ID)
}

func TestDeleteConfigRepoError400(t *testing.T) {
	ctx := context.Background()
	hs := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(400)
		}))
	defer hs.Close()

	testGoCD := gocd.New(
		ctx,
		map[string]string{
			"GoCDURL":      hs.URL,
			"GoCDUser":     os.Getenv("GOCD_USER"),
			"GoCDPassword": os.Getenv("GOCD_PASSWORD"),
		},
		hs.Client(),
		log.NewNopLogger(),
	)

	exampleConfigRepo := &gocd.ConfigRepo{
		ID: "myprefix-one",
	}

	resp, err := testGoCD.DeleteConfigRepo(exampleConfigRepo, "myprefix")
	assert.NotNil(t, err)
	assert.EqualError(t, err, "invalid response status: 400 Bad Request")
	assert.Equal(t, resp.StatusCode, 400)
}

func TestDeleteConfigRepoUnknownHost(t *testing.T) {
	ctx := context.Background()
	hs := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{"error": "error"}`)
		}))
	defer hs.Close()

	testGoCD := gocd.New(
		ctx,
		map[string]string{
			"GoCDURL":      "http://unknownhost:9090/",
			"GoCDUser":     os.Getenv("GOCD_USER"),
			"GoCDPassword": os.Getenv("GOCD_PASSWORD"),
		},
		hs.Client(),
		log.NewNopLogger(),
	)

	exampleConfigRepo := &gocd.ConfigRepo{
		ID: "myprefix-one",
	}

	_, err := testGoCD.DeleteConfigRepo(exampleConfigRepo, "myprefix")
	assert.NotNil(t, err)
	assert.Regexp(t, "error executing http request to delete a gocd config repo: Delete .* no such host", err)
}

func TestDeleteConfigRepoOK(t *testing.T) {
	ctx := context.Background()
	hs := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{
				"message": "The config repo 'myprefix-one' was deleted successfully."
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
		hs.Client(),
		log.NewNopLogger(),
	)

	exampleConfigRepo := &gocd.ConfigRepo{
		ID: "myprefix-one",
	}

	resp, err := testGoCD.DeleteConfigRepo(exampleConfigRepo, "myprefix")
	assert.Nil(t, err)
	if resp == nil {
		t.Fatal("FATAL >>> response is nil")
	}
	assert.Equal(t, 200, resp.StatusCode)
}
