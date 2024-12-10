<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.5.0

## Notes

Thank you to all that contributed with flushing out issues with K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes!

If you've filed an issue please help me verify and close.

Thank you so much for your support and awesome suggestions to make K9s better!!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

I am super excited about this drop of K9s. Lots of cool improvements based on K9s friends excellent feedback!


### Popeye

Turns out [Popeye](https://github.com/derailed/popeye) is in too much flux at present, thus I've decided to remove it from K9s for the time being.

### ContainerView

Added a container view to list all the containers available on a given pod. On a selected pod, you can now press `<enter>` to view all of it's associated containers. Once in container view pressing `<enter>` on a selected container, will show the container logs.

### Resource Traversals

> Ever wanted to know where your pods originated from?

Fear not, K9s has got your back! Some folks have expressed desires to navigate from a deployment to its pods or see which pods are running on a given node. Whether you are starting from a Node, a Deployment, ReplicaSet, DaemonSet or StatefulSet, you can now simply `<enter>` of a selected item a view the associated pods. [Issue #149](https://github.com/derailed/k9s/issues/149)

### RollingBack ReplicaSets

You can now select a ReplicaSet and rollback your Deployment to that version. Enter the command mode via `:rs` to view ReplicaSets, select the replica you want to rollback to and use `Ctrl-B` to rollback your deployment to that revision.

---

## Resolved Bugs

+ [Issue #163](https://github.com/derailed/k9s/issues/163)
+ [Issue #162](https://github.com/derailed/k9s/issues/162)
+ [Issue #39](https://github.com/derailed/k9s/issues/39)
+ [Issue #27](https://github.com/derailed/k9s/issues/27)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
