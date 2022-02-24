# Contributing

## Help wanted

Besides the stuff on the README (features planned / under consideration):

- Windows support for static builds w/go-libtor
- golangci-lint to enforce some of the guidelines below
- User docs
- A fancy landing page for the project

## How to contribute

You'd like to propose a change into this project. That's great, much
appreciated!

A few things I ask of you to keep things running smoothly.

### Commits

This project uses [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)
to automate semantic versioning of releases on merge into the main branch.

You can read that link, but basically:
- Prefix feature commit titles with `feat: `
- Prefix bug fix commit titles with `fix: `
- Prefix stuff that shouldn't need a release with chore:

There's some stuff about breaking changes in there as well but honestly if
you're proposing something in that territory, you'll hear about it :)

I can't say I really like Conventional Commits. But do you know what I like even less?
Having to do releases myself! :)

### Code

Go code should generally follow the
[Go language code review guide](https://github.com/golang/go/wiki/CodeReviewComments)
unless there's a good reason not to (not common).

Adding new dependencies should be carefully considered.

Changes needed in dependencies should be contributed upstream if at all
possible. It takes a village.
