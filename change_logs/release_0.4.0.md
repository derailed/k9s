# Release v0.4.0

## Notes

Thank you to all that contributed with flushing out issues with K9s! I'll try
to mark some of these issues as fixed. But if you don't mind grab the latest
rev and see if we're happier with some of the fixes!

If you've filed an issue please help me verify and close.

Thank you so much for your support and awesome suggestions to make K9s better!!

---

## Change Logs

> NOTE! Lots of changes here, please report any disturbances in the force. Thank you!

1. [Feature #82](https://github.com/derailed/k9s/issues/82)
   1. Added ability to view RBAC policies while in clusterrole or role view.
   2. The RBAC view will auto-refresh just like any K9s views hence showing live RBAC updates
   3. RBAC view supports standard K8s verbs ie get,list,deletecollection,watch,create,patch,update,delete.
   4. Any verbs not in this standard K8s verb list, will end up in the EXTRAS column.
   5. For non resource URLS, we map standard REST verbs to K8s verbs ie post=create patch=update, etc.
   6. Added initial sorts by name and group while in RBAC view.
   7. Usage: To activate, enter command mode via `:cr` or `:ro` for clusterrole(cr)/role(ro), select a row and press `<enter>`
   8. To bail out of the view and return to previous use `p` or `<esc>`
2. One feature that was mentioned in the comments for the RBAC feature above Tx [faheem-cliqz](https://github.com/faheem-cliqz)! was the ability to check RBAC rules for a given user. Namely reverse RBAC lookup
   1. Added a new view, code name *Fu* view to show all the clusterroles/roles associated with a given user.
   2. The view also supports for checking RBAC Fu for a user, a group or an app via a serviceaccount.
   3. To activate: Enter command mode via `:fu` followed by u|g|s:subject + `<enter>`.
      For example: To view user *fred* Fu enter `:fu u:fred` + `<enter>` will show all clusterroles/roles and verbs associated
      with the user *fred*
   4. For group Fu lookup, use the same command as above and substitute `u:fred` with `g:fred`
   5. For ServiceAccount *fred* Fu check: use `s:fred`
3. Eliminated jitter while scrolling tables


---

## Resolved Bugs

+ None
