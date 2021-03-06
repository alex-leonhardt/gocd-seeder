// gocd-seeder scans a GitHub org for repositories that contain a pre-specified "topic" (default: ci-gocd), if a repo is found, it will create a GoCD config repo, which will make GoCD poll the repository for the file "ci.gocd.yaml" and create (a) new pipeline/s basaed on the config in that file.
package main

import (
	"expvar"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/alex-leonhardt/gocd-seeder/gh"
	"github.com/alex-leonhardt/gocd-seeder/gocd"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
)

var (
	versionString string
	buildDateTime = time.Now().UTC()
	buildHost, _  = os.Hostname()
	startTime     = time.Now().UTC()
)

func version() {
	if versionString == "" {
		versionString = "UNKNOWN (Add the following when compiling the app: -ldflags \"-X main.versionString=`git rev-list --max-count=1 --branches master --abbrev-commit`\")"
	}
	fmt.Println(os.Args[0] + "\n")
	fmt.Println("Built on: \t" + fmt.Sprintf("%v", buildDateTime))
	fmt.Println("Build host: \t" + buildHost)
	fmt.Println("Version: \t" + fmt.Sprintf("%s", versionString))
	os.Exit(0)
}

func help() {

	fmt.Printf(
		`Set the following environment vars: 

Required:
=========
GITHUB_API_KEY  (e.g.: 1235436, use GITHUB_SECRETS_PATH when deploying to kubernetes)
GITHUB_ORG      (e.g.: gooflix)

Optional:
=========
GITHUB_TOPIC    (default: ci-gocd)
GOCD_URL        (default: http://localhost:8081)
GOCD_USER       (e.g.: admin, use GOCD_SECRETS_PATH when deploying to kubernetes)
GOCD_PASSWORD   (e.g.: admin, use GOCD_SECRETS_PATH when deploying to kubernetes)
HTTP_STATS_IP   (default: "")
HTTP_STATS_PORT (default: 9090)
LOG_LEVEL       (e.g.: DEBUG)

Kubernetes:
===========

Optional, howver, recommended to use as env vars can be seen in clear in the dashboard :facepalm:

GITHUB_SECRETS_PATH (e.g: /secrets/github)
-- if set, must contain a file "api_key" with the github api key

GOCD_SECRETS_PATH (e.g.: /secrets/gocd)
-- if set, must contain a file "gocd_password" with the password corresponding to the gocd_user
-- if set, must contain a file "gocd_user"     with the username to use to connect to GoCD
`)
	os.Exit(0)
}

// Getenv returns a default when os.Getenv fails to find a value
func Getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return strings.Trim(value, "\n")
}

// ConfigReader provides a interface to be able to swap out the implementation for ReadSecretFromFile
type ConfigReader interface {
	Read() (string, error)
}

// ConfigFileReader implements ConfigReader
type ConfigFileReader struct {
	path string
}

// Read implements the Read() method of ConfigReader (interface)
func (c ConfigFileReader) Read() (string, error) {
	value, err := ioutil.ReadFile(c.path)
	return string(value), err
}

// ReadSecretFromFile takes a ConfigReader, invokes Read() removes the trailing NewLine char if one is present, returns: string, error
func ReadSecretFromFile(reader ConfigReader) (string, error) {
	value, err := reader.Read()
	return strings.Trim(value, "\n"), err
}

// ------------------------------------------------

// Goroutines provide goroutines stats for expvar
func Goroutines() interface{} {
	return runtime.NumGoroutine()
}

// Uptime provide uptime in seconds for expvar
func Uptime() interface{} {
	uptime := time.Since(startTime).Seconds()
	return int64(uptime)
}

// ------------------------------------------------

