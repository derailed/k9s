<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s.png" align="center" width="800" height="auto"/>

# Release v0.40.0

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

* [Glory Box - Portishead](https://www.youtube.com/watch?v=4qQyUi4zfDs)
* [Hit Me With Your Rhythm Stick - Ian Dury And The BlockHeads](https://www.youtube.com/watch?v=0WGVgfjnLqc)
* [Cupidon s'en fout! - George Brassens](https://www.youtube.com/watch?v=a-RlZLfIeKM)
* [Shipbuilding - Elvis Costello](https://www.youtube.com/watch?v=dVhjRqBM5uw)
* [Low Sun - Hermanos Gutierrez](https://www.youtube.com/watch?v=ubaJbw7hkeQ)

---

## A Word From Our Sponsors...

To all the good folks below that opted to `pay it forward` and join our sponsorship program, I salute you!!

* [Panfactum](https://github.com/Panfactum)
* [Bastian Pätzold](https://github.com/bastianpaetzold)
* [Mikita Vazhnik](https://github.com/Vazhnik)
* [Jacob Salway](https://github.com/jacobsalway)
* [Eckard Mühlich](https://github.com/eckardnet)
* [Luke](https://github.com/lukepatrick)
* [tomasbanet](https://github.com/tomasbanet)
* [Robin Opletal](https://github.com/fourstepper)
* [Euroblaze](https://github.com/euroblaze)
* [Jack Daniels](https://github.com/dkr91)
* [decafcode](https://github.com/decafcode)
* [Guillaume Copin](https://github.com/GuillaumeCo)
* [Lokalise](https://github.com/lokalise)
* [Gustavo Bini](https://github.com/gustavobini)
* [JMSwag](https://github.com/JMSwag)
* [Daniel Gospodinow](https://github.com/danielgospodinow)
* [Klaviyo](https://github.com/klaviyo)
* [Paul Farver](https://github.com/PaulFarver)

> Sponsorship cancellations since the last release: **12!** 🥹

## 🎉 Feature Release code name: Colon Blow! 🎈

We are pretty stocked about this drop (hopefully...) as we've fully enabled custom columns support in K9s!
Historically, one could customize the view for a given resource by adding a definition in `views.yaml`.
From there one could change sort order and re-arrange the standard column layout.
Several folks voiced the need to add a column for a given label/annotation or any other fields available on a resource.
To date, this wasn't possible 😳

So... without further ado, let see what we can now do with `Custom Views` ding dang deal!
It all starts with a few new directives available in `views.yaml`

### A Refresher...

Customize a pod view and ensure age, ns and name appear first and sort by age descending.

> NOTE! You no longer need to list out all columns.
> The remaining columns will be automatically filled from the standard columns.

```yaml
# Usual biz...
views:
  v1/pods:                         # specify the gvr you want to customize aka group/version/resource
    sortColumn: AGE:desc           # set the default ordering to ascending (asc) or descending (desc)
    columns:                       # tell the view which columns to display and in which order
      - AGE                        # ensure age, ns and name are the first 3 cols and backfill the rest
      - NAMESPACE
      - NAME
      - READY|H                    # => NEW! Do not display the READY column
      - NODE|W                     # => NEW! Show node column only on wide
      - IP|WR                      # => NEW! Pull the ip column and right align it in wide mode only
```

## Colon Blow!

Say your pods comes standard with a label `blee` and you want to show it while in pod view.

```yaml
# Pull labels/annotations
views:
  v3/freds:
    sortColumn: NAMESPACE:dsc
    columns:
      - NAMESPACE
      - NAME
      - BLEE:.metadata.labels.blee                        # => NEW! Pull values from a label or an annotation using json parser
                                                          # expression similar mechanic as kubectl -o custom-columns
      - ZORG:.spec.zips[?(@.type == 'zorg')].ip|WR        # => NEW! Same deal with a json exp + but align right and show wide only
```

## TLDR...

As you can see the CustomView feature adds a few new semantics on this drop.

You can now use the following shape for columns definition `COL_NAME<:json_parse_expression><|column attributes>`

The `:json_parse_expression` is optional.

The column attributes are as follows:

* `T` -> time column indicator
* `N` -> number column indicator
* `W` -> turns on wide column aka only shows while in wide mode. Defaults to the standard resource definition when present.
* `H` -> Hides the column
* `L` -> Left align (default)
* `R` -> Right align

When certain columns are not present in the custom view, K9s will pull the standard column definition and merge the columns.
This allows user to specify and order which columns they want to see first without having to define every single columns from the default resource representation. If you do not wish to see all these columns you can add them to your custom view definition and either specify `|W` or `|H` to `wide` it or `hide` it.

> 📢 Still work in progress so your mileage may vary!
> This feature will likely need additional TLC.
> Your feedback on this will be much appreciated and we will iterate as usual to ensure it vorks as prescribed... 🙀

---

## Videos Are In The Can!

Please dial [K9s Channel](https://www.youtube.com/channel/UC897uwPygni4QIjkPCpgjmw) for up coming content...

* [K9s v0.40.0 Colon Blow Sneak peek](https://youtu.be/iy6RDozAM4A)
* [K9s v0.31.0 Configs+Sneak peek](https://youtu.be/X3444KfjguE)
* [K9s v0.30.0 Sneak peek](https://youtu.be/mVBc1XneRJ4)
* [Vulnerability Scans](https://youtu.be/ULkl0MsaidU)

---

## Resolved Issues

* [#3064](https://github.com/derailed/k9s/issues/3064) Question: brew formula k9s vs derailed/k9s/k9s
* [#3061](https://github.com/derailed/k9s/issues/3061) k9s not opening active namespace or namespace specified via -n
* [#3044](https://github.com/derailed/k9s/issues/3044) CRDs are loaded incorrectly into metadata registry, cause sporadic "Jump Owner" issues
* [#2995](https://github.com/derailed/k9s/issues/2995) Latest image on quay.io contains "failed" kubectl binary

---

## Contributed PRs

Please be sure to give `Big Thanks!` and `ATTA Girls/Boys!` to all the fine contributors for making K9s better for all of us!!

* [#3065](https://github.com/derailed/k9s/pull/3065) Fixed trimming of favorite namespaces in Config
* [#3063](https://github.com/derailed/k9s/pull/3063) Updating CVE dependencies
* [#3062](https://github.com/derailed/k9s/pull/3062) feat: use kubectl events for plugin watch-events
* [#3060](https://github.com/derailed/k9s/pull/3060) Rename "delete local data" checkbox description in drain dialog
* [#3046](https://github.com/derailed/k9s/pull/3046) Strict unmarshal for plugin files
* [#3045](https://github.com/derailed/k9s/pull/3045) fix: CRD loading: trim group suffix from CRD name
* [#3043](https://github.com/derailed/k9s/pull/3043) Fix K9S_EDITOR
* [#3041](https://github.com/derailed/k9s/pull/3041) Fix Flux trace plugin command
* [#3038](https://github.com/derailed/k9s/pull/2038) fix check e != nil but return a nil value error err
* [#3026](https://github.com/derailed/k9s/pull/3026) Fix typos
* [#3018](https://github.com/derailed/k9s/pull/3018) fix: coloring of rose-pine for values of log options
* [#3017](https://github.com/derailed/k9s/pull/3017) feat: add helm diff plugin
* [#3009](https://github.com/derailed/k9s/pull/3009) fix(argo-rollouts plugin): resolve improper piping in watch command
* [#2996](https://github.com/derailed/k9s/pull/2996) Bump version of netshoot image in debug-container plugin
* [#2994](https://github.com/derailed/k9s/pull/2994) fix kubectl url and fail build on download errors
* [#2986](https://github.com/derailed/k9s/pull/2986) plugin/trace-dns: Trace DNS requests using Inspektor Gadget
* [#2985](https://github.com/derailed/k9s/pull/2985) feat(plugins/crossplane): change to crossplane cli & add crossplane-watch
* [#2986](https://github.com/derailed/k9s/pull/2986) plugin/trace-dns: Trace DNS requests using Inspektor Gadget

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2024 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)