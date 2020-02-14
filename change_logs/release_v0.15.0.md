<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.15.0

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

## Does This FaaS Look Familiar?

The awesome and ever so smart and creative [Alex Ellis](https://github.com/alexellis) of [OpenFaas Fame](https://www.openfaas.com) fame, had pinged me when I had launched K9s to add support for OpenFaas functions. It's been a long time coming indeed, but we now have a very (VERY!) primitive integration with this very cool framework.

The current approach is to enable a few environment variables to tell K9s that you have an OpenFaas cluster available namely:

```shell
OPENFAAS_GATEWAY=http://YOUR_CLUSTER_IP:31112
OPENFAAS_TLS_INSECURE=false
OPENFAAS_JWT_TOKEN=YOUR_TOKEN
```

These will tell K9s that an OpenFaas gateway is available and exposed on a given nodeport.

Next you can navigate to your OpenFaas function view by entering command mode `:openfaas` or using aliases `:ofaas` or `ofa`

If functions are present in the given namespace they will be displayed here just like any other K8s resources.

The following operations are currently supported:

* Describe and YAML to view function definitions (Note: currently yields same results!)
* Enter to view all pods instances associated with the selected function
* Delete a function
* Editing, shelling, logs, etc... are all supported by navigating to the underlying pods

Keep in mind, the paint is way fresh here and this feature could be a complete dud, but figure will give it a rinse on this drop and Alex can pipe in and helps us ironing this out.

> NOTE! It's been a while since I've played with OpenFaas so if some of you are more versed in this space by all means please do land a hand so we can make this feature more awesome!

## Moving Forward!

A few folks had mentioned the eagerness to port-forward directly from a pod or a service. Well now you can! Port Forwarding is now available on both the pod view and services view. Note! at the end of the day, you are still port-fowarding to a container! So the port-forward dialog is a bit different for these views as there might be several container ports available now when looking at this from a pod perspective. So the first field in the dialog is a combo-box that allows one to pick their desired ports. The rest of the dialog works the same as the container port-forward dialog.

## Resolved Bugs/Features/PRs

* [Issue #546](https://github.com/derailed/k9s/issues/546) BREAKING NEWS! Use `t` vs `ctrl-h` now to toggle the K9s header
* [Issue #541](https://github.com/derailed/k9s/issues/541)
* [Issue #227](https://github.com/derailed/k9s/issues/227)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
