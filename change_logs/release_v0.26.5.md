<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s.png" align="center" width="800" height="auto"/>

# Release v0.26.5

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better are, as ever, very much noted and appreciated! Also big thanks to all that have allocated their own time to help others on both slack and on this repo!!

As you may know, K9s is not pimped out by corps with deep pockets, thus if you feel K9s is helping your Kubernetes journey, please consider joining our [sponsorship program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

## Maintenance Release

So it looks like replacing the clipboard package was indeed a dud ;(
While I was not keen on either running with cgo or taking on external dependencies, after further investigation it looks like the clipboard + wsl issue in the old package was [resolved](https://github.com/atotto/clipboard/pull/42). I don't run WSL so I can't test it but if that's not the case please reopen and we will figure out another solution. For the time being, I've opted for the reversal.
Thank you!!

---

## Resolved Issues

* [Issue #1742](https://github.com/derailed/k9s/issues/1770) copy to clipboard throw panic error
* [Issue #1768](https://github.com/derailed/k9s/issues/1768) build fails due to new clipboard package

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2022 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
