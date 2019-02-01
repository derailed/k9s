<img src="assets/k9s.png">

# K9s - A command line application to manage your kubernetes clusters.

A CLI written in GO and curses to interact with your Kubernetes clusters.
The aim of this project is to make it easier to navigate, observe and manage
your clusters.

K9s is a CLI for Kubernetes. It provides a bit more information about your cluster
than *kubectl* while allowing to perform primordial Kubernetes commands with ease.

<br/>

---

[![Go Report Card](https://goreportcard.com/badge/github.com/derailed/k9s)](https://goreportcard.com/report/github.com/derailed/k9s)
[![Build Status](https://travis-ci.org/derailed/k9s.svg?branch=master)](https://travis-ci.org/derailed/k9s)


<br/>

---
## Installation

### Homebrew (OSX)

```shell
brew tap derailed/k9s https://github.com/derailed/k9s-homebrew-tap.git
brew install k9s
```

### Binary Releases

- [Releases](https://github.com/derailed/k9s/releases)



<br/>

---
## Features

> Note: k9s does not have an idiot light. Please be sure to hit the correct command
> sequences to avoid pilot errors. `Are you sure?` not in effect here...

+ k9s uses 2 or 3 letters alias to navigate most Kubernetes resources
+ At any time you can use `?<Enter>` to look up the various commands
+ Use `alias<Enter>` to activate a resource under that alias
+ Use `Esc` to erase previous keystrokes.
+ Use `Q` or `Ctrl-C` to Quit.
+ `Ctrl` sequences are used to view, edit, delete, ssh ...
+ Use `ctx<Enter>` to see and switch between your clusters

<br/>

---
## Video Demo

+ [k9s Demo](https://youtu.be/k7zseUhaXeU)


<br/>

---
## Screen Shots

### Pod View

<img src="assets/screen_1.png">

### Log View

<img src="assets/screen_2.png">

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

k9s sits on top of two very cool GO projects that provides the much needed terminal
support. So big thanks and shootout to the good folks at tcell+tview for
making k9s a reality!!

+ [tcell](https://github.com/gdamore/tcell)
+ [tview](https://github.com/rivo/tview)


<br/>

---
## Contact Information

+ **Email**:   fernand@imhotep.io
+ **Twitter**: [@kitesurfer](https://twitter.com/kitesurfer?lang=en)
+ **Github**:  [k9s](https://github.com/derailed/k9s)
<br/>

---
<img src="assets/imhotep_logo.png" width="32" height="auto"/> © 2018 Imhotep Software LLC.
All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)