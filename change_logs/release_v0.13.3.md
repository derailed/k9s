<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.13.3

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

### XRay Now With Lipstick?

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_xray.png"/>

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/xray_icons.png"/>

Call me old school, but Xray without icons made me a bit sad ;( Just like any engineer would, I do fancy eye candy once in a while...
So I've decided to revive the xray `icon` mode for the some of us that are not stuck with what I'd like to call `Jurassic` terminals.
To date, there was no way to skin the Xray view, so I've added a new xray skin config section that `currently` looks like this:

```yaml
# $HOME/.k9s/skin.yml
k9s:
  body:
    fgColor: dodgerblue
    bgColor: black
    logoColor: orange
  ...
  xray:
      fgColor: blue
      bgColor: black
      cursorColor: aqua
      graphicColor: darkgoldenrod
      # NOTE! Show xray in icon mode. Defaults to false!!
      showIcons: true
```

So if your terminal does not support emoji's we're still cool...

---

## Resolved Bugs/Features

* [Issue #505](https://github.com/derailed/k9s/issues/505)
* [Issue #504](https://github.com/derailed/k9s/issues/504)
* [Issue #503](https://github.com/derailed/k9s/issues/503)
* [Issue #501](https://github.com/derailed/k9s/issues/501)
* [Issue #499](https://github.com/derailed/k9s/issues/499)
* [Issue #493](https://github.com/derailed/k9s/issues/493)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
