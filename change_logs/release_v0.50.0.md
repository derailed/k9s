<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s.png" align="center" width="800" height="auto"/>

# Release v0.50

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

---

## ♫ Sounds Behind The Release ♭

* [Afterimage - Justice](https://www.youtube.com/watch?v=9zBJlLbkfzA)
* [This Is The Day - The The](https://www.youtube.com/watch?v=qBF3YqUzYRc)

## 5-O, 5-0... Spring Cleaning In Effect!

☠️ Careful on this upgrade! 🏴‍☠️
We've gone thru lots of code revamp/refactor on this drop, so mileage may vary!!

### K9s Slow?

It looks like K9s performance took a dive in the wrong direction circa v0.40.x releases.
Took a big perf/cleanup pass to improve perf and think this release should help a lot (famous last words...)

> NOTE! As my dear granny use to say: `You can't cook a great meal without trashing the kitchen`,
> So likely I have broken a few things in the process. So thread carefully and report back!

### Now with Super Column Blow!

By general demand, juice up custom views! In a feature we like to refer to as `Super Column Blow...`
As of this drop, you can go full `Chuck Norris` and sprinkle some of your JQ_FU with you custom views.

For example...

```yaml
# views.yaml
views:
  v1/pods:
    sortColumn: NAME:asc
    columns:
    - AGE
    - NAMESPACE
    - NAME
    - IMG-VERSION:.spec.containers[0].image|split(":")|.[-1]|R # => Grab the main container image name and pull the image version
                                                               # => out into the `IMG-VERSION` right aligned column
```

> NOTE: ☢️ This is very much experimental! Not all JQ queries features are supported!
> (See https://github.com/itchyny/gojq for the details!)

## Videos Are In The Can!

Please dial [K9s Channel](https://www.youtube.com/channel/UC897uwPygni4QIjkPCpgjmw) for up coming content...

* [K9s v0.40.0 -Column Blow- Sneak peek](https://youtu.be/iy6RDozAM4A)
* [K9s v0.31.0 Configs+Sneak peek](https://youtu.be/X3444KfjguE)
* [K9s v0.30.0 Sneak peek](https://youtu.be/mVBc1XneRJ4)
* [Vulnerability Scans](https://youtu.be/ULkl0MsaidU)

---

## Resolved Issues

* [#3226](https://github.com/derailed/k9s/issues/3226) Filter view will show mess when filtering some string
* [#3224](https://github.com/derailed/k9s/issues/3224) Respect kubectl.kubernetes.io/default-container annotation
* [#3222](https://github.com/derailed/k9s/issues/3222) Option to Display Resource Names Without API Version Prefix
* [#3210](https://github.com/derailed/k9s/issues/3210) Description line is buggy

---

## Contributed PRs

Please be sure to give `Big Thanks!` and `ATTA Girls/Boys!` to all the fine contributors for making K9s better for all of us!!

* [#3237](https://github.com/derailed/k9s/pull/3237) fix: List CRDs which has k8s.io in their names
* [#3223](https://github.com/derailed/k9s/pull/3223) Fixed skin config ref of in_the_navy to in-the-navy
* [#3110](https://github.com/derailed/k9s/pull/3110) feat: add splashless option to suppress splash screen on start

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2025 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)