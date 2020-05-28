<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.20.0

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, consider joining our [sponsorhip program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

## ♫ The Sound Behind The Release ♭

And now for something a `beat` different?

I figured, why not share one of the tunes I was spinning when powering thru teh bugs? Might as well share the pain/pleasure roaght?

I've just discovered this Turkish band, that I dig and figured I'll share it with you while you read these release notes...

[Ruh - She Past Away](https://www.youtube.com/watch?v=B7f-opGKOyI)

NOTE! Mind you I grew up on the `The Cure`, so likely not for everyone here 🙀

## PortForward Revisited

While performing port-forwards, it could be convenient to specify a given IP address vs 'localhost'
for the forwarding host. For this reason, we are introducing a configuration setting that allows you to set the host IP address for the port-forward dialog on a per cluster basis. The IP address currently defaults to `localhost`.

Big Thanks and all credits goes to [Stowe4077](https://github.com/Stowe4077) (and that very cute dog!) for raising this issue in the first place!!

In order to change the configuration, edit your k9s config file as follows:

```yaml
k9s:
  ...
  clusters:
    blee:
      namespace:
        active: ""
        favorites:
        - fred
        - default
      view:
        active: po
      portForwardAddress: 1.2.3.4
```

## And We've Got A Floater!

I've been noodling on this feature for a while and thought it might be time to `float` this over to you guys... While operating on a cluster you may ask yourself: "Hum... wonder which resources use configmap `fred`?" Sure a quick grep through your manifests on disk will do fine, but what about the resources actually deployed on your cluster? Well my friends wonder no m'o, K9s knows!
While navigating to your ConfigMap View a new option will appear `UsedBy` pressing `u` will reveal any resources that are currently referencing that ConfigMap. As of this drop, this feature will be available for the usual suspects namely: ConfigMaps, Secrets and ServiceAccounts. K9s scans managing resources and locate references from Env vars, Volumes or ServiceAccounts.

NOTE: This feature is expensive to produce and might take a while to fully resolve on larger clusters! Also K9s referential scans might not be full proof and the paint is still fresh on this one so trade carefully! More resources refs checks will be enabled once we've rinse and repeat on this deal. We hope you'll find this feature useful, if so, please make some noise!

## Lastly...

There has been quick a bit of surgery going on with this drop, so this release could be a bit unstable. Please watch out for that carp overbite! As always, Thank You All for your understanding, support and patience!!

## Resolved Bugs/Features/PRs

- [Issue #734](https://github.com/derailed/k9s/issues/734)
- [Issue #733](https://github.com/derailed/k9s/issues/733)
- [Issue #716](https://github.com/derailed/k9s/issues/716)
- [Issue #693](https://github.com/derailed/k9s/issues/693)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
