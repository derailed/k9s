<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.23.0

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better are as ever very much noted and appreciated!

If you feel K9s is helping your Kubernetes journey, please consider joining our [sponsorhip program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

## ♫ Sound Behind The Release ♭

I figured why not share one of the tunes I was spinning when powering thru teh bugs? Might as well share the pain/pleasure while viewing this release notes!

[On An Island - David Gilmour With Crosby&Nash](https://www.youtube.com/watch?v=kEa__0wtIRo)

## Our Release Heroes...

Please join me in recognizing and applauding this drop contributors that went the extra mile to make sure K9s is better and more useful for all of us!!

Big ATTA BOY/GIRL! in full effect this week to the good folks below for their efforts and contributions to K9s!!

* [Antoine Méausoone](https://github.com/Ameausoone)
* [Michael Albers](https://github.com/michaeljohnalbers)
* [Wi1dcard](https://github.com/wi1dcard)
* [Saskia Keil](https://github.com/SaskiaKeil)
* [Tomasz Lipinski](https://github.com/tlipinski)
* [Emeric Martineau](https://github.com/emeric-martineau)
* [Eldad Assis](https://github.com/eldada)
* [David Arnold](https://github.com/blaggacao)
* [Peter Parente](https://github.com/parente)

## A Word From Our Sponsors...

First off I would like to send a `Big Thank You` to the following generous K9s friends for joining our sponsorship program and supporting this project!

* [William Alexander](https://github.com/carpetfuz)
* [Jiri Valnoha](https://github.com/waldauf)
* [Pavel Tumik](https://github.com/sagor999)
* [Bart Plasmeijer](https://github.com/bplasmeijer)
* [Matt Welke](https://github.com/mattwelke)
* [Stefan Mikolajczyk](https://github.com/stefanmiko)

Contrarily to popular belief, OSS is not free! We've now reached ~9k stars and 300k downloads! As you all know, this project is not pimped out by a big company with deep pockets. K9s is complex and does demand lots of my time. So if this tool is useful to you and benefits you and your organization in your Kubernetes journey, please contribute! Your contribution whether financial, PRs, issues or shout-outs on social/blogs are crucial to keep K9s growing and powerful for all of us. Don't let OSS by individual contributors become an oxymoron!

## Describe/YAML views goes FullMonty

We've added a new option to enable full screen while describing or viewing a resource YAML. Similarly to the full screen toggle option in the log view, pressing `f` will now toggle fullscreen for both YAML and Describe views.

Additionally, the YAML and Describe view are now reactive! YAML/Describe views will now watch for changes to the underlying resource manifests. How cool is that?

Not cool enough for Ya? the YAML view also affords for getting ride of those pesky `managedFields` while viewing a resource. Pressing `m` will toggle visibility on these fields.

## Best Effort... Not!

In this drop, we've added 2 new columns to the Pod/Container views namely `CPU(R:L)` and `MEM(R:L)`. These represents the current request:limit resources specified at either the pod or container level. While in Pod view, you will need to use the `Go Wide` option `Ctrl-W` to see the resources set at the pod level. You can also leverage K9s [Custom Column](https://github.com/derailed/k9s#resource-custom-columns) feature to volunteer them while in Pod view. In the Container view these columns will be available by default.

## Set Container Images

You have now the ability to tweak your container images for experimentation, using the new SetImage binding aka `i`. This feature is available for unmanaged pods, deployments, sts and ds. With a resource selected, pressing `i` will provision an edit dialog listing all init/container images.

NOTE! This is a one shot commands applied directly against your cluster and won't survive a new resource deployment.

## Crumbs On, Crumbs Off, Caterpillar

We've added a new configuration to turn off the crumbs via `crumbsLess` configuration option. You can also toggle the crumbs via the new key option `C`. You can enable/disable this option in your ~/.k9s/config.yml or via command line using `--crumbsless` CLI option.

```yaml
k9s:
  refreshRate: 2
  headless: false
  crumbsless: false
  readOnly: true
  ...
```

## BANG FILTERS!

Some folks have voiced the desire to use inverse filters to refine content while in resource table views. Prepending a `!` to your filter will now enable an inverse filtering operation For example, in order to see all pods that do not contain `fred` in their name, you can now use `/!fred` as your filtering command.

---

## Resolved Issues/Features

* [Issue #906](https://github.com/derailed/k9s/issues/906) Print resources in pod view
* [Issue #901](https://github.com/derailed/k9s/issues/901) Logs page for any pod/container shows Waiting for logs...
* [Issue #900](https://github.com/derailed/k9s/issues/900) Support sort by pending status
* [Issue #895](https://github.com/derailed/k9s/issues/895) Wrong highlight position when filtering logs
* [Issue #892](https://github.com/derailed/k9s/issues/892) tacit kustomize & kpt support
* [Issue #889](https://github.com/derailed/k9s/issues/889) Disable read only config via command line flag
* [Issue #887](https://github.com/derailed/k9s/issues/887) Ability to call out a separate program to parse/filter logs
* [Issue #886](https://github.com/derailed/k9s/issues/886) Full screen mode or remove borders in YAML view for easy copy/paste
* [Issue #884](https://github.com/derailed/k9s/issues/884) Refresh for describe & yaml view
* [Issue #883](https://github.com/derailed/k9s/issues/883) View logs quickly scrolls through entire logs when initially loading
* [Issue #875](https://github.com/derailed/k9s/issues/875) Lazy filter
* [Issue #848](https://github.com/derailed/k9s/issues/848) Support an inverse operator on filtered search
* [Issue #820](https://github.com/derailed/k9s/issues/820) Log file spammed despite K9s not running
* [Issue #794](https://github.com/derailed/k9s/issues/794) Events view

## Resolved PRs

* [PR #909](https://github.com/derailed/k9s/pull/909) Add support for inverse filtering
* [PR #908](https://github.com/derailed/k9s/pull/908) Remove trailing delta from the scale dialog when replicas are in flux
* [PR #907](https://github.com/derailed/k9s/pull/907) Improve docs on sinceSeconds logger option
* [PR #904](https://github.com/derailed/k9s/pull/904) PVC `UsedBy` list irrelevant statefulsets
* [PR #898](https://github.com/derailed/k9s/pull/898) Use config.CallTimeout in APIClient
* [PR #897](https://github.com/derailed/k9s/pull/897) Use DefaultColorer for aliases rendering
* [PR #896](https://github.com/derailed/k9s/pull/896) Allow remove crumbs
* [PR #894](https://github.com/derailed/k9s/pull/894) Execute plugins and pass context
* [PR #891](https://github.com/derailed/k9s/pull/891) Add command to get the latest stable kubectl version and support for KUBECTL_VERSION as Dockerfile ARG
* [PR #847](https://github.com/derailed/k9s/pull/847) Add ability to set container images

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
