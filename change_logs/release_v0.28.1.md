<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s.png" align="center" width="800" height="auto"/>

# Release v0.28.1

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better are, as ever, very much noted and appreciated! Also big thanks to all that have allocated their own time to help others on both slack and on this repo!!

As you may know, K9s is not pimped out by corps with deep pockets, thus if you feel K9s is helping your Kubernetes journey, please consider joining our [sponsorship program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

## ♫ Sounds Behind The Release ♭

* [If Trouble Was Money - Albert Collins](https://www.youtube.com/watch?v=cz6LbWWqX-g)
* [Old Love - Eric Clapton](https://www.youtube.com/watch?v=EklciRHZnUQ)
* [Touch And GO - The Cars](https://www.youtube.com/watch?v=L7Gpr_Auz8Y)

---

## A Word From Our Sponsors...

To all the good folks below that opted to `pay it forward` and join our sponsorship program, I salute you!!

* [Bradley Heilbrun](https://github.com/bheilbrun)

> Sponsorship cancellations since the last release: `2` ;(

---

## Feature Release

### Sanitize Me!

Over time, you might end up with a lot of pod cruft on your cluster. Pods that might be completed, erroring out, etc... Once you've completed your pod analysis it could be useful to clear out these pods from your cluster.

In this drop, we introduce a new command `sanitize` aka `z` available on pod views otherwise known as `The Axe!`. This command performs a clean up of all pods that are in either in completed, crashloopBackoff or failed state. This could be especially handy if you run workflows jobs or commands on your cluster that might leave lots of `turd` pods. Tho this has a `phat` fail safe dialog please be careful with this one as it is a blunt tool!

---

## Resolved Issues

* [Issue #2281](https://github.com/derailed/k9s/issues/2281) Can't run Node shell
* [Issue #2277](https://github.com/derailed/k9s/issues/2277) bulk actions applied to power filters
* [Issue #2273](https://github.com/derailed/k9s/issues/2273) Error when draining node that is cordoned bug
* [Issue #2233](https://github.com/derailed/k9s/issues/2233) Invalid port-forwarding status displayed over the k9s UI

---

## Contributed PRs

Please be sure to give `Big Thanks!` and `ATTA Girls/Boys!` to all the fine contributors for making K9s better for all of us!!

* [PR #2280](https://github.com/derailed/k9s/pull/2280) chore: replace github.com/ghodss/yaml with sigs.k8s.
* [PR #2278](https://github.com/derailed/k9s/pull/2278) README.md: fix typo in netshoot URL
* [PR #2275](https://github.com/derailed/k9s/pull/2275) check if the Node already cordoned when executing Drain
* [PR #2247](https://github.com/derailed/k9s/pull/2247) Delete port forwards when pods get deleted

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2023 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
