<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s.png" align="center" width="800" height="auto"/>

# Release v0.31.0

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

---

## ♫ Sounds Behind The Release ♭

* [Border Crossing - Eek A Mouse](https://www.youtube.com/watch?v=KaAC9dBPcOM)
* [The Weight - The Band](https://www.youtube.com/watch?v=FFqb1I-hiHE)
* [Wonderin' - Neil Young](https://www.youtube.com/watch?v=h0PlwVPbM5k)
* [When Your Lover Has Gone - Louis Armstrong](https://www.youtube.com/watch?v=1tdfIj0fvlA)

---

## A Word From Our Sponsors...

To all the good folks below that opted to `pay it forward` and join our sponsorship program, I salute you!!

* [Jacky Nguyen](https://github.com/nktpro)
* [Eckl, Máté](https://github.com/ecklm)
* [Jörgen](https://github.com/wthrbtn)
* [kmath313](https://github.com/kmath313)
* [a-thomas-22](https://github.com/a-thomas-22)
* [wpbeckwith](https://github.com/wpbeckwith)
* [Dima Altukhov](https://github.com/alt-dima)
* [Shoshin Nikita](https://github.com/ShoshinNikita)
* [Tu Hoang](https://github.com/rebyn)
* [Andreas Frangopoulos](https://github.com/qubeio)

> Sponsorship cancellations since the last release: **7!** 🥹

## Feature Release!

😳 Found a few issues in the neutrino drive...
This is another fairly heavy drop so bracing for impact 😱
Be sure to dial in the v0.31.0 SneakPeek video below for the gory details!

😵 Hopefully we've move the needle in the right direction on this drop... 🤞

Thank you all for your kindness, feedback and assistance in flushing out issues!!

### Hold My Hand...

In this drop, we've added schema validation to ensure various configs are setup as expected.
K9s will now run validation checks on the following configurations:

1. K9s main configuration (config.yaml)
2. Context specific configs (clusterX/contextY/config.yaml)
3. Skins
4. Aliases
5. HotKeys
6. Plugins
7. Views

K9s behavior changed in this release if the main configuration does not match schema expectations.
In the past, the configuration will be validated, updated and saved should validation checks failed. Now the app will stop and report validation issues.

The schemas are set to be a bit loose for the time being. Once we/ve vetted they are cool, we could publish them out (with additional TLC!) so k9s users can leverage them in their favorite editors.

In the meantime, you'll need to keep k9s logs handy, to check for validation errors. The validation messages can be somewhat cryptic at times and so please be sure to include your debug logs and config settings when reporting issues which might be plenty ;(.

### Breaking Bad!

Configuration changes:

1. DRY fullScreenLogs -> fullScreens (k9s root config.yaml)

   ```yaml
   #  $XDG_CONFIG_HOME/k9s/config.yaml
   k9s:
     liveViewAutoRefresh: false
     logger:
       sinceSeconds: -1
       fullScreen: false # => Was fullScreenLogs
     ...
   ```

2. Views Configuration.
   To match other configurations the root is now `views:` vs `k9s: views:`

   ```yaml
   # $XDG_CONFIG_HOME/k9s/views.yaml
   views: # => Was k9s:\n  views:
    v1/pods:
      columns:
        - AGE
        - NAMESPACE
        ...
   ```

### Serenity Now!

   You can now opt in/out of the `reactive ui` feature. This feature enable users to make change to some configurations and see changes reflected live in the ui. This feature is now disabled by default and one must opt-in to enable via `k9s.UI.reactive`
   Reactive UI provides for monitoring various config files on disk and update the UI when changes to those files occur. This is handy while tuning skins, plugins, aliases, hotkeys and benchmarks parameters.

   ```yaml
   # $XDG_CONFIG_HOME/k9s/config.yaml
   k9s:
     liveViewAutoRefresh: false
     UI:
       ...
       reactive: true # => enable/disable reactive UI
     ...
   ```

---

## Videos Are In The Can!

Please dial [K9s Channel](https://www.youtube.com/channel/UC897uwPygni4QIjkPCpgjmw) for up coming content...

* [K9s v0.31.0 Configs+Sneak peek](https://youtu.be/X3444KfjguE)
* [K9s v0.30.0 Sneak peek](https://youtu.be/mVBc1XneRJ4)
* [Vulnerability Scans](https://youtu.be/ULkl0MsaidU)

---

## Resolved Issues

* [#2434](https://github.com/derailed/k9s/issues/2434) readOnly: true in config.yaml doesn't get overridden by readOnly: false in cluster config
* [#2430](https://github.com/derailed/k9s/issues/2430) Referencing a namespace with the name of an alias inside an alias causes infinite loop
* [#2428](https://github.com/derailed/k9s/issues/2428) Boom!! runtime error: invalid memory address or nil pointer dereference - v0.30.8
* [#2421](https://github.com/derailed/k9s/issues/2421) k9s/config.yaml configuration file is overwritten on launch

---

## Contributed PRs

Please be sure to give `Big Thanks!` and `ATTA Girls/Boys!` to all the fine contributors for making K9s better for all of us!!

* [#2433](https://github.com/derailed/k9s/pull/2433) switch contexts only when needed
* [#2429](https://github.com/derailed/k9s/pull/2429) Reference correct configuration ENV var in README
* [#2426](https://github.com/derailed/k9s/pull/2426) Update carvel plugin kick to shift K
* [#2420](https://github.com/derailed/k9s/pull/2420) supports referencing envs in hotkeys
* [#2419](https://github.com/derailed/k9s/pull/2419) fix typo

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2024 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
