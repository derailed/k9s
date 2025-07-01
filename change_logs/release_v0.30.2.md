<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s-xmas.png" align="center" width="800" height="auto"/>

# Release v0.30.2

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

## 🎄 Maintenance Release! 🎄

🎵 `On The eleventh day of Christmas my true love gave to me... More Bugs!!` 🎵

Thank you all for pitching in and help flesh out bugs!!

---

## [!!FEATURE NAME CHANGED!!] Vulnerability Scan Exclusions...

As it seems customary with all k9s new features, folks want to turn them off ;(
The `Vulscan` feature did not get out unscathed ;(
As it was rightfully so pointed out, you may want to opted out scans for images that you do not control.
Tho I think it might be a good idea to run wide open once in a while to see if your cluster has any holes??
For this reason, we've opted to intro an exclusion section under the image scan configuration to exclude certain images from the scans.

Here is a sample configuration:

```yaml
k9s:
  liveViewAutoRefresh: false
  refreshRate: 2
  ui:
    enableMouse: false
    headless: false
    logoless: false
    crumbsless: false
    noIcons: false
  imageScans:
    enable: true
    # MOTE!! Field Name changed!!
    exclusions:
      # Skip scans on these namespaces
      namespaces:
        - ns-1
        - ns-2
      # Skip scans for pods matching these labels
      labels:
        - app:
          - fred
          - blee
          - duh
        - env:
          - dev
```

---

## Videos Are In The Can!

Please dial [K9s Channel](https://www.youtube.com/channel/UC897uwPygni4QIjkPCpgjmw) for up coming content...

* [K9s v0.30.0 Sneak peek](https://youtu.be/mVBc1XneRJ4)
* [Vulnerability Scans](https://youtu.be/ULkl0MsaidU)

---

## Resolved Issues

* [#2374](https://github.com/derailed/k9s/issues/2374) The headless parameter does not function properly (v0.30.1)
* [#2372](https://github.com/derailed/k9s/issues/2372) Unable to set default resource to load (v0.30.1)
* [#2371](https://github.com/derailed/k9s/issues/2371) --write cli option does not work (0.30.X)
* [#2370](https://github.com/derailed/k9s/issues/2370) Wrong list of pods on node (0.30.X)
* [#2362](https://github.com/derailed/k9s/issues/2362) blackList: Use inclusive language alternatives

---

## Contributed PRs

Please be sure to give `Big Thanks!` and `ATTA Girls/Boys!` to all the fine contributors for making K9s better for all of us!!

* [#2375](https://github.com/derailed/k9s/pull/2375) get node filtering params from matching context values
* [#2373](https://github.com/derailed/k9s/pull/2373) fix command line flags not working

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2023 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
