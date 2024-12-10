<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.13.0

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

### GitHub Sponsors

I'd like to personally thank the following folks for their support and efforts with this project as I know some of you have been around since it's inception almost a year ago!

* [Norbert Csibra](https://github.com/ncsibra)
* [Andrew Roth](https://github.com/RothAndrew)
* [James Smith](https://github.com/sedders123)
* [Daniel Koopmans](https://github.com/fsdaniel)

Big thanks in full effect to you all, I am so humbled and honored by your kind actions!

### Dracula Skin

Since we're in the thank you phase, might as well lasso in [Josh Symonds](https://github.com/Veraticus) for contributing the `Dracula` K9s skin that is now available in this repo under the skins directory. Here is a sneak peek of what K9s looks like under that skin. I am hopeful that like minded `graphically` inclined K9sers will contribute cool skins for this project for us to share/use in our Kubernetes clusters.

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/skins/dracula.png"/>

### XRay Vision!

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_xray.png"/>

Since we've launched K9s, we've longed for a view that would display the relationships among resources. For instance, pods may reference configmaps/secrets directly via volumes or indirectly with containers referencing configmaps/secrets via say env vars. Having the ability to know which pods/deployments use a given configmap may involve some serious `kubectl` wizardry. K9s now has xray vision which allows one to view and traverse these relationships/associations as well as check for referential integrity.

For this, we are introducing a new command aka `xray`. Xray initially supports the following resources (more to come later...)

1. Deployments
2. Services
3. StatefulSets
4. DaemonSets

To enable cluster xray vision for deployments simply type `:xray deploy`. You can also enter the resource aliases/shortnames or use the alias `x` for `xray`. Some of the commands available in table view mode are available here ie describe, view, shell, logs, delete, etc...

Xray not only will tell you when a resource is considered `TOAST` ie the resource is in a bad state, but also will tell you if a dependency is actually broken via `TOAST_REF` status. For example a pod referencing a configmap that has been deleted from the cluster.

Xray view also supports for filtering the resources by leveraging regex, labels or fuzzy filters. This affords for getting more of an application `cross-cut` among several resources.

As it stands Xray will check for following resource dependencies:

* pods
* containers
* configmaps
* secrets
* serviceaccounts
* persistentvolumeclaims

Keep in mind these can be expensive traversals and the view is eventually consistent as dependent resources will be lazy loaded.

We hope you'll find this feature useful? Keep in mind this is an initial drop and more will be coming in this area in subsequent releases. As always, your comments/suggestions are encouraged and welcomed.

### Breaking Change Header Toggle

It turns out the 'h' to toggle header was a bad move as it is use by the view navigation. So we changed that shortcut to `Ctrl-h` to toggle the header expansion/collapse.

---

## Resolved Bugs/Features

* [Issue #494](https://github.com/derailed/k9s/issues/494)
* [Issue #490](https://github.com/derailed/k9s/issues/490)
* [Issue #488](https://github.com/derailed/k9s/issues/488)
* [Issue #486](https://github.com/derailed/k9s/issues/486)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
