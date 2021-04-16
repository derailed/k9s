<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.24.7

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better are as ever very much noted and appreciated!

If you feel K9s is helping your Kubernetes journey, please consider joining our [sponsorship program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

## Maintenance Release!

## Disturbance In The Farce.. Windows!

Splendid! So I had to borrow my neighbors kids 20 pounds windows `gaming` laptop for this one ;( Recent K9s drops are looking less than optimal on windows due to dependencies changes.
I was able to narrow it down to named colors are no longer being respected on Windows platforms. I'll keep digging on this but if you find yourself in the situation where K9s is looking less than optimal on Windows, for the short term please either use a custom skin with hex colors or change the stock skin to use hex color values vs named colors. Thank you!

## There are some that call me... Alpha!

K9s is still and will remain an open source software. As such it is free and we will continue to maintain this repo!

That said in order to support our efforts, we've recently launched [K9sAlpha](https://k9salpha.io) which is a freemium version of K9s. K9sAlpha unlocks additional features and enhancements.

If you would like to support us, you can either join our github sponsors or purchase a K9sAlpha license. If you are an active member of our github sponsorship program, you are eligible for a free K9sAlpha license. Please reach out for your shoe-phone and contact us for your personalized license key.

<img src="https://k9salpha.io/assets/k9salpha-blue.png" align="center" width="300" height="auto"/>

---

## Resolved Issues

* [Issue #1067](https://github.com/derailed/k9s/issues/1067) Increase HPA target column display
* [Issue #1061](https://github.com/derailed/k9s/issues/1061) Container shell Windows (Don't do windows so please help verify!)
* [Issue #1060](https://github.com/derailed/k9s/issues/1060) Exception when setting container image

## Resolved PRs

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
