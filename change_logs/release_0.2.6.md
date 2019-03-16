# Release v0.2.6

## Notes

Thank you to all that contributed with flushing out issues with K9s! I'll try
to mark some of these issues as fixed. But if you don't mind grab the latest
rev and see if we're happier with some of the fixes!

If you've filed an issue please help me verify and close.

Thank you so much for your support!!

---

## Change Logs

1. Preliminary drop on sorting by resource columns
2. Add sort by namespace, name and age for all views
3. Add invert sort functionality on all sortable views
4. Add sort on pod views for metrics and most other columns
5. For all other views we will add custom sort on a per request basis


---

## Resolved Bugs

+ [Issue #117](https://github.com/derailed/k9s/issues/117)
  Was filtering out inactive ns which need to be there for all to see anyway!
