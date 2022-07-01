#! /bin/bash

git checkout test
git pull origin test
git submodule init
git submodule update --remote

PROJECT="member"
GitReversion=`git rev-parse HEAD`
BuildTime=`date +'%Y.%m.%d.%H%M%S'`
BuildGoVersion=`go version`

go build -ldflags "-X main.gitReversion=${GitReversion}  -X 'main.buildTime=${BuildTime}' -X 'main.buildGoVersion=${BuildGoVersion}'" -o $PROJECT
# cg
scp -i /opt/data/p3test -P 10087 $PROJECT p3test@34.92.240.177:/home/centos/workspace/cg/$PROJECT
