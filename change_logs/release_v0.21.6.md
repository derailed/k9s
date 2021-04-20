<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.21.6

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better are as ever very much noted and appreciated!

If you feel K9s is helping your Kubernetes journey, please consider joining our [sponsorship program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

## New Skins On The Block. Part Duh!

In this drop, we've added a new skin configuration for table's cursor namely `cursorFgColor` and `cursorBgColor` as well as the ability to skin your dialogs:

```yaml
  # skin.yml
k9s:
  ...
  # Note: You can now skin your dialogs.
  dialog:
    fgColor: *foreground
    bgColor: *background
    buttonFgColor: *foreground
    buttonBgColor: *magenta
    buttonFocusFgColor: white
    buttonFocusBgColor: *cyan
    labelFgColor: *orange
    fieldFgColor: *foreground
  ...
  views:
    table:
      fgColor: *foreground
      bgColor: *background
      # Note! new tags
      cursorFgColor: *foreground
      cursorBgColor: *current_line
      header:
        fgColor: white
        bgColor: *background
        sorterColor: *cyan
  ...
```

## Resolved Bugs/Features/PRs

* [Issue #795](https://github.com/derailed/k9s/issues/795)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
