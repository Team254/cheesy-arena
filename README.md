Cheesy Arena [![Build Status](https://github.com/Team254/cheesy-arena/actions/workflows/test.yml/badge.svg)](https://github.com/Team254/cheesy-arena/actions)
============
A field management system that just works.

For the game-agnostic version, see [Cheesy Arena Lite](https://github.com/Team254/cheesy-arena-lite).

## Key features
**For participants and spectators**

* Same network isolation and security as the official FIRST FMS
* No-lag realtime scoring
* Team stack lights and seven-segment display are replaced by an LCD screen, which shows team info before the match and realtime scoring and timer during the match
* Smooth-scrolling rankings display
* Direct publishing of schedule, results, and rankings to The Blue Alliance

**For scorekeepers and event staff**

* Runs on Windows, macOS, and Linux
* No install prerequisites
* No "pre-start" &ndash; hardware is configured automatically and in the background
* Flexible and quick match schedule generation
* Streamlined realtime score entry
* Reports, results, and logs can be viewed from any computer
* An arbitrary number of auxiliary displays can be set up using any computer with just a web browser, to show rankings, queueing, field status, etc.

## License
Teams may use Cheesy Arena freely for practice, scrimmages, and off-season events. See [LICENSE](LICENSE) for more details.

## Installing

**From a pre-built release**

Download the [latest release](https://github.com/Team254/cheesy-arena/releases). Pre-built packages are available for Linux, macOS (x64 and M1), and Windows.

On recent versions of macOS, you may be prevented from running an app from an unidentified developer; see [these instructions](https://support.apple.com/guide/mac-help/open-a-mac-app-from-an-unidentified-developer-mh40616/mac) on how to bypass the warning.

**From source**

1. Download [Go](https://golang.org/dl/) (version 1.20 or later recommended)
1. Clone this GitHub repository to a location of your choice
1. Navigate to the repository's directory in the terminal
1. Compile the code with `go build`
1. Run the `cheesy-arena` or `cheesy-arena.exe` binary
1. Navigate to http://localhost:8080 in your browser (Google Chrome recommended)

**IP address configuration**

When running Cheesy Arena on a playing field with robots, set the IP address of the computer running Cheesy Arena to 10.0.100.5. By a convention baked into the FRC Driver Station software, driver stations will broadcast their presence on the network to this hardcoded address so that the FMS does not need to discover them by some other method.

When running Cheesy Arena without robots for testing or development, any IP address can be used.

## Under the hood
Cheesy Arena is written using [Go](https://golang.org), a language developed by Google and first released in 2009. Go excels in the areas of concurrency, networking, performance, and portability, which makes it ideal for a field management system.

Cheesy Arena is implemented as a web server, with all human interaction done via browser. The graphical interfaces are implemented in HTML, JavaScript, and CSS. There are many advantages to this approach &ndash; development of new graphical elements is rapid, and no software needs to be installed other than on the server. Client web pages send commands and receive updates using WebSockets.

[Bolt](https://github.com/etcd-io/bbolt) is used as the datastore, and making backups or transferring data from one installation to another is as simple as copying the database file.

Schedule generation is fast because pre-generated schedules are included with the code. Each schedule contains a certain number of matches per team for placeholder teams 1 through N, so generating the actual match schedule becomes a simple exercise in permuting the mapping of real teams to placeholder teams. The pre-generated schedules are checked into this repository and can be vetted in advance of any events for deviations from the randomness (and other) requirements.

Cheesy Arena includes support for, but doesn't require, networking hardware similar to that used in official FRC events. Teams are issued their own SSIDs and WPA keys, and when connected to Cheesy Arena are isolated to a VLAN which prevents any communication other than between the driver station, robot, and event server. The network hardware is reconfigured via SSH and Telnet commands for the new set of teams when each mach is loaded.

## PLC integration
Cheesy Arena has the ability to integrate with an Allen-Bradley PLC setup similar to the one that FIRST uses, to read field sensors and control lights and motors. The PLC hardware travels with the FIRST California fields; contact your FTA for more information.

The PLC code can be found [here](https://github.com/ejordan376/Cheesy-PLC).

## LED hardware
Due to the prohibitive cost of the LEDs and LED controllers used on official fields, for years in which LEDs are mandatory for a proper game experience (such as 2018), Cheesy Arena integrates with [Advatek](https://www.advateklights.com) controllers and LEDs.

## Advanced networking
See the [Advanced Networking wiki page](https://github.com/Team254/cheesy-arena/wiki/Advanced-Networking-Concepts) for instructions on what equipment to obtain and how to configure it in order to support advanced network security.

## Contributing
Cheesy Arena is far from finished! You can help by:

* Checking out the [TODO list](TODO.md), writing a missing feature, and sending a pull request
* Filing any bugs or feature requests using the [issue tracker](https://github.com/Team254/cheesy-arena/issues)
* Contributing documentation to the [wiki](https://github.com/Team254/cheesy-arena/wiki)
* Sending baked goods to [Pat](https://github.com/patfair)

## Acknowledgements
[Several folks](https://github.com/Team254/cheesy-arena/graphs/contributors) have contributed pull requests. Thanks!

In addition, the following individuals have contributed to make Cheesy Arena a reality:

* Tom Bottiglieri
* James Cerar
* Kiet Chau
* Travis Covington
* Nick Eyre
* Patrick Fairbank
* Eugene Fang
* Thad House
* Ed Jordan
* Karthik Kanagasabapathy
* Ken Mitchell
* Andrew Nabors
* Jared Russell
* Ken Schenke
* Austin Schuh
* Colin Wilson
