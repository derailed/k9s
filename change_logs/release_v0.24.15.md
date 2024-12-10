<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.24.15

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better are as ever very much noted and appreciated!

If you feel K9s is helping your Kubernetes journey, please consider joining our [sponsorship program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

## ♫ Sounds Behind The Release ♭

* [Paradise Delay - Marteria, DJ Kose](https://www.youtube.com/watch?v=eM-xTN8ggOs)
* [Fool For Your Stockings - ZZ Top - Sadly this one is a tribute to Dusty Hill ;(](https://www.youtube.com/watch?v=UExKTZ3veB8)

---

### A Word From Our Sponsors...

I want to recognize the following folks that have been kind enough to join our sponsorship program and opted to `pay it forward`!

* [Viacheslav Moskin](https://github.com/viacheslavmoskin)
* [Thomas Peter Bernsten](https://github.com/tpberntsen)
* [EMR-Bear](https://github.com/emrbear)

So if you feel K9s is helping with your productivity while administering your Kubernetes clusters, please consider pitching in as it will go a long way in ensuring a thriving environment for this repo and our K9sers community at large.

Thank you!!

---

## !!BREAKING CHANGE!!... We've moved!

As of this drop, k9s home directory is now configurable via [XDG](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html). Please see the specification depending on your platform of choice. You will now need to set or use the default for `$XDG_CONFIG_HOME` if not already present on your system. This is now the de facto replacement for`HOME/.k9s` as K9s will no longer honor this directory to load artifacts such as config, skins, views, etc... If you have existing customizations, you will need to move those over to your `$XDG_CONFIG_HOME/k9s` dir.

This feature is still fresh and we could have totally missed a piece, so please proceed with caution and keep that issue tracker handy...

Please join me in giving a Big Thank you! to [Arthur](https://github.com/pysen) for making this happen for us!

---

## Resolved Issues

* [Issue #1209](https://github.com/derailed/k9s/issues/1209) K9s - Popeye run instructions
* [Issue #1203](https://github.com/derailed/k9s/issues/1203) K9s does not remember last view I was in when switching contexts
* [Issue #1181](https://github.com/derailed/k9s/issues/1181) Cannot list roles

---

## PRs

* [PR #1213](https://github.com/derailed/k9s/pull/1213) Big Thanks to [Takumasa Sakao](https://github.com/sachaos)!
* [PR #1205](https://github.com/derailed/k9s/pull/1205) Great catch from [David Alger](https://github.com/davidalger)!
* [PR #1198](https://github.com/derailed/k9s/pull/1198) Once again [Takumasa Sakao](https://github.com/sachaos) to the rescue!!
* [PR #1196](https://github.com/derailed/k9s/pull/1196) ATTA Boy! [Daniel Lee Harple](https://github.com/dlh)
* [PR #1025](https://github.com/derailed/k9s/pull/1025) Big Thanks to [Arthur](https://github.com/pysen)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
