<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s.png" align="center" width="800" height="auto"/>

# Release v0.31.6

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

## Maintenance Release!

😱 More aftermath... 😱

Thank you all for pitching in and helping flesh out issues!!

Please make sure to add gory details to issues ie relevant configs, debug logs, etc...

Comments like: `same here!` or `me to!` doesn't really cut it for us to zero in ;(
Everyone has slightly different settings/platforms so every little bits of info helps with the resolves even if seemingly irrelevant.

---

## NOTE

In this drop, we've made k9s a bit more resilient (hopefully!) to configuration issues and in most cases k9s will come up but may exhibit `limp mode` behaviors.
Please double check your k9s logs if things don't work as expected and file an issue with the `gory` details!

☢️ This drop may cause `some disturbance in the farce!` ☢️

Please proceed with caution with this one as we did our best to attempt to address potential context config file corruption by eliminating race conditions.
It's late and I am operating on minimal sleep so I may have hosed some behaviors 🫣
If you experience k9s locking up or misbehaving, as per the above👆 you know what to do now and as customary
we will do our best to address them quickly to get you back up and running!

Thank you for your support, kindness and patience!

---

## Videos Are In The Can!

Please dial [K9s Channel](https://www.youtube.com/channel/UC897uwPygni4QIjkPCpgjmw) for up coming content...

* [K9s v0.31.0 Configs+Sneak peek](https://youtu.be/X3444KfjguE)
* [K9s v0.30.0 Sneak peek](https://youtu.be/mVBc1XneRJ4)
* [Vulnerability Scans](https://youtu.be/ULkl0MsaidU)

---

## Resolved Issues

* [#2476](https://github.com/derailed/k9s/issues/2476) Pods are not displayed for the selected namespace. Hopefully!
* [#2471](https://github.com/derailed/k9s/issues/2471) Shell autocomplete functions do not work correctly

---

## Contributed PRs

Please be sure to give `Big Thanks!` and `ATTA Girls/Boys!` to all the fine contributors for making K9s better for all of us!!

* [#2480](https://github.com/derailed/k9s/pull/2480) Adding system arch to nodes view
* [#2477](https://github.com/derailed/k9s/pull/2477) Shell autocomplete for k8s flags

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2024 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)