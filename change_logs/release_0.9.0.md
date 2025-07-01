<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.9.0

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

A lots of changes here in 0.9.0!! Please watch out for potential disturbance in the force as much code changed on this drop...

Figured, I'll put a quick video out for you to checkout the latest [K9s V0.9.0](https://www.youtube.com/watch?v=bxKfqumjW4I)

### Support K8s 1.16

As you might have heard K8s had a big drop with 1.16 so we've added client/server support this new kubernetes release.

### Alias Alas!

K9s now supports standard kubernetes short name. Major shoutout to [Gustavo](https://github.com/paivagustavo) for making this painful change happen!
With this change is place you can now use all standard K8s short names along with defining your own. You can now define a new alias file aka `alias.yml` in your k9s home directory `$HOME/.k9s`. An alias is made up of a command and a group/version/resource aka GVR specification as follows:

```yaml
alias:
  fred: apps/v1/deployments # Typing fred while in command mode will list out deployments
  pp: v1/pods               # Typing pp while in command mode will list out pods
```

### Plug For Plugins

As of this release and based on some users feedback we've moved the plugin section that used to live in the main K9s configuration file out to its own file. So as of this release we've added a new file in K9s home dir called `plugin.yml`. This is where you will define/share your K9s plugins and define your own commands and menu mnemonics. Here is an example for defining a custom command to show logs.

```yaml
# plugin.yml
plugin:
  fred:
    shortCut: Ctrl-L
    description: "Pod logs"
    scopes:
    - po
    command: /usr/local/bin/kubectl
    background: false
    args:
    - logs
    - -f
    - $NAME
    - -n
    - $NAMESPACE
    - --context
    - $CONTEXT
```

Special K9s env vars you will have access to are currently for your commands or shell scripts are as follows:

* NAMESPACE
* NAME
* CLUSTER
* CONTEXT
* USER
* GROUPS
* COL[0-9+]

I will setup a plugin/alias repo so we can share these with all K9sers. Please ping me if interested in contributing/sharing your commands. Thank you!!

### Aye Aye Capt'ain!!

Hopefully improved overall navigation...

#### Real Estate

This release allows you to maximize screen real estate via 2 combos. First, the command/filter prompt is now hidden. To enter commands or filters you can type `:` or `/` to type your commands. Second, you can toggle the header using `CTRL-H`.

#### Bett'a ShortCuts

You can now use commands like `svc fred` while in command mode to directly navigate to a resource in a given namespace. Likewise to switch contexts you can now enter `ctx blee` to switch out clusters.

#### Sticky Filters

You can now keep filters sticky allowing you to filter a view bases on regex, fuzzy or labels and keep the filter live while switching resources. This provides for a horizontal navigation to view the various resources for a given application. Thank you so much [Nobert](https://github.com/ncsibra) for your continuous awesome feedback!!

### New Resources

Added support for StorageClass, you can now view this resource and describe it directly in K9s. Major shoutout to [Oscar F](https://github.com/fridokus), zero go chops and yet managed to push this PR thru with minimal support. You Sir, blew me away. Thank you!!

---

## Resolved Bugs/Features

* [Issue #318](https://github.com/derailed/k9s/issues/318)
* [Issue #303](https://github.com/derailed/k9s/issues/303)
* [Issue #301](https://github.com/derailed/k9s/issues/301)
* [Issue #300](https://github.com/derailed/k9s/issues/300)
* [Issue #276](https://github.com/derailed/k9s/issues/276)
* [Issue #268](https://github.com/derailed/k9s/issues/268)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
