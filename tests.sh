#!/usr/bin/env bash

rm -rf .tests
mkdir .tests
cd .tests
git init
git config user.name "tester"
git config user.email "tester@test.test"
touch foo

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

# branches 4 and 5. Safe to remove
git branch b4 master
git branch b5 master

### tests
APP=../build/amd64/linux/git-sculpt
git checkout master

error=0
echo "t: empty branch"
$APP br-empty
if [ $? -eq 0 ]; then
    echo PASSED
else
    echo FAILED
    error=$((error+1))
fi

echo "t: merged branch"
$APP br-one-commit
if [ $? -eq 0 ]; then
    echo PASSED
else
    echo FAILED
    error=$((error+1))
fi

echo "t: branch with changes"
$APP br-new
if [ $? -ne 0 ]; then
    echo FAILED
    error=$((error+1))
else
	git checkout br-new && git checkout master
	if [ $? -eq 0 ]; then
    	echo PASSED
	else
		error=$((error+1))
	fi
fi

git branch -D br-new
echo "t: --all mode"
$APP --all
num_br=`git branch | wc -l`
if [ $num_br -eq 1 ]; then
	echo PASSED
else
	echo FAILED
	error=$((error+1))
fi

exit $error

