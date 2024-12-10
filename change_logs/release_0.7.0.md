<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.7.0

## Notes

Thank you to all that contributed with flushing out issues with K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as always very much appreciated!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

### Labor Day Weekend?

I always seem to get this wrong... Does Labor Day weekend mean you get to work on your OSS projects all weekend?

I am very excited about this drop and hopefully won't be unanimous (?) on this? 🐭

For the impatient watch this! [K9s v0.7.0 Features](https://youtu.be/83jYehwlql8)

### Service Traversals

Provided your K8s services are head(Full), you can now navigate to the pods that match the service selector. So you will be able to traverse Pods/Containers directly from a service just like other resources like deployment, cron, sts...

### Moving Forward!

In this drop, we've added support for port-forwarding that allows you to exercise your container from your local machine. To setup a port-forward, from the Pod view drill down to a selected Pod's containers, select the container that exposes the port of interest and hit `Ctrl-F`. A dialog will popup allowing you to configure a localhost port to forward to. Once set up, K9s will take you to the new PortForward view aka `pf`. Pending your terminal feature and container setup, you should be able to pop the forwarded URL directly into your browse. On iTerm2 me think `command+click` does the trick?

Big thanks and ATTABOY! in full effect all week to [Brent](https://github.com/brentco) for filing this initial issue. Please keep in mind, these port-forward babies are a bit expensive to run, so make sure you choose wisely and delete any superfluous PFs!!

This feature is still work in progress. It does require a new config file to help assist with URL configurations. As it stands, your PortForwards are in effect for the current K9s session and will be terminated on exit. Please thread lightly and checkout the README under the Benchmarking section. Your feedback on this as always, is welcome and encouraged!

### Hey now!

This is one feature that I think is, pardon my french.., totally `Bitch'n`! K9s now bundles [Hey](https://github.com/rakyll/hey) CLI tool from the extremely talented Jaana Dogan of Google fame. Hey allows you to benchmark HTTP service endpoints similar to apache bench.

Benchmarking is enabled via keyboard shortcuts `Ctrl-B` and `Alt-B` to activate/cancel a benchmark while in PortForward and Service view. Benchmarking a service assumes the HTTP service is exposed either as NodePort or LoadBalancer. To view your benchmarks, navigate to the new Benchmark view aka `:be<ENTER>` to list your benchmarks and runs statistics.

So you now have the ability to stretch out your cluster legs by benchmarking your pods and services and gather all kind of interesting statistics directly in K9s. Generating load on your resources will help you tune your cluster resources, exercise your auto scalers, compare Canary builds perf, etc...

Please keep in mind, this is very much a moving target at this point and will change. Ingress support will come next once we solidify the existing feature. Also checkout the README for additional configuration for this feature. With the understanding the Full Monty is coming, please help us solidify these features as these are the base ingredients to even cooler things coming down the line...

> NOTE! As with anything in life `Aim small, Miss small!`. You could totally overwhelm K9s with over-zealous benchmarks and port-forwards, so please start small say C:1 N:1000, measure and go from there.

---

## Resolved Bugs/Features

+ [Issue #198](https://github.com/derailed/k9s/issues/198)
+ [Issue #197](https://github.com/derailed/k9s/issues/197)
+ [Issue #195](https://github.com/derailed/k9s/issues/195) Thanks to the awesome [Sebastiaan](https://github.com/tammert). You Rock Sir!!
+ [Issue #194](https://github.com/derailed/k9s/issues/194)
+ [Issue #187](https://github.com/derailed/k9s/issues/187)
+ [Issue #119](https://github.com/derailed/k9s/issues/119) Added `Ctrl-S` shortcut to dump table data as csv and log data as text.
+ [Issue #69](https://github.com/derailed/k9s/issues/69)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
