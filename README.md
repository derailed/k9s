<img src="assets/k9s.png">

# K9s - Kubernetes CLI To Manage Your Clusters In Style!

K9s provides a curses based terminal UI to interact with your Kubernetes clusters.
The aim of this project is to make it easier to navigate, observe and manage
your applications in the wild. K9s continually watches Kubernetes
for changes and offers subsequent commands to interact with observed resources.

---

[![Go Report Card](https://goreportcard.com/badge/github.com/derailed/k9s?)](https://goreportcard.com/report/github.com/derailed/k9s)
[![Build Status](https://travis-ci.com/derailed/k9s.svg?branch=master)](https://travis-ci.com/derailed/k9s)

---

## Installation

K9s is available on Linux, OSX and Windows platforms.

* Binaries for Linux, Windows and Mac are available as tar balls in the [release](https://github.com/derailed/k9s/releases) page.

* For OSX using Homebrew

   ```shell
   brew tap derailed/k9s && brew install k9s
   ```

* Building from source
   K9s was built using go 1.11 or above. In order to build K9 from source you must:
   1. Clone the repo
   2. Set env var *GO111MODULE=on*
   3. Add the following command in your go.mod file

      ```text
      replace (
       github.com/derailed/k9s => MY_K9S_CLONED_GIT_REPO
      )
      ```

   4. Build and run the executable

        ```shell
        go run main.go
        ```

---

## The Command Line

```shell
# List all available CLI options
k9s -h
# To get info about K9s runtime (logs, configs, etc..)
k9s info
# To run K9s in a given namespace
k9s -n mybitchns
# Start K9s in an existing KubeConfig context
k9s --context coolCtx
```

---

## PreFlight Checks

* K9s uses 256 colors terminal mode. On `Nix system make sure TERM is set accordingly.

    ```shell
    export TERM=xterm-256color
    ```

---

## K9s config file ($HOME/.k9s/config.yml)

  K9s keeps its configurations in a dot file in your home directory.

  > NOTE: This is still in flux and will change while in pre-release stage!

  ```yaml
  k9s:
    refreshRate: 2
    logBufferSize: 200
    currentContext: minikube
    currentCluster: minikube
    clusters:
      bitchn:
        namespace:
          active: coolio
          favorites:
          - cassandra
          - default
        view:
          active: po
      minikube:
        namespace:
          active: all
          favorites:
          - all
          - kube-system
          - default
        view:
          active: dp
  ```

---

## Key Bindings

K9s uses aliases to navigate most K8s resources.

| Command               | Result                                             | Example                    |
|-----------------------|----------------------------------------------------|----------------------------|
| `:`alias`<ENTER>`     | View a Kubernetes resource                         | `:po<ENTER>`               |
| '?'                   | Show all command aliases                           | select+<ENTER> to view     |
| `/`filter`ENTER`>     | Filter out a resource view given a filter          | `/bumblebeetuna`           |
| `<Esc>`               | Bails out of command mode                          |                            |
| `d`,`v`, `e`, `l`,... | Key mapping to describe, view, edit, view logs,... | `d` (describes a resource) |
| `:`ctx`<ENTER>`       | To view and switch to another Kubernetes context   | `:`+`ctx`+`<ENTER>`        |
| `q`, `Ctrl-c`         | To bail out of K9s                                 |                            |

---

## Demo Video

1. [K9s Demo](https://youtu.be/k7zseUhaXeU)


## Screenshots

1. Pods

  <img src="assets/screen_po.png">

1. Logs

  <img src="assets/screen_logs.png">

1. Deployments

  <img src="assets/screen_dp.png">


---

## Known Issues

This initial drop is brittle. K9s will most likely blow up if...

1. You don't have enough RBAC fu to manage your cluster
2. Your cluster does not run a metric server.

---

## Disclaimer

This is still work in progress! If there is enough interest in the Kubernetes
community, we will enhance per your recommendations/contributions. Also if you
dig this effort, please let us know that too!

---

## ATTA Girls/Boys!

K9s sits on top of many of opensource projects and libraries. Our *sincere*
appreciations to all the OSS contributors that work nights and weekends
to make this project a reality!


---

## Contact Info

1. **Email**:   fernand@imhotep.io
1. **Twitter**: [@kitesurfer](https://twitter.com/kitesurfer?lang=en)


---

<img src="assets/imhotep_logo.png" width="32" height="auto"/> © 2018 Imhotep Software LLC.
All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)