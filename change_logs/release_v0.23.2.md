<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.23.2

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better are as ever very much noted and appreciated!

If you feel K9s is helping your Kubernetes journey, please consider joining our [sponsorship program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

### Write Mode

K9s is writable by default, meaning you can interact with your cluster and make changes using one shot commands ie edit, delete, scale, etc... There `readOnly` config option that can be specified in the configuration or via a cli arg to override this behavior. In this drop, we're introducing a symmetrical command line arg aka `--write` that overrides a K9s session and make it writable tho the readOnly config option is set to true.

## Inverse Log Filtering

In the last drop, we've introduces reverse filters to filter out resources from table views. Now you will be able to apply inverse filtering on your log views as well via `/!fred`

---

## Resolved Issues/Features

* [Issue #906](https://github.com/derailed/k9s/issues/906) Print resources in pod view. With Feelings. Thanks Claudio!
* [Issue #889](https://github.com/derailed/k9s/issues/889) Disable readOnly config
* [Issue #564](https://github.com/derailed/k9s/issues/564) Invert filter mode on logs

## Resolved PRs

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
