<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9s_small.png" align="right" width="200" height="auto"/>

# Release v0.24.2

## Notes

Thank you to all that contributed with flushing out issues and enhancements for K9s! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make K9s better are as ever very much noted and appreciated!

If you feel K9s is helping your Kubernetes journey, please consider joining our [sponsorship program](https://github.com/sponsors/derailed) and/or make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

On Slack? Please join us [K9slackers](https://join.slack.com/t/k9sers/shared_invite/enQtOTA5MDEyNzI5MTU0LWQ1ZGI3MzliYzZhZWEyNzYxYzA3NjE0YTk1YmFmNzViZjIyNzhkZGI0MmJjYzhlNjdlMGJhYzE2ZGU1NjkyNTM)

## ♫ Sounds Behind The Release ♭

* [ZZ Top - My Head's in Mississippi](https://www.youtube.com/watch?v=Gp2PosHepzg)

## A Word From Our Sponsors...

I would like to extend a `Big Thank You` to the following generous K9s friends for joining our sponsorship program and supporting this project!

* [Tim Orling](https://github.com/moto-timo)
* [Jiri Valnoha](https://github.com/waldauf)
* [Osx2000](https://github.com/osx2000)

## Our Release Heroes

Major ATTA BOY/GIRL! in full effect this week to the good folks below for their efforts and contributions in making sure K9s is better for all of us!

* [Ainslie Hsu](https://github.com/ainslie-hsu)
* [Lucas Teligioridis](https://github.com/lucasteligioridis)
* [Gergely Tankovics](https://github.com/gtankovics)
* [Michal Kuratczyk](https://github.com/mkuratczyk)
* [Simon Caron](https://github.com/simoncaron)

## She Can't Take Much More Capt'n!!

### Background

Thanks to all of you for supporting K9s and being avid fans. I am truly humbled and amazed by your continued kindness and support!! As we're nearing K9s second anniversary, the project has reached over 10k stars and 384k downloads! That said, while these numbers sound stunning, there is another number on this project that is not so and that is number of sponsors 😿.
As I understand it, there are a several organizations leveraging K9s productivity to better their bottom line, without much care for ours...
As you all know, K9s is a complex tool in a continually evolving space and we find ourselves spending a lot of our free time, thinking, experimenting and supporting K9s to continually improve the offering. As it stands, there is currently a very small fraction of you that actively sponsor this project either financially or by filing issues/PRs while the rest are benefiting from these efforts. This just does not sound like a fair deal and if we were in the music business it would be a total outrage!

### There Are Some That Call Me... Alpha!

To this end, I'd like to introduce a new member of the K9s pack, the main dog, aka `k9sAlpha`. This is going to be a licensed version of K9s. The current plan is to offer a tiered license scheme starting at `$10/month` for a license. K9s𝞪 will provide fixes, enhancements, further integrations and a bunch of new features that have been sitting in the back burner...

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/k9salpha.png" align="center" width="300" height="auto"/>

### So what does this entail?

1. The current k9s branch will be in feature freeze
1. K9s𝞪 users will need to purchase a license from our store
1. Active sponsors get a K9s𝞪 license
1. Documentation, binaries, issue trackers, will be provisioned under a new K9s𝞪 site

Given any license schemes are meant to be hacked/broken, we're not going to over complicate things with calling out to license servers and such to ensure the keys are legit.
The current plan is to email out your license keys and trusting our `Gentlemen Agreement` that you will not share or distribute your keys to other folks.
In the current economic climate, if you can't afford a K9s𝞪 license, we will provide you one on a case by case basis.

The process should be simple:

1. Acquire a license
1. Get a key via email
1. Store your key somewhere on disk
1. Download the K9s𝞪 binary
1. Administer your Kubernetes clusters with K9s𝞪
1. Rinse and repeat when your license expires

### K9s𝞪 Needs You!

To this end, I'd like to enlist a few of you to help me validate license keys, K9s𝞪 store and site to ensure the flow well... flows!
If you are so inclined, please reach out for your `shoephones` and send me an email with why you want to participate. Folks with K9s chops in multi clusters env would be preferred.
It should not take too much of your time to ensure all is cool, but want to make sure I have at least another 5 pairs of eyes to help out with the K9s𝞪 drop.
My hope is to get an initial K9s𝞪 revision dropped before Santa comes around...

### Pipe In!

By all means, this is a democracy and not a dictatorship! So... if you have better/other ideas or concerns please pipe in! Open an issue on the repo so we can track, discuss, opiniate and figure out the best course of action that will be fair to both K9s maintainers and users alike.

---

## Resolved Issues/Features

* [Issue #972](https://github.com/derailed/k9s/issues/972) Default color is no longer transparent.
* [Issue #933](https://github.com/derailed/k9s/issues/933) Unable to cordon node.

## Resolved PRs

* [PR #982](https://github.com/derailed/k9s/pull/982) Fix typo
* [PR #976](https://github.com/derailed/k9s/pull/976) Add OneDark color theme
* [PR #975](https://github.com/derailed/k9s/pull/982) Handling non json lines as raw with red color
* [PR #968](https://github.com/dserailed/k9s/pull/968) Disable filtering on help screen ... and broke the build ;)
* [PR #960](https://github.com/derailed/k9s/pull/960) Handle empty port list in PortForward view

---

<img src="https://raw.githubusercontent.com/derailed/k9s/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
