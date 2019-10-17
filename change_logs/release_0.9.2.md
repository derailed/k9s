<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.9.2

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

I am absolutely blown away by your support and excitement about K9s! As I can recall, this is the first drop since we've launched K9s
back in January 2019 that I've seen some many external contributions and PRs. Thank you!! This is both super exciting and humbling.

### Core +1

As you may have noticed, there is a new voice on the project. [Gustavo Silva Paiva](https://github.com/paivagustavo) kindly accepted to become a K9s core member. Gustavo has been following and contributing to K9s for a while now and have patiently plowed thru my code ;( Raising issues, fixing them, improving code and test coverage, he has demonstrated a genuine interest on making sure K9s is better for all of us.

Actually, I can say enough about Gustavo since I don't know him that well yet ;) But I can tell from my interactions with him that he is a great human being and continues the K9s tradition of kindness and respect. Please help me in welcoming him to the K9s pac!

### Breaking Bad

There was an issue with the header toggle mnemonic `Ctrl-H` and it has been changed on this release to just `h`. Thank you for the heads up [Swe Covis](https://github.com/swe-covis)!!

## Merged PRs

* [PR #365](https://github.com/derailed/k9s/pull/365) Fix Alias columns sorting.
* [PR #363](https://github.com/derailed/k9s/issues/363) Change Terminated to Terminating
* [PR #360](https://github.com/derailed/k9s/pull/360) Header toggle while typing commands
* [PR #359](https://github.com/derailed/k9s/pull/359) Add support for CRD v1beta1
* [PR #356](https://github.com/derailed/k9s/pull/356) Remove Object field from CRD yaml
* [PR #347](https://github.com/derailed/k9s/pull/347) Sort node roles
* [PR #346](https://github.com/derailed/k9s/pull/346) Optimize configmap and secret rendering
* [PR #342](https://github.com/derailed/k9s/pull/342) Add copy YAML to clipboard
* [PR #338](https://github.com/derailed/k9s/pull/338) Escape describe text
* [PR #330](https://github.com/derailed/k9s/pull/330) Don't override standard K8s short names
* [PR #324](https://github.com/derailed/k9s/pull/324) Leverage cached client to speed up K9s

---

## Resolved Bugs/Features

* [Issue #361](https://github.com/derailed/k9s/issues/361)
* [Issue #341](https://github.com/derailed/k9s/issues/341)
* [Issue #335](https://github.com/derailed/k9s/issues/335)
* [Issue #331](https://github.com/derailed/k9s/issues/331)
* [Issue #323](https://github.com/derailed/k9s/issues/323)
* [Issue #280](https://github.com/derailed/k9s/issues/280)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
