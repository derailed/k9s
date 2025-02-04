<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.8.2

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

Maintenance release.

In this quick drop, we've opted to nuke any menu shortcut using the infamous `Alt` key. This includes the new pod kill command that is now `Ctrl-K` and for the most part the column sorting shortcuts for CPU% and MEMORY%. My apologizes to all on this fiasco as it turns out I had remapped opt->alt on my local dev machine and space it while trying to offer different key mappings. Will revisit this in the future when things simmer down a bit. Thank you to all that reported on this!

---

## Resolved Bugs/Features

+ Nuked Alt-XXX menu mnemonic [Issue #285](https://github.com/derailed/k9s/issues/285)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
