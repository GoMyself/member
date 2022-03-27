#! /bin/bash

git checkout fat
git pull origin fat
git submodule init
git submodule update --remote

PROJECT="member2"
GitReversion=`git rev-parse HEAD`
BuildTime=`date +'%Y.%m.%d.%H%M%S'`
BuildGoVersion=`go version`

go build -ldflags "-X main.gitReversion=${GitReversion}  -X 'main.buildTime=${BuildTime}' -X 'main.buildGoVersion=${BuildGoVersion}'" -o $PROJECT
# bigbet88
sshpass -p $1 scp $PROJECT root@192.168.80.194:/home/centos/workspace/agame/$PROJECT

