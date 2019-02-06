<img src="assets/k9s.png">

# K9s - Kubernetes CLI To Manage Your Clusters In Style!

K9s provides a curses based terminal UI to interact with your Kubernetes clusters.
The aim of this project is to make it easier to navigate, observe and manage
your applications. This command line app continually watches Kubernetes
for changes and offers subsequent commands to interact with observed resources.


<br/>

---

[![Go Report Card](https://goreportcard.com/badge/github.com/derailed/k9s?)](https://goreportcard.com/report/github.com/derailed/k9s)
[![Build Status](https://travis-ci.com/derailed/k9s.svg?branch=master)](https://travis-ci.com/derailed/k9s)

<br/>

---
## Installation

### Homebrew (OSX)

```shell
brew tap derailed/k9s && brew install k9s
```

### Binary Releases

- [Releases](https://github.com/derailed/k9s/releases)


<br/>

---
## Command Line

To show all available options from the cli use:

```shell
k9s -h
```

<br/>

---
## PreFlight Checks

* K9s uses 256 colors terminal mode. On `Nix system make sure TERM is set accordingly.

    ```shell
    export TERM=xterm-256color
    ```

* For clusters with many namespaces you can either edit ~/.k9s/config.yml or
  go to the namespace(ns) view to switch your default namespace to your namespace
  of choice using *Ctrl-S*witch. K9s will keep your top 5 favorite namespaces.
  Namespaces will get evicted based on your namespace switching frequency.


    ```yaml
    k9s:
      refreshRate:   5   # K9s refresh rate in secs
      logBufferSize: 200 # Size of the logs buffer. Try to keep a sensible default!
      namespace:
        active: myCoolNS # Current active namespace name
        favorites:       # List of your 5 most frequently used namespaces
        - myCoolNS1
        - myCoolNS2
        - all
        - default
        - kube-system
      view:
        active: po       # Active resource view
    ```

* K9s can use **$KUBECONFIG** env var to load cluster information. However we've
  seen hill effects of using this env with multiple files as setting the current
  context may not update the correct file when using this technique.


---
<br/>

---
## Commands

+ K9s uses 2 or 3 letters alias to navigate most K8s resource
+ At any time you can use `?` to look up the various commands
+ Use `alias<Enter>` to activate a resource under that alias
+ `Ctrl` sequences are used to view, edit, delete, ssh ...
+ Use `ctx<Enter>` to switch between clusters
+ Use `Q` or `Ctrl-C` to Quit.

<br/>

---
## Demo Video

+ [k9s Demo](https://youtu.be/k7zseUhaXeU)


<br/>

---
## Screen Shots

### Pod View

<img src="assets/screen_3.png">

### Log View

<img src="assets/screen_4.png">

<br/>

---
## Known Issues...

This initial drop is brittle. k9s will most likely blow up if...

+ Your kube-config file does not live under $HOME/.kube or you use multiple configs
+ You don't have enough RBAC fu to manage your cluster
+ Your cluster does not run a metrics-server

<br/>

---
## Disclaimer

This is still work in progress! If there is enough interest in the Kubernetes
community, we will enhance per your recommendations/contributions. Also if you
dig this effort, please let us know that too!

<br/>

---
## ATTA Girls/Boys!

k9s sits on top of on a lot of opensource projects. So *big thanks* to all the
contributors that made this project a reality!


<br/>

---
## Contact Information

+ **Email**:   fernand@imhotep.io
+ **Twitter**: [@kitesurfer](https://twitter.com/kitesurfer?lang=en)
+ **Web**:  [K9s](https://k9ss.io)

<br/>

---
<img src="assets/imhotep_logo.png" width="32" height="auto"/> © 2018 Imhotep Software LLC.
All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)