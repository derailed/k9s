<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s.png" align="center" width="800" height="auto"/>

# Release v0.50.17

## Notes

🥳🎉 Happy new year fellow k9ers!🎊🍾 Hoping 2026 will bring good health and great success to you and yours...

Thank you to all that contributed with flushing out issues and enhancements for K9s!
I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev
and see if we're happier with some of the fixes!
If you've filed an issue please help me verify and close.

Your support, kindness and awesome suggestions to make K9s better are, as ever, very much noted and appreciated!
Also big thanks to all that have allocated their own time to help others on both slack and on this repo!!

As you may know, K9s is not pimped out by big corporations with deep pockets, thus if you feel K9s is helping in your Kubernetes journey, please consider joining our [sponsorship program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/zt-3360a389v-ElLHrb0Dp1kAXqYUItSAFA)

---

## ♫ Sounds Behind The Release ♭

* [A cool new way - Joe Satriani](https://www.youtube.com/watch?v=4apA948yOF0)
* [Song for you - Ray Charles](https://www.youtube.com/watch?v=CzAkTrDiXxg)
* [Kill the pain - SYZGYX](https://www.youtube.com/watch?v=5XuvMhHZorw&list=RD5XuvMhHZorw&start_radio=1)

---

## Maintenance Release!

Sponsorships are dropping at an alarming rate which puts this project in the red. This is becoming a concern and sad not to mention unsustainable ;( If you dig `k9s` and want to help the project, please consider `paying it forward!` and don't become just another `satisfied, non paying customer!`. K9s does take a lot of my `free` time to maintain, enhance and keep the light on. Many cool ideas are making it straight to the `freezer` as I just can't budget them in.
I know many of you work for big corporations, so please put in the word/work and have them help us out via sponsorships or other means.

Thank you!


## A Word From Our Sponsors...

To all the good folks and orgs below that opted to `pay it forward` and join our sponsorship program, I salute you!!

* [Philomena Yeboah](https://github.com/PhilomenaYeboah1989)
* [Kilian](https://github.com/kaerbr)
* [TVRiddle](https://github.com/TVRiddle)
* [Tom Morelly](https://github.com/FalcoSuessgott)
* [Nikhil Narayen](https://github.com/nnarayen)
* [Andrew Aadland](https://github.com/DaemonDude23)
* [Radek](https://github.com/radvym)
* [Timothée Gerber](https://github.com/TimotheeGerber)
* [Matthias](https://github.com/maetthu)
* [DKB](https://github.com/dkb-bank) ❤️
* [Kraken Tech](https://github.com/kraken-tech)
* [Daniel](https://github.com/sherlock7402)
* [Fred Loucks](https://github.com/fullmetal-fred)
* [Patricia Mascaros](https://github.com/ccong2586)
* [Qube Research & Technologies](https://github.com/qube-rt)
* [Michel Jung](https://github.com/micheljung)
* [Ümüt Özalp](https://github.com/uozalp)
* [Nathan Papapietro](https://github.com/npapapietro)
* [Oleksandr Podze](https://github.com/dasdy)
* [Lee Jones](https://github.com/leejones)
* [tsahlif](https://github.com/tshalif)
* [Jean-Christophe Amiel](https://github.com/jcamiel)
* [Lightspark](https://github.com/lightsparkdev)
* [egs-hub](https://github.com/egs-hub) ❤️
* [Sergey](https://github.com/malsatin)
* [Wynter Inc](https://github.com/copytesting)
* [Jen Norris](https://github.com/tnorris)
* [Joakim-Byg](https://github.com/Joakim-Byg)
* [Oleksandr Podze](https://github.com/dasdy)
* [Lee Jones](https://github.com/leejones)

> Sponsorship cancellations since the last release: **17!** 🥹

## Resolved Issues

* [#3765](https://github.com/derailed/k9s/issues/3765) quay.io docker images not up to date but referenced in README.md
* [#3762](https://github.com/derailed/k9s/issues/3762) Copy multiple selected items
* [#3751](https://github.com/derailed/k9s/issues/3751) Improve visual distinction for cordoned nodes in Node view
* [#3735](https://github.com/derailed/k9s/issues/3735) Cannot decode secret if there is no get permissions for all secrets
* [#3708](https://github.com/derailed/k9s/issues/3708) Editing a single Namespace opens the editor with a list of all Namespaces
* [#3731](https://github.com/derailed/k9s/issues/3731) feat: add neat plugin

* [#3735](https://github.com/derailed/k9s/issues/3735) Cannot decode secret if there is no get permissions for all secrets
* [#3708](https://github.com/derailed/k9s/issues/3708) Editing a single Namespace opens the editor with a list of all Namespaces

---

## Contributed PRs

Please be sure to give `Big Thanks!` and `ATTA Girls/Boys!` to all the fine contributors for making K9s better for all of us!!

* [#3763](https://github.com/derailed/k9s/pull/3763) feat: enable copying multiple resource, namespace names to clipboard
* [#3760](https://github.com/derailed/k9s/pull/3760) fix: Editing a single Namespace opens the editor with a list of all Namespaces
* [#3756](https://github.com/derailed/k9s/pull/3756) feat: Add reconcile plugin for Flux instances
* [#3755](https://github.com/derailed/k9s/pull/3755) fix: panic on 'jump to owner' of reflect.Value.Elem on zero Value
* [#3753](https://github.com/derailed/k9s/pull/3553) feat: add plugins for argo workflows
* [#3750](https://github.com/derailed/k9s/pull/3750) fix: Flux trace plugin shortcut conflict by changing to Shift-Q
* [#3749](https://github.com/derailed/k9s/pull/3749) feat: add dark/light theme inversion using Oklch
* [#3739](https://github.com/derailed/k9s/pull/3739) chore: refine LabelsSelector comment to match function behavior
* [#3738](https://github.com/derailed/k9s/pull/3738) feat: add symlink handle for plugin directory
* [#3720](https://github.com/derailed/k9s/pull/3720) fix(internal/render): ensure object is deep copied before realization in Render method
* [#3704](https://github.com/derailed/k9s/pull/3704) Allow k9s to start without a valid Kubernetes context
* [#3699](https://github.com/derailed/k9s/pull/3699) feat(pulse): map hjkl to navigate as help shows
* [#3697](https://github.com/derailed/k9s/pull/3697) Issue 3667 Fix
* [#3696](https://github.com/derailed/k9s/pull/3696) fix for scale option appearing on non-scalable resources
* [#3690](https://github.com/derailed/k9s/pull/3690) feat: add support for scaling HPA targets
* [#3671](https://github.com/derailed/k9s/pull/3671) fix fails to modify or delete namespaces using RBAC
* [#3669](https://github.com/derailed/k9s/pull/3669) feat: logs column lock
* [#3663](https://github.com/derailed/k9s/pull/3663) Map Q to "Back"
* [#3859](https://github.com/derailed/k9s/pull/3859) fix: update busybox image version to 1.37.0 in configuration files
* [#3650](https://github.com/derailed/k9s/pull/3650) Sort all columns
* [#3458](https://github.com/derailed/k9s/pull/3458) Document how to install on Fedora

---
<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2026 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)#