<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.16.0

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_doc.png" align="center"/>

This is one of these drops that may make you wonder if you'll go from zero to hero or likely the reverse?? Will see how this goes... Please proceed with caution on this one as there could very well be much disturbances in the force...

Lots of code churns so could have totally hose some stuff, but like my GranPappy used to say `can't cook without making a mess!`

## Going Wide?

In this drop, we've enabled a new shortcut namely `wide` as `Ctrl-w`. On table views, you will be able to see more information about the resources such as labels or others depending on the viewed resource. This mnemonic works as a toggle so you can `narrow` the view by hitting it again.

## Zoom, Zoom, Zoom!

While viewing some resources that may contain errors, sorting on columns may not achieve the results you're seeking ie `show me all resources in an error state`. We've added a new option to achieve just that aka `zoom errors` as `ctrl-z`. This works as a toggle and will unveil resources that are need of some TLC on your part ;)

## Does Your Cluster Have A Pulse 💓?

In this drop, we're introducing a brand new view aka `K9s Pulses` 💓. This is a summary view listing the most salient resources in your clusters and their current states. This view tracks two main metrics ie Ok and Toast on a 5sec beat. This view affords cluster activity and failure rates. BTW this is the zero to hero deal 🙀 Hopefully you'll dig it as this was much work to put together and I personally think it's the `ducks nuts`... If you like, please give me some luving on social or via GH sponsors as batteries are running low...

To active, enter command mode by typing in `:pulse` aliases are `pu`, `pulses` or `hz`
To navigate thru the various pulses, you can use `tab`/`backtab` or use the menu index (just like namespaces selectors). Once on a pulse view, you can press `enter` to see the associated resource table view. Pressing `esc` will nav you back.

As I've may have mentioned before, my front-end/UX FU is weak, so I've also added a way for you to skin the charts via skins yaml to your own liking. Please see the skin section below for an example on how to skin the pulses dials. BONUS you should be able to skin K9s live! How cool is that 😻?

NOTE: Pulses are very much experimental and could totally bomb on your clusters! So please thread carefully and please do report (kindly!) back.

## BReaking Bad!

In this drop I've broken a few things (that I know of...), here is the list as I can recall...

1. Toggle header aka `my red headed step child`. Key moved (again!) now `Ctrl-e`
2. Skin yaml layout CHANGED! Moved table and xray sections under views and added charts section.

## Skins Updates!

The skin file format CHANGE! If you are running skins with K9s, please make sure to update your skin file. If not K9s could bomb coming up!

NOTE: I don't think I'll get around to update all the contributed skins in this repo `skins` dir. If you're looking for a way to help out and are UI inclined, please take a peek and make them cool!

```yaml
# my_cluster_skin.yml
# Styles...
foreground: &foreground "#f8f8f2"
background: &background "#282a36"
current_line: &current_line "#44475a"
selection: &selection "#44475a"
comment: &comment "#6272a4"
cyan: &cyan "#8be9fd"
green: &green "#50fa7b"
orange: &orange "#ffb86c"
pink: &pink "#ff79c6"
purple: &purple "#bd93f9"
red: &red "#ff5555"
yellow: &yellow "#f1fa8c"

# Skin...
k9s:
  # General K9s styles
  body:
    fgColor: *foreground
    bgColor: *background
    logoColor: *purple
  # ClusterInfoView styles.
  info:
    fgColor: *pink
    sectionColor: *foreground
  frame:
    # Borders styles.
    border:
      fgColor: *selection
      focusColor: *current_line
    menu:
      fgColor: *foreground
      keyColor: *pink
      # Used for favorite namespaces
      numKeyColor: *purple
    # CrumbView attributes for history navigation.
    crumbs:
      fgColor: *foreground
      bgColor: *current_line
      activeColor: *current_line
    # Resource status and update styles
    status:
      newColor: *cyan
      modifyColor: *purple
      addColor: *green
      errorColor: *red
      highlightcolor: *orange
      killColor: *comment
      completedColor: *comment
    # Border title styles.
    title:
      fgColor: *foreground
      bgColor: *current_line
      highlightColor: *orange
      counterColor: *purple
      filterColor: *pink
  views:
    charts:
      bgColor: *background
      dialBgColor: "#0A2239"
      chartBgColor: "#0A2239"
      defaultDialColors:
        - "#1E3888"
        - "#820101"
      defaultChartColors:
        - "#1E3888"
        - "#820101"
      resourceColors:
        batch/v1/jobs:
          - "#5D737E"
          - "#820101"
        v1/persistentvolumes:
          - "#3E554A"
          - "#820101"
        cpu:
          - "#6EA4BF"
          - "#820101"
        mem:
          - "#17505B"
          - "#820101"
        v1/events:
          - "#073B3A"
          - "#820101"
        v1/pods:
          - "#487FFF"
          - "#820101"
    # TableView attributes.
    table:
      fgColor: *foreground
      bgColor: *background
      cursorColor: *current_line
      # Header row styles.
      header:
        fgColor: *foreground
        bgColor: *background
        sorterColor: *cyan
    # Xray view attributes.
    xray:
      fgColor: *foreground
      bgColor: *background
      cursorColor: *current_line
      graphicColor: *purple
      showIcons: true
    # YAML info styles.
    yaml:
      keyColor: *pink
      colonColor: *purple
      valueColor: *foreground
    # Logs styles.
    logs:
      fgColor: *foreground
      bgColor: *background
```

## Resolved Bugs/Features/PRs

- [Issue #557](https://github.com/derailed/k9s/issues/557)
- [Issue #555](https://github.com/derailed/k9s/issues/555)
- [Issue #554](https://github.com/derailed/k9s/issues/554)
- [Issue #553](https://github.com/derailed/k9s/issues/553)
- [Issue #552](https://github.com/derailed/k9s/issues/552)
- [Issue #551](https://github.com/derailed/k9s/issues/551)
- [Issue #549](https://github.com/derailed/k9s/issues/549) A start with pulses...
- [Issue #540](https://github.com/derailed/k9s/issues/540)
- [Issue #421](https://github.com/derailed/k9s/issues/421)
- [Issue #351](https://github.com/derailed/k9s/issues/351) Solved by Pulses?
- [Issue #25](https://github.com/derailed/k9s/issues/25) Pulses? Oldie but goodie!

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
