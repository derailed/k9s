<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s.png" align="center" width="800" height="auto"/>

# Release v0.26.0

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better are as ever very much noted and appreciated! Also big thanks to all that have allocated their own time to help others on both slack and on this repo!!

If you feel K9s is helping your Kubernetes journey, please consider joining our [sponsorship program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

## ♫ Sounds Behind The Release ♭

* [Sugar Water - Cibo Matto](https://www.youtube.com/watch?v=EN9auBn6Jys)
* [Midnight To Stevens - The Clash](https://www.youtube.com/watch?v=9suQJthS6to)
* [Cool & Proper - Natty Nation](https://www.youtube.com/watch?v=9q337zn7bpI)

---

## Maintenance Release

Please join me in giving a big THANK YOU and ATTA BOY!! to [Aleksei Romanenko](https://github.com/slimus) for allocating his personal time in helping out his fellow K9sers with issues, PRs and slack!!

Also in the last drop, I'd updated k8s API's to the latest which caused some `disturbance in the farce!` and hosed AWS cluster connections in the same swop ;( Please see [Issue#119](https://github.com/derailed/k9s/issues/1619) for `a` resolve... I did not catch it early enough hence the release bump on this drop. My bad!!

---

## Resolved Issues

* [Issue #1655](https://github.com/derailed/k9s/issues/1655) Text not appearing in context windows
* [Issue #1654](https://github.com/derailed/k9s/issues/1654) K9s crash on m1 with index out of range [0] with length 0
* [Issue #1652](https://github.com/derailed/k9s/issues/1652) HPA with custom metrics has "Target%" column showing "unknown/unknown"
* [Issue #1639](https://github.com/derailed/k9s/issues/1639) Helm releases view broken after interacting with 0.25.21

## Resolved PR

* [PR #1656](https://github.com/derailed/k9s/pull/156) Fix PF and RS dialog colors
* [PR #163](https://github.com/derailed/k9s/pull/1636) Fix #1636: can't switch context with --kubeconfig flag

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2021 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
