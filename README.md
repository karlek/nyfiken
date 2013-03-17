nyfiken
=======
Nyfiken means curious in Swedish. Nyfikend is a daemon which will periodically check for updates on a list of URLs and send a notification to the user when it happens. Nyfikenc is client which interacts with the daemon.

Installation
------------
    go install github.com/karlek/nyfiken/cmd/nyfikend
    mkdir ~/.config/nyfiken
	mv $GOPATH/src/github.com/karlek/nyfiken/config.ini $GOPATH/src/github.com/karlek/nyfiken/pages.ini ~/.config/nyfiken

Security
--------
#### Warning: there exists some known security plausible scenarios.
If an attacker can modify a nyfiken pages file; nyfiken can be used to:

    - Perform all web-based attacks based on HTTP requests.
    - Scan the network for web-servers or routers and, via site-specific mail-setting, gain access to the information.

Nyfiken(c/d)
------------
Nyfikenc is a client to access the updated information from nyfikend. It can be used to force the program to check all pages again, clear all logged updates and to open them in a browser.

Nyfiken(c/d) communicates on port `5239` by default.

Nyfikenc Usage
--------------
	$ nyfikenc
	Sorry, no updates :(
	$ nyfikenc -f
	Pages will be checked immediately by your demand.
	$ nyfikenc
	http://example.com/
	http...
	$ nyfikenc -r
	Opening all updates with: /usr/bin/...
	$ nyfikenc -c
	Updates list has been cleared!

Config
------
Settings are defined in either global or site specific scope. Site specific settings will always overwrite the global or default and global will overwrite default settings.

    Default values
    --------------
    // Default interval between updates unless overwritten in config file.
    DefaultInterval = 1 * time.Minute

	// Default permissions to create files: user read and write permissions.
	DefaultFilePerms = os.FileMode(0600)

	// Default newline character
	Newline = "\n"

	// Default port number for nyfikenc/d connection.
	DefaultPortNum = ":5239"

INI
---
The config and pages file is in the INI format.
Sections are case-sensitive.

The config.ini file has two types of sections.

	[settings]
	Contains global settings for the program.

	[mail]
	Contains mail information to send and receive updates.

The pages.ini can have several sections in this format:

	[http://pagename.tld]
    Any number of pages to check when they update. Settings under this section will be referenced as site-specific.

Settings
--------
The settings available are:

[G] - Definable in global scope

[S] - Site specific settings

[G] - Fileperms
---------------
Change what permissions files are created with. Default value is 0600 - user read and write.
Fileperms is parsed as int.

    fileperms = 0777

[G][S] - Interval
-----------------
The time delay before making another request to check if it's been updated.
Interval is parsed by time.ParseDuration: [time#ParseDuration](http://tip.golang.org/pkg/time/#ParseDuration).

    interval = 1m

[G][S] - Strip
--------------
If the page is falsely notified as updated, you can use stripping functions to remove certain parts of the site which triggers these false-positives.

Whitelist

    - numbers
    - attrs
    - html

Numbers is useful for blogs where the numbers of comments are shown, or forums where post count is often updated.

	Numbers removes: [0-9]

Stripping with attrs is often used when session-id, timestamps or CSRF-protection is added to every URL on page.

	Attrs removes: <element attr="remove me"> -> <element>

When you just want the text on the page, strip HTML.

	HTML removes: <b>asdf</b> -> asdf

Strip is parsed as a list where the  `<` character appends to the list (each function on a new line).

    strip < numbers
    strip < html

[S] - Headers
-------------
If the page needs login or maybe has lang_option=English, you can add a special request header.
Headers is parsed as a list where the  `<` character appends to the list (each field name on a new line).

    headers < Cookie: lang_option=English; session_id=58ab1408ddefbe367ba2e808e54ed15a
    headers < User-Agent: Fake

[S] - Sel
---------
If the page has ads that constantly change or maybe has a comment field that you're not interested in you can specifically select a certain part of the page to check on request.
Sel is parsed as a string which follows CSS syntax.

    sel = #Blog1 .date-outer:first-child

[S] - Regexp
------------
If no CSS selector can't be compiled or you want further selection, you can use regular expressions to further check for updates.
Regexp is parsed as string and follows regexp syntax: [regexp/syntax](http://tip.golang.org/pkg/regexp/syntax/)

    regexp = S08E1[0-9]

[S] - Negexp
------------
A regexp string which removes everything that matches it. Efficient for blacklisting common words.
Negexp is parsed as string and follows regexp syntax: [regexp/syntax](http://tip.golang.org/pkg/regexp/syntax/)

    negexp = (hours ago)

[G][S] - Recvmail
-----------------
Decides which mail-address the program should mail updates when pages have been updated.
Mail is parsed as a string.

    recvmail = name@domain.tld

[S] - Threshold
---------------
Sets the percentage of accepted deviation from the last update.
Threshold is parsed as a float.

    threshold = 0.05

Z-level
-------
Further selection is performed in this order.

    - CSS-selection
        - Strip (number, attrs, html)
            - Regexp
                - Negexp

API documentation
-----------------
http://go.pkgdoc.org/github.com/karlek/nyfiken

Public domain
-------------
I hereby release this code into the [public domain](https://creativecommons.org/publicdomain/zero/1.0/).
