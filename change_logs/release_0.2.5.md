# Release v0.2.5

## Notes

Thank you to all that contributed with flushing out issues with K9s! I'll try
to mark some of these issues as fixed. But if you don't mind grab the latest
rev and see if we're happier with some of the fixes!

If you've filed an issue please help me verify and close.

Thank you so much for your support!!

---

## Change Logs

+ Added an actual help view to show available key bindings. Use `<?>` to access it.
+ Changed alias view to now be accessible via key `<a>`
+ Pressing `<enter>` while on the namespace/context views will navigate directly to the pods view.
+ Added resource view breadcrumbs to easily navigate back in history. Use key `<p>` to navigate back.
+ Added configuration `logBufferSize` to limit the size of the log view while viewing chatty or big logs.

---

## Resolved Bugs

+ [Issue #116](https://github.com/derailed/k9s/issues/116)
+ [Issue #113](https://github.com/derailed/k9s/issues/113)
+ [Issue #111](https://github.com/derailed/k9s/issues/111)
+ [Issue #110](https://github.com/derailed/k9s/issues/110)