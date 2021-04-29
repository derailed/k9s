<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.21.0

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

If you feel K9s is helping your Kubernetes journey, please consider joining our [sponsorship program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

## First A Word From Our Sponsors...

First off, I would like to send a `Big Thank You` to the following generous K9s friends for joining our sponsorship program and supporting this project!

* [Remo Eichenberger](https://github.com/remoe)
* [Ken Ahrens](https://github.com/kenahrens)

## Moving Forward!

In this drop, we've added a port-forward indicator to visually see if a port-forward is active on a pod/container. You can also navigate directly to the port-forward view using the new shortcut `f` available in
pod and container view.

## Manifest That!

Ever wanted to manipulate your Kubernetes manifests directly in K9s? `Yes Please!!`

We are introducing a new view namely `directory` aka `dir`. Using this command you can list/traverse a given directory structure containing your Kubernetes manifests using a new `:dir /fred` command.
From there you can view/edit your manifests and also deploy or delete these resources for your cluster directly from K9s. Just like `kubectl` you can apply/delete an entire directory or a single manifest.
How cool is that?

## Resolved Bugs/Features/PRs

* [Issue #778](https://github.com/derailed/k9s/issues/778)
* [Issue #774](https://github.com/derailed/k9s/issues/774)
* [Issue #761](https://github.com/derailed/k9s/issues/761)
* [Issue #759](https://github.com/derailed/k9s/issues/759)
* [Issue #758](https://github.com/derailed/k9s/issues/758)
* [PR #746](https://github.com/derailed/k9s/pull/746) Big Thanks to [Groselt](https://github.com/groselt)!

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
