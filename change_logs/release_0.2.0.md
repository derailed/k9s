# Release v0.2.0

## Notes

Thank you to all that contributed with flushing out issues with K9s! I'll try
to mark some of these issues as fixed. But if you don't mind grab the latest
rev and see if we're happier with some of the fixes!

If you've filed an issue please help me verify and close.

Thank you so much for your support!!

---

## Change Logs

+ [Feature #97](https://github.com/derailed/k9s/issues/97)
  Changed log view to now use kubectl logs shell command.
  There were some issues with the previous implementation with missing info and panics.
  NOTE! User must type Ctrl-C to exit the logs and navigate back to K9s
+ Reordered containers to show spec.containers first vs spec.initcontainers.
+ [Feature #29](https://github.com/derailed/k9s/issues/29)
  Side effect of #97 Log coloring if present, will now show in the terminal.

---

## Resolved Bugs

* [Issue #99](https://github.com/derailed/k9s/issues/99)
* [Issue #100](https://github.com/derailed/k9s/issues/100)