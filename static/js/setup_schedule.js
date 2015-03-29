// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side methods for the schedule generation page.

var blockTemplate = Handlebars.compile($("#blockTemplate").html());
var lastBlockNumber = 0;
var blockMatches = {};

// Adds a new scheduling block to the page.
var addBlock = function(startTime, numMatches, matchSpacingSec) {
  if (!startTime) {
    // Start the next block where the last one left off and use the same spacing.
    var lastStartTime = moment(Date.parse($("#startTime" + lastBlockNumber).val()));
    var lastNumMatches = blockMatches[lastBlockNumber];
    matchSpacingSec = getMatchSpacingSec(lastBlockNumber);
    startTime = moment(lastStartTime + lastNumMatches * matchSpacingSec * 1000);
    numMatches = 10;
  }
  var endTime = moment(startTime + numMatches * matchSpacingSec * 1000);
  lastBlockNumber += 1;
  var matchSpacingMinSec = moment(matchSpacingSec * 1000).format("m:ss");
  var block = blockTemplate({blockNumber: lastBlockNumber, matchSpacingMinSec: matchSpacingMinSec});
  $("#blockContainer").append(block);
  $("#startTimePicker" + lastBlockNumber).datetimepicker({useSeconds: true}).
      data("DateTimePicker").setDate(startTime);
  $("#endTimePicker" + lastBlockNumber).datetimepicker({useSeconds: true}).
      data("DateTimePicker").setDate(endTime);
  updateBlock(lastBlockNumber);
}

// Updates the per-block and global schedule statistics.
var updateBlock = function(blockNumber) {
  var startTime = moment(Date.parse($("#startTime" + blockNumber).val()));
  var endTime = moment(Date.parse($("#endTime" + blockNumber).val()));
  var matchSpacingSec = getMatchSpacingSec(blockNumber);
  var numMatches = Math.floor((endTime - startTime) / matchSpacingSec / 1000);
  var actualEndTime = moment(startTime + numMatches * matchSpacingSec * 1000).format("hh:mm:ss A");
  blockMatches[blockNumber] = numMatches;
  if (matchSpacingSec == "" || isNaN(numMatches) || numMatches <= 0) {
    numMatches = "";
    actualEndTime = "";
    blockMatches[blockNumber] = 0;
  }
  $("#numMatches" + blockNumber).text(numMatches);
  $("#actualEndTime" + blockNumber).text(actualEndTime);

  updateStats();
}

var updateStats = function() {
  // Update total number of matches.
  var totalNumMatches = 0;
  $.each(blockMatches, function(k, v) {
    totalNumMatches += v;
  });
  var matchesPerTeam = Math.floor(totalNumMatches * 6 / numTeams);
  var numExcessMatches = totalNumMatches - Math.ceil(matchesPerTeam * numTeams / 6);
  var nextLevelMatches = Math.ceil((matchesPerTeam + 1) * numTeams / 6) - totalNumMatches;
  $("#totalNumMatches").text(totalNumMatches);
  $("#matchesPerTeam").text(matchesPerTeam);
  $("#numExcessMatches").text(numExcessMatches);
  $("#nextLevelMatches").text(nextLevelMatches);
};

var deleteBlock = function(blockNumber) {
  delete blockMatches[blockNumber];
  $("#block" + blockNumber).remove();
  updateStats();
};

// Dynamically generates and posts a form containing the schedule blocks to the server for population.
var generateSchedule = function() {
  var form = $("#scheduleForm");
  form.attr("method", "POST");
  form.attr("action", "/setup/schedule/generate");
  var addField = function(name, value) {
  var field = $(document.createElement("input"));
    field.attr("type", "hidden");
    field.attr("name", name);
    field.attr("value", value);
    form.append(field);
  }
  var i = 0;
  $.each(blockMatches, function(k, v) {
    addField("startTime" + i, $("#startTime" + k).val());
    addField("numMatches" + i, $("#numMatches" + k).text());
    addField("matchSpacingSec" + i, getMatchSpacingSec(k));
    i++;
  });
  addField("numScheduleBlocks", i);
  form.submit();
};

// Parses the min:sec match spacing field for the given block and returns the number of seconds.
var getMatchSpacingSec = function(blockNumber) {
  var matchSpacingMinSec = $("#matchSpacingMinSec" + blockNumber).val().split(":");
  return parseInt(matchSpacingMinSec[0]) * 60 + parseInt(matchSpacingMinSec[1]);
};
