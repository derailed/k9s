<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s.png" align="center" width="800" height="auto"/>

# Release v0.27.0

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better are, as ever, very much noted and appreciated! Also big thanks to all that have allocated their own time to help others on both slack and on this repo!!

As you may know, K9s is not pimped out by corps with deep pockets, thus if you feel K9s is helping your Kubernetes journey, please consider joining our [sponsorship program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

## Maintenance Release

---

## ♫ Sounds Behind The Release ♭

I'd like to dedicate this release to `Jeff Beck` one of my all time favorite musicians that sadly passed away this last week ;(

* [The Pump - Jeff Beck](https://www.youtube.com/watch?v=xiDYrQp9wFQ)
* [Brush With The Blues - Jeff Beck](https://www.youtube.com/watch?v=O640IGLjnfs)
* [Cause We've Ended As Lovers - Jeff Beck](https://www.youtube.com/watch?v=VC02wGj5gPw)
* [Where Were You - Jeff Beck](https://www.youtube.com/watch?v=howz7gVecjE)
* [Rockabilly Set At Ronnie Scott](https://www.youtube.com/watch?v=_3aIEzXHBWw)

---

## A Word From Our Sponsors...

To all the good folks below that opted to `pay it forward` and join our sponsorship program, I salute you!!

* [Vibin reddy](https://github.com/vibin)
* [Maciek Albin](https://github.com/mckk)
* [Dherraj Yennam](https://github.com/dyennam)
* [Alan Ream](https://github.com/aream2006)
* [djheap](https://github.com/djheap)
* [MaterializeInc](https://github.com/MaterializeInc)
* [Jeff Evans](https://github.com/jeff303)

---

## Resolved Issues

* [Issue #1917](https://github.com/derailed/k9s/issues/1917) Crash on open single ingress from list
* [Issue #1906](https://github.com/derailed/k9s/issues/1680) k9s exits silently if screenDumpDir cannot be created
* [Issue #1661](https://github.com/derailed/k9s/issues/1661) ClusterRole with wrong privilege list display
* [Issue #1680](https://github.com/derailed/k9s/issues/1680) Change pod kill grace period for 0 to 1

## Contributed PRs

Please give `Big Thanks!` and `ATTA Girls/Boys!` to all the fine contributors for making K9s better for all of us!!

* [PR #1910](https://github.com/derailed/k9s/pull/1910) Replace x86_64 to amd64 build
* [PR #1877](https://github.com/derailed/k9s/pull/1877) Bug: portforward custom containers not showing
* [PR #1874](https://github.com/derailed/k9s/pull/1874) Feat: Add noLatestRevCheck config option
* [PR #1872](https://github.com/derailed/k9s/pull/1872) Docs: Add k8s client compatibility matrix
* [PR #1871](https://github.com/derailed/k9s/pull/1871) Bug: update scanSA calls to account for blank service accounts
* [PR #1866](https://github.com/derailed/k9s/pull/1866) Bug: Fix order of arguments for CanI function call
* [PR #1859](https://github.com/derailed/k9s/pull/1859) FEAT: Add vim-like quit force option
* [PR #1849](https://github.com/derailed/k9s/pull/1849) Bug: Fix build date for OSX
* [PR #1847](https://github.com/derailed/k9s/pull/1847) FEAT: Add labels configuration for shell node pod
* [PR #1840](https://github.com/derailed/k9s/pull/1840) FEAT: Add policy view to service accounts
* [PR #1837](https://github.com/derailed/k9s/pull/1837) FEAT: Use default terminal colors for better readability
* [PR #1830](https://github.com/derailed/k9s/pull/1830) FEAT: Plugin support for carvel kapp CR
* [PR #1829](https://github.com/derailed/k9s/pull/1829) FEAT: flux.yml plugin new displays stderr messages

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2022 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