func main() {

	if len(os.Args) > 1 {
		if os.Args[1] == "help" {
			help()
		}
		if os.Args[1] == "version" {
			version()
		}
	}

	// ------------------------------------------------

	githubConfig := map[string]string{
		"GithubAPIKey":     Getenv("GITHUB_API_KEY", ""),
		"GithubOrgMatch":   Getenv("GITHUB_ORG", "ORG_DOES_NOT_EXIST_MUST_SET_VALUE_FROM_ENV"),
		"GithubTopicMatch": Getenv("GITHUB_TOPIC", "ci-gocd"),
	}

	gocdConfig := map[string]string{
		"GoCDURL":      Getenv("GOCD_URL", "http://localhost:8081"),
		"GoCDUser":     Getenv("GOCD_USER", ""),
		"GoCDPassword": Getenv("GOCD_PASSWORD", ""),
	}

	httpConfig := map[string]string{
		"StatsIP":   Getenv("HTTP_STATS_IP", ""),
		"StatsPort": Getenv("HTTP_STATS_PORT", "9090"),
	}

	githubSecretsPath := Getenv("GITHUB_SECRETS_PATH", "")
	gocdSecretsPath := Getenv("GOCD_SECRETS_PATH", "")

	// ------------------------------------------------

	logger := log.NewJSONLogger(os.Stdout)
	logger = log.With(logger, "timestamp", log.DefaultTimestampUTC)
	logger = log.With(logger, "source", log.Caller(5))

	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		logger = level.NewFilter(logger, level.AllowDebug())
	} else {
		logger = level.NewFilter(logger, level.AllowInfo())
	}

	if githubSecretsPath != "" {
		var value string
		var err error
		reader := ConfigFileReader{
			path: githubSecretsPath + "/api_key",
		}
		// read config file and set to GithubAPIKey in githubConfig map
		value, err = ReadSecretFromFile(reader)
		githubConfig["GithubAPIKey"] = value
		if err != nil {
			level.Error(logger).Log("msg", err)
			panic(err)
		}
	}

	if gocdSecretsPath != "" {
		var value string
		var err error
		pwReader := ConfigFileReader{
			path: gocdSecretsPath + "/gocd_password",
		}
		usrReader := ConfigFileReader{
			path: gocdSecretsPath + "/gocd_user",
		}
		// read gocd_password file and set to GoCDPassword in gocdConfig map
		value, err = ReadSecretFromFile(pwReader)
		gocdConfig["GoCDPassword"] = value
		if err != nil {
			level.Error(logger).Log("msg", err)
			panic(err)
		}
		// read gocd_user file and set to GoCDUser in gocdConfig map
		value, err = ReadSecretFromFile(usrReader)
		gocdConfig["GoCDUser"] = value
		if err != nil {
			level.Error(logger).Log("msg", err)
			panic(err)
		}
	}

	// --------------------------------------------------

	defaultHTTPClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	myGithub, err := gh.New(nil, githubConfig, logger, nil)
	if err != nil {
		level.Error(logger).Log("msg", err)
	}

	myGoCD := gocd.New(nil, gocdConfig, defaultHTTPClient, logger)

	doneChan := make(chan bool)
	ticker := time.NewTicker(55 * time.Second)

	// ------------------------------------------------

	expvar.Publish("Uptime", expvar.Func(Uptime))
	expvar.Publish("Goroutines", expvar.Func(Goroutines))

	go func() {
		level.Info(logger).Log("msg", http.ListenAndServe(fmt.Sprintf("%s:%s", httpConfig["StatsIP"], httpConfig["StatsPort"]), nil))
	}()

	// ------------------------------------------------

	level.Info(logger).Log("version", versionString, "msg", "app started")

	go func() {

		for {

			// keep pulling repos and add them as they are created ...
			foundGitHubRepos, err := myGithub.Repos()

			if err != nil {
				level.Error(logger).Log("msg", errors.Wrap(err, "error retrieving github repos"))
			}

			// -------------------------------------
			if foundGitHubRepos != nil {

				for _, repo := range foundGitHubRepos {

					_, err := myGoCD.GetConfigRepo(repo, githubConfig["GithubOrgMatch"])

					if err != nil {

						if err.Error() != "404 Not Found" {
							level.Warn(logger).Log("msg", errors.Wrap(err, "error retrieving gocd config repo for "+*repo.FullName))
						}

						if err.Error() == "404 Not Found" {

							newRepoConfig, err := myGoCD.CreateConfigRepo(repo, githubConfig["GithubOrgMatch"])

							if err != nil {

								level.Error(logger).Log("msg", errors.Wrap(err, "error creating config repo for "+*repo.FullName))
								continue
							}

							level.Info(logger).Log("msg", "created "+newRepoConfig.ID)
						}

					}

				}

				// -------------------------------------

				// get all gocd config repos
				foundGoCDConfigRepos, err := myGoCD.GetConfigRepos()
				if err != nil {
					level.Error(logger).Log("msg", errors.Wrap(err, "error retrieving all config repos from gocd"))
				}

				err = gocd.Reconcile(myGoCD, logger, githubConfig["GithubOrgMatch"], foundGoCDConfigRepos, foundGitHubRepos)
				if err != nil {
					level.Error(logger).Log("msg", errors.Wrap(err, "error reconciling gocd config repos with github repos"))
				}

			}

			level.Debug(logger).Log("msg", fmt.Sprintf("found repo count: %v", len(foundGitHubRepos)))
			// -------------------------------------

			// use a ticker to continue, and a done channel to break out, it's neater
			select {
			case <-doneChan:
				level.Info(logger).Log("msg", "shutting down goroutine")
				break
			case <-ticker.C:
				level.Debug(logger).Log("msg", "ticker still ticking")
			}

		}
	}()

	// ------------------------------------------------

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	signal := <-signals // this blocks until a signal was caught

	// stop the ticker in the go routine
	level.Info(logger).Log("msg", fmt.Sprintf("received %v; shutting down", signal))
	ticker.Stop()
	time.Sleep(1 * time.Second)
	doneChan <- true
	time.Sleep(1 * time.Second)
	level.Debug(logger).Log("numGoRoutines", runtime.NumGoroutine())
	level.Info(logger).Log("msg", "good bye")

}
