<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.8.3

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

### NetworkPolicy

NetworkPolicy resource is now available under the command `np` while in command mode. It behaves like other K9s resource views. You will get a bit more information in K9s vs `kubectl` as it includes information about ingress and egress rules.

### Arrrggg! New CLI Argument

There isa new K9s command option available on the CLI that affords for launching K9s with a given resource. For example using `k9s -c svc` willlaunch K9s with a preloaded service view. You can use the same aliases as you would while in K9s to navigate a resources. For all supports resource aliases please view the `Alias View` using `Ctrl-A`.

### CRDS!

We've beefed up CRD support to allow users to navigate to the CRD instances view more readily. So you can now navigate between CRD schema and CRD instances by just hitting `ENTER` while in the `crd` view.

---

## Resolved Bugs/Features

+ CRD Navigation [Issue #295](https://github.com/derailed/k9s/issues/295)
+ Terminal colors [Issue #294](https://github.com/derailed/k9s/issues/294)
+ Help menu typo [Issue #291](https://github.com/derailed/k9s/issues/291)
+ NetworkPolicy Support [Issue #289](https://github.com/derailed/k9s/issues/289)
+ Scaling replicas start count [Issue #288](https://github.com/derailed/k9s/issues/288)
+ CLI command arg support [Issue #283](https://github.com/derailed/k9s/issues/283)
+ YAML screen dump support [Issue #275](https://github.com/derailed/k9s/issues/275)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
