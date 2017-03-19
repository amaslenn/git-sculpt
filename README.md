[![Travis CI](https://travis-ci.org/amaslenn/git-sculpt.svg)](https://travis-ci.org/amaslenn/git-sculpt/builds) [![Code Climate](https://lima.codeclimate.com/github/amaslenn/git-sculpt/badges/gpa.svg)](https://lima.codeclimate.com/github/amaslenn/git-sculpt) [![Issue Count](https://lima.codeclimate.com/github/amaslenn/git-sculpt/badges/issue_count.svg)](https://lima.codeclimate.com/github/amaslenn/git-sculpt)

# git-sculpt
Tool for removing merged local branches. Extremely useful in case when changes are incorporated into mainline by `rebase`.

**Tool removes branch only if it thinks it is safe.**

## Quick installation
Go to the [latest release](https://github.com/amaslenn/git-sculpt/releases/latest) to download binaries.

## Examples
Remove single branch `feature1`:
```sh
git sculpt feature1
```

Remove branch `feature2` using **develop** as base:
```sh
git sculpt --base develop feature2
```

Remove all local branches (keeps all not safe for removal):
```sh
git sculpt --all
git sculpt --all -i		# will ask confirmation for removal
```

## Build
The following command will trigger cross platform build for `amd64` for macOS, Linux and Windows.
```sh
make
```

# How it works
Similar to what real `rebase` does: search for **merge base**, then calculate `patch-id` for all commits in _local_ branch and try to find all of them in the _base_ branch (`master` by default).

