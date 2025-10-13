<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s.png" align="center" width="800" height="auto"/>

# Release v0.32.6

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s!
I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev
and see if we're happier with some of the fixes!
If you've filed an issue please help me verify and close.

Your support, kindness and awesome suggestions to make K9s better are, as ever, very much noted and appreciated!
Also big thanks to all that have allocated their own time to help others on both slack and on this repo!!

As you may know, K9s is not pimped out by corps with deep pockets, thus if you feel K9s is helping your Kubernetes journey,
please consider joining our [sponsorship program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

## Maintenance Release!

---

## Videos Are In The Can!

Please dial [K9s Channel](https://www.youtube.com/channel/UC897uwPygni4QIjkPCpgjmw) for up coming content...

* [K9s v0.31.0 Configs+Sneak peek](https://youtu.be/X3444KfjguE)
* [K9s v0.30.0 Sneak peek](https://youtu.be/mVBc1XneRJ4)
* [Vulnerability Scans](https://youtu.be/ULkl0MsaidU)

---

## Resolved Issues

* [#2947](https://github.com/derailed/k9s/issues/2947) CTRL+Z causes k9s to crash
* [#2938](https://github.com/derailed/k9s/issues/2938) Critical Vulnerability CVE-2024-41110 in v26.0.1 of docker included in k9s
* [#2929](https://github.com/derailed/k9s/issues/2929) conflicting plugins shortcuts
* [#2896](https://github.com/derailed/k9s/issues/2896) Add a plugin to disable/enable a keda ScaledObject
* [#2811](https://github.com/derailed/k9s/issues/2811) Dockerfile build step fails due to misaligned Go versions (1.21.5 vs 1.22.0)
* [#2767](https://github.com/derailed/k9s/issues/2767) Manually triggered jobs don't get automatically cleaned up
* [#2761](https://github.com/derailed/k9s/issues/2761) Enable "jump to owner" for more kinds
* [#2754](https://github.com/derailed/k9s/issues/2754) Plugins not loaded/shown in UI
* [#2747](https://github.com/derailed/k9s/issues/2747) Combining context and namespace switching only works sporadically (e.g. ":pod foo-ns @ctx-dev")
* [#2746](https://github.com/derailed/k9s/issues/2746) k9s does not display "[::]" string in its logs
* [#2738](https://github.com/derailed/k9s/issues/2738) "Faults" view should show all Terminating pods

---

## Contributed PRs

Please be sure to give `Big Thanks!` and `ATTA Girls/Boys!` to all the fine contributors for making K9s better for all of us!!

* [#2937](https://github.com/derailed/k9s/pull/2937) Adding Argo Rollouts plugin version for PowerShell
* [#2935](https://github.com/derailed/k9s/pull/2935) fix: show all terminating pods in Faults view (#2738)
* [#2933](https://github.com/derailed/k9s/pull/2933) chore: broken url in build-status tag in the readme.md
* [#2932](https://github.com/derailed/k9s/pull/2932) fix: add kubeconfig if k9s is launched with --kubeconfig
* [#2930](https://github.com/derailed/k9s/pull/2930) fixed conflicting plugin shortcuts, and added 2 new plugins
* [#2927](https://github.com/derailed/k9s/pull/2927) Fix "Mark Range": reduce maximum namespaces in favorites, fix shadowing of ctrl+space
* [#2926](https://github.com/derailed/k9s/pull/2926) chore(plugins,remove-finalizers): make sure the resources api group is respected
* [#2921](https://github.com/derailed/k9s/pull/2921) feat: Add plugins for kubectl node-shell
* [#2920](https://github.com/derailed/k9s/pull/2920) eat: added StartupProbes status (S) to the PROBES column in the container render
* [#2914](https://github.com/derailed/k9s/pull/2914) Adding eks-node-viewer plugin
* [#2898](https://github.com/derailed/k9s/pull/2898) Add argocd plugin to community plugins
* [#2896](https://github.com/derailed/k9s/pull/2896) feat(2896): Add toggle keda plugin
* [#2890](https://github.com/derailed/k9s/pull/2890) Update README.md
* [#2881](https://github.com/derailed/k9s/pull/2881) Fix Mark-Range command: ensure that NS Favorite doesn't exceed the limit
* [#2861](https://github.com/derailed/k9s/pull/2861) chore: fix function name
* [#2856](https://github.com/derailed/k9s/pull/2856) fix internal/render/hpa.go merge issue
* [#2848](https://github.com/derailed/k9s/pull/2848) Include sidecar containers requests and limits
* [#2844](https://github.com/derailed/k9s/pull/2844) Update README GO Version Required
* [#2830](https://github.com/derailed/k9s/pull/2830) update tview to fix log escaping problem completely
* [#2822](https://github.com/derailed/k9s/pull/2822) Adding HolmesGPT plugin
* [#2821](https://github.com/derailed/k9s/pull/2821) Add a spark-operator plugin
* [#2817](https://github.com/derailed/k9s/pull/2817) Add comment about Escape keybinding
* [#2812](https://github.com/derailed/k9s/pull/2812) fix: align build image Go version with go.mod
* [#2795](https://github.com/derailed/k9s/pull/2795) add new plugin current-ctx-terminal
* [#2791](https://github.com/derailed/k9s/pull/2791) Add leading space to Kubernetes context suggestions
* [#2789](https://github.com/derailed/k9s/pull/2789) Create kubectl-get-in-shell.yaml
* [#2788](https://github.com/derailed/k9s/pull/2788) Update README.md plugin format
* [#2787](https://github.com/derailed/k9s/pull/2787) Update helm-purge.yaml
* [#2786](https://github.com/derailed/k9s/pull/2786) Update README.md with plugin dangerous field
* [#2780](https://github.com/derailed/k9s/pull/2780) install copyright file into correct location
* [#2775](https://github.com/derailed/k9s/pull/2775) fix freebsd build failure
* [#2780](https://github.com/derailed/k9s/pull/2780) install copyright file into correct location
* [#2772](https://github.com/derailed/k9s/pull/2772) proper handle OwnerReference for manually created job
* [#2771](https://github.com/derailed/k9s/pull/2771) feat: add duplik8s plugin
* [#2770](https://github.com/derailed/k9s/pull/2770) feat: allow plugins block in plugin files
* [#2765](https://github.com/derailed/k9s/pull/2765) fix: Shellin -> ShellIn
* [#2763](https://github.com/derailed/k9s/pull/2763) enable "jump to owner" for more kinds
* [#2755](https://github.com/derailed/k9s/pull/2755) Loki plugin
* [#2751](https://github.com/derailed/k9s/pull/2751) container logs should be escaped when printed
* [#2750](https://github.com/derailed/k9s/pull/2750) fix: should switching ctx before ns

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2024 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)