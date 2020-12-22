<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.23.0

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better are as ever very much noted and appreciated!

If you feel K9s is helping your Kubernetes journey, please consider joining our [sponsorhip program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

## ♫ Sounds Behind The Release ♭

I figured why not share one of the tunes I was spinning when powering thru teh bugs? Might as well share the pain/pleasure while viewing these release notes!

* [On An Island - David Gilmour With Crosby&Nash](https://www.youtube.com/watch?v=kEa__0wtIRo)
* [Cause We've Ended As Lovers - Jeff Beck](https://www.youtube.com/watch?v=VC02wGj5gPw)

## Our Release Heroes

Please join me in recognizing and applauding this drop contributors that went the extra mile to make sure K9s is better and more useful for all of us!!

Big ATTA BOY/GIRL! in full effect this week to the good folks below for their efforts and contributions to K9s!!

* [Michael Albers](https://github.com/michaeljohnalbers)
* [Wi1dcard](https://github.com/wi1dcard)
* [Saskia Keil](https://github.com/SaskiaKeil)
* [Tomasz Lipinski](https://github.com/tlipinski)
* [Antoine Méausoone](https://github.com/Ameausoone)
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

Contrarily to popular belief, OSS is not free! We've now reached ~9k stars and 300k downloads! As you all know, this project is not pimped out by a big company with deep pockets and a big dev team. K9s is complex and does demand lots of my time. So if this tool is useful to you and your organization and part of your daily Kubernetes flow, please contribute! Your contribution whether financial, PRs, issues or shout-outs on social/blogs are crucial to keep K9s growing and powerful for all of us. Don't let OSS by individual contributors become an oxymoron!

## Describe/YAML goes FullMonty!!

We've added a new option to enable full screen while describing or viewing a resource YAML. Similarly to the full screen toggle option in the log view, pressing `f` will now toggle full-screen for both YAML and Describe views.

Additionally, the YAML and Describe view are now reactive! YAML/Describe views will now watch for changes to the underlying resource manifests. I'll admit this was a feature I was missing, but decided to punt as it required a bit of re-org to make it happen correctly. So BIG thanks to [Fabian-K](https://github.com/Fabian-K) for entering this issue and for the boost!!

Not cool enough for Ya? the YAML view now also affords for getting ride of those pesky `managedFields` while viewing a resource. Use the `m` key to toggle visibility on the managedFields.

## Best Effort... Not!

In this drop, we've added 2 new columns namely `CPU/R:L` and `MEM/R:L`. These represents the current request:limit specified on containers. They are available in node, pod and container views. While in Pod view, you will need to volunteer them and use the `Go Wide` option `Ctrl-W` to see the columns. These columns will be display by default for Node/Container views. In the node view, they tally the total amount of resources for all pods hosted a given node. If that's inadequate, you can also leverage K9s [Custom Column](https://github.com/derailed/k9s#resource-custom-columns) feature to volunteer them or not.

## Set Container Images

You will have the ability to tweak your container images for experimentation, using the new SetImage binding aka `i`. This feature is available for un-managed pods, deployments, statefulsets and daemonsets. With a resource selected, pressing `i` will provision an edit dialog listing all init/container images. So you will have to ability to tweak the images and update your containers. Big Thanks to [Antoine Méausoone](https://github.com/Ameausoone) for making this feature available to all of us!!

NOTE! This is a one shot commands applied directly against your cluster and won't survive a new resource deployment.

## Crumbs On...Crumbs Off, Caterpillar

We've added a new configuCCration to turn off the crumbs via `crumbsLess` configuration option. You can also toggle the crumbs via the new key option `Ctrl-g`. You can enable/disable this option in your ~/.k9s/config.yml or via command line using `--crumbsless` CLI option.

```yaml
k9s:
  refreshRate: 2
  headless: false
  crumbsless: false
  readOnly: true
  ...
```

## BANG FILTERS!

Some folks have voiced the desire to use inverse filters to refine content while in resource table views. Appending a `!` to your filter will now enable an inverse filtering operation For example, in order to see all pods that do not contain `fred` in their name, you can now use `/!fred` as your filtering command. If you dig this implementation, please make sure to give a big thank you to [Michael Albers](https://github.com/michaeljohnalbers) for the swift implementation!

## New Conf On the Block...

In this release, we've made some changes to the retry policies when things fail on your cluster and the api-server is suffering from an hearing impediment. The current policy was to check for connection issues every 15secs and retry 15 times before exiting K9s. This rules were not configurable and could yield for overtaxing the api-server. So we've implemented exponential back-off so that K9s can attempt to remediate or bail out of the session if not.
To this end, there is a new config option namely `maxConnRetry` to will be added to your K9s config to set the retry policy. The default is currently set to 5 retries.

NOTE: This is likely an ongoing story and more will come based on your feedback!

Sample K9s configuration

```yaml
k9s:
  refreshRate: 2
  # Set the maximum attempt to reconnect with the api-server in case of failures.
  maxConnRetry: 5
  ...
```

## 🏁 Start Your Engines...

As you can see, this is a pretty big drop and likely we've created some new issues in the process 🙀

Please make sure to file issues/PRs if things are not working as expected so we can improve on these features.

👻 Happy Halloween To All!! (as if 2020 is not scary enough 🙈)

Thank you all for your being fans and supporting K9s!!

---

## Resolved Issues/Features

* [Issue #906](https://github.com/derailed/k9s/issues/906) Print resources in pod view
* [Issue #903](https://github.com/derailed/k9s/issues/903) Slow down reconnection rate on auth failures
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
