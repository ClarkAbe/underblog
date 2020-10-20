#!/bin/bash

GitUrl="your git url"

rm -rf ./.tmp
mkdir ./.tmp
cp -r ./public/* ./.tmp
cd ./.tmp
git init
git add --all
git commit -m "up commit"
git remote add origin $GitUrl
git push -f origin master
rm -rf ./.tmp/.git