// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the scoring interface.

var websocket;
let alliance;
let nearSide;

// True when scoring controls in general should be available
let scoringAvailable = false;
// True when the commit button should be available
let commitAvailable = false;
// True when teleop-only scoring controls should be available
let inTeleop = false;
// True when post-auto and in edit auto mode
let editingAuto = false;

// Whether the most recent match has been committed
let committed = false;

const endgameDialog = $("#endgame-dialog")[0];

const showEndgameDialog = function() {
  endgameDialog.showModal();
}

const closeEndgameDialog = function() {
  endgameDialog.close();
}

const closeEndgameDialogIfOutside = function(event) {
  if (event.target === endgameDialog) {
    closeEndgameDialog();
  }
}

// Handles a websocket message to update the teams for the current match.
const handleMatchLoad = function(data) {
  $("#matchName").text(data.Match.LongName);
  if (alliance === "red") {
    $(".team-1 .team-num").text(data.Match.Red1);
    $(".team-2 .team-num").text(data.Match.Red2);
    $(".team-3 .team-num").text(data.Match.Red3);
  } else {
    $(".team-1 .team-num").text(data.Match.Blue1);
    $(".team-2 .team-num").text(data.Match.Blue2);
    $(".team-3 .team-num").text(data.Match.Blue3);
  }
};

// Handles a websocket message to update the match status.
const handleMatchTime = function(data) {
  switch (matchStates[data.MatchState]) {
    case "AUTO_PERIOD":
    case "PAUSE_PERIOD":
      scoringAvailable = true;
      commitAvailable = false;
      inTeleop = false;
      editingAuto = false;
      committed = false;
      break;
    case "TELEOP_PERIOD":
      scoringAvailable = true;
      commitAvailable = false;
      inTeleop = true;
      committed = false;
      break;
    case "POST_MATCH":
      if (!committed) {
        scoringAvailable = true;
        commitAvailable = true;
        inTeleop = true;
      }
      break;
    default:
      scoringAvailable = false;
      commitAvailable = false;
      inTeleop = false;
      editingAuto = false;
      committed = false;
  }
  updateUIMode();
};

const updateUIMode = function() {
  // Push mode changes to the UI
  $(".scoring-button").prop('disabled', !scoringAvailable);
  $(".scoring-teleop-button").prop('disabled', !(inTeleop && scoringAvailable));
  $("#commit").prop('disabled', !commitAvailable);
  $("#edit-auto").prop('disabled', !(inTeleop && scoringAvailable));
  $("main").attr("data-editing-auto", editingAuto);
  $("#edit-auto").text(editingAuto ? "Save Auto" : "Edit Auto");
}

const endgameStatusNames = [
  "None",
  "Park",
  "Shallow",
  "Deep",
];

// Handles a websocket message to update the realtime scoring fields.
const handleRealtimeScore = function(data) {
  let realtimeScore;
  if (alliance === "red") {
    realtimeScore = data.Red;
  } else {
    realtimeScore = data.Blue;
  }
  const score = realtimeScore.Score;

  for (let i = 0; i < 3; i++) {
    const i1 = i + 1;
    $(`#auto-status-${i1}>.team-text`).text(score.LeaveStatuses[i] ? "Leave" : "None");
    $(`#auto-status-${i1}`).attr("data-selected", score.LeaveStatuses[i]);
    $(`#endgame-status-${i1}>.team-text`).text(endgameStatusNames[score.EndgameStatuses[i]]);
    $(`#endgame-status-${i1}`).attr("data-selected", endgameStatusNames[score.EndgameStatuses[i]] != "None");
    for (let j = 0; j < endgameStatusNames.length; j++) {
      $(`#endgame-input-${i1} .endgame-${j}`).attr("data-selected", j == score.EndgameStatuses[i]);
    }
  }

  for (let i = 0; i < 12; i++) {
    const i1 = i + 1;
    for (let j = 0; j < 3; j++) {
      const j2 = j + 2;
      $(`#reef-column-${i1}`).attr(`data-l${j2}-scored`, score.Reef.Branches[j][i]);
      $(`#reef-column-${i1}`).attr(`data-l${j2}-auto-scored`, score.Reef.AutoBranches[j][i]);
    }
  }

  $(`#barge .counter-value`).text(score.BargeAlgae);
  $(`#processor .counter-value`).text(score.ProcessorAlgae);
  if (nearSide) {
    $(`#trough .counter-value`).text(score.Reef.TroughNear);
    $(`#trough .counter-auto-value`).text(score.Reef.AutoTroughNear);
  } else {
    $(`#trough .counter-value`).text(score.Reef.TroughFar);
    $(`#trough .counter-auto-value`).text(score.Reef.AutoTroughFar);
  }
};

// Websocket message senders for various buttons
const handleCounterClick = function(command, adjustment) {
  websocket.send(command, {Adjustment: adjustment, Current: !editingAuto, Autonomous: !inTeleop || editingAuto, NearSide: nearSide});
}

const handleLeaveClick = function(teamPosition) {
  websocket.send("leave", {TeamPosition: teamPosition});
}

const handleEndgameClick = function(teamPosition, endgameStatus) {
  websocket.send("endgame", {TeamPosition: teamPosition, EndgameStatus: endgameStatus});
}

const handleReefClick = function(reefPosition, reefLevel) {
  websocket.send("reef", {ReefPosition: reefPosition, ReefLevel: reefLevel, Current: !editingAuto, Autonomous: !inTeleop || editingAuto, NearSide: nearSide});
}

// Sends a websocket message to indicate that the score for this alliance is ready.
const commitMatchScore = function() {
  websocket.send("commitMatch");

  committed = true;
  scoringAvailable = false;
  commitAvailable = false;
  inTeleop = false;
  editingAuto = false;
  updateUIMode();
};

const toggleEditAuto = function() {
  editingAuto = !editingAuto;
  updateUIMode();
}

$(function() {
  position = window.location.href.split("/").slice(-1)[0];
  [alliance, side] = position.split("_");
  $(".container").attr("data-alliance", alliance);
  nearSide = side === "near";

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/panels/scoring/" + alliance + "/websocket", {
    matchLoad: function(event) { handleMatchLoad(event.data); },
    matchTime: function(event) { handleMatchTime(event.data); },
    realtimeScore: function(event) { handleRealtimeScore(event.data); },
  });
});
