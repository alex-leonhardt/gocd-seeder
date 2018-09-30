/*
GoCD-Seeder scans a GitHub org for repositories that contain a pre-specified "topic" (default: ci-gocd), if a repo is found, it will create a GoCD config repo, which will make GoCD poll the repository for the file "ci.gocd.yaml" and create (a) new pipeline/s basaed on the config in that file.
*/

package main

import (
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/alex-leonhardt/gocd-seeder/gh"
	"github.com/alex-leonhardt/gocd-seeder/gocd"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
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
GITHUB_API_KEY  (e.g.: 1235436)
GITHUB_ORG      (e.g.: gooflix)

Optional:
=========
GITHUB_TOPIC    (default: ci-gocd)
GOCD_URL        (default: http://localhost:8081)
GOCD_USER       (e.g.: admin)
GOCD_PASSWORD   (e.g.: admin)
HTTP_STATS_IP   (default: "")
HTTP_STATS_PORT (default: 9090)
LOG_LEVEL       (e.g.: DEBUG)
`)
	os.Exit(0)
}

// helper func to return a default when os.Getenv fails to find a value
func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

// ------------------------------------------------

// provide goroutines stats for expvar
func goroutines() interface{} {
	return runtime.NumGoroutine()
}

// provide uptime in seconds for expvar
func uptime() interface{} {
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
		"GithubAPIKey":     os.Getenv("GITHUB_API_KEY"),
		"GithubOrgMatch":   os.Getenv("GITHUB_ORG"),
		"GithubTopicMatch": getenv("GITHUB_TOPIC", "ci-gocd"),
	}

	gocdConfig := map[string]string{
		"GoCDURL":      getenv("GOCD_URL", "http://localhost:8081"),
		"GoCDUser":     os.Getenv("GOCD_USER"),
		"GoCDPassword": os.Getenv("GOCD_PASSWORD"),
	}

	httpConfig := map[string]string{
		"StatsIP":   getenv("HTTP_STATS_IP", ""),
		"StatsPort": getenv("HTTP_STATS_PORT", "9090"),
	}

	// ------------------------------------------------

	logger := log.NewJSONLogger(os.Stdout)
	logger = log.With(logger, "timestamp", log.DefaultTimestampUTC)
	logger = log.With(logger, "source", log.Caller(3))

	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		logger = level.NewFilter(logger, level.AllowDebug())
	} else {
		logger = level.NewFilter(logger, level.AllowInfo())
	}

	myGithub := gh.New(githubConfig)
	myGoCD := gocd.New(gocdConfig)

	hc := &http.Client{
		Timeout: 10 * time.Second,
	}

	doneChan := make(chan bool)
	ticker := time.NewTicker(55 * time.Second)

	// ------------------------------------------------

	expvar.Publish("Uptime", expvar.Func(uptime))
	expvar.Publish("Goroutines", expvar.Func(goroutines))

	go func() {
		level.Info(logger).Log("msg", http.ListenAndServe(fmt.Sprintf("%s:%s", httpConfig["StatsIP"], httpConfig["StatsPort"]), nil))
	}()

	// ------------------------------------------------

	go func() {

		for {

			// keep pulling repos and add them as they are created ...
			foundGitHubRepos, err := myGithub.Repos()

			if err != nil {
				level.Error(logger).Log("msg", err)
			}

			// -------------------------------------

			for _, repo := range foundGitHubRepos {

				_, err := myGoCD.GetConfigRepo(hc, repo, githubConfig["GithubOrgMatch"])
				if err != nil {

					if err.Error() != "404 Not Found" {
						level.Warn(logger).Log("msg", err)
					}

					if err.Error() == "404 Not Found" {
						newRepoConfig, err := myGoCD.CreateConfigRepo(hc, repo, githubConfig["GithubOrgMatch"])

						if err != nil {
							level.Error(logger).Log("msg", err)
							continue
						}

						level.Info(logger).Log("msg", "created "+newRepoConfig.ID)
					}

				}

			}

			// -------------------------------------

			// get all gocd config repos
			foundGoCDConfigRepos, err := myGoCD.GetConfigRepos(hc)
			if err != nil {
				level.Error(logger).Log("msg", err)
			}

			err = gocd.Reconcile(logger, myGoCD, githubConfig["GithubOrgMatch"], hc, foundGoCDConfigRepos, foundGitHubRepos)
			if err != nil {
				level.Error(logger).Log("msg", err)
			}

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
