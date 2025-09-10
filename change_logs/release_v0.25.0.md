<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.25.0

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better are as ever very much noted and appreciated!

If you feel K9s is helping your Kubernetes journey, please consider joining our [sponsorship program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

## ♫ Sounds Behind The Release ♭

* [High Fidelity - By Elvis Costello (Yup! he started is career as a computer operator. Can u tell??)](https://www.youtube.com/watch?v=DJS-2kacmpU)
* [Walk With A Big Stick - Foster The People](https://www.youtube.com/watch?v=XMY1VMTyl8s)
* [Beirut - Steps Ahead -- Love this band!! with the ever so talented and sadly late Michael Brecker ;(](https://www.youtube.com/watch?v=UExKTZ3veB8)

---

### A Word From Our Sponsors...

I want to recognize the following folks that have been kind enough to join our sponsorship program and opted to `pay it forward`!

* [Andrew Regan](https://github.com/poblish)
* [Bruno Brito](https://github.com/brunohbrito)
* [ScubaDrew](https://github.com/ScubaDrew)
* [mike-code](https://github.com/mike-code)
* [Andrew Aadland](https://github.com/DaemonDude23)
* [Michael Albers](https://github.com/michaeljohnalbers)

So if you feel K9s is helping with your productivity while administering your Kubernetes clusters, please consider pitching in as it will go a long way in ensuring a thriving environment for this repo and our K9sers community at large.

Also please take some time and give a huge shoot out to all the good folks below that have spent time plowing thru the code to help improve K9s for all of us!

Thank you!!

---

## Personal Note...

I had so many distractions this cycle so expect some `disturbance in the farce!` on this drop.
To boot rat holed quiet a bit on improving speed. So I might have drop some stuff on the floor in the process...
Please report back if that's the case and we will address shortly. Tx!!

## Port It Forward??

Ever been in a situation where you need to constantly port-forward on a given pod with multiple containers or exposing multiple ports? If so it might be cumbersome to have to type in the full container:port specification to activate a forward. If you fall in this use cases, you can now specify which container and port you would rather port-forward to by default. In this drop, we introduce a new annotation that you can use to specify and container/port to forward to by default. If set, the port-forward dialog will know to default to your settings.

> NOTE: you can either use a container port name or number in your annotation!

```yaml
# Pod fred
apiVersion: v1
kind: Pod
metadata:
  name: fred
  annotations:
    k9scli.io/auto-portforwards: zorg::5556        # => will default to container zorg port 5556 and local port 5566. No port-forward dialog will be shown.
    # Or...
    k9scli.io/portforward: bozo::6666:p1           # => launches the port-forward dialog selecting default port-forward on container bozo port named p1(8081)
                                                   # mapping to local port 6666.
    ...
spec:
  containers:
  - name: zorg
    ports:
    - name: p1
      containerPort: 5556
    ...
  - name: bozo
    ports:
    - name: p1
      containerPort: 8081
    - name: p2
      containerPort: 5555
    ...
```

The annotation value must specify a container to forward to as well as a local port and container port. The container port may be specified as either a port number or port name. If the local port is omitted then the local port will default to the container port number. Here are a few examples:

1. bozo::http      - creates a pf on container `bozo` with port name http. If http specifies port number 8080 then the local port will be 8080 as well.
2. bozo::9090:http - creates a pf on container `bozo` mapping local port 9090->http(8080)
3. bozo::9090:8080 - creates a pf on container `bozo` mapping local port 9090->8080

---

## Resolved Issues

* [Issue #1299](https://github.com/derailed/k9s/issues/1299) After upgrade to 0.24.15 sorting shortcuts not working
* [Issue #1298](https://github.com/derailed/k9s/issues/1298) Install K9s through go get reporting ambiguous import error
* [Issue #1296](https://github.com/derailed/k9s/issues/1296) Crash when clicking between border of K9s and terminal pane
* [Issue #1289](https://github.com/derailed/k9s/issues/1289) Homebrew calling bottle :unneeded is deprecated! There is no replacement
* [Issue #1273](https://github.com/derailed/k9s/issues/1273) Not loading config from correct default location when XDG_CONFIG_HOME is unset
* [Issue #1268](https://github.com/derailed/k9s/issues/1268) Age sorting wrong for years
* [Issue #1258](https://github.com/derailed/k9s/issues/1258) Configurable or recent use based port-forward
* [Issue #1257](https://github.com/derailed/k9s/issues/1257) Why is the latest chocolatey on 0.24.10
* [Issue #1243](https://github.com/derailed/k9s/issues/1243) Port forward fails in kind on windows 10

---

## PRs

* [PR #1300](https://github.com/derailed/k9s/pull/1300) move from io/ioutil to io/os packages
* [PR #1287](https://github.com/derailed/k9s/pull/1287) Add missing styles to kiss
* [PR #1286](https://github.com/derailed/k9s/pull/1286) Some small color modifications
* [PR #1284](https://github.com/derailed/k9s/pull/1284) Fix a small typo which comes from cluster view info
* [PR #1271](https://github.com/derailed/k9s/pull/1271) Removed cursor colors that are too light to read
* [PR #1266](https://github.com/derailed/k9s/pull/1266) Skin to preserve your terminal session background color
* [PR #1264](https://github.com/derailed/k9s/pull/1205) Adding note on popeye config
* [PR #1261](https://github.com/derailed/k9s/pull/1261) Blurry logo
* [PR #1250](https://github.com/derailed/k9s/pull/1250) Gruvbox dark skin
* [PR #1249](https://github.com/derailed/k9s/pull/1249) Node shell pod tolerate all taints
* [PR #1232](https://github.com/derailed/k9s/pull/1232) Add red skin for production env
* [PR #1227](https://github.com/derailed/k9s/pull/1227) Add abbreviation ReadWriteOncePod PV access mode

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
