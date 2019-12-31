<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.7.1

## Notes

Thank you to all that contributed with flushing out issues with K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as always very much appreciated!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

### AfterMath

Looks like I've broken some stuff in the excitement of 0.7.0! As I ran thru the video this am, I noticed the last minute screen dumps might not be a viable feature. As [Norbert](https://github.com/ncsibra) correctly points out, in issue #187 (Thanks Norbert!!), retrieving screen dumps was a pain. So I've put together a quick ScreenDump view alias `sd` to view the screen snapshots and allows to pop your editor of choice upon selection to view the screen dump file content.

NOTE: You will need to use an EDITOR env var to tell K9s which editor you want to use.

```shell
# For example...
export EDITOR=vim
```

This is a quick turn around, hopefully I did not hose anything else in the process ;(

---

## Resolved Bugs/Features

+ [Issue #200](https://github.com/derailed/k9s/issues/200)
+ [Issue #187](https://github.com/derailed/k9s/issues/187)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
