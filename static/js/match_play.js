// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the match play page.

var websocket;
let scoreIsReady;
let isReplay;
const lowBatteryThreshold = 8;

// Sends a websocket message to load the specified match.
const loadMatch = function (matchId) {
  websocket.send("loadMatch", {matchId: matchId});
}

// Sends a websocket message to load the results for the specified match into the display buffer.
const showResult = function (matchId) {
  websocket.send("showResult", {matchId: matchId});
}

// Sends a websocket message to load all teams into their respective alliance stations.
const substituteTeams = function (team, position) {
  const teams = {
    Red1: getTeamNumber("R1"),
    Red2: getTeamNumber("R2"),
    Red3: getTeamNumber("R3"),
    Blue1: getTeamNumber("B1"),
    Blue2: getTeamNumber("B2"),
    Blue3: getTeamNumber("B3"),
  };

  websocket.send("substituteTeams", teams);
};

// Sends a websocket message to toggle the bypass status for an alliance station.
const toggleBypass = function (station) {
  websocket.send("toggleBypass", station);
};

// Sends a websocket message to start the match.
const startMatch = function () {
  websocket.send("startMatch",
    {muteMatchSounds: $("#muteMatchSounds").prop("checked")});
};

// Sends a websocket message to abort the match.
const abortMatch = function () {
  websocket.send("abortMatch");
};

// Sends a websocket message to signal to the volunteers that they may enter the field.
const signalVolunteers = function () {
  websocket.send("signalVolunteers");
};

// Sends a websocket message to signal to the teams that they may enter the field.
const signalReset = function () {
  websocket.send("signalReset");
};

// Sends a websocket message to commit the match score and load the next match.
const commitResults = function () {
  websocket.send("commitResults");
};

// Sends a websocket message to discard the match score and load the next match.
const discardResults = function () {
  websocket.send("discardResults");
};

// Switches the audience display to the match intro screen.
const showOverlay = function () {
  $("input[name=audienceDisplay][value=intro]").prop("checked", true);
  setAudienceDisplay();
  $("#showOverlay").prop("disabled", true);
}

// Switches the audience display to the final score screen.
const showFinalScore = function () {
  $("input[name=audienceDisplay][value=score]").prop("checked", true);
  setAudienceDisplay();
  $("#showFinalScore").prop("disabled", true);
}

// Sends a websocket message to change what the audience display is showing.
const setAudienceDisplay = function () {
  websocket.send("setAudienceDisplay", $("input[name=audienceDisplay]:checked").val());
};

// Sends a websocket message to change what the alliance station display is showing.
const setAllianceStationDisplay = function () {
  websocket.send("setAllianceStationDisplay", $("input[name=allianceStationDisplay]:checked").val());
};

// Sends a websocket message to start the timeout.
const startTimeout = function () {
  const duration = $("#timeoutDuration").val().split(":");
  let durationSec = parseFloat(duration[0]);
  if (duration.length > 1) {
    durationSec = durationSec * 60 + parseFloat(duration[1]);
  }
  websocket.send("startTimeout", durationSec);
};

