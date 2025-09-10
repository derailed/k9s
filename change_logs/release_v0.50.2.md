<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s.png" align="center" width="800" height="auto"/>

# Release v0.50.2

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s!
I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev
and see if we're happier with some of the fixes!
If you've filed an issue please help me verify and close.

Your support, kindness and awesome suggestions to make K9s better are, as ever, very much noted and appreciated!
Also big thanks to all that have allocated their own time to help others on both slack and on this repo!!

As you may know, K9s is not pimped out by corps with deep pockets, thus if you feel K9s is helping your Kubernetes journey,
please consider joining our [sponsorship program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

## 5-0, 5-0 HotFix!

It looks like we've broken a few (more) things in the clean up process 😳
This is what you get for trying to refresh a ~10 year old code base 🙀
Apologizes for the `disruption in the farce`. Hopefully much happier on v0.50.2...
Are we there yet? Crossing fingers AND toes...

☠️ Careful on this upgrade! 🏴‍☠️
We've gone thru lots of code revamp/refactor in the v0.50.0, so mileage may vary...

---

## Resolved Issues

* [#3267](https://github.com/derailed/k9s/issues/3267) Show some output or message when no resources are found
* [#3266](https://github.com/derailed/k9s/issues/3266) Command alias :dp fails with "no resource meta defined for deployments" error
* [#3264](https://github.com/derailed/k9s/issues/3264) can't execute get(y) or describe(d) in StorageClass view
* [#3260](https://github.com/derailed/k9s/issues/3260) yaml view of pod will crash the app (Boom!! cannot deep copy int. (Maybe??)

---
<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2025 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)