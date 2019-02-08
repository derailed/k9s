# Release v0.1.3

<br/>

---
## Notes

Thank you to all that contributed with flushing out issues with K9s! I'll try
to mark some of these issues as fixed. But if you don't mind grab the latest
rev and see if we're happier with some of the fixes!

If you've file an issue please help me verify and close.

Thank you so much for your support!!

<br/>

---
## Change Logs

<br/>

+ IMPORTANT: Changed HotKeys to single chars for most non destructive operations
  For **command** mode use the <:> key
  For **search** mode use the </> key
+ Revert Delete to Ctrl-D. (Sorry for the brain fart on this!)
+ IMPORTANT! Breaking change! The K9s config has changed to handle multi-clusters.
  If K9s does not launch, please move over .k9s/config.yml.
+ Added Resource for ReplicaController
+ Added auth support for cloud provider using the same auth options as kubectl

---
## Resolved Bugs

+ [Issue #50](https://github.com/derailed/k9s/issues/50)
+ [Issue #44](https://github.com/derailed/k9s/issues/44)
+ [Issue #42](https://github.com/derailed/k9s/issues/42)
+ [Issue #38](https://github.com/derailed/k9s/issues/38)
+ [Issue #36](https://github.com/derailed/k9s/issues/36)
+ [Issue #28](https://github.com/derailed/k9s/issues/28)
+ [Issue #24](https://github.com/derailed/k9s/issues/24)
+ [Issue #24](https://github.com/derailed/k9s/issues/3)
