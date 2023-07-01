// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the match play page.

var websocket;
let scoreIsReady;
let isReplay;
const lowBatteryThreshold = 8;

// Sends a websocket message to load the specified match.
const loadMatch = function(matchId) {
  websocket.send("loadMatch", { matchId: matchId });
}

// Sends a websocket message to load the results for the specified match into the display buffer.
const showResult = function(matchId) {
  websocket.send("showResult", { matchId: matchId });
}

// Sends a websocket message to load a team into an alliance station.
const substituteTeam = function(team, position) {
  websocket.send("substituteTeam", { team: parseInt(team), position: position })
};

// Sends a websocket message to toggle the bypass status for an alliance station.
const toggleBypass = function(station) {
  websocket.send("toggleBypass", station);
};

// Sends a websocket message to start the match.
const startMatch = function() {
  websocket.send("startMatch",
      { muteMatchSounds: $("#muteMatchSounds").prop("checked") });
};

// Sends a websocket message to abort the match.
const abortMatch = function() {
  websocket.send("abortMatch");
};

// Sends a websocket message to signal to the volunteers that they may enter the field.
const signalVolunteers = function() {
  websocket.send("signalVolunteers");
};

// Sends a websocket message to signal to the teams that they may enter the field.
const signalReset = function() {
  websocket.send("signalReset");
};

// Sends a websocket message to commit the match score and load the next match.
const commitResults = function() {
  websocket.send("commitResults");
};

// Sends a websocket message to discard the match score and load the next match.
const discardResults = function() {
  websocket.send("discardResults");
};

// Sends a websocket message to change what the audience display is showing.
const setAudienceDisplay = function() {
  websocket.send("setAudienceDisplay", $("input[name=audienceDisplay]:checked").val());
};

// Sends a websocket message to change what the alliance station display is showing.
const setAllianceStationDisplay = function() {
  websocket.send("setAllianceStationDisplay", $("input[name=allianceStationDisplay]:checked").val());
};

// Sends a websocket message to start the timeout.
const startTimeout = function() {
  const duration = $("#timeoutDuration").val().split(":");
  let durationSec = parseFloat(duration[0]);
  if (duration.length > 1) {
    durationSec = durationSec * 60 + parseFloat(duration[1]);
  }
  websocket.send("startTimeout", durationSec);
};

const confirmCommit = function() {
  if (isReplay || !scoreIsReady) {
    // Show the appropriate message(s) in the confirmation dialog.
    $("#confirmCommitReplay").css("display", isReplay ? "block" : "none");
    $("#confirmCommitNotReady").css("display", scoreIsReady ? "none" : "block");
    $("#confirmCommitResults").modal("show");
  } else {
    commitResults();
  }
};

// Sends a websocket message to specify a custom name for the current test match.
const setTestMatchName = function() {
  websocket.send("setTestMatchName", $("#testMatchName").val());
};

