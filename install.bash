#!/bin/bash -e

HERE=$(cd $(dirname "$0"); pwd)

GOVERSION=1.6

for step in ${HERE}/sysdeps/*; do
	$step
done

. ${HOME}/.bash_profile

gobrew versions | grep ${GOVERSION} || gobrew install ${GOVERSION}
gobrew use ${GOVERSION}

make -C ${HERE} all

SHIBBOLETH=e0f9a3ae6cf2f6470bfc002c4d7b40cae0fae49cb2b267b1de03ac6d5f45be75

grep ${SHIBBOLETH} ${HOME}/.bash_profile || cat >>~/.bash_profile <<EOF

# The following line marks this file as having been installed by
# github.com/cmars/tools.
# ${SHIBBOLETH}

export PATH=${HERE}/bin:\$PATH

if [ -f "\${HOME}/.bashrc" ]; then
	. \${HOME}/.bashrc
fi

if [ -d "\${HOME}/.bash_profile.d" ]; then
	for i in \${HOME}/.bash_profile.d/*.bash; do
		. \$i
	done
	unset i
fi
EOF
