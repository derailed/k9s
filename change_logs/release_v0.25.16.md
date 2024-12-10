<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s.png" align="center" width="800" height="auto"/>

# Release v0.25.16

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better are as ever very much noted and appreciated!

If you feel K9s is helping your Kubernetes journey, please consider joining our [sponsorship program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

### A Word From Our Sponsors...

I want to recognize the following folks that have been kind enough to join our sponsorship program and opted to `pay it forward`!

* [Sebastian Racs](https://github.com/sebracs)
* [Timothy C. Arland](https://github.com/tcarland)
* [Julie Ng](https://github.com/julie-ng)

So if you feel K9s is helping with your productivity while administering your Kubernetes clusters, please consider pitching in as it will go a long way in ensuring a thriving environment for this repo and our K9sers community at large.

Also please take some time and give a huge shoot out to all the good folks below that have spent time plowing thru the code to help improve K9s for all of us!

Thank you!!

---

## ♫ Sounds Behind The Release ♭

[Blue Christmas - Fats Domino](https://www.youtube.com/watch?v=7jeo09zAskc)
[Mele Kalikimaka - Bing Crosby](https://www.youtube.com/watch?v=hEvGKUXW0iI)
[Cause - Rodriguez -- Spreading The Holiday Cheer! 🤨](https://www.youtube.com/watch?v=oKFkc19T3Dk)

---

## 🎅🎄 !!Merry Christmas To All!! 🎄🎅

I hope you will take this time of the year to relax, re-source and spend quality time with your loved ones. I know it's been a `tad rocky` of recent ;( as I've gotten seriously slammed with work in the last few months...
The fine folks here on this channel have been nothing but kind, patient and willing to help, this humbles me! I feel truly blessed to be affiliated with our great `k9sers` community!
Next month, we'll celebrate our anniversary as we've started out in this venture back in Jan 2019 (Yikes!) so get crack'in and iron out those bow ties already!!

Best wishes for great health, happiness and continued success for 2022 to you all!!

-Fernand

---

## A Christmas Story...

As of this drop, we've added a new feature to override the sort column and order for a given Kubernetes resource. This feature piggy backs of custom column views and add a new attribute namely `sortColumn`. For example say you'd like to set the default sort for pods to age descending vs name/namespace, you can now do the following in your `views.yml` file in the k9s config directory:

NOTE: This file is live thus you can nav to your fav resource, change the column config and view the resource columns and sort changes... Woot!!

```yaml
k9s:
  views:
    v1/endpoints:
      columns:
        - NAME
        - NAMESPACE
        - ENDPOINTS
        - AGE
    v1/pods:
      sortColumn: AGE:desc  # => suffix [:asc|:desc] for ascending or descending order.
    v1/services:
      ...
```

---

## Resolved Issues

* [Issue #1398](https://github.com/derailed/k9s/issues/1398) Pod logs containing brackets not in k9s logs output
* [Issue #1397](https://github.com/derailed/k9s/issues/1397) Regression: k9s no longer starts in current context namespace since v0.25.12
* [Issue #1358](https://github.com/derailed/k9s/issues/1358) Namespaces list is empty
* [Issue #956](https://github.com/derailed/k9s/issues/956) Feature request : Default column sort (by resource view)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2021 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
