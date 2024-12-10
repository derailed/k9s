<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>


# Release v0.4.1

## Notes

Thank you to all that contributed with flushing out issues with K9s! I'll try
to mark some of these issues as fixed. But if you don't mind grab the latest
rev and see if we're happier with some of the fixes!

If you've filed an issue please help me verify and close.

Thank you so much for your support and awesome suggestions to make K9s better!!

---

## Change Logs

### o Subject View

   You can now view users/groups that are bound by RBAC rules without having to type to full subject name.
   To activate use the following command mode

   ```text
   # For users
   :usr
   # For groups
   :grp
   ```

   These commands will pull all the available cluster and role bindings associated with these subject types.
   Use select + `<enter>` to see the associated RBAC policy rules.
   You can also filter/sort, like in any other K9s views with the added bonus of auto updates when new user/group bindings come into your clusters.

   To see ServiceAccount RBAC policies, you can navigate to the serviceaccount view aka `:sa` and select + `<enter>` to view the associated policy rules.

### o ~~FuView~~ is now PolicyView

  The Fu command has been deprecated for pol(icy) command to see all RBAC policies available on a subject. You can use `pol` (instead of `fu`) to list out RBAC policies associated with a
  user/group or serviceaccount.

  ```text
  # To list out all the RBAC policies associated with user `fernand`
  :pol u:fernand
  ```

### Enter. Yes Please!

  Pressing `<enter>` on most resource views will now describe the resource by default.

---

## Resolved Bugs

+ RBAC long subject names [Issue #143](https://github.com/derailed/k9s/issues/143)
+ Support HPA v1 [Issue #140](https://github.com/derailed/k9s/issues/140)
  > NOTE: Describe on v1 HPA is busted just like it is when running kubectl v1.13
  > against an older cluster.

---


<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
