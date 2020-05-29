<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.20.2

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, consider joining our [sponsorhip program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

Maintenance Release! Fixing a few issue in the v0.20 aftermath ;(

## Selection Marker

In this drop, we're adding the ability to set row marks ranges. There are situations where you've filtered a resource and need to delete part or all of the rows. In previous releases, you had to mark each rows one by one. Now you have the ability to select a beginning and an end range and all rows in between will now be marked! To mark a single row, you can use `space`. To select rows between your initial mark to the current selection use `Ctrl-space`. To nuke all marked rows use `Ctrl-\`. All credits and ATTA BOY goes to [Ryan Richard](https://github.com/cfryanr) for suggesting this feature!!

## Logs Got Some TLC!

Per [Raman Gupta](https://github.com/rocketraman) excellent suggestion, we've added a way to add a separator to your chatty logs to easily see the latest incoming logs. While in log view you can now press `m` to add the separator to the log stream. If you don't care about the log history and just want to see the latest incoming logs, pressing `c` will clear out the log viewer.

## Resolved Bugs/Features/PRs

- [Issue #741](https://github.com/derailed/k9s/issues/741)
- [Issue #740](https://github.com/derailed/k9s/issues/740)
- [Issue #739](https://github.com/derailed/k9s/issues/739)
- [Issue #727](https://github.com/derailed/k9s/issues/727)
- [Issue #723](https://github.com/derailed/k9s/issues/723)
- [PR #725](https://github.com/derailed/k9s/pull/725) Big Thanks To [Soupyt](https://github.com/soupyt)!!

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