const confirmCommit = function () {
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
const setTestMatchName = function () {
  websocket.send("setTestMatchName", $("#testMatchName").val());
};

// Returns the integer team number entered into the team number input box for the given station, or 0 if it is empty.
const getTeamNumber = function (station) {
  const teamId = $(`#status${station} .team-number`).val().trim();
  return teamId ? parseInt(teamId) : 0;
}

// Handles a websocket message to update the team connection status.
const handleArenaStatus = function (data) {
  // Update the team status view.
  $.each(data.AllianceStations, function (station, stationStatus) {
    const wifiStatus = stationStatus.WifiStatus;
    $("#status" + station + " .radio-status").text(wifiStatus.TeamId);

    if (stationStatus.DsConn) {
      // Format the driver station status box.
      const dsConn = stationStatus.DsConn;
      $("#status" + station + " .ds-status").attr("data-status-ok", dsConn.DsLinked);
      if (dsConn.DsLinked) {
        $("#status" + station + " .ds-status").text(wifiStatus.MBits.toFixed(2) + "Mb");
      } else {
        $("#status" + station + " .ds-status").text("");
      }

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
    }

    // Format the radio status box according to whether the AP is configured with the correct SSID and the connection
    // status of the robot radio.
    const expectedTeamId = stationStatus.Team ? stationStatus.Team.Id : 0;
    let radioStatus = 0;
    if (expectedTeamId === wifiStatus.TeamId) {
      if (wifiStatus.RadioLinked || stationStatus.DsConn?.RobotLinked) {
        radioStatus = 2;
      } else {
        radioStatus = 1;
      }
    }
    $(`#status${station} .radio-status`).attr("data-status-ternary", radioStatus);

    if (stationStatus.EStop) {
      $("#status" + station + " .bypass-status").attr("data-status-ok", false);
      $("#status" + station + " .bypass-status").text("ES");
    } else if (stationStatus.AStop) {
      $("#status" + station + " .bypass-status").attr("data-status-ok", true);
      $("#status" + station + " .bypass-status").text("AS");
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
      $("#showOverlay").prop("disabled", true);
      $("#introRadio").prop("disabled", true);
      $("#showFinalScore").prop("disabled", true);
      $("#scoreRadio").prop("disabled", true);
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
      $("#showOverlay").prop("disabled", true);
      $("#introRadio").prop("disabled", true);
      $("#showFinalScore").prop("disabled", true);
      $("#scoreRadio").prop("disabled", true);
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
      $("#showOverlay").prop("disabled", true);
      $("#introRadio").prop("disabled", true);
      $("#showFinalScore").prop("disabled", false);
      $("#scoreRadio").prop("disabled", false);
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
      $("#showOverlay").prop("disabled", false);
      $("#introRadio").prop("disabled", false);
      $("#showFinalScore").prop("disabled", false);
      $("#scoreRadio").prop("disabled", false);
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

  $("#accessPointStatus").attr("data-status", data.AccessPointStatus);
  $("#switchStatus").attr("data-status", data.SwitchStatus);
  $("#redSCCStatus").attr("data-status", data.RedSCCStatus);
  $("#blueSCCStatus").attr("data-status", data.BlueSCCStatus);

  if (data.PlcIsHealthy) {
    $("#plcStatus").text("Connected");
    $("#plcStatus").attr("data-ready", true);
  } else {
    $("#plcStatus").text("Not Connected");
    $("#plcStatus").attr("data-ready", false);
  }
  $("#fieldEStop").attr("data-ready", !data.FieldEStop);
  $.each(data.PlcArmorBlockStatuses, function (name, status) {
    $("#plc" + name + "Status").attr("data-ready", status);
  });
};

// Handles a websocket message to update the teams for the current match.
const handleMatchLoad = function (data) {
  isReplay = data.IsReplay;

  fetch("/match_play/match_load")
    .then(response => response.text())
    .then(html => $("#matchListColumn").html(html));

  $("#matchName").text(data.Match.LongName);
  $("#testMatchName").val(data.Match.LongName);
  $("#testMatchSettings").toggle(data.Match.Type === matchTypeTest);
  $.each(data.Teams, function (station, team) {
    const teamId = $(`#status${station} .team-number`);
    teamId.val(team ? team.Id : "");
    teamId.prop("disabled", !data.AllowSubstitution);
  });
  $("#playoffRedAllianceInfo").html(formatPlayoffAllianceInfo(data.Match.PlayoffRedAlliance, data.RedOffFieldTeams));
  $("#playoffBlueAllianceInfo").html(formatPlayoffAllianceInfo(data.Match.PlayoffBlueAlliance, data.BlueOffFieldTeams));

  $("#substituteTeams").prop("disabled", true);
  $("#showOverlay").prop("disabled", false);
  $("#introRadio").prop("disabled", false);
  $("#muteMatchSounds").prop("checked", false);
}

// Handles a websocket message to update the match time countdown.
const handleMatchTime = function (data) {
  translateMatchTime(data, function (matchState, matchStateText, countdownSec) {
    $("#matchState").text(matchStateText);
    $("#matchTime").text(countdownSec);
  });
};

// Handles a websocket message to update the match score.
const handleRealtimeScore = function (data) {
  $("#redScore").text(data.Red.ScoreSummary.Score);
  $("#blueScore").text(data.Blue.ScoreSummary.Score);
};

// Handles a websocket message to populate the final score data.
const handleScorePosted = function (data) {
  let matchName = data.Match.LongName;
  if (matchName) {
    $("#showFinalScore").prop("disabled", false);
    $("#scoreRadio").prop("disabled", false);
  } else {
    matchName = "None"
  }
  $("#savedMatchName").html(matchName);
}

// Handles a websocket message to update the audience display screen selector.
const handleAudienceDisplayMode = function (data) {
  $("input[name=audienceDisplay]:checked").prop("checked", false);
  $("input[name=audienceDisplay][value=" + data + "]").prop("checked", true);
};

// Handles a websocket message to signal whether the referee and scorers have committed after the match.
const handleScoringStatus = function (data) {
  scoreIsReady = data.RefereeScoreReady;
  for (const status of Object.values(data.PositionStatuses)) {
    if (!status.Ready) {
      scoreIsReady = false;
      break;
    }
  }
  $("#refereeScoreStatus").attr("data-ready", data.RefereeScoreReady);
  updateScoreStatus(data, "red_near", "#redNearScoreStatus", "Red Near");
  updateScoreStatus(data, "red_far", "#redFarScoreStatus", "Red Far");
  updateScoreStatus(data, "blue_near", "#blueNearScoreStatus", "Blue Near");
  updateScoreStatus(data, "blue_far", "#blueFarScoreStatus", "Blue Far");
};

// Helper function to update a badge that shows scoring panel commit status.
const updateScoreStatus = function (data, position, element, displayName) {
  const status = data.PositionStatuses[position];
  $(element).text(`${displayName} ${status.NumPanelsReady}/${status.NumPanels}`);
  $(element).attr("data-present", status.NumPanels > 0);
  $(element).attr("data-ready", status.Ready);
};

// Handles a websocket message to update the alliance station display screen selector.
const handleAllianceStationDisplayMode = function (data) {
  $("input[name=allianceStationDisplay]:checked").prop("checked", false);
  $("input[name=allianceStationDisplay][value=" + data + "]").prop("checked", true);
};

// Handles a websocket message to update the event status message.
const handleEventStatus = function (data) {
  if (data.CycleTime === "") {
    $("#cycleTimeMessage").text("Last cycle time: Unknown");
  } else {
    $("#cycleTimeMessage").text("Last cycle time: " + data.CycleTime);
  }
  $("#earlyLateMessage").text(data.EarlyLateMessage);
};

const formatPlayoffAllianceInfo = function (allianceNumber, offFieldTeams) {
  if (allianceNumber === 0) {
    return "";
  }
  let allianceInfo = `<b>Alliance ${allianceNumber}</b>`;
  if (offFieldTeams.length > 0) {
    allianceInfo += ` (not on field: ${offFieldTeams.map(team => team.Id).join(", ")})`;
  }
  return allianceInfo;
}

$(function () {
  // Activate tooltips above the status headers.
  const tooltipTriggerList = document.querySelectorAll("[data-bs-toggle=tooltip]");
  const tooltipList = [...tooltipTriggerList].map(element => new bootstrap.Tooltip(element));

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/match_play/websocket", {
    allianceStationDisplayMode: function (event) {
      handleAllianceStationDisplayMode(event.data);
    },
    arenaStatus: function (event) {
      handleArenaStatus(event.data);
    },
    audienceDisplayMode: function (event) {
      handleAudienceDisplayMode(event.data);
    },
    eventStatus: function (event) {
      handleEventStatus(event.data);
    },
    matchLoad: function (event) {
      handleMatchLoad(event.data);
    },
    matchTime: function (event) {
      handleMatchTime(event.data);
    },
    matchTiming: function (event) {
      handleMatchTiming(event.data);
    },
    realtimeScore: function (event) {
      handleRealtimeScore(event.data);
    },
    scorePosted: function (event) {
      handleScorePosted(event.data);
    },
    scoringStatus: function (event) {
      handleScoringStatus(event.data);
    },
  });
});
