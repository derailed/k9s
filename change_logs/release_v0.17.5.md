<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.17.5

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, please consider sponsoring 👆us or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/story/this_is_fine_300.png" align="center" width="500" height="auto"/>

## Thresholds Reloaded!

In the previous k9s release, we've introduced the notion of thresholds to provide with an alert mechanism when either the cpu or memory goes high on your clusters. Looking at the current solution, we felt we needed a bit more granularity in the severity levels thanks to [Eldad Assis](https://github.com/eldada) feedback on this one! So here is the new configuration for cluster thresholds. Please keep in mind this feature is still in flux!

```yaml
# $HOME/.k9s/config.yml
k9s:
  refreshRate: 2
  headless: false
  ...
  # Specify resources thresholds in percent - defaults: critical=90, warn=70
  thresholds:
    cpu:
      critical: 85
      warn: 75
    memory:
      critical: 80
      warn: 70
  ...
```

## Resolved Bugs/Features/PRs

- [Issue #604](https://github.com/derailed/k9s/issues/604)
- [Issue #601](https://github.com/derailed/k9s/issues/601) Thank you [Christian Vent](https://github.com/christian-vent)
- [Issue #598](https://github.com/derailed/k9s/issues/598)  `Ctrl-l` will now trigger the benchmarking toggle!
- [Issue #593](https://github.com/derailed/k9s/issues/593)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
