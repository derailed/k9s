<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.17.0

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/beach.png" align="center"/>

## Custom Columns? Yes Please!!

[SneakCast v0.17.0 on The Beach! - Yup! sound is sucking but what a setting!](https://youtu.be/7S33CNLAofk)

In this drop I've reworked the rendering engine to provide for custom columns support. Now, you should be able to not only tell K9s which columns you would like to display but also which order they should be in.

To surface this feature, you will need to create a new configuration file, namely `$HOME/.k9s/views.yml`. This file leverages GVR (Group/Version/Resource) to configure the associated table view columns. If no GVR is found for a view the default rendering will take over (ie what we have now). Going wide will add all the remaining columns that are available on the given resource after your custom columns. To boot, you can edit your views config file and tune your resources views live!

> NOTE: This is experimental and will most likely change as we iron this out!

Here is a sample views configuration that customize a pods and services views.

```yaml
# $HOME/.k9s/views.yml
k9s:
  views:
    v1/pods:
      columns:
        - AGE
        - NAMESPACE
        - NAME
        - IP
        - NODE
        - STATUS
        - READY
    v1/services:
      columns:
        - AGE
        - NAMESPACE
        - NAME
        - TYPE
        - CLUSTER-IP
```

## Resolved Bugs/Features/PRs

- [Issue #581](https://github.com/derailed/k9s/issues/581)
- [Issue #576](https://github.com/derailed/k9s/issues/576)
- [Issue #574](https://github.com/derailed/k9s/issues/574)
- [Issue #573](https://github.com/derailed/k9s/issues/573)
- [Issue #571](https://github.com/derailed/k9s/issues/571)
- [Issue #566](https://github.com/derailed/k9s/issues/566)
- [Issue #563](https://github.com/derailed/k9s/issues/563)
- [Issue #562](https://github.com/derailed/k9s/issues/562)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
