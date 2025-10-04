<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s.png" align="center" width="800" height="auto"/>

# Release v0.50.10

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s!
I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev
and see if we're happier with some of the fixes!
If you've filed an issue please help me verify and close.

Your support, kindness and awesome suggestions to make K9s better are, as ever, very much noted and appreciated!
Also big thanks to all that have allocated their own time to help others on both slack and on this repo!!

As you may know, K9s is not pimped out by corps with deep pockets, thus if you feel K9s is helping your Kubernetes journey,
please consider joining our [sponsorship program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/zt-3360a389v-ElLHrb0Dp1kAXqYUItSAFA)

## Maintenance Release!

---

## A Word From Our Sponsors...

To all the good folks below that opted to `pay it forward` and join our sponsorship program, I salute you!!

* [rufusshrestha](https://github.com/rufusshrestha)
* [Ovidijus Balkauskas](https://github.com/Stogas)
* [Konrad Konieczny](https://github.com/Psyhackological)
* [Serit Tromsø](https://github.com/serit)
* [Dennis](https://github.com/dennisTGC)
* [LinPr](https://github.com/LinPr)
* [franzXaver987](https://github.com/franzXaver987)
* [Drew Showalter](https://github.com/one19)
* [Sandylen](https://github.com/Sandylen)
* [Uriah Carpenter](https://github.com/uriahcarpenter)
* [Vector Group](https://github.com/vectorgrp)
* [Stefan Roman](https://github.com/katapultcloud)
* [Phillip](https://github.com/Loki-Afro)
* [Lasse Bang Mikkelsen](https://github.com/lassebm)

> Sponsorship cancellations since the last release: **19!** 🥹

---

## Resolved Issues

* [#3541](https://github.com/derailed/k9s/issues/3541) ServiceAccount RBAC Rules not displayed if RoleBinding subject doesn't specify namespace
* [#3535](https://github.com/derailed/k9s/issues/3535) Current Release process will cause code changes been reverted
* [#3525](https://github.com/derailed/k9s/issues/3525) k9s suspends when launching foreground plugin
* [#3495](https://github.com/derailed/k9s/issues/3495) Regression: filtering no long works with aliases
* [#3478](https://github.com/derailed/k9s/issues/3478) High Disk and CPU usage when imageScans Is enabled in K9s
* [#3470](https://github.com/derailed/k9s/issues/3470) Aliases for pods with unequal (!=) label filters not working
* [#3466](https://github.com/derailed/k9s/issues/3466) Shared GPU (nvidia.com/gpu.shared) is shown as n/a on K9s node view
* [#3455](https://github.com/derailed/k9s/issues/3455) memory command not found

---

## Contributed PRs

Please be sure to give `Big Thanks!` and `ATTA Girls/Boys!` to all the fine contributors for making K9s better for all of us!!

* [#3558](https://github.com/derailed/k9s/pull/3558) refactor(duplik8s): consolidate duplicate resource commands and updat…
* [#3555](https://github.com/derailed/k9s/pull/3555) feat: add dup plugin
* [#3543](https://github.com/derailed/k9s/pull/3543) Make "flux trace" more generic
* [#3536](https://github.com/derailed/k9s/pull/3536) Add flux-operator resources to flux plugin
* [#3528](https://github.com/derailed/k9s/pull/3528) feat(plugins): add pvc debug container plugin
* [#3517](https://github.com/derailed/k9s/pull/3517) Feature/refresh rate
* [#3516](https://github.com/derailed/k9s/pull/3516) Fixes flickering/jumping issue in context suggestions caused by inconsistent spacing behavior
* [#3515](https://github.com/derailed/k9s/pull/3515) Fix/suppress init no resources warning
* [#3513](https://github.com/derailed/k9s/pull/3513) fix: Color PV row according to its STATUS column
* [#3513](https://github.com/derailed/k9s/pull/3513) fix: Color PV row according to its STATUS column
* [#3505](https://github.com/derailed/k9s/pull/3505) docs: Add installation method with gah
* [#3503](https://github.com/derailed/k9s/pull/3503) fix(logs): enhance log streaming with retry mechanism and error handling
* [#3489](https://github.com/derailed/k9s/pull/3489) feat: Add context deletion functionality
* [#3487](https://github.com/derailed/k9s/pull/3487) fsupport core group resources in k9s/plugins/watch-events.yaml
* [#3485](https://github.com/derailed/k9s/pull/3485) Add disable-self-subject-access-reviews flag to disable can-i check…
* [#3464](https://github.com/derailed/k9s/pull/3464) fix: get-all command in get all plugin

---
<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2025 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)#