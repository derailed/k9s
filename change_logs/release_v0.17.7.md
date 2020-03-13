<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.17.7

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, please consider sponsoring 👆us or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

## 🙀(PLUGIN-19)

[BR]eaking bad on K9s plugins! In previous releases, we used the COL<INDEX> semantic to reference view column data in the plugin extensions. In this drop, we've axed this in favor of column name vs column index. This makes K9s plugin more readable and usable. Also, in light of custom columns, this old semantic just did not jive to well. To boot, all columns available on the viewed resource, regardless of display preferences or order are now free game to plugin authors. So for folks currently leveraging K9s plugins, this drop will break you I am hopeful you guys dig this approach betta??

Here is a sample plugin file that highlights the new functionality. Please see the updated docs for additional information!

```yaml
plugin:
  toggleCronJob:
    shortCut: Ctrl-T
    scopes:
      - cj
    description: Suspend/Resume
    command: kubectl
    background: true
    args:
      - patch
      - cronjobs
      - $NAME
      - -n
      - $NAMESPACE
      - --context
      - $CONTEXT
      - -p
      - '{"spec" : {"suspend" : $!COL-SUSPEND }}' # => Used to be COL3!
```

## Resolved Bugs/Features/PRs

- [Issue #616](https://github.com/derailed/k9s/issues/616)
- [Issue #615](https://github.com/derailed/k9s/issues/615)
- [Issue #614](https://github.com/derailed/k9s/issues/614)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
