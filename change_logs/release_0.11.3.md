<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.11.3

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_helm.png" align="center" width="300" height="auto"/>

Maintenance Release!

### Speedy Gonzales?

In this drop, we took a bit of a perf pass in light of recent issues and thanks to [Chris Werner Rau](https://github.com/cwrau) pushing me and keeping me up to speed, I've digged a bit deeper and found that there might be some seamingly innocent calls that sucked a bit of cycles during K9s refreshes. Long story short, I think this drop will improve perf by a factor of ~10x in some instances. Typically the initial load will be slower but subsequent loads should be much faster. Famous last words right? Anyhow, can't really take credit for this one as the awesome [Gustavo Silva Paiva](https://github.com/paivagustavo) suggested doing this a while back, but since I was already in flight with the refactor decided to punt until back online. And there we are...

Hopefully these findings will coalesce with yours?? If not, please forward bulk Prozac patches at the address below ;)

Thanks Chris! Was up all night trying to figure out what was the deal with K9s and your specific clusters. Hopefully this time for sure??

---

## Resolved Bugs/Features

* [Issue #475](https://github.com/derailed/k9s/issues/475)
* [Issue #473](https://github.com/derailed/k9s/issues/473)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
