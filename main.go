/*
gocd-seeder
*/
package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alex-leonhardt/gocd-seeder/gh"
	"github.com/alex-leonhardt/gocd-seeder/gocd"
)

func main() {

	github := gh.GH{
		APIKey: os.Getenv("GITHUB_API_KEY"),
	}

	myGoCD := gocd.New()

	hc := &http.Client{
		Timeout: 5 * time.Second,
	}

	var shutdown = false
	var grace = 65 * time.Second

	go func() {

		for {

			// when we receive a signal, we set shutdown to true, which will break the endless loop
			// and we should be good to stop as we should no longer be doing any processing, the
			// timing is critical, as it _must_ be at least larger than the poll interval
			if shutdown {
				break
			}

			// keep pulling repos and add them as they are created ...
			foundRepos, err := github.Repos("ci-gocd")
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
						newRepoConfig, err := myGoCD.CreateConfigRepo(hc, repo)

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

	shutdown = true
	log.Printf("Received %v. Shutting down. %v grace period.\n", signal, grace)
	time.Sleep(grace)
	log.Println("Good bye.")

}
