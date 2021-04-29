<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.11.0

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_helm.png" align="center" width="300" height="auto"/>

## Change Logs

### Anyone At The Helm?

K9s now offers preliminary support for Helm 3 charts! It's been a long time coming and I know a few early users had mentioned the need, but I wanted to see where Helm3 was going first. This is a preview release to see how we fair in Helm land. Besides managing your installed charts, you will be able to perform the following operations:

* Uninstall a chart
* View chart release notes
* View deployed manifests

#### How to use?

Simply enter `:charts` K9s alias command to view the deployed Helm3 charts on your cluster.

If you're using Helm3 in your current clusters, please give it a rip and also pipe in for potential features/enhancements. Mind the gap here as the paint on this feature is totally fresh...

### Bring Out Your Deads...

There are also a few bugs fixes from the refactor aftermath that made this drop. I know this was a bit of a brutal transition, so thank you all for your patience and for filing issues! I am hopeful that K9s will stabilize quickly so we can move on to bigger things.

---

## Resolved Bugs/Features

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
