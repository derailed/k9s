<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.4.5

## Notes

Thank you to all that contributed with flushing out issues with K9s! I'll try
to mark some of these issues as fixed. But if you don't mind grab the latest
rev and see if we're happier with some of the fixes!

If you've filed an issue please help me verify and close.

Thank you so much for your support and awesome suggestions to make K9s better!!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

### Multi containers

  There was an [issue](https://github.com/derailed/k9s/issues/135) where we ran into limitations with the container
  selection keyboard shortcuts only allowing up to 10 containers. In this release, we've changed to a pick list vs the menu
  to select containers for both shell and logs access. This gives K9s the ability to select up to 26 containers now. This
  is not in any way an *encouragement* to have so many containers per pods!!

### Alias View ShortCut

  The change above entailed having to move the alias shortcut to `A` vs `a` as the pick list shortcuts conflicted with
  the alias view keyboard activation.


---

## Resolved Bugs

+ [Issue #152](https://github.com/derailed/k9s/issues/152)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
