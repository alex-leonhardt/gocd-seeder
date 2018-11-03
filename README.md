[![GoDoc](https://godoc.org/github.com/alex-leonhardt/gocd-seeder?status.svg)](https://godoc.org/github.com/alex-leonhardt/gocd-seeder)

# GOCD-SEEDER
A GoCD-Seeder scans a GitHub org for repositories that contain a pre-specified "topic" (default: ci-gocd), if a repo is found, it will create a GoCD config repo, which will make GoCD poll the repository for the file "ci.gocd.yaml" and create (a) new pipeline/s basaed on the config in that file.

## DOCKER

#### Docker Hub

You can now use the below command to just pull down the latest `tag`.

_NOTE: as of 3/11/18 -- this was previously `master` (so `:latest` will correspond to `f8c3fafed4555ee25053f9fdc0ae368d0657e6e4`) but I believe it's better to use tags so it's not a moving target when you rely on a specific version_


```
docker pull alexleonhardt/gocd-seeder:<tag>
```
`tag` will correspond to tags in github repo (https://github.com/alex-leonhardt/gocd-seeder/releases)

#### Build own

Multi-stage docker build, it's using alpine as the base OS so we can get a shell if we needed to debug the app for some reason.

```
docker build --no-cache -t local/gocd-seeder:<tag> .
```

To check version and help you can use: 

```
docker run --rm -ti local/gocd-seeder:<tag> version
docker run --rm -ti local/gocd-seeder:<tag> help
```

## NTHs

_(nice to haves)_

- metric (counter) of repos created, deleted

## BUILD

```
go build -ldflags "-X main.versionString=`git rev-list --max-count=1 --branches master --abbrev-commit`"
```

## TEST

Some tests will require to have a valid Github token as it wasn't possible to use _httptest_.

To run the tests: 

```
go test ./... -cover
```

or with some more detail 

```
go test -v ./... -cover
```


## METRICS

A metrics endpoint is running by default on port `:9090` and is reachable via `http://<IP|localhost>:9090/debug/vars`; metrics are provided via `expvar` - you can use things like

- [Datadog](https://docs.datadoghq.com/integrations/go_expvar/)
- [expvarmon](https://github.com/divan/expvarmon)
  ```shell
  expvarmon -ports="9090" -vars "Goroutines,Uptime,mem:memstats.Alloc,mem:memstats.Sys,mem:memstats.HeapAlloc,mem:memstats.HeapInuse,duration:memstats.PauseNs,duration:memstats.PauseTotalNs"
  ```
- Prometheus (maybe coming soon at some point if it is useful, `expvar` seems quick and cheap right now)

to monitor the app's memory, gc, goroutines & uptime

### Config

Use env variables to set the stats port and IP to listen on:
```
export HTTP_STATS_PORT=<PORT>
export HTTP_STATS_IP=<IP|localhost>
```

the defaults are: 

```
HTTP_STATS_PORT=9090
HTTP_STATS_IP=""
```

which will make the endpoint available on all interfaces, on port 9090

## REQUIREMENTS

- GoCD 18.x (contains yaml plugin by default)

## CONTRIBUTE

Contributions through PRs are more than welcome, please also update the necessary tests as part of the submitted changes.
