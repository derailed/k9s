<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.23.0

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better are as ever very much noted and appreciated!

If you feel K9s is helping your Kubernetes journey, please consider joining our [sponsorhip program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

## ♫ Sound Behind The Release ♭

I figured, why not share one of the tunes I was spinning when powering thru teh bugs? Might as well share the pain/pleasure while viewing this release notes!

[On An Island - David Gilmour With Crosby&Nash](https://www.youtube.com/watch?v=kEa__0wtIRo)

## A Word From Our Sponsors...

First off, I would like to send a `Big Thank You` to the following generous K9s friends for joining our sponsorship program and supporting this project!

* [Martin Kemp](https://github.com/MartiUK)

Contrarily to popular belief, OSS is not free! We've now reached ~9k stars and 300k downloads! As you all know, this project is not pimped out by a big company with deep pockets and a large team. K9s is complex and does demand a lot of my time. So if this tool is useful to you and part of your daily lifecycle, please contribute! Your contribution whether financial, PRs, issues or shout-outs on social/blogs are crucial to keep K9s growing and powerful for all of us. Don't let OSS by individual contributors become an oxymoron!
## Best Effort... Not!

In this drop, we've added 2 new columns to the Pod/Container views namely `CPU(R:L)` and `MEM(R:L)`. These represents the current request/limit resources specified at either the pod or container level. While in Pod view, you will need to use the wide command `Ctrl-W` to see the resources set at the pod level or you can use K9s column customization feature to volunteer them by default.

## Container Images

You have now the ability to tweak your container images for experimentation, using the new SetImage binding aka `i`. This feature is available for standalone pods, deployments, sts and ds. With a resource selected, pressing `i` will provision an edit dialog listing all init/container images.

NOTE! This is a one shot commands applied directly against your cluster and won't survive a new resource deployment.

Big `ATTA Boy!` in effect to [Antoine Méausoone](https://github.com/Ameausoone) for putting forth the effort to make this feature available to all of us!!

## Resolved Issues/Features

* [Issue #906](https://github.com/derailed/k9s/issues/906) Print resources in pod view
* [Issue #900](https://github.com/derailed/k9s/issues/900) Support sort by pending status
* [Issue #895](https://github.com/derailed/k9s/issues/895) Wrong highlight position when filtering logs
* [Issue #892](https://github.com/derailed/k9s/issues/892) tacit kustomize & kpt support
* [Issue #889](https://github.com/derailed/k9s/issues/889) Disable read only config via command line flag
*
* [Issue #886](https://github.com/derailed/k9s/issues/886) Full screen mode or remove borders in YAML view for easy copy/paste
* [Issue #887](https://github.com/derailed/k9s/issues/887) Ability to call out a separate program to parse/filter logs
* [Issue #884](https://github.com/derailed/k9s/issues/884) Refresh for describe & yaml view
* [Issue #883](https://github.com/derailed/k9s/issues/883) View logs quickly scrolls through entire logs when initially loading
* [Issue #875](https://github.com/derailed/k9s/issues/875) Lazy filter
* [Issue #848](https://github.com/derailed/k9s/issues/848) Support an inverse operator on filtered search

## Resolved PRs

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
