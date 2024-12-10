<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.19.5

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, consider joining our [sponsorship program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

## A Word From Out Sponsors...

First off, I would like to send a `Big Thank You` to the following generous K9s friends for joining our sponsorship program and supporting this project!

* [Tommy Dejbjerg Pedersen](https://github.com/tpedersen123)
* [Matt Welke](https://github.com/mattwelke)

## Disruption In The Force

During this drop, I've gotten totally slammed by other forces ;( I've had so many disruptions that affected my `quasi` normal flow hence this drop might be a bit wonky ;( So please proceed with caution!!

As always please help me flush/report issues and I'll address them promptly! Thank you so much for your understanding and patience!! 🙏👨‍❤️‍👨😍

## Improved Node Shell Usability

In this drop we've changed the configuration of the node shell action that lets you shell into nodes. Big thanks to [Patrick Decat](https://github.com/pdecat) for helping us flesh out this beta feature! We've added configuration to not only customize the image but also the resources and namespace on how to run these K9s pods on your clusters. The new configuration is set at the cluster scope level.

Here is an example of the new pod shell config options:

```yaml
# $HOME/.k9s/config.yml
k9s:
  clusters:
    blee:
      featureGates:
        # You must enable the nodeShell feature gate to enable shelling into nodes
        nodeShell: true
      # NEW! You can now tune the pod specification: currently image, namespace and resources
      shellPod:
        image: cool_kid_admin:42
        namespace: blee
        limits:
          cpu: 100m
          memory: 100Mi
```

## Resolved Bugs/Features/PRs

* [Issue #714](https://github.com/derailed/k9s/issues/714)
* [Issue #713](https://github.com/derailed/k9s/issues/713)
* [Issue #708](https://github.com/derailed/k9s/issues/708)
* [Issue #707](https://github.com/derailed/k9s/issues/707)
* [Issue #705](https://github.com/derailed/k9s/issues/705)
* [Issue #704](https://github.com/derailed/k9s/issues/704)
* [Issue #702](https://github.com/derailed/k9s/issues/702)
* [Issue #700](https://github.com/derailed/k9s/issues/700) Fingers and toes crossed ;)
* [Issue #694](https://github.com/derailed/k9s/issues/694)
* [Issue #663](https://github.com/derailed/k9s/issues/663) Partially - should be better launching in a given namespace ie k9s -n fred??
* [Issue #702](https://github.com/derailed/k9s/issues/702)
* [PR #709](https://github.com/derailed/k9s/pull/709) All credits goes to [Namco](https://github.com/namco1992)!!
* [PR #706](https://github.com/derailed/k9s/pull/706) Big Thanks to [M. Tarık Yurt](https://github.com/mtyurt)!
* [PR #704](https://github.com/derailed/k9s/pull/704) Atta Boy!! [psvo](https://github.com/psvo)
* [PR #696](https://github.com/derailed/k9s/pull/696) Thank you! Credits to [Christian Köhn](https://github.com/ckoehn)
* [PR #691](https://github.com/derailed/k9s/pull/691) Mega Thanks To [Pavel Tumik](https://github.com/sagor999)!

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
