#! /bin/bash

git checkout main
git pull origin main
git submodule init
git submodule update

DIR=$(pwd)
PROJECT="member"
GitReversion=`git rev-parse HEAD`
BuildTime=`date +'%Y.%m.%d.%H%M%S'`
BuildGoVersion=`go version`

go build -ldflags "-X main.gitReversion=${GitReversion}  -X 'main.buildTime=${BuildTime}' -X 'main.buildGoVersion=${BuildGoVersion}'" -o $PROJECT
# shellcheck disable=SC2164
cd /opt/deploy/cg/$PROJECT
git checkout master
git pull origin master
mv $DIR/$PROJECT /opt/deploy/cg/$PROJECT
git commit -am "${GitReversion}"
git push

