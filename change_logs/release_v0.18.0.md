<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.18.0

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better is as ever very much noticed and appreciated!

Also if you dig this tool, please consider sponsoring 👆us or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

---

## GH Sponsors

Big `ThankYou` to the following folks that I've decided to dig in and give back!! 👏🙏🎊
Thank you for your gesture of kindness and for supporting K9s!!

* [Bob Johnson](https://github.com/bbobjohnson)
* [Poundex](https://github.com/Poundex)
* [thllxb](https://github.com/thllxb)

If you've contributed $25 or more please reach out to me on slack with your earth coordinates so I can send you your K9s swags!

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/shirts/k9s_front.png" align="right" width="200" height="auto"/>
<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/shirts/k9s_back.png" align="right" width="200" height="auto"/>

---

## AutoSuggestions

K9s command mode now provides for autocomplete. Suggestions are pulled from available kubernetes resources and custom aliases. The command mode supports the following commands:

| Key                 | Description                              |
|---------------------|------------------------------------------|
| ⬆️ ⬇️               | Navigate up or down thru the suggestions |
| `Ctrl-w`, `Ctrl-u`  | Clear out the command                    |
| `Tab`, `Ctrl-f`, ➡️ | Accept the suggestion                    |

## Logs Revisited

Breaking Change! This drop changes how logs are viewed and configured. The log view now support for pulling logs based on the log timeline current settings are: all, 1m, 5m, 15m and 1h. The following log configuration is in effect as of this drop:

```yaml
# $HOME/.k9s/config.yml
k9s:
  refreshRate: 2
  readOnly: false
  # NOTE: New logger configuration!
  logger:
    tail:          200 # Tail the last 100 lines. Default 100
    buffer:       5000 # Max number of lines displayed in the view. Default 1000
    sinceSeconds:  900 # Displays the last x seconds from the logs timeline. Default 5m
  ...
```

## Resolved Bugs/Features/PRs

* [Issue #628](https://github.com/derailed/k9s/issues/628)
* [Issue #623](https://github.com/derailed/k9s/issues/623)
* [Issue #622](https://github.com/derailed/k9s/issues/622)
* [Issue #565](https://github.com/derailed/k9s/issues/565)

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
