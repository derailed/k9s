<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s.png" align="center" width="800" height="auto"/>

# Release v0.29.0

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s!
I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev
and see if we're happier with some of the fixes!
If you've filed an issue please help me verify and close.

Your support, kindness and awesome suggestions to make K9s better are, as ever, very much noted and appreciated!
Also big thanks to all that have allocated their own time to help others on both slack and on this repo!!

As you may know, K9s is not pimped out by corps with deep pockets, thus if you feel K9s is helping your Kubernetes journey,
please consider joining our [sponsorship program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

## ♫ Sounds Behind The Release ♭

* [Snowbound - Donald Fagen](https://www.youtube.com/watch?v=bj8ZdBdKsfo)
* [Pilgrim - Eric Clapton](https://www.youtube.com/watch?v=8V9tSQuIzbQ)
* [Lucky Number - Lene Lovich](https://www.youtube.com/watch?v=KnIJOO__jVo)

---

## 🦃 Happy (Belated!) ThanksGiving To All! 🦃

Hope you and yours had a wonderful holiday!!
Hopefully this drop won't be a cold turkey 😳

I'd like to take this opportunity to honor two very special folks:

* [Alexandru Placinta](https://github.com/placintaalexandru)
* [Jayson Wang](https://github.com/wjiec)

These guys have been relentless in fishing out bugs, helping out with support and addressing issues, not to mention enduring my code! 🙀
They dedicate a lot of their time to make `k9s` better for all of us!
So if you happen to run into them live/virtual, please be sure to `Thank` them and give them a huge hug! 🤗

I am thankful for all of you for being kind, patient, understanding and one of the coolest OSS community on the web!!

Feeling blessed and ever so humbled to be part of it.

Thank you!!

---

## A Word From Our Sponsors...

To all the good folks below that opted to `pay it forward` and join our sponsorship program, I salute you!!

* [Marco Stuurman](https://github.com/fe-ax)
* [Paul Sweeney](https://github.com/Kolossi)
* [Cayla Fauver](https://github.com/cayla)
* [alemanek](https://github.com/alemanek)
* [Danske Commodities A/S](https://github.com/DanskeCommodities)

> Sponsorship cancellations since the last release: **8** ;(

---

## 🎉 Feature Release 🎈👯

---

### Breaking Bad!

WARNING! There are breaking change on this drop!

1. NodeShell configuration has moved up in the k9s config file from the context section to the top level config.
More than likely, one uses the same nodeShell image with all the fixins to introspect nodes no matter the cluster. This update DRY's up k9s config and still allows one to opt in/out of nodeShell via the context specific feature gate.
Please see README for the details.

   > NOTE: If you haven't customize the shellPod images on your contexts, the app will move the nodeShell config section to
   > it's new location and update your clusters information accordingly.
   > If not, you will need to edit the nodeShell section and manage it from a single location!

1. Log view used to default to the last 5mins aka `sinceSeconds: 300`.
   Changed the default to tail logs instead aka `sinceSeconds: -1`

1. Skins loading changed! In this release, we do away with the context specific skin files. You can now directly specify the skin to use for a given cluster directly in the k9s config file under the cluster configuration. K9s now expects a skins directory in the k9s config home with your skin files. You can use your custom skins and copy them to the `skins` directory or use the contributes skins found on this repo root.
Specify the name of the skin in the config file and now your cluster will load the specified skin.

For example: create a `skins` dir your k9s config home and add one_dark.yml skin file from this repo. Then edit your k9s config file as follows:

```yaml
k9s:
  ...
  clusters:
    fred:
      # Override the default skin and use this skin for this cluster.
      skin: one_dark # -> Look for a skin file in ~/.config/k9s/skins/one_dark.yml
      namespace:
        ...
      view:
        active: pod
      featureGates:
        nodeShell: false
      portForwardAddress: localhost
```

The `fred` cluster will now load with the specified skin name. Rinse and repeat for other clusters of your liking. In the case where neither the skin dir or skin file are present, k9s will still honor the global skin aka `skin.yml` in your k9s config home directory to skin all your clusters.

---

### Walk Of SHelm...

Added a `Releases` view to Helm!

This provides the ability for Helm users to manage their releases directly from k9s.
You can now press `enter` on a selected Helm install and view all associated releases.
While in the releases view, you can also rollback an install to a previous revision.

---

### Spock! Are You Out Of Your VulScan Mind?

Tired of having malignant folks shoot holes in your prod clusters or failing compliance testing?

Added ability to run image vulnerability scans directly from k9s. You can now monitor your security stance in dev/staging/... clusters
prior to proclaiming `It's Open Season...` in prod!

As it stands Pod, Deployment, StatefulSet, DaemonSet, CronJob, Job views will feature a new column for Vulnerability Scan aka `VS`.

> NOTE! This feature is gated so you'll need to manually opt in/out by modifying your k9s config file like so:

```yaml
k9s:
  liveViewAutoRefresh: false
  enableImageScan: true # <- Yes Please!!
  headless: false
  ...
```

Once enabled, a new column `VS` (aka Vulnerability Score) should be present on the aforementioned views where you will see your vulnerability scores (*Still work in progress!!*).
The `VS` column displays a bit vector aka Sev-1|Sev-2|Sev-3|Sev-4|Sev-5|Sev-Unknown. When the bit is high it indicate the presence of the severity in the scans. Higher order bits = Higher severity
For instance, the following vector `110001` indicates the presence of both critical (Sev-1) and high (Sev-2) and an unclassified severity (aka Sev-Unknown) issues in the scan. Sev-U indicates no classification currently exist in our vulnerability database.

The image scans are run async, rendering the views eventually consistent, hence you may have to give the scores a few cycles for the dust to settle...
Once the caches are primed, subsequent loads should be faster 🤞

You can sort the views by vulnerability score using `ShiftV`.
Additionally, you can view the full scans report by pressing `v` on a selected resource.

I've synced my entire Thanksgiving holiday break on this ding dang deal, so hopefully it works for most of you??
Also if you dig this new feature, please make some noise! 😍

💘 This is an experimental feature and likely will require additional TLC 💘

> NOTE! The lib we use to scan for vulnerabilities only supports macOS and Linux!!
> NOTE: I have yet to test this feature on larger clusters, so likely this may break??
> Please take these reports with a grain of salt as likely your mileage will vary and help us
> validate the accuracy of the report ie if we cry `Wolf`, is it actually there?

The paint is still fresh on this deal!!

### Do You Tube?

My plan is to begin (again!) putting out short k9s episodes with how-tos, tips, tricks and features previews.

Please dial [K9s Channel](https://www.youtube.com/channel/UC897uwPygni4QIjkPCpgjmw) for up coming content...

The first drop should be up by the time you read this!

* [Vulnerability Scans](https://youtu.be/ULkl0MsaidU)

---

## Resolved Issues

* [#2308](https://github.com/derailed/k9s/issues/2308) Unable to list CRs for crd with only list and get verb without watch verb
* [#2301](https://github.com/derailed/k9s/issues/2301) Add imagePullPolicy and imagePullSecrets on shell_pod for internal registry uses
* [#2298](https://github.com/derailed/k9s/issues/2298) Weird color after plugin usage
* [#2297](https://github.com/derailed/k9s/issues/2297) Select nodes with space does not work anymore
* [#2290](https://github.com/derailed/k9s/issues/2290) Provide release assets for freebsd amd64/arm64
* [#2283](https://github.com/derailed/k9s/issues/2283) Adding auto complete in search bar
* [#2219](https://github.com/derailed/k9s/issues/2219) Add tty: true to the node shell pod manifest
* [#2167](https://github.com/derailed/k9s/issues/2167) Show wrong Configmap data
* [#2166](https://github.com/derailed/k9s/issues/2166) Taint count for the nodes view
* [#2165](https://github.com/derailed/k9s/issues/2165) Restart counter for init containers
* [#2162](https://github.com/derailed/k9s/issues/2162) Make edit work when describing a resource
* [#2154](https://github.com/derailed/k9s/issues/2154) Help and h command does not work if typed into cmdbuff
* [#2036](https://github.com/derailed/k9s/issues/2036) Crashed while do filtering
* [#2009](https://github.com/derailed/k9s/issues/2009) Ctrl-s: Name of file (Describe-....)
* [#1513](https://github.com/derailed/k9s/issues/1513) Problem regarding showing the logs - it hangs/slow on pods which are running for long time
  NOTE: Better but not cured! Perf improvements while viewing large cm (7k lines) from 26s->9s
* [#568](https://github.com/derailed/k9s/issues/568) Allow both .yaml and .yml yaml config files

---

## Contributed PRs

Please be sure to give `Big Thanks!` and `ATTA Girls/Boys!` to all the fine contributors for making K9s better for all of us!!

* [#2322](https://github.com/derailed/k9s/pull/2322) Check if the service provides selectors
* [#2319](https://github.com/derailed/k9s/pull/2319) Proper handling of help commands (fixes #2154)
* [#2315](https://github.com/derailed/k9s/pull/2315) Fix namespace suggestion error on context switch
* [#2313](https://github.com/derailed/k9s/pull/2313) Should not clear screen when executing plugin command
* [#2310](https://github.com/derailed/k9s/pull/2310) chore: Mot recommended to use k8s.io/kubernetes as a dependency
* [#2303](https://github.com/derailed/k9s/pull/2303) Clean up items
* [#2301](https://github.com/derailed/k9s/pull/2301) feat: Add imagePullSecrets and imagePullPolicy configuration for shellpod
* [#2289](https://github.com/derailed/k9s/pull/2289) Clean up issues introduced in #2125
* [#2288](https://github.com/derailed/k9s/pull/2288) Fix merge issues from PR #2168
* [#2284](https://github.com/derailed/k9s/issues/2284) Allow both .yaml and .yml yaml config files

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2023 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
