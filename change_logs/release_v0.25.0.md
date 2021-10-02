<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.25.0

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better are as ever very much noted and appreciated!

If you feel K9s is helping your Kubernetes journey, please consider joining our [sponsorship program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

## ♫ Sounds Behind The Release ♭

* [High Fidelity - By Elvis Costello (yup! he started as a computer operator. Can u tell?)](https://www.youtube.com/watch?v=DJS-2kacmpU)
* [Walk With A Big Stick - Foster The People](https://www.youtube.com/watch?v=XMY1VMTyl8s)
* [Beirut - Steps Ahead -- Love this band!! with the ever so talented and sadly late Michael Brecker ;(](https://www.youtube.com/watch?v=UExKTZ3veB8)

---

### A Word From Our Sponsors...

I want to recognize the following folks that have been kind enough to join our sponsorship program and opted to `pay it forward`!

* [Andrew Regan](https://github.com/poblish)
* [Astraea](https://github.com/s22s)
* [DataRoots](https://github.com/datarootsio)

So if you feel K9s is helping with your productivity while administering your Kubernetes clusters, please consider pitching in as it will go a long way in ensuring a thriving environment for this repo and our k9ers community at large.

Thank you!!

---

## Forward That!

Ever been in a situation where you need to constantly port-forward on a given pod with multiple containers exposing multiple ports? If so it might be cumbersome to have to type in the full container:port specification to activate a forward. If you fall in this use cases, you can now specify which container and port you would rather port-forward to by default. In this drop, we introduce a new annotation that you can use to specify and container/port to forward to by default. If set the port-forward dialog will know default to your settings.

> NOTE: you can either use a port name or number in your annotation.

```yaml
# Pod fred
apiVersion: v1
kind: Pod
metadata:
  name: fred
  annotations:
    k9s.imhotep.io/default-portforward-container: bozo:p1 # => will default to container bozo port named p1
    # Or...
    k9s.imhotep.io/default-portforward-container: bozo:8081 # => will default to container bozo port number 8081
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

---

## Resolved Issues

* [Issue #1268](https://github.com/derailed/k9s/issues/1268) Age sorting wrong for years
* [Issue #1258](https://github.com/derailed/k9s/issues/1258) Configurable or recent use based port-forward
* [Issue #1257](https://github.com/derailed/k9s/issues/1257) Why is the latest chocolatey on 0.24.10

---

## PRs

* [PR #1266](https://github.com/derailed/k9s/pull/1266) Skin to preserve your terminal session background color
* [PR #1264](https://github.com/derailed/k9s/pull/1205) Adding note on popeye config
* [PR #1261](https://github.com/derailed/k9s/pull/1261) Blurry logo
* [PR #1250](https://github.com/derailed/k9s/pull/1250) Gruvbox dark skin
* [PR #1249](https://github.com/derailed/k9s/pull/1249) Node shell pod tolerate all taints
* [PR #1232](https://github.com/derailed/k9s/pull/1232) Add red skin for production env
* [PR #1227](https://github.com/derailed/k9s/pull/1227) Add abbreviation ReadWriteOncePod PV access mode

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
