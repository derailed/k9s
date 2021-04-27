<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.20.4

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, consider joining our [sponsorship program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

## PersistentVolumeClaims Reference Tracking

In continuation with the resource usage check feature added in v0.20, we've added reference checks on the PVC view. If you ever wonder which resources on your cluster are referencing a given PVC, simply press `u` for `UsedBy` and k9s will tell you.

## New Config On The Block

Some folks voiced concerns with K9s config dir littering their home directory with yet another `.dir`. In this drop, we're introducing a new env variable `K9SCONFIG` that tells K9s where to look for its configurations. If `K9SCONFIG` is not set K9s will look in the usual place aka `$HOME/.k9s`.

## Resolved Bugs/Features/PRs

- [Issue #754](https://github.com/derailed/k9s/issues/754)
- [Issue #753](https://github.com/derailed/k9s/issues/753)
- [Issue #743](https://github.com/derailed/k9s/issues/743)
- [Issue #728](https://github.com/derailed/k9s/issues/728)
- [Issue #718](https://github.com/derailed/k9s/issues/718)
- [Issue #643](https://github.com/derailed/k9s/issues/643)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
