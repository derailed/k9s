# Release v0.1.1

<br/>

---
## Notes

Thank you to all that contributed with flushing out issues with K9s! I'll try
to mark some of these issues as fixed. But if you don't mind grab the latest
rev and see if we're happier with some of the fixes!

<br/>

---
## Change Logs

+ Added config file to tracks K9s configuration ~/.k9s/config.yml
+ Change log file location to use Go tmp dir stdlib package.
  Check the log destination and config file location using
    ```shell
    k9s info
    ```
+ Removed 9 namespaces limitation by allowing user to manage namespaces using
  the namespace view or the dotfile configuration.
+ Updated keyboard navigation on log view. Up/Down, PageUp/PageDown
+ Added configuration to manage buffer size while viewing container logs
+ Added fail early countermeasures. Hopefully will help us figure out non starts??
+ Beefed up CLI arguments
+ Changed help command to just ?
+ Changed back command to just Esc
+ Added filtering feature to trim down viewed resources
  Use **/**term or **Esc** to kill filtering

<br/>

---
## Resolved Bugs

+ [Issue 17] Multi user log usage. Added user descriptor on log files
+ [Issue 18] Non starts due to color. Added preflight item on README.
+ [Issue 13] ? does not do anything.
+ [Issue 8] Don't reset selection after deletion.
+ [Issue 1,7] Limit available namespaces. Added config file to manage top 5 namespaces
  and also added a switch command while in the namespace resource view.
+ [Issue 6] Sorting/filtering. Added preliminary filtering capability. Raw search
  on table item using /filter_me command. Use Esc to turn off filtering.
+ [Issue 5] Scrolling in log view. Added up/down/pageUp/pageDown.
+ [Issue 3] No output when failing. Added fail early countermeasures. Hopefully
  will give us a heads up now to track down config issues??
