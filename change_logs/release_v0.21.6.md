<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.21.6

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better are as ever very much noted and appreciated!

If you feel K9s is helping your Kubernetes journey, please consider joining our [sponsorhip program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

## First A Word From Our Sponsors...

First off, I would like to send a `Big Thank You` to the following generous K9s friends for joining our sponsorship program and supporting this project!

* [Drew](https://github.com/ScubaDrew)
* [Vladimir Rybas](https://github.com/vrybas)

Contrarily to popular belief, OSS is not free! We've now reached 8k stars and 270k downloads! As you all know, this project is not pimped out by a big company with deep pockets or a large team. This project is complex and does demand a lot of my time. So if k9s is useful to you and part of your daily lifecycle. Please contribute! Your contribution whether financial, PRs, issues or shout-outs on social/blogs are crucial to keep K9s growing and powerful for all of us!
Don't let OSS by individual contributors become an oxymoron...

## New Skins On The Block!

In this drop, big thanks are in effect for [Dan Mikita](https://github.com/danmikita) for contributing a new K9s [solarized theme](https://github.com/derailed/k9s/tree/master/skins)!

Also we've added a new skin configuration for table's cursor namely `cursorFgColor` and `cursorBgColor`:

```yaml
  # skin.yml
  ...
  views:
    table:
      fgColor: *foreground
      bgColor: *background
      cursorFgColor: *foreground
      cursorBgColor: *current_line
      header:
        fgColor: white
        bgColor: *background
        sorterColor: *cyan
  ...
```

## Resolved Bugs/Features/PRs

Maintenance Release

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
