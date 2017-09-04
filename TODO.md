Cheesy Arena To-Do List
=======================

### Features for FRC parity
* Event wizard to guide scorekeeper through running an event
* Awards tracking and publishing
* Elimination bracket report and audience screen
* Interface for viewing logs (right now it's CSV files in Excel)
* Log/report to see which teams have successfully connected to the field
* Quality of service
* Twitter publishing

### Public-facing features
* Fancier graphics and animations for alliance station display
* Ability to yank the match data from the Internet for an existing event, for use just in webcast overlays
* GameSense-style next match screen with robot photos

### Scorekeeper-facing features
* Ability to unscore a match and reset it to non-played status
* Logging console on Match Play page for errors and warnings
* Team/field timeout tracking and overlay
* Allow reordering of sponsor slides in the setup page
* Automatic creation of lower thirds for awards
* Persist schedule blocks after schedule generation, in case the schedule needs to be tweaked and re-run

### Features for other volunteers
* Referee interface: add timer starting at field reset to track time limit for calling timeouts/backups
* Mobile compatibility for announcer display

### Cheesy Arena Lite - a game-agnostic version
* Configurable match period timing
* Realtime scoring: just a simple single number input, plus API
* Final score screen: just show point total and remove breakdowns
* Genericize logos
* Manual input of match name
* Remove match scheduling and team standings functionality

### Development tasks
* Generate more schedules and find an automated way to evaluate them
* Clean up sponsor carousel JavaScript and make it load new slides asynchronously without needing a reload of the audience display page
* Refactor websockets to reduce code repetition between displays with similar functions
* Show non-modal dialog with websocket-returned errors
* JavaScript unit testing
* Fix Handlebars and golang html/template confict
* Set up [Travis continuous integration](https://travis-ci.org)
* [Selenium](http://www.seleniumhq.org) testing
