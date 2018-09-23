/*
gocd-seeder
*/
package main

import (
	"log"
	"net/http"
	"os"
	_ "os/signal"
	"time"

	"github.com/alex-leonhardt/gocd-seeder/gh"
	"github.com/alex-leonhardt/gocd-seeder/gocd"
)

func main() {

	// e2d89483a6f6b3f3a58df277b7334dc1f3b7b174
	github := gh.GH{
		APIKey: os.Getenv("GITHUB_API_KEY"),
	}

	myGoCD := gocd.New()

	hc := &http.Client{
		Timeout: 5 * time.Second,
	}

	go func() {

		for {

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

			time.Sleep(60 * time.Second)
		}
	}()

	for {
		time.Sleep(10 * time.Second)
	}

}
