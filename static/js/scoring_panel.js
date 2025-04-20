// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the scoring interface.

var websocket;
let alliance;

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
    case "PRE_MATCH":
      // Pre-match message state is set in handleRealtimeScore().
      $("#postMatchMessage").hide();
      $("#commitMatchScore").hide();
      break;
    case "POST_MATCH":
      $("#postMatchMessage").hide();
      $("#commitMatchScore").css("display", "flex");
      break;
    default:
      $("#postMatchMessage").hide();
      $("#commitMatchScore").hide();
  }
};

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
    for (let j = 0; j < endgameStatusNames.length; j++) {
      $(`#endgame-input-${i1} .endgame-${j}`).attr("data-selected", j == score.EndgameStatuses[i]);
    }
    $(`#parkTeam${i1}`).attr("data-value", score.EndgameStatuses[i] === 1);
    $(`#stageSide0Team${i1}`).attr("data-value", score.EndgameStatuses[i] === 2);
    $(`#stageSide1Team${i1}`).attr("data-value", score.EndgameStatuses[i] === 3);
    $(`#stageSide2Team${i1}`).attr("data-value", score.EndgameStatuses[i] === 4);
    $(`#stageSide${i}Microphone`).attr("data-value", score.MicrophoneStatuses[i]);
    $(`#stageSide${i}Trap`).attr("data-value", score.TrapStatuses[i]);
  }
};

// Handles an element click and sends the appropriate websocket message.
const handleClick = function(command, teamPosition = 0, stageIndex = 0) {
  websocket.send(command, {TeamPosition: teamPosition, StageIndex: stageIndex});
};

// Sends a websocket message to indicate that the score for this alliance is ready.
const commitMatchScore = function() {
  websocket.send("commitMatch");
  $("#postMatchMessage").css("display", "flex");
  $("#commitMatchScore").hide();
};

$(function() {
  alliance = window.location.href.split("/").slice(-1)[0];
  $(".container").attr("data-alliance", alliance);

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/panels/scoring/" + alliance + "/websocket", {
    matchLoad: function(event) { handleMatchLoad(event.data); },
    matchTime: function(event) { handleMatchTime(event.data); },
    realtimeScore: function(event) { handleRealtimeScore(event.data); },
  });
});
