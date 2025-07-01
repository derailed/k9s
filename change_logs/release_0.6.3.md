<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.6.3

## Notes

Thank you to all that contributed with flushing out issues with K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes!

If you've filed an issue please help me verify and close.

Thank you so much for your support and awesome suggestions to make K9s better!!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

### Performance... With feelings!

Ran thru another perf pass and hope I've pushed the needle in the right direction? K9s is now leveraging informers which I think came out of CRDs work. Our initial assessments shows numbers to μsecond updates, down from milliseconds 🎉. Hopefully the outputs are still correct as I have the tendency to make things much faster with incorrect results ;( We hope to hear back from you with a report from your clusters and assessments and brace for good news? This was a deep cycle thru K9s core and more perf will be gained, once we get a chance to vet this new strategy. I'd like to take this opportunity to thank you all for your patience and incredible kindness and support. We certainly hope this drop won't turn out to be a dud as I am fresh out of prozac patches 😩

---

## Resolved Bugs

+ [Issue #176](https://github.com/derailed/k9s/issues/171) Fingers crossed it's a better drop 🙏🐭?

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
