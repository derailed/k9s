<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.19.0

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, please consider sponsoring 👆us or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

## A Word From Our Sponsors...

It makes me always very happy to hear folks are digging this effort and using K9s daily! If you feel this way please tell us and consider joining our [sponsorship](https://github.com/sponsors/derailed) program.

Big Thank You! to [hornbech](https://github.com/hornbech) for joining our sponsors!

## K8s v1.18.0 Support

As you might have heard, the good Kubernetes folks just dropped some big features in this new release. ATTA Girls/Boys!! We've (painfully) updated K9s to now link with the latest and greatest apis. Likely more work will need to take place here as I am still trying to catch up with the latest enhancements. This is great to see and excellent for all our Kubernetes friends!

## Oh Biffs'em And Buffs'em Popeye!

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_popeye.png" align="center" width="400" height="auto"/>

As you may know, I am the author of [Popeye](https://popeyecli.io) a Kubernetes cluster linter/sanitizer. Popeye scans your clusters live and reports potential issues, things like: referential integrity, misconfiguration, resource usage, etc...
In this drop, we've integrated K9s and Popeye to produce what I believe is a killer combo. Not only can you manage/observe your cluster resources in the wild, but you can now assert that your resources are indeed cool and potentially get rid of dead weights that might add up to your monthly cloud service bills. How cool is that?

In order to run your sanitization and produce reports, you can enter a new command `:popeye`. Once your cluster sanitization is complete, you can use familiar keyboard shortcuts to sort columns and view the sanitization reports by pressing `enter` on a given resource linter. Popeye also supports a configuration file namely `spinach.yml`, this file provides for customizing what resources get scanned as well as setting different severity levels to your own company policies. Please read the Popeye docs on how to customize your reports. The spinach.yml file will be read from K9s home directory `$HOME/.k9s/MY_CLUSTER_CONTEXT_NAME_spinach.yml`

NOTE! This is very much still experimental, so you may experience some disturbances in the force! And remember PRs are always open ;)

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/popeye/sanitizers.png" align="center" width="400" height="auto"/>
<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/popeye/report.png" align="center" width="400" height="auto"/>

## Command History Support

K9s now supports for command history. Entering command mode via `:` you can now up/down arrow to navigate thru your command history. Pressing `tab` or `ctrl-e` or `->` will activate the selected command upon `enter`.

## K9s Icons

Some terminals often don't offer icon support. In this release there is a new option `noIcons` available to enable/disable K9s icons. By default this option is set `false`. You can now set your icon preference in the K9s config file as follows:

```yaml
# $HOME/.k9s/config.yml
k9s:
  refreshRate: 2
  headless: false
  readOnly: false
  noIcons: true  # Enable/Disable K9s icons display.
```

## Videos!

* [video v0.19.0](https://www.youtube.com/watch?v=kj-WverKZ24)
* [video v0.18.0](https://www.youtube.com/watch?v=zMnD5e53yRw)

## Resolved Bugs/Features/PRs

* [Issue #647](https://github.com/derailed/k9s/issues/647)
* [Issue #645](https://github.com/derailed/k9s/issues/645)
* [Issue #640](https://github.com/derailed/k9s/issues/640)
* [Issue #639](https://github.com/derailed/k9s/issues/639)
* [Issue #635](https://github.com/derailed/k9s/issues/635)
* [Issue #634](https://github.com/derailed/k9s/issues/634) Thank you!! [David Němec](https://github.com/davidnemec)
* [Issue #626](https://github.com/derailed/k9s/issues/626)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
