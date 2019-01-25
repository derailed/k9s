# Contributing to tview

First of all, thank you for taking the time to contribute.

The following provides you with some guidance on how to contribute to this project. Mainly, it is meant to save us all some time so please read it, it's not long.

Please note that this document is work in progress so I might add to it in the future.

## Issues

- Please include enough information so everybody understands your request.
- Screenshots or code that illustrates your point always helps.
- It's fine to ask for help. But you should have checked out the [documentation](https://godoc.org/github.com/rivo/tview) first in any case.
- If you request a new feature, state your motivation and share a use case that you faced where you needed that new feature. It should be something that others will also need.

## Pull Requests

If you have a feature request, open an issue first before sending me a pull request. It may save you from writing code that will get rejected. If your case is strong, there is a good chance that I will add the feature for you.

I'm very picky about the code that goes into this repo. So if you violate any of the following guidelines, there is a good chance I won't merge your pull request.

- There must be a strong case for your additions/changes, such as:
  - Bug fixes
  - Features that are needed (see "Issues" above; state your motivation)
  - Improvements in stability or performance (if readability does not suffer)
- Your code must follow the structure of the existing code. Don't just patch something on. Try to understand how `tview` is currently designed and follow that design. Your code needs to be consistent with existing code.
- If you're adding code that increases the work required to maintain the project, you must be willing to take responsibility for that extra work. I will ask you to maintain your part of the code in the long run.
- Function/type/variable/constant names must be as descriptive as they are right now. Follow the conventions of the package.
- All functions/types/variables/constants, even private ones, must have comments in good English. These comments must be elaborate enough so that new users of the package understand them and can follow them. Provide examples if you have to.
- Your changes must not decrease the project's [Go Report](https://goreportcard.com/report/github.com/rivo/tview) rating.
- No breaking changes unless there is absolutely no other way.
