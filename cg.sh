#! /bin/bash

#git checkout main
#git pull origin main
#git submodule init
#git submodule update --remote

PROJECT="member"
GitReversion=`git rev-parse HEAD`
BuildTime=`date +'%Y.%m.%d.%H%M%S'`
BuildGoVersion=`go version`

go build -ldflags "-X main.gitReversion=${GitReversion}  -X 'main.buildTime=${BuildTime}' -X 'main.buildGoVersion=${BuildGoVersion}'" -o $PROJECT
# cg
scp -i /home/gocloud-yiy-rich -P 10087 $PROJECT p3test@34.92.240.177:/home/centos/workspace/cg/member/member_cg
ssh -i /home/gocloud-yiy-rich -p 10087 p3test@34.92.240.177 "sh /home/centos/workspace/cg/member/cg.sh"
