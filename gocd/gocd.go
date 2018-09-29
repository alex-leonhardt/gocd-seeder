package gocd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
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
	Links         map[string]map[string]string `json:"_links,omitempty"`
	ID            string                       `json:"id"`
	PluginID      string                       `json:"plugin_id"`
	Material      repoMaterial                 `json:"material"`
	Configuration []interface{}                `json:"configuration,omitempty"`
}

// AllConfigRepos contains the response from GoCD containing all config repos
type AllConfigRepos struct {
	Links    map[string]map[string]string `json:"_links,omitempty"`
	Embedded map[string][]ConfigRepo      `json:"_embedded,omitempty"`
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

	// not entirely sure why this gets an EOF error when doing this the same way as GetConfigRepo
	// so for now we'll read in the entire response, and then unmarshal
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	repos := AllConfigRepos{}
	err = json.Unmarshal(body, &repos)
	if err != nil {
		return nil, err
	}

	return repos.Embedded["config_repos"], nil
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
func (g *GoCD) DeleteConfigRepo(hc *http.Client, repo *ConfigRepo, prefix string) error {
	if prefix != "" {
		prefix = fmt.Sprintf("%s-", prefix)
	}

	headers := http.Header{
		"Accept":       []string{"application/vnd.go.cd.v1+json"},
		"Content-Type": []string{"application/json"},
	}

	req, err := http.NewRequest(http.MethodDelete, g.URL+"/"+repo.ID, nil)
	if err != nil {
		return err
	}
	req.Header = headers
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode > 399 {
		return fmt.Errorf("%v", resp.Status)
	}

	return nil
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

// Reconcile ensures that repos that have been removed from Github, or are no longer found when
// they had the topic to match removed, are also removed from GoCD
func Reconcile(logger log.Logger, g *GoCD, prefix string, hc *http.Client, gocdRepos []ConfigRepo, ghRepos []*github.Repository) error {

	githubSeen := map[string]bool{}
	for _, ghRepo := range ghRepos {
		githubSeen[*ghRepo.Name] = true
	}

	for _, gocdRepo := range gocdRepos {
		if githubSeen[gocdRepo.Material.Attributes.Name] != true ||
			!githubSeen[gocdRepo.Material.Attributes.Name] {
			err := g.DeleteConfigRepo(hc, &gocdRepo, prefix)
			if err != nil {
				return err
			}
			level.Info(logger).Log("msg", "removed gocd config repo for "+gocdRepo.Material.Attributes.Name+" ("+gocdRepo.Material.Attributes.URL+")")
		}
	}

	return nil
}
