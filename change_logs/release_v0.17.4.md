<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.17.4

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, please consider sponsoring 👆us or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/story/this_is_fine.png" align="center" width="500" height="auto"/>

## Pulses Part Duh!

In this drop, we've updated pulses to now show used/allocatable resources for cpu and mem as recommended by the awesome and kind [Eldad Assis](https://github.com/eldada)! We've also added the concept of threshold to alert you when things in your clusters are going south. These currently come in the shape of cpu and mem thresholds. They are set at the cluster level. K9s will now let you know when these limits are reached or surpassed. As it stands, the k9s logo will change color and a flash message will appear to let you know which resource thresold was exceeded. Once the load subsumes the logo/flash will return to their orginal states.

In order to override the default thresholds (cpu/mem: 80% ), you will need to modify your `$HOME/.k9s/config.yml` using the new config section named `thresholds` as follows:

```yaml
# $HOME/.k9s/config.yml
k9s:
  refreshRate: 2
  headless: false
  ...
  # Specify resources thresholds percentages
  thresholds:
    cpu:    80 # default is 80
    memory: 55 # default is 80
  ...
```

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/story/pulses_tripped.png" align="center" width="500" height="auto"/>

## Resolved Bugs/Features/PRs

- [Issue #596](https://github.com/derailed/k9s/issues/596)
- [Issue #593](https://github.com/derailed/k9s/issues/593)
- [Issue #560](https://github.com/derailed/k9s/issues/560)
  - NOTE!! All credits here goes to [Bruno Meneguello](https://github.com/bkmeneguello) and [Michael Cristina](https://github.com/mcristina422) for making this possible in K9s!

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
