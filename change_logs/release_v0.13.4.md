<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.13.4

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

Maintenance Release!

## GH Sponsors

A Big Thank You to the following folks that I've decided to dig in and give back!! 👏🙏🎊
Thank you for your gesture of kindness and for supporting K9s!! (not to mention for replenishing my liquids during oh-dark-thirty hours 🍺🍹🍸)

* [w11d](https://github.com/w11d)
* [vglen](https://github.com/vglen)

## CPU/MEM Metrics

A small change here based on [Benjamin](https://github.com/binarycoded) excellent PR! We've added 2 new columns for pod/container views to indicate percentages of resources request/limits if set on the containers. The columns have been renamed to represent the resources requests/limits as follows:

| Name   | Description                    | Sort Keys |
|--------|--------------------------------|-----------|
| %CPU/R | Percentage of requested cpu    | shift-x   |
| %MEM/R | Percentage of requested memory | shift-z   |
| %CPU/L | Percentage of limited cpu      | ctrl-x    |
| %MEM/L | Percentage of limited memory   | ctrl-z    |

---

## Resolved Bugs/Features

* [Issue #507](https://github.com/derailed/k9s/issues/507) ??May be??
* [PR #489](https://github.com/derailed/k9s/issues/489) ATTA Boy! [Benjamin](https://github.com/binarycoded)
* [PR #491](https://github.com/derailed/k9s/issues/491) Big Thanks! [Bjoern](https://github.com/bjoernmichaelsen)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
