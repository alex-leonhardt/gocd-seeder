FROM golang:alpine AS build

ADD . /go/src/github.com/alex-leonhardt/gocd-seeder
RUN apk add git \
    && cd /go/src/github.com/alex-leonhardt/gocd-seeder \
    && go get -v ./... \
    && find /go \
    && go build -v -ldflags "-X main.versionString=`git rev-list --max-count=1 --branches master --abbrev-commit`" -o /go/bin/gocd-seeder \
    && chmod +x /go/bin/gocd-seeder \
    && ls /go/bin/gocd-seeder

FROM alpine
WORKDIR /app
RUN ls -al /app \
    && apk update \
    && apk add ca-certificates \
    && rm -rf /var/cache/apk/*

COPY --from=build /go/bin/gocd-seeder /app/
ENTRYPOINT ["./gocd-seeder"]
