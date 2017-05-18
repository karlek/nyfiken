nyfiken
=======

Nyfiken means curious in Swedish. Nyfikend is a daemon which will periodically check for updates on a list of URLs and send a notification to the user when it happens. Nyfiken is client which interacts with the daemon.

Installation
------------
```fish
$ go install github.com/karlek/nyfiken/cmd/nyfikend
$ go install github.com/karlek/nyfiken/cmd/nyfiken
$ mkdir ~/.config/nyfiken  
$ mv $GOPATH/src/github.com/karlek/nyfiken/config.ini $GOPATH/src/github.com/karlek/nyfiken/pages.ini ~/.config/nyfiken
```

Security
--------

#### Warning: there exists some known security plausible scenarios.
If an attacker can modify a nyfiken pages file; nyfiken can be used to:

    - Perform all web-based attacks based on HTTP requests.
    - Scan the network for web-servers or routers and, via site-specific mail-setting, gain access to the information.

Nyfikend
--------
Nyfiken is a client which access the updated information from nyfikend. It can be used to force the program to check all pages again, clear all logged updates and to open them in a browser.

Nyfiken communicates on port `5239` by default.

Nyfiken Usage
--------------
```fish
$ nyfiken
Sorry, no updates :(
$ nyfiken -f
Pages will be checked immediately by your demand.
$ nyfiken
http://example.org/
http...
$ nyfiken -r
Opening all updates with: /usr/bin/browser
$ nyfiken -c
Updates list has been cleared!
```

API documentation
-----------------
http://go.pkgdoc.org/github.com/karlek/nyfiken

Public domain
-------------
I hereby release this code into the [public domain](https://creativecommons.org/publicdomain/zero/1.0/).
