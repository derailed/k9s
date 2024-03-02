<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s.png" align="center" width="800" height="auto"/>

# Release v0.32.0

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

A lot of refactors, perf improvements (crossing fingers+toes!) and general spring cleaning items in this release.
Thus I expect a bit of `disturbance in the farce` given the major code churns, so please beware!

---

## Videos Are In The Can!

Please dial [K9s Channel](https://www.youtube.com/channel/UC897uwPygni4QIjkPCpgjmw) for up coming content...

* [K9s v0.31.0 Configs+Sneak peek](https://youtu.be/X3444KfjguE)
* [K9s v0.30.0 Sneak peek](https://youtu.be/mVBc1XneRJ4)
* [Vulnerability Scans](https://youtu.be/ULkl0MsaidU)

---

## A Word From Our Sponsors...

To all the good folks below that opted to `pay it forward` and join our sponsorship program, I salute you!!

* [Justin Reid](https://github.com/jmreid)
* [Danni](https://github.com/danninov)
* [Robert Krahn](https://github.com/rksm)
* [Hao Ke](https://github.com/kehao95)
* [PH](https://github.com/raphael-com-ph)

> Sponsorship cancellations since the last release: **9!!** 🥹

---

## Resolved Issues

* [#2569](https://github.com/derailed/k9s/issues/2569) k9s panics on start if the main config file (config.yml) is owned by root
* [#2568](https://github.com/derailed/k9s/issues/2568) kube context in running k9s is no longer sticky, during kubectx context switch
* [#2560](https://github.com/derailed/k9s/issues/2560) Namespace/Settings keeps resetting
* [#2557](https://github.com/derailed/k9s/issues/2557) [Feature]: Sort CRDs by their group
* [#1462](https://github.com/derailed/k9s/issues/1462) k9s running very slowly when opening namespace with 13k pods (maybe??)

---

## Contributed PRs

Please be sure to give `Big Thanks!` and `ATTA Girls/Boys!` to all the fine contributors for making K9s better for all of us!!

* [#2564](https://github.com/derailed/k9s/pull/2564) Add everforest skins
* [#2558](https://github.com/derailed/k9s/pull/2558) feat: sort by role in node list view
* [#2554](https://github.com/derailed/k9s/pull/2554) Added context to the debug command for debug-container plugin
* [#2554](https://github.com/derailed/k9s/pull/2554) Correctly respect the KUBECACHEDIR env var
* [#2546](https://github.com/derailed/k9s/pull/2546) Use configured log fgColor to print log markers

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2024 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)