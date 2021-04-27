<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.8.1

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

### FuzzBuster!

So it looks like going all fuzzy was a mistake as we've lost some nice searchability feature with the regex counterpart. No worries tho Fuzzy is still around! The logic for searching will default to regex like all prior K9s version. To enable fuzzy logic, I figured we will use the same idea as we did with label filters using `/-lapp=bobo` but instead using `/-fpromset`

### Location, Location, Location!

There was a few issues related to screen `real estate` with K9s or more specifically the lack of it! Some folks flat out decided not to use K9s just because of the ASCII Logo ;( WTF! In this drop, I'd like to introduce a new presentation mode aka `Headless`.

Using the following command you can now run K9s headless:

```shell
k9s --headless # => Launch K9s without the header rows
```

NOTE! If you forgot your K9s shortcuts already, fear not! I've also updated the help menu so `?` will remind you of all the available options.

Lastly if you really dig the headless mode, you can sneak an extra `headless: true` in your ./k9s/config.yml like so:

```yaml
k9s:
  refreshRate: 2
  headless: false
  ...
```

### Menu Shortcuts

Some folks correctly pointed out the issue with the `Alt-XXX`. Totally my bad as my external mac keyboard unlike my MBP keyboard shows `option` and `alt` as the same key. So I've added a check to make sure the correct mnemonic is displayed based on you OS. Big Thanks for the call out to Ming, Eldad, Raman and Andrew!! Hopefully it did not hose the menu options in the process... 🙏

---

## Resolved Bugs/Features

+ [Issue #286](https://github.com/derailed/k9s/issues/286)
+ [Issue #285](https://github.com/derailed/k9s/issues/285)
+ [Issue #270](https://github.com/derailed/k9s/issues/270)
+ [Issue #223](https://github.com/derailed/k9s/issues/223)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
