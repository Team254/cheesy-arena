// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the announcer display.

var websocket;
var teamTemplate = Handlebars.compile($("#teamTemplate").html());
var matchResultTemplate = Handlebars.compile($("#matchResultTemplate").html());
Handlebars.registerHelper("eachMapEntry", function(context, options) {
  var ret = "";
  $.each(context, function(key, value) {
    var entry = {"key": key, "value": value};
    ret = ret + options.fn(entry);
  });
  return ret;
});

// Handles a websocket message to hide the score dialog once the next match is being introduced.
var handleAudienceDisplayMode = function(targetScreen) {
  // Hide the final results so that they aren't blocking the current teams when the announcer needs them most.
  if (targetScreen === "intro" || targetScreen === "match") {
    $("#matchResult").modal("hide");
  }
};

// Handles a websocket message to update the teams for the current match.
var handleMatchLoad = function(data) {
  $("#matchName").text(data.MatchType + " Match " + data.Match.DisplayName);
  $("#red1").html(teamTemplate(formatTeam(data.Teams["R1"])));
  $("#red2").html(teamTemplate(formatTeam(data.Teams["R2"])));
  $("#red3").html(teamTemplate(formatTeam(data.Teams["R3"])));
  $("#blue1").html(teamTemplate(formatTeam(data.Teams["B1"])));
  $("#blue2").html(teamTemplate(formatTeam(data.Teams["B2"])));
  $("#blue3").html(teamTemplate(formatTeam(data.Teams["B3"])));
};

// Handles a websocket message to update the match time countdown.
var handleMatchTime = function(data) {
  translateMatchTime(data, function(matchState, matchStateText, countdownSec) {
    $("#matchState").text(matchStateText);
    $("#matchTime").text(getCountdown(data.MatchState, data.MatchTimeSec));
  });
};

// Handles a websocket message to update the match score.
var handleRealtimeScore = function(data) {
  $("#redScore").text(data.Red.ScoreSummary.Score - data.Red.ScoreSummary.HangarPoints);
  $("#blueScore").text(data.Blue.ScoreSummary.Score - data.Blue.ScoreSummary.HangarPoints);
};

// Handles a websocket message to populate the final score data.
var handleScorePosted = function(data) {
  $.each(data.RedFouls, function(i, foul) {
    Object.assign(foul, data.RulesViolated[foul.RuleId]);
  });
  $.each(data.BlueFouls, function(i, foul) {
    Object.assign(foul, data.RulesViolated[foul.RuleId]);
  });

  var redRankings = {};
  redRankings[data.Match.Red1] = getRankingText(data.Match.Red1, data.Rankings);
  redRankings[data.Match.Red2] = getRankingText(data.Match.Red2, data.Rankings);
  redRankings[data.Match.Red3] = getRankingText(data.Match.Red3, data.Rankings);
  var blueRankings = {};
  blueRankings[data.Match.Blue1] = getRankingText(data.Match.Blue1, data.Rankings);
  blueRankings[data.Match.Blue2] = getRankingText(data.Match.Blue2, data.Rankings);
  blueRankings[data.Match.Blue3] = getRankingText(data.Match.Blue3, data.Rankings);

  $("#scoreMatchName").text(data.MatchType + " Match " + data.Match.DisplayName);
  $("#redScoreDetails").html(matchResultTemplate({score: data.RedScoreSummary, fouls: data.RedFouls,
      rulesViolated: data.RulesViolated, cards: data.RedCards, rankings: redRankings}));
  $("#blueScoreDetails").html(matchResultTemplate({score: data.BlueScoreSummary, fouls: data.BlueFouls,
    rulesViolated: data.RulesViolated, cards: data.BlueCards, rankings: blueRankings}));
  $("#matchResult").modal("show");

  // Activate tooltips above the foul listings.
  $("[data-toggle=tooltip]").tooltip({"placement": "top"});
};

// Replaces newlines in team fields with HTML line breaks.
var formatTeam = function(team) {
  if (team) {
    team.Accomplishments = team.Accomplishments.replace(/[\r\n]+/g, "<br />");
  }
  return team;
};

// Returns the string to be displayed to indicate change in rank.
var getRankingText = function(teamId, rankings) {
  var ranking = rankings[teamId];
  if (ranking === undefined || ranking.Rank === 0) {
    return "";
  }
  var arrow = "";
  if (ranking.Rank > ranking.PreviousRank && ranking.PreviousRank > 0) {
    arrow = "&#11015;";
  } else if (ranking.Rank < ranking.PreviousRank) {
    arrow = "&#11014;";
  }
  var previousRank = "";
  if (ranking.PreviousRank > 0) {
    previousRank = " (was " + ranking.PreviousRank + ")";
  }
  return ranking.Rank + arrow + previousRank;
};

$(function() {
  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/announcer/websocket", {
    audienceDisplayMode: function(event) { handleAudienceDisplayMode(event.data); },
    matchLoad: function(event) { handleMatchLoad(event.data); },
    matchTime: function(event) { handleMatchTime(event.data); },
    matchTiming: function(event) { handleMatchTiming(event.data); },
    realtimeScore: function(event) { handleRealtimeScore(event.data); },
    scorePosted: function(event) { handleScorePosted(event.data); }
  });

  // Make the score blink.
  setInterval(function() {
    var blinkOn = $("#savedMatchResult").attr("data-blink") === "true";
    $("#savedMatchResult").attr("data-blink", !blinkOn);
  }, 500);
});
