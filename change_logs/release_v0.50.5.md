<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s.png" align="center" width="800" height="auto"/>

# Release v0.50.5

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

## Resolved Issues

* [#3328](https://github.com/derailed/k9s/issues/3328) Pod overview shows wrong number of running containers with sidecar init-container
* [#3309](https://github.com/derailed/k9s/issues/3309) [0.50.4] k9s crashes when attempting to load logs
* [#3301](https://github.com/derailed/k9s/issues/3301) Port Forward deleted without UI notification when forwarding to wrong port
* [#3294](https://github.com/derailed/k9s/issues/3294) [0.50.4] k9s crashes when filtering based on labels
* [#3278](https://github.com/derailed/k9s/issues/3278) k9s doesn't honor the --namespace parameter

---

## Contributed PRs

Please be sure to give `Big Thanks!` and `ATTA Girls/Boys!` to all the fine contributors for making K9s better for all of us!!

* [#3311](https://github.com/derailed/k9s/pull/3311) Fix concurrent read writes
* [#3310](https://github.com/derailed/k9s/pull/3310) fix: use full path of date to avoid conflict
* [#3308](https://github.com/derailed/k9s/pull/3308) Show replicasets from deployment view
* [#3300](https://github.com/derailed/k9s/pull/3300) fix: truncate label selector input to max length
* [#3296](https://github.com/derailed/k9s/pull/3296) fix: update time format in logging to 24-hour format

---
<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2025 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)