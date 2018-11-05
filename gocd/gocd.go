package gocd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
)

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

// GoCD provides GoCD funcs
type GoCD struct {
	URL      string
	User     string
	Password string
	hc       *http.Client
	logger   log.Logger
}

// ConfigRepoInterface provides implementations that interact with GoCD
type ConfigRepoInterface interface {
	GetConfigRepos() ([]ConfigRepo, error)
	GetConfigRepo(*github.Repository, string) (ConfigRepo, error)
	CreateConfigRepo(*github.Repository, string) (ConfigRepo, error)
	DeleteConfigRepo(*ConfigRepo, string) (*http.Response, error)
}

/*

************************************************************************************************

 */

// NewRequest creates a new request to the GoCD server and returns it, it populates it with the necessary headers and auth creds
func (g *GoCD) NewRequest(verb string, path string, headers http.Header, body io.Reader) (*http.Request, error) {

	if headers == nil {
		headers = http.Header{
			"Accept":       []string{"application/vnd.go.cd.v1+json"},
			"Content-Type": []string{"application/json"},
		}
	}

	if path != "" {
		path = g.URL + "/" + path
	} else {
		path = g.URL
	}

	req, err := http.NewRequest(verb, path, body)

	if g.User != "" && g.Password != "" {
		req.SetBasicAuth(g.User, g.Password)
	}

	req.Header = headers

	if err != nil {
		return nil, errors.Wrap(err, "error creating http request")
	}

	return req, nil
}

// GetConfigRepos populates the GoCD struct with config repos
func (g *GoCD) GetConfigRepos() ([]ConfigRepo, error) {

	req, err := g.NewRequest(http.MethodGet, "", nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating http request for GetConfigRepos")
	}

	resp, err := g.hc.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error doing http request")
	}
	defer resp.Body.Close()

	// not entirely sure why this gets an EOF error when doing this the same way as GetConfigRepo
	// so for now we'll read in the entire response, and then unmarshal
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading response body")
	}
	repos := AllConfigRepos{}
	err = json.Unmarshal(body, &repos)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshaling json from response body")
	}

	return repos.Embedded["config_repos"], nil
}

// GetConfigRepo retrieves an existing config repo
func (g *GoCD) GetConfigRepo(repo *github.Repository, prefix string) (ConfigRepo, error) {

	if prefix != "" {
		prefix = fmt.Sprintf("%s-", prefix)
	}

	id := fmt.Sprintf("%s%s", prefix, *repo.Name)

	req, err := g.NewRequest("GET", id, nil, nil)
	if err != nil {
		return ConfigRepo{}, errors.Wrap(err, "error creating request to retrieve gocd config repo")
	}

	resp, err := g.hc.Do(req)
	if resp == nil {
		return ConfigRepo{}, errors.Wrap(err, "error retrieving a response from gocd")
	}
	if err != nil || resp.StatusCode > 399 {
		if resp.StatusCode > 399 {
			return ConfigRepo{}, errors.New(resp.Status)
		}
		return ConfigRepo{}, errors.Wrap(err, "error executing request to retrieve gocd config repo")
	}
	defer resp.Body.Close()

	var cfgrepo ConfigRepo
	jd := json.NewDecoder(resp.Body)
	jd.Decode(&cfgrepo)
	return cfgrepo, nil
}

// CreateConfigRepo creates a previously non-existent config repo
func (g *GoCD) CreateConfigRepo(repo *github.Repository, prefix string) (ConfigRepo, error) {

	if prefix != "" {
		prefix = fmt.Sprintf("%s-", prefix)
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
		return ConfigRepo{}, errors.Wrap(err, "error marshalling json to create gocd config repo")
	}

	req, err := g.NewRequest("POST", "", nil, bytes.NewBuffer(postBody))
	if err != nil {
		return ConfigRepo{}, errors.Wrap(err, "error creating http post request")
	}

	resp, err := g.hc.Do(req)

	if err != nil || resp.StatusCode > 399 {
		if resp.StatusCode > 399 {
			if err == nil {
				err = fmt.Errorf("invalid response status")
			}
			msg, _ := ioutil.ReadAll(resp.Body)
			return ConfigRepo{}, errors.Wrap(err, fmt.Sprintf("%v %v", resp.Status, string(msg)))
		}

		return ConfigRepo{}, errors.Wrap(err, "error executing http post request")
	}

	var cfgrepo ConfigRepo
	jd := json.NewDecoder(resp.Body)
	jd.Decode(&cfgrepo)

	return cfgrepo, nil
}

// DeleteConfigRepo removes a config repo from GoCD
func (g *GoCD) DeleteConfigRepo(repo *ConfigRepo, prefix string) (*http.Response, error) {
	if prefix != "" {
		prefix = fmt.Sprintf("%s-", prefix)
	}

	req, err := g.NewRequest("DELETE", repo.ID, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating new http request")
	}

	resp, err := g.hc.Do(req)
	if err != nil {
		return resp, errors.Wrap(err, "error executing http request to delete a gocd config repo")
	}
	if resp.StatusCode > 399 {
		return resp, errors.Wrap(fmt.Errorf(resp.Status), "invalid response status")
	}

	return resp, nil
}

/*

************************************************************************************************

 */

// New returns a GoCD Client
func New(ctx context.Context, config map[string]string, hc *http.Client, logger log.Logger) ConfigRepoInterface {
	return &GoCD{
		URL:      config["GoCDURL"] + "/go/api/admin/config_repos",
		User:     config["GoCDUser"],
		Password: config["GoCDPassword"],
		hc:       hc,
		logger:   logger,
	}
}

// Reconcile ensures that repos that have been removed from Github, or are no longer found when
// they had the topic to match removed, are also removed from GoCD
func Reconcile(g ConfigRepoInterface, logger log.Logger, prefix string, gocdRepos []ConfigRepo, ghRepos []*github.Repository) error {

	githubSeen := map[string]bool{}
	for _, ghRepo := range ghRepos {
		githubSeen[*ghRepo.Name] = true
	}

	for _, gocdRepo := range gocdRepos {
		if githubSeen[gocdRepo.Material.Attributes.Name] != true ||
			!githubSeen[gocdRepo.Material.Attributes.Name] {
			_, err := g.DeleteConfigRepo(&gocdRepo, prefix)
			if err != nil {
				return errors.Wrap(err, "error deleting config repo "+gocdRepo.ID)
			}
			level.Info(logger).Log("msg", fmt.Sprintf("removed gocd config repo %s for %s (%s)", gocdRepo.ID, gocdRepo.Material.Attributes.Name, gocdRepo.Material.Attributes.URL))
		}
	}

	return nil
}
