<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.22.0

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better are as ever very much noted and appreciated!

If you feel K9s is helping your Kubernetes journey, please consider joining our [sponsorship program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

Maintenance Release!

## A Word From Our Sponsors...

First off, I would like to send a `Big Thank You` to the following generous K9s friends for joining our sponsorship program and supporting this project!

* [Martin Kemp](https://github.com/MartiUK)

Contrarily to popular belief, OSS is not free! We've now reached ~9k stars and 300k downloads! As you all know, this project is not pimped out by a big company with deep pockets and a large team. K9s is complex and does demand a lot of my time. So if this tool is useful to you and part of your daily lifecycle, please contribute! Your contribution whether financial, PRs, issues or shout-outs on social/blogs are crucial to keep K9s growing and powerful for all of us. Don't let OSS by individual contributors become an oxymoron!

## I Should've known better

Seems like I've broken the golden rule ie never add a feature without providing an option to turn it off ;( It looks like enable mouse support for K9s had unexpected side effects. So in this drop, we're introducing a new configuration aka `enableMouse` that defaults to `false`. You can opt-in mouse support, by enabling it in the K9s config file. That said when mouse support is enabled, you can still use terminal selection using either `Shift/Option` for Windows/Mac.

```yaml
# $HOME/.k9s/config.yml
k9s:
  refreshRate: 2
  enableMouse: true # Defaults to false if not set
  headless: false
  ...
```

## Resolved Issues/Features

* [Issue #874](https://github.com/derailed/k9s/issues/874) Latest version broke selecting text by mouse

## Resolved PRs

* [PR #877](https://github.com/derailed/k9s/pull/877) Change character used for X in RBAC view. Thank you! [Torjus](https://github.com/torjue)
* [PR #876](https://github.com/derailed/k9s/pull/876) Migrate to new sortorder import path. Big thanks to [fbbommel](https://github.com/fvbommel)
* [PR #873](https://github.com/derailed/k9s/pull/873) Fix default logger config, same as README. Thank you! [darklore](https://github.com/darklore)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