// Handles a websocket message to update the team connection status.
const handleArenaStatus = function(data) {
  // Update the team status view.
  $.each(data.AllianceStations, function(station, stationStatus) {
    const wifiStatus = data.TeamWifiStatuses[station];
    $("#status" + station + " .radio-status").text(wifiStatus.TeamId);

    if (stationStatus.DsConn) {
      // Format the driver station status box.
      const dsConn = stationStatus.DsConn;
      $("#status" + station + " .ds-status").attr("data-status-ok", dsConn.DsLinked);
      if (dsConn.DsLinked) {
        $("#status" + station + " .ds-status").text(wifiStatus.MBits.toFixed(2)  + "Mb");
      } else {
        $("#status" + station + " .ds-status").text("");
      }
      // Format the radio status box according to the connection status of the robot radio.
      const radioOkay = stationStatus.Team && stationStatus.Team.Id === wifiStatus.TeamId && wifiStatus.RadioLinked;
      $("#status" + station + " .radio-status").attr("data-status-ok", radioOkay);

      // Format the robot status box.
      const robotOkay = dsConn.BatteryVoltage > lowBatteryThreshold && dsConn.RobotLinked;
      $("#status" + station + " .robot-status").attr("data-status-ok", robotOkay);
      if (stationStatus.DsConn.SecondsSinceLastRobotLink > 1 && stationStatus.DsConn.SecondsSinceLastRobotLink < 1000) {
        $("#status" + station + " .robot-status").text(stationStatus.DsConn.SecondsSinceLastRobotLink.toFixed());
      } else {
        $("#status" + station + " .robot-status").text(dsConn.BatteryVoltage.toFixed(1) + "V");
      }
    } else {
      $("#status" + station + " .ds-status").attr("data-status-ok", "");
      $("#status" + station + " .robot-status").attr("data-status-ok", "");
      $("#status" + station + " .robot-status").text("");

      // Format the robot status box according to whether the AP is configured with the correct SSID.
      const expectedTeamId = stationStatus.Team ? stationStatus.Team.Id : 0;
      if (wifiStatus.TeamId === expectedTeamId) {
        if (wifiStatus.RadioLinked) {
          $("#status" + station + " .radio-status").attr("data-status-ok", true);
        } else {
          $("#status" + station + " .radio-status").attr("data-status-ok", "");
        }
      } else {
        $("#status" + station + " .radio-status").attr("data-status-ok", false);
      }
    }

    if (stationStatus.Estop) {
      $("#status" + station + " .bypass-status").attr("data-status-ok", false);
      $("#status" + station + " .bypass-status").text("ES");
    } else if (stationStatus.Bypass) {
      $("#status" + station + " .bypass-status").attr("data-status-ok", false);
      $("#status" + station + " .bypass-status").text("B");
    } else {
      $("#status" + station + " .bypass-status").attr("data-status-ok", true);
      $("#status" + station + " .bypass-status").text("");
    }
  });

  // Enable/disable the buttons based on the current match state.
  switch (matchStates[data.MatchState]) {
    case "PRE_MATCH":
      $("#startMatch").prop("disabled", !data.CanStartMatch);
      $("#abortMatch").prop("disabled", true);
      $("#signalVolunteers").prop("disabled", false);
      $("#signalReset").prop("disabled", false);
      $("#fieldResetRadio").prop("disabled", false);
      $("#commitResults").prop("disabled", true);
      $("#discardResults").prop("disabled", true);
      $("#editResults").prop("disabled", true);
      $("#startTimeout").prop("disabled", false);
      break;
    case "START_MATCH":
    case "WARMUP_PERIOD":
    case "AUTO_PERIOD":
    case "PAUSE_PERIOD":
    case "TELEOP_PERIOD":
      $("#startMatch").prop("disabled", true);
      $("#abortMatch").prop("disabled", false);
      $("#signalVolunteers").prop("disabled", true);
      $("#signalReset").prop("disabled", true);
      $("#fieldResetRadio").prop("disabled", true);
      $("#commitResults").prop("disabled", true);
      $("#discardResults").prop("disabled", true);
      $("#editResults").prop("disabled", true);
      $("#startTimeout").prop("disabled", true);
      break;
    case "POST_MATCH":
      $("#startMatch").prop("disabled", true);
      $("#abortMatch").prop("disabled", true);
      $("#signalVolunteers").prop("disabled", false);
      $("#signalReset").prop("disabled", false);
      $("#fieldResetRadio").prop("disabled", false);
      $("#commitResults").prop("disabled", false);
      $("#discardResults").prop("disabled", false);
      $("#editResults").prop("disabled", false);
      $("#startTimeout").prop("disabled", true);
      break;
    case "TIMEOUT_ACTIVE":
      $("#startMatch").prop("disabled", true);
      $("#abortMatch").prop("disabled", false);
      $("#signalVolunteers").prop("disabled", true);
      $("#signalReset").prop("disabled", true);
      $("#fieldResetRadio").prop("disabled", false);
      $("#commitResults").prop("disabled", true);
      $("#discardResults").prop("disabled", true);
      $("#editResults").prop("disabled", true);
      $("#startTimeout").prop("disabled", true);
      break;
    case "POST_TIMEOUT":
      $("#startMatch").prop("disabled", true);
      $("#abortMatch").prop("disabled", true);
      $("#signalVolunteers").prop("disabled", true);
      $("#signalReset").prop("disabled", true);
      $("#fieldResetRadio").prop("disabled", false);
      $("#commitResults").prop("disabled", true);
      $("#discardResults").prop("disabled", true);
      $("#editResults").prop("disabled", true);
      $("#startTimeout").prop("disabled", true);
      break;
  }

  if (data.PlcIsHealthy) {
    $("#plcStatus").text("Connected");
    $("#plcStatus").attr("data-ready", true);
  } else {
    $("#plcStatus").text("Not Connected");
    $("#plcStatus").attr("data-ready", false);
  }
  $("#fieldEstop").attr("data-ready", !data.FieldEstop);
  $.each(data.PlcArmorBlockStatuses, function(name, status) {
    $("#plc" + name + "Status").attr("data-ready", status);
  });
};

