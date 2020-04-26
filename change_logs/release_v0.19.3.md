<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.19.3

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, consider joining our [sponsorhip program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

## Look Who Is Back?

Thanks to the good Helm folks, we're now back on par with the Helm charts support feature. As you may recall, when we've updated to K8s v1.18, the Helm feature took one for the team ;( as they had yet to upgrade to the latest k8s rev. So K9s Helm chart feature is back in this drop! On that note, we've added new aliases to allow you to view your currently installed Helm charts aka `helm` | `hm` | `chart` | `charts`.

## Boh-Bye Windows 386!

As of this drop, I've decided to axe Windows 386 support. Our good friend [Guy Barrette](https://github.com/guybarrette) reported K9s Windows-386 binary is triping his virus scanner. After double checking my installed shas/binaries/dependencies/etc... and performing vulnerabily scans on various win-i386 K9s binaries, I just could not figure out which dependencies are causing the exec to bomb on the scans??

Note: This does not necessary entails that there is a deliberate or malicious intent with the software, but likely a false positive thrown by the Windows virus scanner. This has been [reported](https://golang.org/doc/faq#virus) with other GO binaries on windows as well ;(

That said, I've repeatdly scanned the K9s Windows-x64 and ended up with a clean bill of health on every single scans. So I've decided to drop the 386 windows support for the time being. If that causes you some grief, please land a hand as I am fresh out of ideas...

## And Now For Something A Bit More... Controversial?

There has been a lot of requests for K9s to support shelling directly into cluster nodes. I was resisting the temptation to support this useful feature as depending on your cluster hosting solution, this involved less than ideal solutions. My clusters are provisioned in a multitude of platforms ranging from bare metal to cloud vendor self/managed hosting. I wanted the same experience shelling into an GKE/AWS node as a local KiND cluster node. To this end, we've opted to experimentally support shelling into nodes using the following approach:

1. While in the Node view, we are introducing a new `s` mnemonic to shell into nodes on your cluster.
2. K9s will spin up a `k9s-shell` pod in the `default` namespace with an official Busybox container running in `privileged` mode. This may require extra RBAC and PSPs (This will need Docs!)
3. Once shelled-in, you can poke around any of your nodes.
4. Upon exiting the node shell, K9s will automatically delete the `k9s-shell` pod for that node.

This feature is `OPT-IN` only ie you will need to manually enable the feature gate to make this functionality available to K9s on a specific cluster as follows:

```yaml
# $HOME/.k9s/config.yml
k9s:
  ...
  clusters:
    fred:
      namespace:
        active: "default"
        favorites:
        - default
      view:
        active: po
      featureGates:
        nodeShell: true # Defaults to false!
```

Please let us know if you dig this feature? This very much experimental and we're open to your suggestions. Thank you!

## New Sheriff In Town K9S_EDITOR

As you may know K9s currently uses your `EDITOR` env var to launch an editor while editing a k8s resource or viewing a screen dump or a performance benchmark. So folks voiced they are using some editors that require different CLI args when editing k8s resources vs files on disk. In this drop, we're introducing a new env var `K9S_EDITOR` to provide an affordance to deal with these discrepancies. If you are using emacs/vi/nano no action should be required. K9s will now check for `K9S_EDITOR` existence to view K9s artifacts such as screen_dumps. K9s still honors `KUBE_EDITOR` or `EDITOR` for K8s resource edits. K9s will fallback to the `EDITOR` env var if `K9S_EDITOR` is not set.

## Resolved Bugs/Features/PRs

* [Issue #669](https://github.com/derailed/k9s/issues/669)
* [Issue #677](https://github.com/derailed/k9s/issues/677)
* [Issue #673](https://github.com/derailed/k9s/issues/673)
* [Issue #671](https://github.com/derailed/k9s/issues/671)
* [Issue #670](https://github.com/derailed/k9s/issues/670)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
