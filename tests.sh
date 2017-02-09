#!/usr/bin/env bash

rm -rf .tests
mkdir .tests
cd .tests
git init
touch > foo

# 1 commit
git add foo
git commit -m "add foo"

# branch 1. Safe to remove (empty)
git checkout -b br-empty master

# branch 2. Safe to remove (merged with ff)
git checkout -b br-one-commit master
echo err >> foo
git commit -am "change foo"
git checkout master
git merge --ff-only br-one-commit

# branch 3. Not safe o remove
git checkout -b br-new master
echo bar >> foo
git commit -am "change foo"

### tests
PATH=$PATH:../build/amd64/linux/
git checkout master

echo "t: empty branch"
git sculpt -d br-empty
if [ $? -eq 0 ]; then
    echo PASSED
else
    echo FAILED
fi

echo "t: merged branch"
git sculpt -d br-one-commit
if [ $? -eq 0 ]; then
    echo PASSED
else
    echo FAILED
fi

echo "t: branch with changes"
git sculpt -d br-new
if [ $? -eq 0 ]; then
    echo FAILED
else
    echo PASSED
fi

