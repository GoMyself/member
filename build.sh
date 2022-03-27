#! /bin/bash

git checkout main
git pull origin main
git submodule init
git submodule update

PROJECT="member"
GitReversion=`git rev-parse HEAD`
BuildTime=`date +'%Y.%m.%d.%H%M%S'`
BuildGoVersion=`go version`

go build -ldflags "-X main.gitReversion=${GitReversion}  -X 'main.buildTime=${BuildTime}' -X 'main.buildGoVersion=${BuildGoVersion}'" -o $PROJECT
mv $PROJECT /opt/data/deploy/va/$PROJECT
# shellcheck disable=SC2164
cd /opt/data/deploy/va/$PROJECT

git pull
git commit -am "${GitReversion}"
git push

