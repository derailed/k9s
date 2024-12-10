<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.24.8

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better are as ever very much noted and appreciated!

If you feel K9s is helping your Kubernetes journey, please consider joining our [sponsorship program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

## Maintenance Release!

### NodeShell args

In this drop, we've added additional configurations to the k9s node shell so you override the command and args on the node shell containers.

```yaml
    # $HOME/.k9s/config.yml
    ...
    minikube:
      view:
        active: pod
      featureGates:
        nodeShell: true
      shellPod:
        image: busybox:1.31
        # New!
        command: ["/bin/sh", "-c"]
        # New!
        args: ["ls -al"]
        namespace: default
        limits:
          cpu: 100m
          memory: 100Mi
     ...
```

---

## Resolved Issues

* [Issue #1106](https://github.com/derailed/k9s/issues/1106) Remove padding while in full screen
* [Issue #1104](https://github.com/derailed/k9s/issues/1104) Config args for shellPod
* [Issue #1102](https://github.com/derailed/k9s/issues/1102) Explicitly announce no metrics are available
* [Issue #1097](https://github.com/derailed/k9s/issues/1097) Delete resource dialog stopped working
* [Issue #1093](https://github.com/derailed/k9s/issues/1094) Leading comma in command column
* [Issue #1094](https://github.com/derailed/k9s/issues/1094) Screendumps empty on EKS
* [Issue #1060](https://github.com/derailed/k9s/issues/1060) Exception when setting container image
* [Issue #1081](https://github.com/derailed/k9s/issues/1081) Color issue on startup
* [Issue #1078](https://github.com/derailed/k9s/issues/1078) Nord skin
* [Issue #1075](https://github.com/derailed/k9s/issues/1075) Crash on mouse click out of main window
* [Issue #1070](https://github.com/derailed/k9s/issues/1070) lose cursor on windows 10
* [Issue #1068](https://github.com/derailed/k9s/issues/1068) Build error 0.24.7
* [Issue #1063](https://github.com/derailed/k9s/issues/1063) Weird colour scheme on windows

## Resolved PRs

* [PR #1101](https://github.com/derailed/k9s/pull/1101) propagate insecure-skip-tls-verify

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
