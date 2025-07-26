// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Author: ian@yann.io (Ian Thompson)
//
// Client-side logic for the scoring interface.

var websocket;
let alliance;
let nearSide;
let committed = false;

// True when scoring controls in general should be available
let scoringAvailable = false;
// True when the commit button should be available
let commitAvailable = false;
// True when teleop-only scoring controls should be available
let inTeleop = false;
// True when post-auto and in edit auto mode
let editingAuto = false;

let localFoulCounts = {
  "red-minor": 0,
  "blue-minor": 0,
  "red-major": 0,
  "blue-major": 0,
}

// Handle controls to open/close the endgame dialog
const endgameDialog = $("#endgame-dialog")[0];
const showEndgameDialog = function () {
  endgameDialog.showModal();
}
const closeEndgameDialog = function () {
  endgameDialog.close();
}
const closeEndgameDialogIfOutside = function (event) {
  if (event.target === endgameDialog) {
    closeEndgameDialog();
  }
}

const foulsDialog = $("#fouls-dialog")[0];
const showFoulsDialog = function () {
  foulsDialog.showModal();
}
const closeFoulsDialog = function () {
  foulsDialog.close();
}
const closeFoulsDialogIfOutside = function (event) {
  if (event.target === foulsDialog) {
    closeFoulsDialog();
  }
}

// Handles a websocket message to update the teams for the current match.
const handleMatchLoad = function (data) {
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

const renderLocalFoulCounts = function () {
  for (const foulType in localFoulCounts) {
    const count = localFoulCounts[foulType];
    $(`#foul-${foulType} .fouls-local`).text(count);
  }
}

const resetFoulCounts = function () {
  localFoulCounts["red-minor"] = 0;
  localFoulCounts["blue-minor"] = 0;
  localFoulCounts["red-major"] = 0;
  localFoulCounts["blue-major"] = 0;
  renderLocalFoulCounts();
}

const addFoul = function (alliance, isMajor) {
  const foulType = `${alliance}-${isMajor ? "major" : "minor"}`;
  localFoulCounts[foulType] += 1;
  websocket.send("addFoul", {Alliance: alliance, IsMajor: isMajor});
  renderLocalFoulCounts();
}

// Handles a websocket message to update the match status.
const handleMatchTime = function (data) {
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
      resetFoulCounts();
  }
  updateUIMode();
};

// Switch in and out of autonomous editing mode
const toggleEditAuto = function () {
  editingAuto = !editingAuto;
  updateUIMode();
}

// Clear any local ephemeral state that is not maintained by the server
const resetLocalState = function () {
  committed = false;
  editingAuto = false;
  updateUIMode();
}

// Refresh which UI controls are enabled/disabled
const updateUIMode = function () {
  $(".scoring-button").prop('disabled', !scoringAvailable);
  $(".scoring-teleop-button").prop('disabled', !(inTeleop && scoringAvailable));
  $("#commit").prop('disabled', !commitAvailable);
  $("#edit-auto").prop('disabled', !(inTeleop && scoringAvailable));
  $(".container").attr("data-scoring-auto", (!inTeleop || editingAuto) && scoringAvailable);
  $(".container").attr("data-in-teleop", inTeleop && scoringAvailable);
  $("#edit-auto").text(editingAuto ? "Save Auto" : "Edit Auto");
}

const endgameStatusNames = [
  "None",
  "Park",
  "Shallow",
  "Deep",
];

// Handles a websocket message to update the realtime scoring fields.
const handleRealtimeScore = function (data) {
  let realtimeScore;
  if (alliance === "red") {
    realtimeScore = data.Red;
  } else {
    realtimeScore = data.Blue;
  }
  const score = realtimeScore.Score;

  for (let i = 0; i < 3; i++) {
    const i1 = i + 1;
    $(`#auto-status-${i1} > .team-text`).text(score.LeaveStatuses[i] ? "Leave" : "None");
    $(`#auto-status-${i1}`).attr("data-selected", score.LeaveStatuses[i]);
    $(`#endgame-status-${i1} > .team-text`).text(endgameStatusNames[score.EndgameStatuses[i]]);
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

  const l1Total = score.Reef.TroughNear + score.Reef.TroughFar;
  $("#l1-total-count").text(l1Total);
  
  $(`#barge .counter-value`).text(score.BargeAlgae);
  $(`#processor .counter-value`).text(score.ProcessorAlgae);

  if (nearSide) {
    $(`#trough .counter-value`).text(score.Reef.TroughNear);
    $(`#trough .counter-auto-value`).text(score.Reef.AutoTroughNear);
  } else {
    $(`#trough .counter-value`).text(score.Reef.TroughFar);
    $(`#trough .counter-auto-value`).text(score.Reef.AutoTroughFar);
  }

  redFouls = data.Red.Score.Fouls || [];
  blueFouls = data.Blue.Score.Fouls || [];
  $(`#foul-blue-minor .fouls-global`).text(blueFouls.filter(foul => !foul.IsMajor).length)
  $(`#foul-blue-major .fouls-global`).text(blueFouls.filter(foul => foul.IsMajor).length)
  $(`#foul-red-minor .fouls-global`).text(redFouls.filter(foul => !foul.IsMajor).length)
  $(`#foul-red-major .fouls-global`).text(redFouls.filter(foul => foul.IsMajor).length)
};

// Websocket message senders for various buttons
const handleCounterClick = function (command, adjustment) {
  websocket.send(command, {
    Adjustment: adjustment,
    Current: true,
    Autonomous: !inTeleop || editingAuto,
    NearSide: nearSide
  });
}
const handleLeaveClick = function (teamPosition) {
  websocket.send("leave", {TeamPosition: teamPosition});
}
const handleEndgameClick = function (teamPosition, endgameStatus) {
  websocket.send("endgame", {TeamPosition: teamPosition, EndgameStatus: endgameStatus});
}
const handleReefClick = function (reefPosition, reefLevel) {
  websocket.send("reef", {
    ReefPosition: reefPosition,
    ReefLevel: reefLevel,
    Current: !editingAuto,
    Autonomous: !inTeleop || editingAuto,
    NearSide: nearSide
  });
}

// Sends a websocket message to indicate that the score for this alliance is ready.
const commitMatchScore = function () {
  websocket.send("commitMatch");

  committed = true;
  scoringAvailable = false;
  commitAvailable = false;
  inTeleop = false;
  editingAuto = false;
  updateUIMode();
};

$(function () {
  position = window.location.href.split("/").slice(-1)[0];
  [alliance, side] = position.split("_");
  $(".container").attr("data-alliance", alliance);
  nearSide = side === "near";
  resetLocalState();

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/panels/scoring/" + position + "/websocket", {
    matchLoad: function (event) {
      handleMatchLoad(event.data);
    },
    matchTime: function (event) {
      handleMatchTime(event.data);
    },
    realtimeScore: function (event) {
      handleRealtimeScore(event.data);
    },
    resetLocalState: function (event) {
      resetLocalState();
    },
  });
});
