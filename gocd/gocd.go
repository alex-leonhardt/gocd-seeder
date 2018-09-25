package gocd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/go-github/github"
)

// GoCD provides GoCD funcs
type GoCD struct {
	URL      string
	User     string
	Password string
}

type repoAttributes struct {
	URL        string `json:"url"`
	Name       string `json:"name,omitempty"`
	Branch     string `json:"branch"`
	AutoUpdate bool   `json:"auto_update"`
}

type repoMaterial struct {
	Type       string         `json:"type"`
	Attributes repoAttributes `json:"attributes"`
}

// ConfigRepo is a representation of a GoCD config repo
type ConfigRepo struct {
	Links          map[string]string `json:"_links,omitempty"`
	ID             string            `json:"id"`
	PluginID       string            `json:"plugin_id"`
	Material       repoMaterial      `json:"material"`
	Configurations []interface{}     `json:"configurations,omitempty"`
}

/*

************************************************************************************************

 */

// ConfigReposRetriever retrieves all gocd config repositories
type ConfigReposRetriever interface {
	GetConfigRepos(*http.Client) ([]ConfigRepo, error)
}

// ConfigRepoRetriever retrieves a gocd config repository
type ConfigRepoRetriever interface {
	GetConfigRepo(*http.Client, *github.Repository, string) (ConfigRepo, error)
}

// ConfigRepoCreator creates a gocd config repo
type ConfigRepoCreator interface {
	CreateConfigRepo(*http.Client, *github.Repository, string) (ConfigRepo, error)
}

// ConfigRepoDeleter deletes a gocd config repo
type ConfigRepoDeleter interface {
	DeleteConfigRepo(*http.Client, *github.Repository, string) error
}

/*

************************************************************************************************

 */

// GetConfigRepos populates the GoCD struct with config repos
func (g *GoCD) GetConfigRepos(hc *http.Client) ([]ConfigRepo, error) {
	headers := http.Header{
		"Accept":       []string{"application/vnd.go.cd.v1+json"},
		"Content-Type": []string{"application/json"},
	}

	req, err := http.NewRequest("GET", g.URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header = headers
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var repos []ConfigRepo
	jd := json.NewDecoder(resp.Body)
	jd.Decode(&repos)

	return repos, fmt.Errorf("not implemented")
}

// GetConfigRepo retrieves an existing config repo
func (g *GoCD) GetConfigRepo(hc *http.Client, repo *github.Repository, prefix string) (ConfigRepo, error) {

	if prefix != "" {
		prefix = fmt.Sprintf("%s-", prefix)
	}

	headers := http.Header{
		"Accept":       []string{"application/vnd.go.cd.v1+json"},
		"Content-Type": []string{"application/json"},
	}

	id := fmt.Sprintf("%s%s", prefix, *repo.Name)
	req, err := http.NewRequest(http.MethodGet, g.URL+"/"+id, nil)
	if err != nil {
		return ConfigRepo{}, err
	}
	req.Header = headers
	resp, err := hc.Do(req)

	if resp == nil {
		return ConfigRepo{}, err
	}
	if err != nil || resp.StatusCode > 399 {
		if resp.StatusCode > 399 {
			return ConfigRepo{}, fmt.Errorf(resp.Status)
		}
		return ConfigRepo{}, err
	}
	defer resp.Body.Close()

	var cfgrepo ConfigRepo
	jd := json.NewDecoder(resp.Body)
	jd.Decode(&cfgrepo)
	return cfgrepo, nil
}

// CreateConfigRepo creates a previously non-existent config repo
func (g *GoCD) CreateConfigRepo(hc *http.Client, repo *github.Repository, prefix string) (ConfigRepo, error) {

	if prefix != "" {
		prefix = fmt.Sprintf("%s-", prefix)
	}

	headers := http.Header{
		"Accept":       []string{"application/vnd.go.cd.v1+json"},
		"Content-Type": []string{"application/json"},
	}

	newRepoConfig := ConfigRepo{
		ID:       fmt.Sprintf("%s%s", prefix, *repo.Name),
		PluginID: "yaml.config.plugin",
		Material: repoMaterial{
			Type: "git",
			Attributes: repoAttributes{
				AutoUpdate: true,
				Branch:     "master",
				Name:       *repo.Name,
				URL:        *repo.CloneURL,
			},
		},
	}

	postBody, err := json.Marshal(newRepoConfig)
	if err != nil {
		return ConfigRepo{}, err
	}

	req, err := http.NewRequest(http.MethodPost, g.URL, bytes.NewBuffer(postBody))
	req.Header = headers

	// set basic user/pass for auth to GoCD
	if g.User != "" && g.Password != "" {
		req.SetBasicAuth(g.User, g.Password)
	}

	if err != nil {
		return ConfigRepo{}, err
	}

	req.Header = headers
	resp, err := hc.Do(req)

	if err != nil || resp.StatusCode > 399 {
		if resp.StatusCode > 399 {
			msg, _ := ioutil.ReadAll(resp.Body)
			return ConfigRepo{}, fmt.Errorf("%v %v", resp.Status, string(msg))
		}

		return ConfigRepo{}, err
	}

	var cfgrepo ConfigRepo
	jd := json.NewDecoder(resp.Body)
	jd.Decode(&cfgrepo)

	return cfgrepo, nil
}

// DeleteConfigRepo removes a config repo from GoCD
func (g *GoCD) DeleteConfigRepo(hc *http.Client, repo *github.Repository, prefix string) error {
	if prefix != "" {
		prefix = fmt.Sprintf("%s-", prefix)
	}
	return fmt.Errorf("not implemented")
}

/*

************************************************************************************************

 */

// New returns a GoCD Client
func New(config map[string]string) *GoCD {
	return &GoCD{
		URL:      config["GoCDURL"] + "/go/api/admin/config_repos",
		User:     config["GoCDUser"],
		Password: config["GoCDPassword"],
	}
}
