// Copyright 2014 Team 254. All Rights Reserved.
// Author: nick@team254.com (Nick Eyre)
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side methods for the pit display.

var initial_dwell_ms = 3000;
var scroll_ms_per_row = 700;  // How long in milliseconds it takes to scroll a height of one row.
var static_update_interval_ms = 10000;  // How long between updates if not scrolling.
var standingsTemplate = Handlebars.compile($("#standingsTemplate").html());
var rankingsData;
var rankingsIteration = 0;
var prevHighestPlayedMatch;

// Loads the JSON rankings data from the event server.
var getRankingsData = function(callback) {
  $.getJSON("/api/rankings", function(data) {
    rankingsData = data;
    if (callback) {
      callback(data);
    }
  });
};

// Updates the rankings in place and initiates scrolling if they are long enough to require it.
var updateStaticRankings = function() {
  getRankingsData(function() {
    var rankingsHtml = standingsTemplate(rankingsData);
    $("#rankings2").html(rankingsHtml);
    $("#scroller").css("transform", "translate(0px, -2px);");
    prevHighestPlayedMatch = rankingsData.HighestPlayedMatch;
    setHighestPlayedMatch(rankingsData.HighestPlayedMatch);
    if ($("#rankings2").height() > $("#container").height()) {
      // Initiate scrolling.
      setTimeout(cycleRankings, initial_dwell_ms);
    } else {
      // Rankings are too short; just update in place.
      setTimeout(updateStaticRankings, static_update_interval_ms);
    }
  });
};

// Seamlessly copies the newer table contents to the older one, resets the scrolling, and loads new data.
var cycleRankings = function() {
  // Overwrite the top data with the bottom data and reset the scrolling back up to the top of the top table.
  $("#rankings1").html($("#rankings2").html());
  $("#scroller").css({ transform: "translate(0px, -1px);" });

  // Load new data into the now out-of-sight bottom table.
  var rankingsHtml = standingsTemplate(rankingsData);
  $("#rankings2").html(rankingsHtml);

  // Delay updating the "Standings as of" message by one cycle because the tables are always one cycle behind
  // the data loading.
  setHighestPlayedMatch(prevHighestPlayedMatch);
  prevHighestPlayedMatch = rankingsData.HighestPlayedMatch;

  if ($("#rankings1").height() > $("#container").height()) {
    // Kick off another scrolling animation.
    var scrollDistance = $("#rankings1").height() + parseInt($("#rankings1").css("border-bottom-width"));
    var scrollTime = scroll_ms_per_row * $("#rankings1 tr").length;
    $("#scroller").transition({y: -scrollDistance}, scrollTime, "linear", cycleRankings);

    // Set the data to be reloaded two seconds before the scrolling terminates.
    var reloadDataTime = Math.max(0, scrollTime - 2000);
    setTimeout(getRankingsData, reloadDataTime);
  } else {
    // The rankings got shorter for whatever reason, so revert to static updating.
    setTimeout(updateStaticRankings, static_update_interval_ms);
  }
};

// Updates the "Standings as of" message with the given value, or blanks it out if there is no data yet.
var setHighestPlayedMatch = function(highestPlayedMatch) {
  if (highestPlayedMatch == "") {
    $("#highestPlayedMatch").text("");
  } else {
    $("#highestPlayedMatch").text("Standings as of Qualification Match " + highestPlayedMatch);
  }
};

$(function() {
  updateStaticRankings();
});
