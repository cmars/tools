#!/bin/bash -xe

HERE=$(cd $(dirname "$0"); pwd)

${HERE}/install.bash
. ${HOME}/.bash_profile

sudo apt-get update -y
sudo apt-get install -y openjdk-7-jre-headless build-essential nacl-tools jq git bzr mercurial

sudo apt-add-repository -y ppa:ubuntu-lxc/lxd-stable
sudo apt-add-repository -y ppa:juju/devel
sudo apt-add-repository -y ppa:juju/stable
sudo apt-get update -y
sudo apt-get install -y lxd charm charm-tools

sudo gpasswd -a ubuntu lxd

export GOPATH=$HOME/gopath
go get -d -t github.com/juju/juju

pushd $GOPATH/src/github.com/juju/juju
git checkout juju-2.0-beta3
godeps -u dependencies.tsv
GOBIN=$HOME/tools/bin make install
popd

