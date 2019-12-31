# Release v0.3.1

## Notes

Thank you to all that contributed with flushing out issues with K9s! I'll try
to mark some of these issues as fixed. But if you don't mind grab the latest
rev and see if we're happier with some of the fixes!

If you've filed an issue please help me verify and close.

Thank you so much for your support!!

---

## Change Logs

1. Refactored a lot of code! So please watch for disturbence in the force!
1. Changed cronjob and job aliases names to `cj` and `jo` respectively
1. *JobView*: Added new columns
   1. Completions
   2. Containers
   3. Images
1. *NodeView* Added the following columns:
   1. Available CPU/Mem
   2. Capacity CPU/Mem
1. *NodeView* Added sort fields for cpu and mem

---

## Resolved Bugs

+ [Issue #133](https://github.com/derailed/k9s/issues/133)
+ [Issue #132](https://github.com/derailed/k9s/issues/132)
+ [Issue #129](https://github.com/derailed/k9s/issues/129) The easiest bug fix to date ;)
