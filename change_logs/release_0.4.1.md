# Release v0.4.1

## Notes

Thank you to all that contributed with flushing out issues with K9s! I'll try
to mark some of these issues as fixed. But if you don't mind grab the latest
rev and see if we're happier with some of the fixes!

If you've filed an issue please help me verify and close.

Thank you so much for your support and awesome suggestions to make K9s better!!

---

## Change Logs

### Subject View

   You can now view users/groups that are bound by RBAC rules without having to type to full subject name.
   To activate use the following command mode

   ```text
   # for users
   :usr
   # for groups
   :grp
   ```

   These commands will pull all the available cluster and role binding associated with these subject types.
   You can then select and `<enter>` to see the associated policies.
   You can also filter/sort like in any other K9s views with the added bonus of auto updates when new
   users/groups binding come into your clusters.

   To see ServiceAccount RBAC policies, you can now navigate to the serviceaccount view aka `:sa` and press `<enter>`
   to view the associated policy rules.

### Fu View

  Has been renamed policy view to see all RBAC policies available on a subject.
  You can now use `pol` (instead of `fu`) to list out RBAC policies associated with a
  user/group or serviceaccount.

### Enter now has a meaning!

  Pressing `<enter>` on most resource views will now describe the resource by default.

---

## Resolved Bugs

+ [Issue #143](https://github.com/derailed/k9s/issues/143)
+ [Issue #140](https://github.com/derailed/k9s/issues/140)
  NOTE! Describe on v1 HPA is busted just like it is when running v 1.13 of
  kubectl against a v1.12 cluster.
