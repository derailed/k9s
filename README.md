<img src="assets/k9s.png" alt="k9s">

## K9s - Kubernetes CLI To Manage Your Clusters In Style!

K9s provides a terminal UI to interact with your Kubernetes clusters.
The aim of this project is to make it easier to navigate, observe and manage
your applications in the wild. K9s continually watches Kubernetes
for changes and offers subsequent commands to interact with your observed resources.

---

## Announcement

<center>
<img src="https://raw.githubusercontent.com/imhotepio/k9salpha/master/assets/k9sa_blue_300.png" alt="k9salpha" width="300"/>
<h1>K9sAlpha RC-0 Is Out!</h1>
</center>

<br/>

Fresh off the press [K9sAlpha](https://k9salpha.io) is now available!
Please read the details in the docs and checkout the new repo.

- Store: [K9sAlpha Store](https://store.k9salpha.io).
- Screencast: [K9sAlpha-v1.0.0-rc.0](https://www.youtube.com/watch?v=hLYK0oPLOIY&t=787s)

> NOTE: Upon purchase, in order to activate your license, please send us a valid user name so we can generate your personalized license key. All licenses are valid for a whole year from the date of purchase.

For all other cases, please reach out to us so we can discuss your needs:

- Corporate licenses
- Education
- Non Profit
- Active K9s sponsors
- Long term K9s supporters and contributors
- Can't afford it
- Others...

---

[![Go Report Card](https://goreportcard.com/badge/github.com/derailed/k9s?)](https://goreportcard.com/report/github.com/derailed/k9s)
[![golangci badge](https://github.com/golangci/golangci-web/blob/master/src/assets/images/badge_a_plus_flat.svg)](https://golangci.com/r/github.com/derailed/k9s)
[![codebeat badge](https://codebeat.co/badges/89e5a80e-dfe8-4426-acf6-6be781e0a12e)](https://codebeat.co/projects/github-com-derailed-k9s-master)
[![Build Status](https://travis-ci.com/derailed/k9s.svg?branch=master)](https://travis-ci.com/derailed/k9s)
[![Docker Repository on Quay](https://quay.io/repository/derailed/k9s/status "Docker Repository on Quay")](https://quay.io/repository/derailed/k9s)
[![release](https://img.shields.io/github/release-pre/derailed/k9s.svg)](https://github.com/derailed/k9s/releases)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/mum4k/termdash/blob/master/LICENSE)
[![Releases](https://img.shields.io/github/downloads/derailed/k9s/total.svg)](https://github.com/derailed/k9s/releases)

---

## Documentation

Please refer to our [K9s documentation](https://k9scli.io) site for installation, usage, customization and tips.

## Slack Channel

Wanna discuss K9s features with your fellow `K9sers` or simply show your support for this tool?

* Channel: [K9ersSlack](https://k9sers.slack.com/)
* Invite: [K9slackers Invite](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

## Installation

K9s is available on Linux, macOS and Windows platforms.

* Binaries for Linux, Windows and Mac are available as tarballs in the [release](https://github.com/derailed/k9s/releases) page.

* Via Homebrew for macOS or LinuxBrew for Linux

   ```shell
   brew install k9s
   ```

* Via [MacPorts](https://www.macports.org)

   ```shell
   sudo port install k9s
   ```

* On Arch Linux

  ```shell
  pacman -S k9s
  ```

* On OpenSUSE Linux distribution

  ```shell
  zypper install k9s
  ```

* Via [Scoop](https://scoop.sh) for Windows

  ```shell
  scoop install k9s
  ```

* Via [Chocolatey](https://chocolatey.org/packages/k9s) for Windows

  ```shell
  choco install k9s
  ```

* Via a GO install

  ```shell
  # NOTE: The dev version will be in effect!
  go get -u github.com/derailed/k9s
  ```

* Via [Webi](https://webinstall.dev) for Linux and macOS

  ```shell
  curl -sS https://webinstall.dev/k9s | bash
  ```

* Via [Webi](https://webinstall.dev) for Windows

  ```shell
  curl.exe -A MS https://webinstall.dev/k9s | powershell
  ```
---

## Building From Source

 K9s is currently using go v1.14 or above. In order to build K9 from source you must:

 1. Clone the repo
 2. Build and run the executable

      ```shell
      make build && ./execs/k9s
      ```

---

## Running with Docker

### Running the official Docker image

  You can run k9s as a Docker container by mounting your `KUBECONFIG`:

  ```shell
  docker run --rm -it -v $KUBECONFIG:/root/.kube/config quay.io/derailed/k9s
  ```

  For default path it would be:

  ```shell
  docker run --rm -it -v ~/.kube/config:/root/.kube/config quay.io/derailed/k9s
  ```

### Building your own Docker image

  You can build your own Docker image of k9s from the [Dockerfile](Dockerfile) with the following:

  ```shell
  docker build -t k9s-docker:0.1 .
  ```

  You can get the latest stable `kubectl` version and pass it to the `docker build` command with the `--build-arg` option.
  You can use the `--build-arg` option to pass any valid `kubectl` version (like `v1.18.0` or `v1.19.1`).

  ```shell
  KUBECTL_VERSION=$(make kubectl-stable-version 2>/dev/null)
  docker build --build-arg KUBECTL_VERSION=${KUBECTL_VERSION} -t k9s-docker:0.1 .
  ```

  Run your container:

  ```shell
  docker run --rm -it -v ~/.kube/config:/root/.kube/config k9s-docker:0.1
  ```

---

## PreFlight Checks

* K9s uses 256 colors terminal mode. On `Nix system make sure TERM is set accordingly.

    ```shell
    export TERM=xterm-256color
    ```

* In order to issue manifest edit commands make sure your EDITOR env is set.

    ```shell
    # Kubectl edit command will use this env var.
    export EDITOR=my_fav_editor
    # Should your editor deals with streamed vs on disk files differently, also set...
    export K9S_EDITOR=my_fav_editor
    ```

* K9s prefers recent kubernetes versions ie 1.16+

---

## The Command Line

```shell
# List all available CLI options
k9s help
# To get info about K9s runtime (logs, configs, etc..)
k9s info
# To run K9s in a given namespace
k9s -n mycoolns
# Start K9s in an existing KubeConfig context
k9s --context coolCtx
# Start K9s in readonly mode - with all cluster modification commands disabled
k9s --readonly
```

## Logs

Given the nature of the ui k9s does produce logs to a specific location. To view the logs and turn on debug mode, use the following commands:

```shell
k9s info
# Will produces something like this
#  ____  __.________
# |    |/ _/   __   \______
# |      < \____    /  ___/
# |    |  \   /    /\___ \
# |____|__ \ /____//____  >
#         \/            \/
#
# Configuration:   /Users/fernand/.k9s/config.yml
# Logs:            /var/folders/8c/hh6rqbgs5nx_c_8k9_17ghfh0000gn/T/k9s-fernand.log
# Screen Dumps:    /var/folders/8c/hh6rqbgs5nx_c_8k9_17ghfh0000gn/T/k9s-screens-fernand

# To view k9s logs
tail -f /var/folders/8c/hh6rqbgs5nx_c_8k9_17ghfh0000gn/T/k9s-fernand.log

# Start K9s in debug mode
k9s -l debug
```

## Key Bindings

K9s uses aliases to navigate most K8s resources.

| Action                                                         | Command                       | Comment                                                                |
|----------------------------------------------------------------|-------------------------------|------------------------------------------------------------------------|
| Show active keyboard mnemonics and help                        | `?`                           |                                                                        |
| Show all available resource alias                              | `ctrl-a`                      |                                                                        |
| To bail out of K9s                                             | `:q`, `ctrl-c`                |                                                                        |
| View a Kubernetes resource using singular/plural or short-name | `:`po⏎                        | accepts singular, plural, short-name or alias ie pod or pods           |
| View a Kubernetes resource in a given namespace                | `:`alias namespace⏎           |                                                                        |
| Filter out a resource view given a filter                      | `/`filter⏎                    | Regex2 supported ie `fred|blee` to filter resources named fred or blee |
| Inverse regex filer                                            | `/`! filter⏎                  | Keep everything that *doesn't* match.                                  |
| Filter resource view by labels                                 | `/`-l label-selector⏎         |                                                                        |
| Fuzzy find a resource given a filter                           | `/`-f filter⏎                 |                                                                        |
| Bails out of view/command/filter mode                          | `<esc>`                       |                                                                        |
| Key mapping to describe, view, edit, view logs,...             | `d`,`v`, `e`, `l`,...         |                                                                        |
| To view and switch to another Kubernetes context               | `:`ctx⏎                       |                                                                        |
| To view and switch to another Kubernetes context               | `:`ctx context-name⏎          |                                                                        |
| To view and switch to another Kubernetes namespace             | `:`ns⏎                        |                                                                        |
| To view all saved resources                                    | `:`screendump or sd⏎          |                                                                        |
| To delete a resource (TAB and ENTER to confirm)                | `ctrl-d`                      |                                                                        |
| To kill a resource (no confirmation dialog!)                   | `ctrl-k`                      |                                                                        |
| Launch pulses view                                             | `:`pulses or pu⏎              |                                                                        |
| Launch XRay view                                               | `:`xray RESOURCE [NAMESPACE]⏎ | RESOURCE can be one of po, svc, dp, rs, sts, ds, NAMESPACE is optional |
| Launch Popeye view                                             | `:`popeye or pop⏎             | See https://popeyecli.io                                               |

---

## Screenshots

1. Pods
      <img src="assets/screen_po.png"/>
1. Logs
      <img src="assets/screen_logs.png"/>
1. Deployments
      <img src="assets/screen_dp.png"/>

---

---

## Demo Videos/Recordings

* [k9s Kubernetes UI - A Terminal-Based Vim-Like Kubernetes Dashboard](https://youtu.be/boaW9odvRCc)
* [K9s v0.21.3](https://youtu.be/wG8KCwDAhnw)
* [K9s v0.19.X](https://youtu.be/kj-WverKZ24)
* [K9s v0.18.0](https://www.youtube.com/watch?v=zMnD5e53yRw)
* [K9s v0.17.0](https://www.youtube.com/watch?v=7S33CNLAofk&feature=youtu.be)
* [K9s Pulses](https://asciinema.org/a/UbXKPal6IWpTaVAjBBFmizcGN)
* [K9s v0.15.1](https://youtu.be/7Fx4XQ2ftpM)
* [K9s v0.13.0](https://www.youtube.com/watch?v=qaeR2iK7U0o&t=15s)
* [K9s v0.9.0](https://www.youtube.com/watch?v=bxKfqumjW4I)
* [K9s v0.7.0 Features](https://youtu.be/83jYehwlql8)
* [K9s v0 Demo](https://youtu.be/k7zseUhaXeU)

---

## K9s Configuration

  K9s keeps its configurations in a .k9s directory in your home directory `$HOME/.k9s/config.yml`.

  > NOTE: This is still in flux and will change while in pre-release stage!

  ```yaml
  # $HOME/.k9s/config.yml
  k9s:
    # Represents ui poll intervals. Default 2secs
    refreshRate: 2
    # Number of retries once the connection to the api-server is lost. Default 15.
    maxConnRetry: 5
    # Enable mouse support. Default false
    enableMouse: true
    # Set to true to hide K9s header. Default false
    headless: false
    # Set to true to hide K9s crumbs. Default false
    crumbsless: false
    # Indicates whether modification commands like delete/kill/edit are disabled. Default is false
    readOnly: false
    # Toggles icons display as not all terminal support these chars.
    noIcons: false
    # Logs configuration
    logger:
      # Defines the number of lines to return. Default 100
      tail: 200
      # Defines the total number of log lines to allow in the view. Default 1000
      buffer: 500
      # Represents how far to go back in the log timeline in seconds. Setting to -1 will show all available logs. Default is 5min.
      sinceSeconds: 300
      # Go full screen while displaying logs. Default false
      fullScreenLogs: false
      # Toggles log line wrap. Default false
      textWrap: false
      # Toggles log line timestamp info. Default false
      showTime: false
    # Indicates the current kube context. Defaults to current context
    currentContext: minikube
    # Indicates the current kube cluster. Defaults to current context cluster
    currentCluster: minikube
    # Persists per cluster preferences for favorite namespaces and view.
    clusters:
      coolio:
        namespace:
          active: coolio
          favorites:
          - cassandra
          - default
        view:
          active: po
        featureGates:
          # Toggles NodeShell support. Allow K9s to shell into nodes if needed. Default false.
          nodeShell: false
        # Provide shell pod customization of feature gate is enabled
        shellPod:
          # The shell pod image to use.
          image: killerAdmin
          # The namespace to launch to shell pod into.
          namespace: fred
          # The resource limit to set on the shell pod.
          limits:
            cpu: 100m
            memory: 100Mi
        # The IP Address to use when launching a port-forward.
        portForwardAddress: 1.2.3.4
      kind:
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

## Node Shell

By enabling the nodeShell feature gate on a given cluster, K9s allows you to shell into your cluster nodes. Once enabled, you will have a new `s` for `shell` menu option while in node view. K9s will launch a pod on the selected node using a special k9s_shell pod. Furthermore, you can refine your shell pod by using a custom docker image preloaded with the shell tools you love. By default k9s uses a BusyBox image, but you can configure it as follows:

```yaml
# $HOME/.k9s/config.yml
k9s:
  clusters:
    # Configures node shell on cluster blee
    blee:
      featureGates:
        # You must enable the nodeShell feature gate to enable shelling into nodes
        nodeShell: true
      # You can also further tune the shell pod specification
      shellPod:
        image: cool_kid_admin:42
        namespace: blee
        limits:
          cpu: 100m
          memory: 100Mi
```

---

## Command Aliases

In K9s, you can define your very own command aliases (shortnames) to access your resources. In your `$HOME/.k9s` define a file called `alias.yml`. A K9s alias defines pairs of alias:gvr. A gvr (Group/Version/Resource) represents a fully qualified Kubernetes resource identifier. Here is an example of an alias file:

```yaml
# $HOME/.k9s/alias.yml
alias:
  pp: v1/pods
  crb: rbac.authorization.k8s.io/v1/clusterrolebindings
```

Using this alias file, you can now type pp/crb to list pods or ClusterRoleBindings respectively.

---

## HotKey Support

Entering the command mode and typing a resource name or alias, could be cumbersome for navigating thru often used resources. We're introducing hotkeys that allows a user to define their own hotkeys to activate their favorite resource views. In order to enable hotkeys please follow these steps:

1. Create a file named `$HOME/.k9s/hotkey.yml`
2. Add the following to your `hotkey.yml`. You can use resource name/short name to specify a command ie same as typing it while in command mode.

      ```yaml
      # $HOME/.k9s/hotkey.yml
      hotKey:
        # Hitting Shift-0 navigates to your pod view
        shift-0:
          shortCut:    Shift-0
          description: Viewing pods
          command:     pods
        # Hitting Shift-1 navigates to your deployments
        shift-1:
          shortCut:    Shift-1
          description: View deployments
          command:     dp
        # Hitting Shift-2 navigates to your xray deployments
        shift-2:
          shortCut:    Shift-2
          description: Xray Deployments
          command:     xray deploy
      ```

 Not feeling so hot? Your custom hotkeys will be listed in the help view `?`. Also your hotkey file will be automatically reloaded so you can readily use your hotkeys as you define them.

 You can choose any keyboard shortcuts that make sense to you, provided they are not part of the standard K9s shortcuts list.

> NOTE: This feature/configuration might change in future releases!

---

## Resource Custom Columns

[SneakCast v0.17.0 on The Beach! - Yup! sound is sucking but what a setting!](https://youtu.be/7S33CNLAofk)

You can change which columns shows up for a given resource via custom views. To surface this feature, you will need to create a new configuration file, namely `$HOME/.k9s/views.yml`. This file leverages GVR (Group/Version/Resource) to configure the associated table view columns. If no GVR is found for a view the default rendering will take over (ie what we have now). Going wide will add all the remaining columns that are available on the given resource after your custom columns. To boot, you can edit your views config file and tune your resources views live!

> NOTE: This is experimental and will most likely change as we iron this out!

Here is a sample views configuration that customize a pods and services views.

```yaml
# $HOME/.k9s/views.yml
k9s:
  views:
    v1/pods:
      columns:
        - AGE
        - NAMESPACE
        - NAME
        - IP
        - NODE
        - STATUS
        - READY
    v1/services:
      columns:
        - AGE
        - NAMESPACE
        - NAME
        - TYPE
        - CLUSTER-IP
```

---

## Plugins

K9s allows you to extend your command line and tooling by defining your very own cluster commands via plugins. K9s will look at `$HOME/.k9s/plugin.yml` to locate all available plugins. A plugin is defined as follows:

* Shortcut option represents the key combination a user would type to activate the plugin
* Confirm option (when enabled) lets you see the command that is going to be executed and gives you an option to confirm or prevent execution
* Description will be printed next to the shortcut in the k9s menu
* Scopes defines a collection of resources names/short-names for the views associated with the plugin. You can specify `all` to provide this shortcut for all views.
* Command represents ad-hoc commands the plugin runs upon activation
* Background specifies whether or not the command runs in the background
* Args specifies the various arguments that should apply to the command above

K9s does provide additional environment variables for you to customize your plugins arguments. Currently, the available environment variables are as follows:

* `$NAMESPACE` -- the selected resource namespace
* `$NAME` -- the selected resource name
* `$CONTAINER` -- the current container if applicable
* `$FILTER` -- the current filter if any
* `$KUBECONFIG` -- the KubeConfig location.
* `$CLUSTER` the active cluster name
* `$CONTEXT` the active context name
* `$USER` the active user
* `$GROUPS` the active groups
* `$POD` while in a container view
* `$COL-<RESOURCE_COLUMN_NAME>` use a given column name for a viewed resource. Must be prefixed by `COL-`!

### Example

This defines a plugin for viewing logs on a selected pod using `ctrl-l` for shortcut.

```yaml
# $HOME/.k9s/plugin.yml
plugin:
  # Defines a plugin to provide a `ctrl-l` shortcut to tail the logs while in pod view.
  fred:
    shortCut: Ctrl-L
    confirm: false
    description: Pod logs
    scopes:
    - pods
    command: kubectl
    background: false
    args:
    - logs
    - -f
    - $NAME
    - -n
    - $NAMESPACE
    - --context
    - $CONTEXT
```

> NOTE: This is an experimental feature! Options and layout may change in future K9s releases as this feature solidifies.

---

## Benchmark Your Applications

K9s integrates [Hey](https://github.com/rakyll/hey) from the brilliant and super talented [Jaana Dogan](https://github.com/rakyll). `Hey` is a CLI tool to benchmark HTTP endpoints similar to AB bench. This preliminary feature currently supports benchmarking port-forwards and services (Read the paint on this is way fresh!).

To setup a port-forward, you will need to navigate to the PodView, select a pod and a container that exposes a given port. Using `SHIFT-F` a dialog comes up to allow you to specify a local port to forward. Once acknowledged, you can navigate to the PortForward view (alias `pf`) listing out your active port-forwards. Selecting a port-forward and using `CTRL-B` will run a benchmark on that HTTP endpoint. To view the results of your benchmark runs, go to the Benchmarks view (alias `be`). You should now be able to select a benchmark and view the run stats details by pressing `<ENTER>`. NOTE: Port-forwards only last for the duration of the K9s session and will be terminated upon exit.

Initially, the benchmarks will run with the following defaults:

* Concurrency Level: 1
* Number of Requests: 200
* HTTP Verb: GET
* Path: /

The PortForward view is backed by a new K9s config file namely: `$HOME/.k9s/bench-<k8s_context>.yml` (note: extension is `yml` and not `yaml`). Each cluster you connect to will have its own bench config file, containing the name of the K8s context for the cluster. Changes to this file should automatically update the PortForward view to indicate how you want to run your benchmarks.

Here is a sample benchmarks.yml configuration. Please keep in mind this file will likely change in subsequent releases!

```yaml
# This file resides in $HOME/.k9s/bench-mycontext.yml
benchmarks:
  # Indicates the default concurrency and number of requests setting if a container or service rule does not match.
  defaults:
    # One concurrent connection
    concurrency: 1
    # Number of requests that will be sent to an endpoint
    requests: 500
  containers:
    # Containers section allows you to configure your http container's endpoints and benchmarking settings.
    # NOTE: the container ID syntax uses namespace/pod-name:container-name
    default/nginx:nginx:
      # Benchmark a container named nginx using POST HTTP verb using http://localhost:port/bozo URL and headers.
      concurrency: 1
      requests: 10000
      http:
        path: /bozo
        method: POST
        body:
          {"fred":"blee"}
        header:
          Accept:
            - text/html
          Content-Type:
            - application/json
  services:
    # Similarly you can Benchmark an HTTP service exposed either via NodePort, LoadBalancer types.
    # Service ID is ns/svc-name
    default/nginx:
      # Set the concurrency level
      concurrency: 5
      # Number of requests to be sent
      requests: 500
      http:
        method: GET
        # This setting will depend on whether service is NodePort or LoadBalancer. NodePort may require vendor port tunneling setting.
        # Set this to a node if NodePort or LB if applicable. IP or dns name.
        host: A.B.C.D
        path: /bumblebeetuna
      auth:
        user: jean-baptiste-emmanuel
        password: Zorg!
```

---

## K9s RBAC FU

On RBAC enabled clusters, you would need to give your users/groups capabilities so that they can use K9s to explore their Kubernetes cluster. K9s needs minimally read privileges at both the cluster and namespace level to display resources and metrics.

These rules below are just suggestions. You will need to customize them based on your environment policies. If you need to edit/delete resources extra Fu will be necessary.

> NOTE! Cluster/Namespace access may change in the future as K9s evolves.
> NOTE! We expect K9s to keep running even in atrophied clusters/namespaces. Please file issues if this is not the case!

### Cluster RBAC scope

```yaml
---
# K9s Reader ClusterRole
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: k9s
rules:
  # Grants RO access to cluster resources node and namespace
  - apiGroups: [""]
    resources: ["nodes", "namespaces"]
    verbs: ["get", "list", "watch"]
  # Grants RO access to RBAC resources
  - apiGroups: ["rbac.authorization.k8s.io"]
    resources: ["clusterroles", "roles", "clusterrolebindings", "rolebindings"]
    verbs: ["get", "list", "watch"]
  # Grants RO access to CRD resources
  - apiGroups: ["apiextensions.k8s.io"]
    resources: ["customresourcedefinitions"]
    verbs: ["get", "list", "watch"]
  # Grants RO access to metric server (if present)
  - apiGroups: ["metrics.k8s.io"]
    resources: ["nodes", "pods"]
    verbs: ["get", "list", "watch"]

---
# Sample K9s user ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: k9s
subjects:
  - kind: User
    name: fernand
    apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: k9s
  apiGroup: rbac.authorization.k8s.io
```

### Namespace RBAC scope

If your users are constrained to certain namespaces, K9s will need to following role to enable read access to namespaced resources.

```yaml
---
# K9s Reader Role (default namespace)
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: k9s
  namespace: default
rules:
  # Grants RO access to most namespaced resources
  - apiGroups: ["", "apps", "autoscaling", "batch", "extensions"]
    resources: ["*"]
    verbs: ["get", "list", "watch"]
  # Grants RO access to metric server
  - apiGroups: ["metrics.k8s.io"]
    resources: ["pods", "nodes"]
    verbs:
      - get
      - list
      - watch

---
# Sample K9s user RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: k9s
  namespace: default
subjects:
  - kind: User
    name: fernand
    apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: Role
  name: k9s
  apiGroup: rbac.authorization.k8s.io
```

---

## Skins

Example: Dracula Skin ;)

<img src="assets/skins/dracula.png" alt="Dracula Skin">

You can style K9s based on your own sense of look and style. Skins are YAML files, that enable a user to change the K9s presentation layer. K9s skins are loaded from `$HOME/.k9s/skin.yml`. If a skin file is detected then the skin would be loaded if not the current stock skin remains in effect.

You can also change K9s skins based on the cluster you are connecting too. In this case, you can specify the skin file name as `$HOME/.k9s/mycontext_skin.yml`
Below is a sample skin file, more skins are available in the skins directory in this repo, just simply copy any of these in your user's home dir as `skin.yml`.

Colors can be defined by name or uing an hex representation. Of recent, we've added a color named `default` to indicate a transparent background color to preserve your terminal background color settings if so desired.

> NOTE: This is very much an experimental feature at this time, more will be added/modified if this feature has legs so thread accordingly!


> NOTE: Please see [K9s Skins](https://k9scli.io/topics/skins/) for a list of available colors.


```yaml
# Skin InTheNavy...
k9s:
  # General K9s styles
  body:
    fgColor: dodgerblue
    bgColor: '#ffffff'
    logoColor: '#0000ff'
  # ClusterInfoView styles.
  info:
    fgColor: lightskyblue
    sectionColor: steelblue
  frame:
    # Borders styles.
    border:
      fgColor: dodgerblue
      focusColor: aliceblue
    # MenuView attributes and styles.
    menu:
      fgColor: darkblue
      keyColor: cornflowerblue
      # Used for favorite namespaces
      numKeyColor: cadetblue
    # CrumbView attributes for history navigation.
    crumbs:
      fgColor: white
      bgColor: steelblue
      activeColor: skyblue
    # Resource status and update styles
    status:
      newColor: '#00ff00'
      modifyColor: powderblue
      addColor: lightskyblue
      errorColor: indianred
      highlightcolor: royalblue
      killColor: slategray
      completedColor: gray
    # Border title styles.
    title:
      fgColor: aqua
      bgColor: white
      highlightColor: skyblue
      counterColor: slateblue
      filterColor: slategray
  views:
    # TableView attributes.
    table:
      fgColor: blue
      bgColor: darkblue
      cursorColor: aqua
      # Header row styles.
      header:
        fgColor: white
        bgColor: darkblue
        sorterColor: orange
    # YAML info styles.
    yaml:
      keyColor: steelblue
      colonColor: blue
      valueColor: royalblue
    # Logs styles.
    logs:
      fgColor: white
      bgColor: black
```

---

## Known Issues

This is still work in progress! If something is broken or there's a feature
that you want, please file an issue and if so inclined submit a PR!

K9s will most likely blow up if...

1. You're running older versions of Kubernetes. K9s works best on Kubernetes latest.
2. You don't have enough RBAC fu to manage your cluster.

---

## ATTA Girls/Boys!

K9s sits on top of many open source projects and libraries. Our *sincere*
appreciations to all the OSS contributors that work nights and weekends
to make this project a reality!

---

## Meet The Core Team!

* [Fernand Galiana](https://github.com/derailed)
  * <img src="assets/mail.png" width="16" height="auto" alt="email"/>  fernand@imhotep.io
  * <img src="assets/twitter.png" width="16" height="auto" alt="twitter"/> [@kitesurfer](https://twitter.com/kitesurfer?lang=en)

We always enjoy hearing from folks who benefit from our work!

## Contributions Guideline

* File an issue first prior to submitting a PR!
* Ensure all exported items are properly commented
* If applicable, submit a test suite against your PR

---

<img src="assets/imhotep_logo.png" width="32" height="auto" alt="Imhotep"/> &nbsp;© 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
