Cheesy Arena
============
A field management system that just works.

##Key Features
Check out a [video overview](http://video.team254.com/watch/Z5ZWI2cDqsvVe--AjHhePAHlOhLK8MT0) of Cheesy Arena's functionality.

**For participants and spectators:**

* No-lag goal/pedestal lighting and realtime scoring
* Team stack lights and sevent-segment display are replaced by an LCD screen, which shows team info before the match and realtime scoring and timer during the match
* Smooth-scrolling rankings display
* Direct publishing of schedule, results, and rankings to The Blue Alliance

**For scorekeepers:**

* Runs on Windows, Mac OS X, and Linux
* No install prerequisites
* No "pre-start" &ndash; hardware is configured automatically and in the background
* Flexible and quick match schedule generation
* Streamlined realtime score entry
* Reports, results, and logs can be viewed from any computer

##License
Teams may use Cheesy Arena freely for practice, scrimmages, and off-season events. See [LICENSE](LICENSE) for more details.

##Installation and use
**Via binaries:**

1. Download the [latest release](https://github.com/Team254/cheesy-arena/releases) for OS X or Windows
1. Unzip the file
1. On Mac OS X, run `cheesy-arena.command`. On Windows, run `cheesy-arena.exe`.
1. Navigate to http://localhost:8080 in your browser (Google Chrome recommended)

**From source:**

1. Download [Go](http://golang.org/doc/install)
1. Set up your [Go workspace](http://golang.org/doc/code.html)
1. If you're using Windows and don't already have a working version of GCC (needed to compile a dependency), install [TDM-GCC](http://tdm-gcc.tdragon.net).
1. Download the Cheesy Arena source and dependencies with `go get github.com/Team254/cheesy-arena`
1. Compile the code with `go build`
1. Run the `cheesy-arena` or `cheesy-arena.exe` binary
1. Navigate to http://localhost:8080 in your browser (Google Chrome recommended)

##Under the hood
Cheesy Arena is written using [Go](http://golang.org), a relatively new language developed by Google. Go excels in the areas of concurrency, networking, performance, and portability, which makes it ideal for a field management system.

Cheesy Arena is implemented as a web server, with all human interaction done via browser. The graphical interfaces are implemented in HTML, JavaScript, and CSS. There are many advantages to this approach &ndash; development of new graphical elements is rapid, and no software needs to be installed other than on the server. Client web pages send commands and receive updates using WebSockets.

SQLite3 is used as the datastore, and making backups or transferring data from one installation to another is as simple as copying the database file.

Schedule generation is fast because pregenerated schedules are included with the code. Each schedule contains a certain number of matches per team for placeholder teams 1 through N, so generating the actual match schedule becomes a simple exercise in permuting the mapping of real teams to placeholder teams. The pregenerated schedules are checked into this repository and can be vetted in advance of any events for deviations from the randomness (and other) requirements.

Cheesy Arena includes support for, but doesn't require, networking hardware similar to that used in official FRC events. Teams are issued their own SSIDs and WPA keys, and when connected to Cheesy Arena are isolated to a VLAN which prevents any communication other than between the driver station, robot, and event server. The network hardware is configured via Telnet commands for the new set of teams when each mach is loaded.

## LED hardware
Due to the prohibitive cost of the LEDs and LED controllers used on official fields, a custom solution was developed for Chezy Champs using consumer-grade LED strips and embedded microcontrollers. The bill of materials, control board schematics, and embedded source code will be provided in an upcoming release.

## Advanced networking
See the [Advanced Networking wiki page](wiki/Advanced-Networking) for instructions on what equipment to obtain and how to configure it in order to support advanced network security.

##Contributing
Cheesy Arena is far from finished! You can help by:

* Checking out the [TODO list](TODO.md), writing a missing feature, and sending a pull request
* Filing any bugs or feature requests using the [issue tracker](https://github.com/Team254/cheesy-arena/issues)
* Contributing documentation to the [wiki](https://github.com/Team254/cheesy-arena/wiki)
* Sending baked goods to Pat

##Acknowledgements
The following individuals contributed to make Cheesy Arena a reality:

* Tom Bottiglieri
* James Cerar
* Travis Covington
* Nick Eyre
* Patrick Fairbank
* Eugene Fang
* Karthik Kanagasabapathy
* Ken Mitchell
* Andrew Nabors
* Jared Russell
* Austin Schuh
* Colin Wilson