<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.13.6

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

### GH Sponsorships

WOOT!! Big Thank you in this release to [shiv3](https://github.com/shiv3) for your contributions and support for K9s!
Duly noted and so much appreciated!!

---

### Bow Or Stern?

Some of you had voiced wanting to enable the multi pod logger [Stern](https://github.com/wercker/stern) from the good folks at [Wercker](https://github.com/wercker). Well now you can!

To make this work the awesome [Tuomo Syvänperä](https://github.com/syvanpera) contributed a PR to enable to plug this in with K9s. Thank you Tuomo!!
By default the filter will be set to the currently selected pod. If you need to change the filter, simply filter the pod view to using your own regex and that's the filter K9s will use. Here is a sample plugin that defines a new K9s shortcut to launch Stern provided of course it is installed on your box...

```yaml
# K9s plugin.yml
plugin:
  stern:
    shortCut: Ctrl-L
    description: "Logs (Stern)"
    scopes:
      - pods
    command: /usr/local/bin/stern # NOTE! Look for the command at this location.
    background: false
    args:
    - --tail
    - 50
    - $FILTER # NOTE! Pulls the filter out of the pod view.
    - -n
    - $NAMESPACE
    - --context
    - $CONTEXT
```

---

## Resolved Bugs/Features/PRs

* [Issue #507](https://github.com/derailed/k9s/issues/507)
* [PR #510](https://github.com/derailed/k9s/pull/510) Thank you!! [Vimal Kumar](https://github.com/vimalk78)
* [PR #340](https://github.com/derailed/k9s/pull/340) ATTA Boy! [Tuomo Syvänperä](https://github.com/syvanpera)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
