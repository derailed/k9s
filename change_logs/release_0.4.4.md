<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.4.4

## Notes

Thank you to all that contributed with flushing out issues with K9s! I'll try
to mark some of these issues as fixed. But if you don't mind grab the latest
rev and see if we're happier with some of the fixes!

If you've filed an issue please help me verify and close.

Thank you so much for your support and awesome suggestions to make K9s better!!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

### Exiting K9s

  There are a few debates about drathers on K9s key bindings. I have caved in
  and decided to give up my beloved 'q' for quit which will no longer be bound. As of this release quitting K9s must be done via `:q` or `ctrl-c`.

### Container Logs

  [Feature #147](https://github.com/derailed/k9s/issues/147). The default behavior was to pick the first available container. Which meant if the pod has an init container, the log view would choose that.
  The view will now choose the first non init container. Most likely it
  would be the wrong choice in pod's sidecar scenarios, but for the time
  being showing log on one of the init containers just did not make much sense. You can still pick other containers via the menu options. We will implement a better solution for this soon...

### Delete Dialog

  [Feature #146](https://github.com/derailed/k9s/issues/146) Tx @dperique!
  Pressing `<enter>` on the delete dialog would delete the resource. Now
  `cancel` is the default button. Hence you must use `<tab>` or `->` to
  select `OK` then press `<enter>` to delete.

---

## Resolved Bugs

+ None

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
