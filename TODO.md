Cheesy Arena To-Do List
=======================

### Features for FRC parity
* Event wizard to guide scorekeeper through running an event
* Elimination bracket report and audience screen
* Interface for viewing logs (right now it's CSV files in Excel)

### Public-facing features
* Fancier graphics and animations for alliance station display
* Ability to yank the match data from the Internet for an existing event, for use just in webcast overlays
* GameSense-style next match screen with robot photos

### Scorekeeper-facing features
* Ability to unscore a match and reset it to non-played status

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
* JavaScript unit testing
* Fix Handlebars and golang html/template confict
* [Selenium](http://www.seleniumhq.org) testing
