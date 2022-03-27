#! /bin/bash

git checkout fat
git pull origin fat
git submodule init
git submodule update --remote

PROJECT="member"
GitReversion=`git rev-parse HEAD`
BuildTime=`date +'%Y.%m.%d.%H%M%S'`
BuildGoVersion=`go version`

go build -ldflags "-X main.gitReversion=${GitReversion}  -X 'main.buildTime=${BuildTime}' -X 'main.buildGoVersion=${BuildGoVersion}'" -o $PROJECT
# cg
sshpass -p $1 scp $PROJECT testp3@34.92.240.177:/home/centos/cg/$PROJECT

