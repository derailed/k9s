<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.4.7

## Notes

Thank you to all that contributed with flushing out issues with K9s! I'll try
to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes!

If you've filed an issue please help me verify and close.

Thank you so much for your support and awesome suggestions to make K9s better!!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

### Popeye Support

Managing and operating a cluster is the wild is hard and getting harder.
I've created [Popeye](https://github.com/derailed/popeye) to help with cluster sanitation and best practices. Since K9s folks are so awesome, you're getting a sneak peek! I figured why not integrate it with K9s? No need to install yet another CLI right? Provided I did not mess this up too much, you should now be able to use command mod `:popeye` to access Popeye sanitizer reports within
K9s and scan your clusters. You can read more about it [here](https://medium.com/@fernand.galiana/k8s-clusters-oh-biff-em-popeye-637e9312963)
and if you like so give it a clap or two ;)

NOTE: In a K9s environment, if you'd like to specify a spinach config file, you must set it in your $HOME/.k9s/spinach.yml.

NOTE: There is a bit more that need to be done on this integration to be complete. Please file an issue if something does not work as expected.

NOTE: Popeye will run its own course and K9s is just a viewer for it, so if you'd like additional sanitation or find Popeye related issues, please tune to the corresponding repo!

Let us know if you dig it? And share your before/after clusters scores!

---

## Resolved Bugs

+ Great find! Thank you @swe-covis! Moved alias view to `Ctrl-A` [Issue #156](https://github.com/derailed/k9s/issues/156)
+ Added toggle autoscroll via `s` key [Issue #155](https://github.com/derailed/k9s/issues/155)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
