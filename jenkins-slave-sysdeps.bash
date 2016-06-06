#!/bin/bash -xe

HERE=$(cd $(dirname "$0"); pwd)

${HERE}/install.bash
. ${HOME}/.bash_profile

sudo apt-get update -y
sudo apt-get install -y default-jre-headless build-essential nacl-tools jq git bzr mercurial

sudo apt-add-repository -y ppa:ubuntu-lxc/lxd-stable
sudo apt-add-repository -y ppa:juju/devel
sudo apt-add-repository -y ppa:juju/stable
sudo apt-get update -y
sudo apt-get install -y lxd charm charm-tools juju

sudo gpasswd -a ubuntu lxd
