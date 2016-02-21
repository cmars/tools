#!/bin/bash -e

GOVERSION=1.6

for step in sysdeps/*; do
	$step
done

. ${HOME}/.bash_profile

gobrew versions | grep ${GOVERSION} || gobrew install ${GOVERSION}
gobrew use ${GOVERSION}

make all

grep TOOLS ${HOME}/.bash_profile || cat >>~/.bash_profile <<EOF
export PATH=$(pwd)/bin:\$PATH
EOF
