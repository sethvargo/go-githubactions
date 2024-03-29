# GitHub Actions SDK (Go)

[![Go Reference](https://pkg.go.dev/badge/github.com/sethvargo/go-githubactions.svg)](https://pkg.go.dev/github.com/sethvargo/go-githubactions)
[![unit](https://github.com/sethvargo/go-githubactions/actions/workflows/unit.yml/badge.svg)](https://github.com/sethvargo/go-githubactions/actions/workflows/unit.yml)

This library provides an SDK for authoring [GitHub Actions][gh-actions] in Go. It has no external dependencies and provides a Go-like interface for interacting with GitHub Actions' build system.


## Installation

Download the library:

```text
$ go get -u github.com/sethvargo/go-githubactions/...
```


## Usage

The easiest way to use the library is by importing it and invoking the functions
at the root:

```go
import (
  "github.com/sethvargo/go-githubactions"
)

func main() {
  val := githubactions.GetInput("val")
  if val == "" {
    githubactions.Fatalf("missing 'val'")
  }
}
```

You can also create an instance with custom fields that will be included in log messages:

```go
import (
  "github.com/sethvargo/go-githubactions"
)

func main() {
  actions := githubactions.WithFieldsMap(map[string]string{
    "file": "myfile.js",
    "line": "100",
  })

  val := actions.GetInput("val")
  if val == "" {
    actions.Fatalf("missing 'val'")
  }
}
```

For more examples and API documentation, please see the [Go docs][godoc].


## Publishing

There are multiple ways to publish GitHub Actions written in Go:

-   [Composite actions](https://github.com/FerretDB/github-actions/blob/2ae30fd2cdb635d8aefdaf9f770257e156c9f77b/extract-docker-tag/action.yml)
-   [Pre-compiled binaries with a shim](https://full-stack.blend.com/how-we-write-github-actions-in-go.html)
-   Docker containers (see below)

By default, GitHub Actions expects actions to be written in Node.js. For other languages like Go, you need to provide a `Dockerfile` and entrypoint instructions in an `action.yml` file:

```dockerfile
# your-repo/Dockerfile
FROM golang:1.18
WORKDIR /src
COPY . .
RUN go build -o /bin/app .
ENTRYPOINT ["/bin/app"]
```

```yaml
# your-repo/action.yml
name: My action
author: My name
description: My description

runs:
  using: docker
  image: Dockerfile
```

And then users can import your action by the repository name:

```yaml
# their-repo/.github/workflows/thing.yml
steps:
- name: My action
  uses: username/repo@latest
```

However, this will clone the entire repo and compile the Go code each time the action runs. Worse, it uses the Go base container which is a few hundred MBs and includes a ton of unnecessary things.

Fortunately, GitHub Actions can also source from a Docker container directly from Docker Hub:

```yaml
steps:
- name: My action
  uses: docker://username/repo:latest
```

Now we can precompile and publish our Go Action as a Docker container, but we need to make it much, much smaller first. This can be achieved using multi-stage Docker builds:

```dockerfile
FROM golang:1.18 AS builder

ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

RUN apt-get -qq update && \
  apt-get -yqq install upx

WORKDIR /src
COPY . .

RUN go build \
  -ldflags "-s -w -extldflags '-static'" \
  -o /bin/app \
  . \
  && strip /bin/app \
  && upx -q -9 /bin/app

RUN echo "nobody:x:65534:65534:Nobody:/:" > /etc_passwd



FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc_passwd /etc/passwd
COPY --from=builder --chown=65534:0 /bin/app /app

USER nobody
ENTRYPOINT ["/app"]
```

The first step, uses a fat container to build, strip, and compress the compiled Go binary. Then, in the second step, the compiled and compressed binary is copied into a scratch (bare) container along with some SSL certificates and a `nobody` user in which to execute the container.

This will usually produce an image that is less than 10MB in size, making for
much faster builds.


[gh-actions]: https://github.com/features/actions
[godoc]: https://godoc.org/github.com/sethvargo/go-githubactions
