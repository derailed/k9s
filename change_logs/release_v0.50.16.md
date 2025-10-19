<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s.png" align="center" width="800" height="auto"/>

# Release v0.50.16

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s!
I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev
and see if we're happier with some of the fixes!
If you've filed an issue please help me verify and close.

Your support, kindness and awesome suggestions to make K9s better are, as ever, very much noted and appreciated!
Also big thanks to all that have allocated their own time to help others on both slack and on this repo!!

As you may know, K9s is not pimped out by big corporations with deep pockets, thus if you feel K9s is helping in your Kubernetes journey, please consider joining our [sponsorship program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/zt-3360a389v-ElLHrb0Dp1kAXqYUItSAFA)

## Maintenance Release!

Sponsorships are dropping at an alarming rate which puts this project in the red. This is becoming a concern and sad not to mention unsustainable ;( If you dig `k9s` and want to help the project, please consider `paying it forward!` and don't become just another `satisfied, non paying customer!`. K9s does take a lot of my `free` time to maintain, enhance and keep the light on. Many cool ideas are making it straight to the `freezer` as I just can't budget them in.
I know many of you work for big corporations, so please put in the word/work and have them help us out via sponsorships or other means.

Thank you!

### Warp Speed Scotty!

As of this drop, we are introducing `namespace warp` via shortcut `w`.
This affords to view all resources of that type based on the currently selected resource namespace.
This command is only available on namespaced resources.
For example, if you are in pod view and select pod-xxx in namespace `bozo`, hitting `w` will `warp`
you to view all pods in namespace `bozo`.

## Resolved Issues

* [#3629](https://github.com/derailed/k9s/issues/3629) vulnerability in k9s project
* [#3621](https://github.com/derailed/k9s/issues/3621) Switching to ":Deploy" sends you to deployments from namespace "deploy"
* [#3620](https://github.com/derailed/k9s/issues/3620) Trying to show pod yaml using custom views.yaml crashes k9s
* [#3608](https://github.com/derailed/k9s/issues/3608) k9s crashes when :namespaces used
* [#3601](https://github.com/derailed/k9s/issues/3601) Can't delete namespace
* [#3595](https://github.com/derailed/k9s/issues/3595) Toggle Namespace Filter in Pods View with 'n' Key
* [#3576](https://github.com/derailed/k9s/issues/3576) Custom alias/view not working anymore since v0.50.10

---

## Contributed PRs

Please be sure to give `Big Thanks!` and `ATTA Girls/Boys!` to all the fine contributors for making K9s better for all of us!!

* [#3625](https://github.com/derailed/k9s/pull/3625) fix: debug-container plugin when KUBECONFIG has multiple files
* [#3623](https://github.com/derailed/k9s/pull/3623) bugfix: fix panic in BenchmarkPodRender by using NewPod() constructor
* [#3619](https://github.com/derailed/k9s/pull/3619) feat: plugin to list all resources by namespace
* [#3605](https://github.com/derailed/k9s/pull/3605) browser: do not prevent redraw when connection unavailable
* [#3600](https://github.com/derailed/k9s/pull/3600) fix(shell): set linux when OS detection fails
* [#3588](https://github.com/derailed/k9s/pull/3588) fix: do not error out of shellIn if OS detection fails


---
<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2025 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)#