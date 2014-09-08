Cheesy Arena To-Do List
=======================

###Features for FRC parity
* Event wizard to guide scorekeeper through running an event
* Awards tracking and publishing
* Elimination bracket report and audience screen
* Interface for viewing logs (right now it's CSV files in Excel)
* Ability to edit match result before committing it
* Configurable match period timing (for test/practice matches only)
* Block driver station port through AP to facilitate on-field tethering
* Quality of service
* Twitter publishing

###Public-facing features
* Fancier graphics and animations for alliance station display
* Ability to yank the match data from the Internet for an existing event, for use just in webcast overlays
* GameSense-style next match screen with robot photos

###Scorekeeper-facing features
* Ability to unscore a match and reset it to non-played status
* Role-based cookie authentication
* Ability to mute match sounds from match play screen
* Logging console on Match Play page for errors and warnings
* Schedule generation takes match cycle time in min:sec instead of just seconds
* Team/field timeout tracking and overlay
* Make lower third show/hide commands use websockets instead of POST so that the scrolling doesn't reset when the page reloads
* Allow reordering of lower thirds and sponsor slides in their respective setup pages
* Automatic creation of lower thirds for awards

###Features for other volunteers
* Referee interface: add timer starting at field reset to track time limit for calling timeouts/backups
* Referee interface: have separate fouls for tech/non-tech for each applicable rule instead of the extra variable
* Mobile compatibility for FTA and announcer displays
* Automatic download of recent accomplishments (needs better TBA API)

###Development tasks
* Generate more schedules and find an automated way to evaluate them
* Clean up sponsor carousel JavaScript and make it load new slides asynchronously without needing a reload of the audience display page
* Refactor websockets to reduce code repetition between displays with similar functions
* Refactor to reduce usage of global variables
* Show non-modal dialog with websocket-returned errors
* JavaScript unit testing
* Fix Handlebars and golang html/template confict
* Set up [Travis continuous integration](https://travis-ci.org)
* [Selenium](http://www.seleniumhq.org) testing
