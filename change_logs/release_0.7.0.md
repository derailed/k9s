<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.7.0

## Notes

Thank you to all that contributed with flushing out issues with K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes!

If you've filed an issue please help me verify and close.

Thank you so much for your support and awesome suggestions to make K9s better!!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

### Labor Day Weekend?

I've always seem to get this wrong. Does Labor Day weekend mean you get to work on you OSS projects all weekend? I am very excited about this drop and hopefully won't be hunamimous on that??

### Service Traversals

Provided your K8s service is head(Full), you can now navigate to the pods that match the service selector. So you will be able to traverse Pods/Containers directly from a service just like other resources like deployment, cron, sts...

### Moving Forward!

K9s supports port-forwarding. Provided a pod's container exposes a port, you can navigate to the container view, select a container with an exposed port and activate a port-forward directly from K9s without needing to shell out. I think that's very handy. This was indeed a long time coming... Big Thank You and ATTABoy!! to [Brent](https://github.com/brentco) for this great issue.

That said, these babies are a bit expensive to run, so make sure you choose wisely and delete any superflous forwards! To access the port-forward view directly use `:pf<enter>`.
BONUS: Pending your terminal of choice, you might even be able to pop the forwarded URL directly into your browse. On iTerm I think `command+click` does the trick.

This feature is very much still work in progress, thinks like basic auth, http verbs, headers, etc... are coming next, so please thread lightly and checkout the README under the Benchmarking section.

### Hey now!

This is one feature that I think, pardon my french `Bitch'n`. K9s now encompassed [Hey](https://github.com/rakyll/hey) from the totaly brilliant and kind Jaana Dogan of Google fame.
So along with the port-forward feature, you can now benchmark your containers and gather some interesting metrics that may help you configure resources, auto scalers, compare Canary builds, etc...

That said, this feature is still a moving target, as much configuration still needs to be tuned to make it totally killer. Please checkout the README on how to configure this feature. There are many more improvements that need to happen notably bench'ing service, ingress, etc and will come in subsequent K9s drops...

We think this port-forward/bench combo is a killer feature! Hopefully you'll agree. With the understanding the full-monty is coming soon, please help us mold and solidify these features as they are the base ingredients to setup for benchmarking services and ingresses...

> NOTE! Has with anything in life `Aim small, Miss small!`. You could totally hose K9s with over ambitious benchmarks, so start small say C:1 N:100, measure and go from there.
---

## Resolved Bugs

+ [Issue #198](https://github.com/derailed/k9s/issues/198)
+ [Issue #197](https://github.com/derailed/k9s/issues/197)
+ [Issue #195](https://github.com/derailed/k9s/issues/195) Thanks [Sebastiaan](https://github.com/tammert)!!
+ [Issue #194](https://github.com/derailed/k9s/issues/194)
+ [Issue #69](https://github.com/derailed/k9s/issues/69)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
