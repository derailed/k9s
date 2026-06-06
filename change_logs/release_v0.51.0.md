<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s.png" align="center" width="800" height="auto"/>

# Release v0.51.0

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s!
I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev
and see if we're happier with some of the fixes!
If you've filed an issue please help me verify and close.

Your support, kindness and awesome suggestions to make K9s better are, as ever, very much noted and appreciated!
Also big thanks to all that have allocated their own time to help others on both slack and on this repo!!

As you may know, K9s is not pimped out by big corporations with deep pockets, thus if you feel K9s is helping in your Kubernetes journey, please consider joining our [sponsorship program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/zt-3360a389v-ElLHrb0Dp1kAXqYUItSAFA)

---

## ♫ Sounds Behind The Release ♭

* [Aprieta - Vincen Garcia](https://www.youtube.com/watch?v=ldQ6hpg9BD0&list=RDldQ6hpg9BD0&start_radio=1)
* [Graham Chapman - John Cleese](https://www.youtube.com/watch?v=Bm2XPkqENaw)
* [Kill the pain - SYZGYX](https://www.youtube.com/watch?v=5XuvMhHZorw&list=RD5XuvMhHZorw&start_radio=1)

---

## Maintenance Release!

Please help me welcome Ümüt Özalp as a core contributor to K9s!
Ümüt has been instrumental in helping this project grow.
I trust you will help Ümüt triage issues and prs reviews and show him
the kindness and patience all k9sers are famous for!

Sponsorships are dropping at an alarming rate which puts this project in the red.
This is becoming a concern and sad not to mention unsustainable ;(
If you dig `k9s` and want to help the project, please consider `paying it forward!` and
don't become just another `satisfied, non paying customer!`.
K9s does take a lot of my `free` time to maintain, enhance and keep the light on.
Many cool ideas are making it straight to the `freezer` as I just can't budget them in.
I know many of you work for big corporations, so please put in the word/work and have
them help us out via sponsorships or other means.

Thank you!

---

## Contributed PRs

Please be sure to give `Big Thanks!` and `ATTA Girls/Boys!` to all the fine contributors for making K9s better for all of us!!

* [#4026](https://github.com/derailed/k9s/pull/4026) fix(xray): disable edit/delete actions in XRay view when readonly mode is enabled
* [#4024](https://github.com/derailed/k9s/pull/4024) Fix 'J'umping to owner of cluster scoped resources
* [#4005](https://github.com/derailed/k9s/pull/4005) Fix pod status for sidecar init containers
* [#4001](https://github.com/derailed/k9s/pull/4001) Adjust namespace handling for RBAC checks in CanForResource and CanForInstance
* [#3997](https://github.com/derailed/k9s/pull/3997) chore: fix wrong function name in comment
* [#3993](https://github.com/derailed/k9s/pull/3993) fix(browser): show syncing status instead of spurious no-resources warning
* [#3989](https://github.com/derailed/k9s/pull/3989) perf: skip reconcile cycle when informer data is unchanged
* [#3988](https://github.com/derailed/k9s/pull/3988) perf: raise default client QPS from 5 to 50
* [#3987](https://github.com/derailed/k9s/pull/3987) fix: paginate metrics API calls to prevent timeout on large clusters
* [#3986](https://github.com/derailed/k9s/pull/3986) perf: batch Hydrate workers to eliminate per-item goroutine overhead
* [#3917](https://github.com/derailed/k9s/pull/3917) Respect wide columns in default view
* [#3911](https://github.com/derailed/k9s/pull/3911) fix: reset styles before loading skin on context switch
* [#3908](https://github.com/derailed/k9s/pull/3908) fix: populate pod count in Node.Get() for single-node view
* [#3902](https://github.com/derailed/k9s/pull/3902) Add OSC52 clipboard backend with native fallback
* [#3888](https://github.com/derailed/k9s/pull/3888) feat: add One Light skin
* [#3879](https://github.com/derailed/k9s/pull/3879) feat: allow users to cycle pulse grid selection forwards and backwards
* [#3873](https://github.com/derailed/k9s/pull/3873) feat: enhance pvc-shell configuration with dynamic inputs and RWO support
* [#3872](https://github.com/derailed/k9s/pull/3872) Add default confirm:true for plugins with inputs
* [#3871](https://github.com/derailed/k9s/pull/3871) internal/render: prevent index out of range panic in initContainerStats
* [#3865](https://github.com/derailed/k9s/pull/3865) Handle blank PVC capacities for the purpose of sorting
* [#3854](https://github.com/derailed/k9s/pull/3854) feat: enhance debug container configuration with input fields
* [#3851](https://github.com/derailed/k9s/pull/3851) Use *grey instead of grey in black-and-wtf.yaml
* [#3839](https://github.com/derailed/k9s/pull/3839) fix: optimize context switching to reduce redundant API calls
* [#3823](https://github.com/derailed/k9s/pull/3823) feat: add resize PVC plugin for dynamic storage resizing
* [#3821](https://github.com/derailed/k9s/pull/3821) feat: add support for plugin input fields
* [#3817](https://github.com/derailed/k9s/pull/3817) Fix Readme: Ubuntu installation command not working
* [#3798](https://github.com/derailed/k9s/pull/3798) Fix boom on Jumping Owner in rare cases
* [#3797](https://github.com/derailed/k9s/pull/3797) feat: add extra hints for column navigation in table view
* [#3792](https://github.com/derailed/k9s/pull/3792) fix: adjust resource access checks for namespace resources
* [#3783](https://github.com/derailed/k9s/pull/3783) fix: avoid logging errors when no context is configured
* [#3780](https://github.com/derailed/k9s/pull/3780) feat: add selected color to table header

---
<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2026 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)#