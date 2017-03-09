[![Travis CI](https://travis-ci.org/amaslenn/git-sculpt.svg)](https://travis-ci.org/amaslenn/git-sculpt/builds) [![Code Climate](https://lima.codeclimate.com/github/amaslenn/git-sculpt/badges/gpa.svg)](https://lima.codeclimate.com/github/amaslenn/git-sculpt) [![Issue Count](https://lima.codeclimate.com/github/amaslenn/git-sculpt/badges/issue_count.svg)](https://lima.codeclimate.com/github/amaslenn/git-sculpt)

# git-sculpt
Tool for removing merged local branches. Extremely useful in case when changes are incorporated into mainline by `rebase`.

## Examples
Check if branch `feature1` in **master**:  
```sh
git sculpt feature1
```

Check if branch `feature2` in **develop**:
```sh
git sculpt --base develop feature2
```

Check if branch `feature3` in **master** and delete if it is safe:  
```sh
git sculpt -d feature3
```

## Build
The following command will trigger cross platform build for `amd64` for macOS, Linux and Windows.
```sh
make
```

## Install
Simply copy pre-build versions (or compile by yourself) to your `$PATH`. `git` will automatically use it for `git sculpt`.

# How it works
Similar to what real `rebase` does: search for **merge base**, then calculate `patch-id` for all commits in _local_ branch and try to find all of them in the _base_ branch (`master` by default).

