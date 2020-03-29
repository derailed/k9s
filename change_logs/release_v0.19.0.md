<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.19.0

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, please consider sponsoring 👆us or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

## A Word From Our Sponsors...

It makes me always very happy to hear folks are digging this effort and using K9s daily! If you feel this way please tell us and consider joining our [sponsorship](https://github.com/sponsors/derailed) program.
Big Thanks! to [hornbech](https://github.com/hornbech) for joining our sponsors!

## K8s v1.18.0 Released!

As you might have heard, the good Kubernetes folks just dropped some big features in this new release. ATTA Girls/Boys!! We've (painfully) updated K9s to now link with the latest and greatest apis. Likely more work will need to take place here as I am still trying to catch up with the latest enhancements. This is great to see and excellent for all our Kubernetes friends!

## Oh Biffs'em And Buffs'em Popeye!

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_popeye.png" align="center" width="400" height="auto"/>

As you may know, I am the author of [Popeye](https://popeyecli.io) a Kubernetes cluster linter/sanitizer. Popeye scans your live clusters and reports potential issues, such as: referential integrity, misconfiguration, resource usage, etc.. In this drop, we've integrated K9s and Popeye to produce what I believe is a killer two punches combo. Not only can you observe your cluster resources in the wild but you can now assert that your resources are indeed legit or get ride of dead weights that might add up to your montly cloud services bills. How cool is that?

In order to run your sanitization and produce reports, you can enter the command mode and type `popeye`. Once your cluster sanitization is complete, you can use familiar keyboard shortcuts to sort columms and view the sanitization reports by pressing `enter` on a given resource linter. Popeye also support a configuration file namely `spinach.yml`, this file affords for customizing what resources get scanned as well as setting different severity levels to your own company policies. Please read the Popeye docs on how to customize your reports. The spinach.yml file will be read from K9s home directory `$HOME/.k9s/spinach.yml`

NOTE! This is very much still experimental, so you may experience some disturbances in the force!

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/popeye/sanitizations.png" align="center" width="400" height="auto"/>
<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/popeye/reports.png" align="center" width="400" height="auto"/>

## Videos!

Forgot to link my glorious in the previous release [video](https://www.youtube.com/watch?v=zMnD5e53yRw)!!

## Resolved Bugs/Features/PRs

* [Issue #635](https://github.com/derailed/k9s/issues/635) Thank you!! [David Němec](https://github.com/davidnemec)
* [Issue #634](https://github.com/derailed/k9s/issues/634)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
