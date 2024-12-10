<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.10.0

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

First off, Happy 2020 to you and yours!! Best wishes for good health and good fortune!

This release represents a major overall of K9s core. It's been a long time coming and indeed a long day in the saddle. There has been many code changes and hopefully improvements from previous releases. I think some of it is better but I've probably borked a bunch of functionality in the process ;( I look to you to help me flesh out issues and bugs, so we can move on to bigger and exciting features in 2020! Please do thread lightly on this one and make sure to keep a previous release handy just in case... This was a boatload of work to make this happen, my hope is you'll enjoy some of the improvements... In any case, and as always, if you feel they're better ways or imperfections by all means please pipe in!

I would also like to take this opportunity to thank all of you for your kind PRs and issues and for your support and patience with K9s. I understand this release might be a bit torked, but I will work hard to make sure we reach stability quickly in the next few drops. Thank you for your understanding!!

## VatDoesDisDo?

Most of the refactors are around K8s resource fetching and viewing as well as navigation changes. Based on our observations this release might load resources a bit slower than usual but should make navigation much faster once the cache is primed. We've made some improvements to be more consistent with navigation, menus and shortcuts management. We've got ride off the breadcrumbs navigation ie no more `p` to nav back. Crumbs are now just tracking a natural resource navigation ie pod -> containers -> logs and no longer commands history. Each new command will now load a brand new breadcrumb. You can press `<esc>` to nav back to the previous page. We're also introducing a new hotkeys feature, that efforts creating shortcuts to navigate to your favorite resources ie shift-0 -> view pods, shift-1 -> view deployments (See HotKey section below). I know there were many outstanding PRS (Thank you to all that I've submitted!) and given the extent of the changes, I've resolved to incorporate them in manually vs having to deal with merge conflicts.

## Custom Skins Per Cluster

In this release, we've added support for skins at the cluster level. Do you want K9s to look differently based on which cluster you're connecting to? All you'll need is to name the skin file in the K9s home directory as follows `mycluster_skin.yml`. If no cluster specific skin file is found, the standard `skin.yml` file will be loaded if present. Please checkout the `skins` directory in this repo or PR me if you have cool skins you'd like to share with your fellow K9sers as they will be featured in these release notes and in the project README.

## Hot(Ness)?

Feeling like you want to be able to quickly switch around your favorite resources with your very own shortcut? Wouldn't it be dandy to navigate to your deployments via a shortcut vs entering a command `:deploy`? Here is what you'll need to do to add HotKeys to your K9s sessions:

1. In your .k9s home directory create a file named `hotkey.yml`
2. For example add the following to your `hotkey.yml`. You can use short names or resource name to specify a command ie same as typing it in command mode.

      ```yaml
      hotKey:
        shift-0:
          shortCut: Shift-0
          description: View pods
          command: pods
        shift-1:
          shortCut: Shift-1
          description: View deployments
          command: dp
        shift-2:
          shortCut: Shift-2
          description: View services
          command: service
        shift-3:
          shortCut: Shift-3
          description: View statefulsets
          command: statefulsets
      ```

 Not feeling too `Hot`? No worried, your custom hotkeys list will be listed in the help view.`<?>`.

 You can choose any keyboard shortcuts that make sense to you, provided they are not part of the standard K9s shortcuts list.

## PullRequests

* [PR #447](https://github.com/derailed/k9s/pull/447) K9s MacPorts support. Thank you! [Nils Breunese](https://github.com/breun)
* [PR #446](https://github.com/derailed/k9s/pull/446) Same key invert sort. Big thanks!! [James Hiew](https://github.com/jameshiew)
* [PR #445](https://github.com/derailed/k9s/pull/445) Use `?` to toggle help. Major thanks!! [Ramz](https://github.com/ageekymonk)
* [PR #443](https://github.com/derailed/k9s/pull/443) Hex color skin support. Great work! [Gavin Ray](https://github.com/gavinray97)
* [PR #442](https://github.com/derailed/k9s/pull/442) Full screen/Wrap support on log view. ATTA BOY! [Shiv3](https://github.com/shiv3)
* [PR #412](https://github.com/derailed/k9s/pull/412) Simplify cruder interface. ATTA BOY!! (as always)[Gustavo Silva Paiva](https://github.com/paivagustavo)
* [PR #350](https://github.com/derailed/k9s/pull/350) Sanitize file name before saving. All credits to [Tuomo Syvänperä](https://github.com/syvanpera)

---

## Resolved Bugs/Features

* [Issue #437](https://github.com/derailed/k9s/issues/437) Error when viewing cluster role on a role binding.
* [Issue #434](https://github.com/derailed/k9s/issues/434) Same key `?` toggle help.
* [Issue #432](https://github.com/derailed/k9s/issues/432) Add address field to port forwards.
* [Issue #431](https://github.com/derailed/k9s/issues/431) Add LimitRange resource support.
* [Issue #430](https://github.com/derailed/k9s/issues/430) Add HotKey support.
* [Issue #426](https://github.com/derailed/k9s/issues/426) Address slow scroll while in table view.
* [Issue #417](https://github.com/derailed/k9s/issues/417) Ensure code lints correctly. Thank you Gustavo!!
* [Issue #415](https://github.com/derailed/k9s/issues/415) Add provisions to support longer clusterinfo/namespace header.
* [Issue #408](https://github.com/derailed/k9s/issues/408) Same key toggle inverse sort.
* [Issue #402](https://github.com/derailed/k9s/issues/402) Add `all` support to plugin scope.
* [Issue #401](https://github.com/derailed/k9s/issues/401) Add support for custom plugins on all views.
* [Issue #397](https://github.com/derailed/k9s/issues/397) Support HPA v2beta1 + v2beta2.

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
