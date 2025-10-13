<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s.png" align="center" width="800" height="auto"/>

# Release v0.40.6

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

## Maintenance Release!

### Breaking change

Moved `portForwardAddress` out of clusterXXX/contextYYY/config.yaml and into the main K9s config file.
This is a global preference based on your setup vs a cluster/context specific attribute.
K9s will nag you in the logs if a specific context config still contains this attribute but should not prevent the configuration load.

### Column Blow Reloaded!

We've added another property to the custom view. You can now also specify namespace specific column definition for a given resource.
For instance, view pods in any namespace using one configuration and view pods in `fred` namespace using an alternate configuration.

```yaml
# views.yaml
views:
  # Using this for all pods...
  v1/pods:
    columns:
      - AGE
      - NAMESPACE|WR                                     # => 🌚 Specifies the NAMESPACE column to be right aligned and only visible while in wide mode
      - ZORG:.metadata.labels.fred\.io\.kubernetes\.blee # => 🌚 extract fred.io.kubernetes.blee label into it's own column
      - BLEE:.metadata.annotations.blee|R                # => 🌚 extract annotation blee into it's own column and right align it
      - NAME
      - IP
      - NODE
      - STATUS
      - READY
      - MEM/RL|S                                         # => 🌚 Overrides std resource default wide attribute via `S` for `Show`
      - '%MEM/R|'                                        # => NOTE! column names with non alpha names need to be quoted as columns must be strings!

   # Use this instead for pods in namespace `fred`
   v1/pods@fred:                                         # => 🌚 New v0.40.6! Customize columns for a given resource and namespace!
    columns:
      - AGE
      - NAMESPACE|WR
```

Additionally, we've added a new column attribute aka `Show` -> `S`. This allows you to now override the default resource column `wide` attribute when set.


---

## Videos Are In The Can!

Please dial [K9s Channel](https://www.youtube.com/channel/UC897uwPygni4QIjkPCpgjmw) for up coming content...

* [K9s v0.40.0 -Column Blow- Sneak peek](https://youtu.be/iy6RDozAM4A)
* [K9s v0.31.0 Configs+Sneak peek](https://youtu.be/X3444KfjguE)
* [K9s v0.30.0 Sneak peek](https://youtu.be/mVBc1XneRJ4)
* [Vulnerability Scans](https://youtu.be/ULkl0MsaidU)

---

## Resolved Issues

* [#3179](https://github.com/derailed/k9s/issues/3179) Resource name with full api or group displayed (somewhere and sometimes)
* [#3178](https://github.com/derailed/k9s/issues/3178) Cronjobs with the same name in different namespaces appear together
* [#3176](https://github.com/derailed/k9s/issues/3176) Trigger all marked cronjobs
* [#3162](https://github.com/derailed/k9s/issues/3162) Context configs: context directory created under wrong cluster after context switch
* [#3161](https://github.com/derailed/k9s/issues/3161) Force wide-only columns to appear outside of wide view
* [#3147](https://github.com/derailed/k9s/issues/3147) Prompt style is overriden by body
* [#3139](https://github.com/derailed/k9s/issues/3139) CPU/R:L and MEM/R:L columns invalid in views.yaml
* [#3138](https://github.com/derailed/k9s/issues/3138) Subresources are not shown correctly in the RBAC view

---

## Contributed PRs

Please be sure to give `Big Thanks!` and `ATTA Girls/Boys!` to all the fine contributors for making K9s better for all of us!!

* [#3182](https://github.com/derailed/k9s/pull/3182) fix: Use the latest version when downloading the Ubuntu deb file
* [#3168](https://github.com/derailed/k9s/pull/3168) fix(history): handle cases where special commands add their command their command to the history
* [#3159](https://github.com/derailed/k9s/pull/3159) Added hard contrast gruvbox skins
* [#3149](https://github.com/derailed/k9s/pull/3149) fix: Pass grv on gotoResource as a String to fix non-default apiGroup list
* [#3149](https://github.com/derailed/k9s/pull/3149) Add externalsecrets plugin
* [#3140](https://github.com/derailed/k9s/pull/3140) fix: Avoid false positive matches in enableRegion (#3093)


<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2025 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)