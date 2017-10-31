# Contribution Guidelines

## Tests

Code should be submitted with a reasonable suite of tests, and code should be 
designed with testability in mind.

## Commit Messages

Commit messages should explain **why** code is changing, configuration is added,
or new types or packages are introduced. A good commit message is always worth it. 

## Code Review

Code should be submitted as a single, or at most a few, squashed commit(s) via a 
merge/pull request. Only open a request with fully designed and tested solution 
that you would be comfortable merging. PRs are not for designing solutions, 
that's what issues are for.  

If code requires special local testing, provide a test plan in a PR comment (not 
the commit message or merge request description). Step by step instructions or
a script are ideal.

Overall, you should strive to make reviewing efficient for your colleagues.

## Changelog

The changelog is a valuable resource. It is maintained in **CHANGELOG.md**. Most
PRs should include edits to the changelog to describe bug fixes, API changes,
or new features added. Reading the changelog should tell a story of the things
that are changed, added, or fixed from release to release. 

## Style Guide

Functions should take as few parameters as possible. If many parameters are 
required, consider introducing a new type that logically groups the data.

Large blocks of commented out code should not be checked in.

Avoid the use of global variables. Prefer a dependency injection style that uses
a mix of interfaces and concrete types.

## Releases and Tagging

To create a release, do the following:

1. Pick a version number for the release. Our example below is **0.1.0**
2. Add that version number as a header in CHANGELOG.md and write a description
   of what's changed there. MR that update to CHANGELOG.md to **develop** branch. 
3. When the MR of changes is accepted, **one team member** does the following
   from the command line:

```
git fetch origin
git checkout master
git reset --hard origin/develop
git tag -a v0.1.0 -m 'version 0.1.0'
git push origin --tags
```

If you ever need to delete a tag, you can do this:

```
git tag -d v1.0.4
git push origin :refs/tags/v1.0.4
```

