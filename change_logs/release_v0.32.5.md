<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s.png" align="center" width="800" height="auto"/>

# Release v0.32.5

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

* [#2734](https://github.com/derailed/k9s/issues/2734) Incorrect pod containers displayed when using custom resource columns
* [#2733](https://github.com/derailed/k9s/issues/2733) Toggle Wide and Toggle Faults broken for PDB view
* [#2656](https://github.com/derailed/k9s/issues/2656) nil pointer dereference when switching contexts
* [#2617](https://github.com/derailed/k9s/issues/2617) Plugin command execution output

---

## Contributed PRs

Please be sure to give `Big Thanks!` and `ATTA Girls/Boys!` to all the fine contributors for making K9s better for all of us!!

* [#2736](https://github.com/derailed/k9s/pull/2736) fix view sorting being reset
* [#2732](https://github.com/derailed/k9s/pull/2732) use policy/v1 instead of policy/v1beta1
* [#2728](https://github.com/derailed/k9s/pull/2728) feat: add pool col to node view
* [#2718](https://github.com/derailed/k9s/pull/2718) fix: jump to namespaceless owner reference
* [#2711](https://github.com/derailed/k9s/pull/2711) Add plugins for argo-rollouts
* [#2700](https://github.com/derailed/k9s/pull/2700) feat: allow jumping to the owner of the resource
* [#2699](https://github.com/derailed/k9s/pull/2699) Added cert-manager and openssl plugins
* [#2711](https://github.com/derailed/k9s/pull/2711) Add plugins for argo-rollouts
* [#2698](https://github.com/derailed/k9s/pull/2698) fix: job color based on failures (#2686)
* [#2685](https://github.com/derailed/k9s/pull/2685) feat: support cluster and cmp view
* [#2678](https://github.com/derailed/k9s/pull/2678) fix: do not hard-code path to kubectl in jq plugin
* [#2676](https://github.com/derailed/k9s/pull/2676) Add kanagawa skin
* [#2666](https://github.com/derailed/k9s/pull/2666) save config when closing k9s with ctrl-c
* [#2644](https://github.com/derailed/k9s/pull/2644) Allow overwriting plugin output with command's stdout

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2024 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)