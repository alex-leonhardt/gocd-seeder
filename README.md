# gocd-seeder
A tiny app to create GoCD pipelines using the yaml config plugin, this would run in paralllel with your GoCD server


## WIP

- add env/args to configure user/pass for GoCD admin/api
- add config options for matching `org` and `topic`
- add filter for repos by org
- use ENV vars for config, and args to override ENV
-- most are env vars, but need to add acouple + need to be able to override as args
- use structured logging go-kit/log
- remove repo when repo present and `topic` was removed
- output a metric of repos created/deleted (or make a http endpoint available)
- add `-version` arg
- clean up
-- write tests
-- make use of the interfaces etc. etc.
-- write logging and metric decorators and wrap func calls where appropriate
-- ensure to only output on startup once and when errors occur with more helpful messages
- provide a Dockerfile to create image, make it avail in docker hub