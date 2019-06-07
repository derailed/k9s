<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.7.6

## Notes

Thank you to all that contributed with flushing out issues with K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as always very much appreciated!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

### MultiLogs Initial Support

This is an experimental enhancement to allow to view logs for associated resources ie view logs for all containers in a pod or view container logs for pods fronted by a service, deployment, etc... directly in K9s. We've contemplated integrating the excellent `stern` CLI for this which is more full featured than the current implementation, but decided that shelling out for logs was at this juncture not ideal. Based on your feedback, we might revisit in future releases should this feature be a total dud...

### Delete Dialog

The resource delete dialog was updated to provide affordance for force and cascade deletes. This should now provide an on par behavior with the `kubectl` CLI. Cascade and force options are checkboxes, please use `<ENTER>` to toggle the flags.

---

## Resolved Bugs/Features

+ [Feature #193](https://github.com/derailed/k9s/issues/193)
+ [Issue #205](https://github.com/derailed/k9s/issues/205)
+ [Issue #212](https://github.com/derailed/k9s/issues/212)
+ [Issue #215](https://github.com/derailed/k9s/issues/215)
+ [Issue #220](https://github.com/derailed/k9s/issues/220)


---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
