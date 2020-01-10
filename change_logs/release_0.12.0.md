<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.12.0

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_helm.png" align="center" width="300" height="auto"/>

This was a long week in the saddle, you guys have been so awesome and supportive thru these last few drops. Thank you!!

### Searchable Logs

There has been quiet a few demands for this feature. It should now be generally available in this drop. It works the same as the resource view ie `/fred`, you can also specify a fuzzy filter using `/-f blee-duh`. The paint is still fresh on that deal and not super confident that it will work nominaly as I had to rework the logs to enable. So totally possible I've hosed something in the process.

### APIServer Dud

At times, it could be you've lost your api server connection while K9s was up which resulted in the `K9s screen of death` or in other words a hosed terminal session ;(. K9s should now detect this condition and close out. Once again no super sure about this implementation on that deal. So if you see K9s close out under normal condition, that means I would need to go back to the drawing board.

### FullScreen Logs

I've been told having a flag to set fullScreen mode preference while viewing the logs would be `awesome`. Thanks [Fardin Khanjani](https://github.com/fardin01)!
So there is now a new K9s config flag available to set your fullsreen logs `drathers` in your .k9s/config.yml. This flag defaults to false if not set.

Here is a snippet:

```yaml
# .k9s/config.yml
k9s:
  refreshRate: 2
  headless: false
  currentContext: crashandburn666
  currentCluster: slowassnot
  fullScreenLogs: true
  ...
```

---

## Resolved Bugs/Features

* [Issue #484](https://github.com/derailed/k9s/issues/484)
* [Issue #481](https://github.com/derailed/k9s/issues/481)
* [Issue #480](https://github.com/derailed/k9s/issues/480)
* [Issue #479](https://github.com/derailed/k9s/issues/479)
* [Issue #477](https://github.com/derailed/k9s/issues/477)
* [Issue #476](https://github.com/derailed/k9s/issues/476)
* [Issue #468](https://github.com/derailed/k9s/issues/468)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