// Handles a websocket message to update the teams for the current match.
const handleMatchLoad = function(data) {
  isReplay = data.IsReplay;

  fetch("/match_play/match_load")
    .then(response => response.text())
    .then(html => $("#matchListColumn").html(html));

  $("#matchName").text(data.Match.LongName);
  $("#testMatchName").val(data.Match.LongName);
  $("#testMatchSettings").toggle(data.Match.Type === matchTypeTest);
  $.each(data.Teams, function(station, team) {
    const teamId = $(`#status${station} .team-number`);
    teamId.val(team ? team.Id : "");
    teamId.prop("disabled", !data.AllowSubstitution);
  });
  $("#playoffRedAllianceInfo").html(formatPlayoffAllianceInfo(data.Match.PlayoffRedAlliance, data.RedOffFieldTeams));
  $("#playoffBlueAllianceInfo").html(formatPlayoffAllianceInfo(data.Match.PlayoffBlueAlliance, data.BlueOffFieldTeams));
}

// Handles a websocket message to update the match time countdown.
const handleMatchTime = function(data) {
  translateMatchTime(data, function(matchState, matchStateText, countdownSec) {
    $("#matchState").text(matchStateText);
    $("#matchTime").text(countdownSec);
  });
};

// Handles a websocket message to update the match score.
const handleRealtimeScore = function(data) {
  $("#redScore").text(data.Red.ScoreSummary.Score);
  $("#blueScore").text(data.Blue.ScoreSummary.Score);
};

// Handles a websocket message to populate the final score data.
const handleScorePosted = function(data) {
  let matchName = data.Match.LongName;
  if (!matchName) {
    matchName = "None"
  }
  $("#savedMatchName").html(matchName);
}

// Handles a websocket message to update the audience display screen selector.
const handleAudienceDisplayMode = function(data) {
  $("input[name=audienceDisplay]:checked").prop("checked", false);
  $("input[name=audienceDisplay][value=" + data + "]").prop("checked", true);
};

// Handles a websocket message to signal whether the referee and scorers have committed after the match.
const handleScoringStatus = function(data) {
  scoreIsReady = data.RefereeScoreReady && data.RedScoreReady && data.BlueScoreReady;
  $("#refereeScoreStatus").attr("data-ready", data.RefereeScoreReady);
  $("#redScoreStatus").text("Red Scoring " + data.NumRedScoringPanelsReady + "/" + data.NumRedScoringPanels);
  $("#redScoreStatus").attr("data-ready", data.RedScoreReady);
  $("#blueScoreStatus").text("Blue Scoring " + data.NumBlueScoringPanelsReady + "/" + data.NumBlueScoringPanels);
  $("#blueScoreStatus").attr("data-ready", data.BlueScoreReady);
};

// Handles a websocket message to update the alliance station display screen selector.
const handleAllianceStationDisplayMode = function(data) {
  $("input[name=allianceStationDisplay]:checked").prop("checked", false);
  $("input[name=allianceStationDisplay][value=" + data + "]").prop("checked", true);
};

// Handles a websocket message to update the event status message.
const handleEventStatus = function(data) {
  if (data.CycleTime === "") {
    $("#cycleTimeMessage").text("Last cycle time: Unknown");
  } else {
    $("#cycleTimeMessage").text("Last cycle time: " + data.CycleTime);
  }
  $("#earlyLateMessage").text(data.EarlyLateMessage);
};

const formatPlayoffAllianceInfo = function(allianceNumber, offFieldTeams) {
  if (allianceNumber === 0) {
    return "";
  }
  let allianceInfo = `<b>Alliance ${allianceNumber}</b>`;
  if (offFieldTeams.length > 0) {
    allianceInfo += ` (not on field: ${offFieldTeams.map(team => team.Id).join(", ")})`;
  }
  return allianceInfo;
}

$(function() {
  // Activate tooltips above the status headers.
  $("[data-toggle=tooltip]").tooltip({"placement": "top"});

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/match_play/websocket", {
    allianceStationDisplayMode: function(event) { handleAllianceStationDisplayMode(event.data); },
    arenaStatus: function(event) { handleArenaStatus(event.data); },
    audienceDisplayMode: function(event) { handleAudienceDisplayMode(event.data); },
    eventStatus: function(event) { handleEventStatus(event.data); },
    matchLoad: function(event) { handleMatchLoad(event.data); },
    matchTime: function(event) { handleMatchTime(event.data); },
    matchTiming: function(event) { handleMatchTiming(event.data); },
    realtimeScore: function(event) { handleRealtimeScore(event.data); },
    scorePosted: function(event) { handleScorePosted(event.data); },
    scoringStatus: function(event) { handleScoringStatus(event.data); },
  });
});
