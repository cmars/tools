# usso-login

Authenticate to [Ubuntu SSO](https://login.ubuntu.com) from the command-line.

This is useful for browserless, command-line interaction with APIs that use
Ubuntu SSO OAuth/OpenID for authentication. For example, Launchpad or Juju
Charms.

Consider this experimental and unsupported. Pull requests welcome.

# Build Dependencies

Developed and tested with Go 1.5.

# Build

## I'm feeling lucky

This might not work if any of the dependencies introduce breaking changes.

    $ go get github.com/cmars/usso-login

## With known-good versions of dependencies

    $ go get launchpad.net/godeps
    $ go get -d github.com/cmars/usso-login
    $ (cd $GOPATH; $GOPATH/bin/godeps -u dependencies.tsv)
    $ go install github.com/cmars/usso-login

# Typical Usage

## Log in and store an Ubuntu SSO cookie

    $ usso-login
    Email: azurediamond@hotmail.com
    Password: *******
    Two-Factor Auth: ******
    $ 

You'll be prompted for email, password, and possibly two-factor auth if you
have it set up. The cookie is then stored in `$HOME/.usso-cookies`.

## Non-interactive Use

If you're not using two-factor auth, you can log in non-interactively with
environment variables. This is useful when you don't have a terminal.

    $ export USSO_EMAIL=azurediamond@hotmail.com
    $  export USSO_PASSWORD=hunter2  # prefix with a space to keep it out of your .bash_history
    $ ... usso-login will use environment variables for these instead of prompting ...

## Using for command-line authentication

    $ export BROWSER=/path/to/usso-login
    $ charm publish /my/awesome/charm

...

    Email: azurediamond@hotmail.com
    Password: *******
    Two-Factor Auth: ******

...

# Known Issues

Reading input from terminal will fail if stdin is not hooked up to a terminal.
When using as a $BROWSER, in some cases the process is launched without stdin.
The best way to work around this is to log in ahead of time to store the
cookie.

Depends on the page contents and behavior of Ubuntu SSO. If that changes, this
will probably break.

# Debugging

    $ export USSO_DEBUG=1

Will print requests and responses to the terminal to help debug HTTP interactions.

---

Copyright 2015 Casey Marshall

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
