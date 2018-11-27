[![GoDoc](https://godoc.org/github.com/alex-leonhardt/gocd-seeder?status.svg)](https://godoc.org/github.com/alex-leonhardt/gocd-seeder) [![Build Status](https://travis-ci.com/alex-leonhardt/gocd-seeder.svg?branch=master)](https://travis-ci.com/alex-leonhardt/gocd-seeder)

# GOCD-SEEDER
A GoCD-Seeder scans a GitHub org for repositories that contain a pre-specified "topic" (default: ci-gocd), if a repo is found, it will create a GoCD config repo, which will make GoCD poll the repository for the file "ci.gocd.yaml" and create (a) new pipeline/s basaed on the config in that file.

## REQUIREMENTS

- GoCD 18.x (contains yaml plugin by default)
- Go 1.11.x (if you're building the binary yourself)

## DOCKER

### Docker Hub

You can now use the below command to just pull down the latest `tag`.

Releases: https://hub.docker.com/r/alexleonhardt/gocd-seeder/tags/


```
docker pull alexleonhardt/gocd-seeder:<tag>
```

`tag` will correspond to tags in github repo (https://github.com/alex-leonhardt/gocd-seeder/releases)

### Build own docker image

Multi-stage docker build, it's using alpine as the base OS so we can get a shell if we needed to debug the app for some reason.

```
docker build --no-cache -t local/gocd-seeder:<tag> .
```

To check version and help you can use: 

```
docker run --rm -ti local/gocd-seeder:<tag> version
docker run --rm -ti local/gocd-seeder:<tag> help
```

# BUILD the app

```
go build -ldflags "-X main.versionString=`git rev-list --max-count=1 --branches master --abbrev-commit`"
```

# TEST the app

Some tests will require to have a valid Github token as it wasn't possible to use _httptest_.

To run the tests: 

```
go test ./... -cover
```

or with some more detail 

```
go test -v ./... -cover
```

# CONFIGURATION

## Kubernetes (and other)

If you're running on kubernetes or a container orchestrator that is capable to mounting a _secret_ into a container as file, or you're able to provide the secret safely into the container using other means, it is **highly recommended** to use the below environment variables to point to the path where the secrets' values are stored.

| env var name | example |  contains |
| ------------ | ------- | --------- |
| GITHUB_SECRETS_PATH | `/secrets/github` | must contain a file "api_key" with the github api key |
| GOCD_SECRETS_PATH   | `/secrets/gocd`  | must contain a file "gocd_password" with the password corresponding to the gocd_user; <br> must contain a file "gocd_user" with the username to use to connect to GoCD |

**NOTE**: *If you set the above variables, and also set e.g. `GITHUB_API_KEY`, the file path will be preferred, this is counterintuitive but is (hopefully) more secure this way.*

## Environment vars

| Required | example | Note |
| -------- | ------- | ---- |
| GITHUB_API_KEY | `1235436adfdsfsadf` | use `GITHUB_SECRETS_PATH` when deploying to kubernetes; this variable will not be used if `GITHUB_SECRETS_PATH` is set |
| GITHUB_ORG     | `gooflix` | |

<br><br>

| Optional | default  |   |
| -------- | -------- | - |
| GITHUB_TOPIC    | `ci-gocd` | |
| GOCD_URL        | `http://localhost:8081` | |
| GOCD_USER       | `admin` | use GOCD_SECRETS_PATH when deploying to kubernetes or orchestrators that support mounting a secret as file |
| GOCD_PASSWORD   | `admin` | use GOCD_SECRETS_PATH when deploying to kubernetes or orchestrators that support mounting a secret as file |
| HTTP_STATS_IP   | default: `""` | the interface to listen on (to serve `/debug/vars` only) |
| HTTP_STATS_PORT | default: `9090` | the port to listen on (to serve `/debug/vars` only) |
| LOG_LEVEL       | default: `<none>` | available: `DEBUG` - this will enable additional log statements to be printed out; useful when debugging issues during development or initial setting up |


# METRICS

A metrics endpoint is running by default on port `:9090` and is reachable via `http://<IP|localhost>:9090/debug/vars`; metrics are provided via `expvar` - you can use things like

- [Datadog](https://docs.datadoghq.com/integrations/go_expvar/)
- [expvarmon](https://github.com/divan/expvarmon)
  ```shell
  expvarmon -ports="9090" -vars "Goroutines,Uptime,mem:memstats.Alloc,mem:memstats.Sys,mem:memstats.HeapAlloc,mem:memstats.HeapInuse,duration:memstats.PauseNs,duration:memstats.PauseTotalNs"
  ```
- Prometheus (maybe coming soon at some point if it is useful, `expvar` seems quick and cheap right now)

to monitor the app's memory, gc, goroutines & uptime

# CONTRIBUTE

Contributions through PRs are more than welcome, please also update the necessary tests as part of the submitted changes.

