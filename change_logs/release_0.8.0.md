<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.8.0

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

Pretty excited about this drop! I am as ever humbled by all the cool comments and suggestions you guys are coming up with.
There are a few features that were requested that are simply excellent! Thank you all for your support, feedback and observations 👏

Now that said, some features might be more or less baked, so there might be some disturbance in the force with this drop since much code churned. So please file issues or PRs 🥰 if you notice anything that no longer works as expected.

### Client Update

In the mist of the next Kubernetes 1.16 drop, deprecating some old apis, we've decided to update K9s to support 1.15.1 client. We don't forsee any issues here but please make sure all is cool with this K9s drop on your clusters. If not please let us know so we can address. Thank you!!

### Scaling Pods

This was feature #12 filed by [Tyler Lewis](https://github.com/alairock) many moons ago. So big thanks to Tyler!! To be honest I was on the fence with this feature as I am not a big fan of one offs when it comes to cluster management. However I think it's a great way to validate adequate HPA settings while putting your cluster under load and use K9s to figure out what reasonable number of pods might be. Now this feature was not my own implementation so all kudos on this one goes to [Nathan Piper](https://github.com/nathanpiper) for spending the time to make this a reality for all of us. So many thanks to you Nathan!!
By Nathan's implementation you can now leverage the `s` shortcut for scale deployments, replication controllers and statefulsets. Very cool!

### FuzzBuzz!

Another enhancement request came this time from [Arthur Koziel](https://github.com/arthurk) and I think you guys will dig this one. So big thanks to Arthur for this report!! K9s now leverages a fuzzy finder to be able to search for resources. Previous implementation just used regex to locate matches. For example with this addition you can now type `promse` while in search mode `/` to locate all prometheus-server-5d5f6db7cc-XXX pods. That's so cool! Once this implementation is vetted, we will enable fuzzy searching on other views as well.

### ClipBoarding

This feature comes out of [Raman Gupta](https://github.com/rocketraman) report. Thank you Raman!! This allows a K9s operator to now just hit `c` while on a resource table view to copy the currently selected resource name to the clipboard. This allows you to navigate between K9s and other tools to search, grep/etc.. thru the currently selected resource. We may want to improve on this some but the basic implementation is now available.

### OldiesButGoodies?

So the initial few releases of K9s did not have any failsafe counter measures while deleting resources. So we've beefed the deletion logic to make sure you did not inadvertently blow something away by leveraging
dialogs. This was totally a reasonable thing to do! However in case of managed pods, one may want to quickly cycle on or more pod perhaps to pickup a new image or configuration. For this purpose we've introduced an alternate deletion mechanism to delete pod under `alt-k` for kill. Thanks to my fellow frenchma [ftorto](https://github.com/ftorto) for this one ;)

### HairPlugs!

This one is cool! I think this thought came about from (Markus)[https://github.com/Makusi75]. Thank you Markus!! This feature allows K9s users to now customize K9s with their own plugin commands. You will be able to add a new menu shortcut to the K9s menu and fire off a custom command on a selected resource. Some of you might be leveraging kubectl plugins and now you will be able to fire these off directly from K9s along with many other shell commands.

In order to specify a custom plugin command, you will need to modify your .k9s/config.yml file. For example here is a sample extension to list out all the pods in the `fred` namespace while in a pod or deployment view. When this plugin is available a new command `<alt-p>` will show only while in pod and deploy view.

```yaml
plugins:
  cmd1:
    # The menu mnemonic to trigger the command. Valid values are [a-z], Shift-[A-Z], Ctrl-[A-Z] or Alt-[A-Z]
    # Note! Mind the cases!!!
    shortCut: Alt-P
    scopes: # View names are typically matching the resource shortname ie po for pod, deploy for deployment, svc for service etc... If no shortname is available use the resource name.
    - po
    - deploy
    description: ViewPods # => Name to show on K9s menu
    command: kubectl      # => The binary to use. Must be on your $PATH.
    # Arguments on per line preceded with a dash! This will run > kubectl get pods -n fred
    args:
    - get
    - pods
    - -n
    - fred
```

Ok so this is pretty cool but what if I want to run a command to leverage the current pod name, namespace, container or other? You bet! Here is a more elaborated example. Say per Markus's report, I want to run my ksniff kubectl plugin from within K9s. So now I can hit `S` while in container view with a selected pod and sniff out incoming traffic. Here is an example plugin config for this.

```yaml
plugins:
  ksniff:
    # Enable `S` on the K9s menu while in container view
    shortCut: Shift-S
    scopes:
    - co
    description: Sniff
    # NOTE! Ksniff has been installed as a kubectl extension!
    command: kubectl
    # Run this command in the background so that I can still do K9s stuff...
    background: true
    args:
    - sniff
    # Use a K9s env var to extract the pod name from the current view.
    - $POD
    - -n
    # Use K9s current namespace
    - $NAMESPACE
    # Oh and pick out the container name from column 0 on that table. Nice!!
    - -c
    - $COL-0 # Use $COL-[0-9] to pick up the value from the desired resource table column.
```

NOTE: This is experimental and the schema/behavior WILL change in the future, so please thread lightly!

### That's a wrap!

We hope you will find some of these features useful on your day to day work with K9s. We know they are now more vendors coming into this space. Hence more choices for you to assess which of these tools makes you most happy and productive. My goal is to continue improving, speeding up and stabilizing K9s. My fuel is to see folks using it, file reports, contribute and seeing that occasional ATTA BOY! (which I must say is much more rewarding to me than money or fame...).

Many thanks to all of you for your time, ideas, contributions and support!!

---

## Resolved Bugs/Features

+ [Issue #274](https://github.com/derailed/k9s/issues/274)
+ [Issue #273](https://github.com/derailed/k9s/issues/273)
+ [Issue #272](https://github.com/derailed/k9s/issues/272)
+ [Issue #271](https://github.com/derailed/k9s/issues/271)
+ [Issue #267](https://github.com/derailed/k9s/issues/267)
+ [Issue #247](https://github.com/derailed/k9s/issues/247)
+ [Issue #203](https://github.com/derailed/k9s/issues/203)
+ [Issue #12](https://github.com/derailed/k9s/issues/12) Thank you Nathan!!

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
