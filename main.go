/*
gocd-seeder
*/
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alex-leonhardt/gocd-seeder/gh"
	"github.com/alex-leonhardt/gocd-seeder/gocd"
)

func help() {

	fmt.Printf(
		`Set the following env vars: 

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
`)
	os.Exit(0)
}

func main() {

	if len(os.Args) > 1 {
		if os.Args[1] == "help" {
			help()
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
				log.Println("ready to shutdown")
				shutdownStatus.finished = true
				break
			}

			// keep pulling repos and add them as they are created ...
			foundRepos, err := myGithub.Repos()

			if err != nil {
				log.Println(err)
			}

			for _, repo := range foundRepos {

				_, err := myGoCD.GetConfigRepo(hc, repo)
				if err != nil {

					if err.Error() != "404 Not Found" {
						log.Println(err)
					}

					if err.Error() == "404 Not Found" {
						newRepoConfig, err := myGoCD.CreateConfigRepo(hc, repo, githubConfig["GithubOrgMatch"])

						if err != nil {
							log.Println(err)
							continue
						}

						log.Println("created", newRepoConfig)
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
	log.Printf("received %v. shutting down. %vs grace period.\n", signal, grace)

	for i := 0; i <= grace; i++ {
		if shutdownStatus.finished == true {
			break
		}
		time.Sleep(1 * time.Second)
	}

	log.Println("good bye")

}
