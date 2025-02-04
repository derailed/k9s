# Release v0.1.2

<br/>

---
## Notes

Thank you to all that contributed with flushing out issues with K9s! I'll try
to mark some of these issues as fixed. But if you don't mind grab the latest
rev and see if we're happier with some of the fixes!

<br/>

---
## Change Logs

+ Navigation changed! Thanks to [Teppei Fukuda](https://github.com/knqyf263) for
  hinting about the different modes ie command vs navigation. Now in order to
  navigate to a specific kubernetes resource you need to issue this command
  to say see all pods (using key `>`):

    ```text
    >po<ENTER>
    ```
+ Similarly to filter on a given resource you can use `/` and type your filter.
+ In both instances `<ESC>` will back you out of command mode and into navigation mode.

<br/>

---
## Resolved Bugs

+ [Issue #23](https://github.com/derailed/k9s/issues/23)
+ [Issue #19](https://github.com/derailed/k9s/issues/19)
