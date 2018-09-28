/*
gocd-seeder
*/
package main

import (
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
GITHUB_TOPIC (e.g.: ci-gocd)
GITHUB_API_KEY (e.g.: 1235436)
GITHUB_ORG (e.g.: gooflix)

GOCD_URL (e.g.: http://localhost:8081)


Optional:
=========
GOCD_USER (e.g.: admin)
GOCD_PASSWORD (e.g.: admin)
LOG_LEVEL (e.g.: DEBUG)
`)
	os.Exit(0)
}

func main() {

	if len(os.Args) > 1 {
		if os.Args[1] == "help" {
			help()
		}
		if os.Args[1] == "version" {
			version()
		}
	}

	githubConfig := map[string]string{
		"GithubAPIKey":     os.Getenv("GITHUB_API_KEY"),
		"GithubOrgMatch":   os.Getenv("GITHUB_ORG"),
		"GithubTopicMatch": os.Getenv("GITHUB_TOPIC"),
	}

	gocdConfig := map[string]string{
		"GoCDURL":      os.Getenv("GOCD_URL"),
		"GoCDUser":     os.Getenv("GOCD_USER"),
		"GoCDPassword": os.Getenv("GOCD_PASSWORD"),
	}

	logger := log.NewJSONLogger(os.Stdout)
	logger = log.With(logger, "timestamp", log.DefaultTimestampUTC)

	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		logger = level.NewFilter(logger, level.AllowDebug())
	} else {
		logger = level.NewFilter(logger, level.AllowInfo())
	}

	myGithub := gh.New(githubConfig)
	myGoCD := gocd.New(gocdConfig)

	hc := &http.Client{
		Timeout: 15 * time.Second,
	}

	type shutdown struct {
		started  bool
		finished bool
	}

	var shutdownStatus = shutdown{
		started:  false,
		finished: false,
	}

	var grace = 65

	go func() {

		for {

			// when we receive a signal, we set shutdown to true, which will break the endless loop
			// and we should be good to stop as we should no longer be doing any processing, the
			// timing is critical, as it _must_ be at least larger than the poll interval
			if shutdownStatus.started {
				level.Info(logger).Log("msg", "ready to shutdown")
				shutdownStatus.finished = true
				break
			}

			// keep pulling repos and add them as they are created ...
			foundRepos, err := myGithub.Repos()

			if err != nil {
				level.Error(logger).Log("err", err)
			}

			for _, repo := range foundRepos {

				_, err := myGoCD.GetConfigRepo(hc, repo, githubConfig["GithubOrgMatch"])
				if err != nil {

					if err.Error() != "404 Not Found" {
						level.Warn(logger).Log("err", err)
					}

					if err.Error() == "404 Not Found" {
						newRepoConfig, err := myGoCD.CreateConfigRepo(hc, repo, githubConfig["GithubOrgMatch"])

						if err != nil {
							level.Error(logger).Log("err", err)
							continue
						}

						level.Info(logger).Log("created", newRepoConfig.ID)
					}

				}

			}

			time.Sleep(55 * time.Second)

		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	signal := <-signals

	shutdownStatus.started = true
	level.Info(logger).Log("msg", fmt.Sprintf("received %v. shutting down. %vs grace period", signal, grace))

	for i := 0; i <= grace; i++ {
		if shutdownStatus.finished == true {
			break
		}
		time.Sleep(1 * time.Second)
	}

	level.Debug(logger).Log("numGoRoutines", runtime.NumGoroutine())
	level.Info(logger).Log("msg", "good bye")

}
