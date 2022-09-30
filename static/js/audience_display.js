// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Author: nick@team254.com (Nick Eyre)
//
// Client-side methods for the audience display.

var websocket;
var transitionMap;
const transitionQueue = [];
let transitionInProgress = false;
var currentScreen = "blank";
var redSide;
var blueSide;
var currentMatch;
var overlayCenteringHideParams;
var overlayCenteringShowParams;
var allianceSelectionTemplate = Handlebars.compile($("#allianceSelectionTemplate").html());
var sponsorImageTemplate = Handlebars.compile($("#sponsorImageTemplate").html());
var sponsorTextTemplate = Handlebars.compile($("#sponsorTextTemplate").html());

// Constants for overlay positioning. The CSS is the source of truth for the values that represent initial state.
const overlayCenteringTopUp = "-130px";
const overlayCenteringBottomHideParams = {queue: false, bottom: $("#overlayCentering").css("bottom")};
const overlayCenteringBottomShowParams = {queue: false, bottom: "0px"};
const overlayCenteringTopHideParams = {queue: false, top: overlayCenteringTopUp};
const overlayCenteringTopShowParams = {queue: false, top: "50px"};
const eventMatchInfoDown = "30px";
const eventMatchInfoUp = $("#eventMatchInfo").css("height");
const logoUp = "10px";
const logoDown = $("#logo").css("top");
const scoreIn = $(".score").css("width");
const scoreMid = "135px";
const scoreOut = "255px";
const scoreFieldsOut = "40px";
const scoreLogoTop = "-350px";
const bracketLogoTop = "-780px";
const bracketLogoScale = 0.75;

// Handles a websocket message to change which screen is displayed.
var handleAudienceDisplayMode = function(targetScreen) {
  transitionQueue.push(targetScreen);
  executeTransitionQueue();
};

// Sequentially executes all transitions in the queue. Returns without doing anything if another invocation is already
// in progress.
const executeTransitionQueue = function() {
  if (transitionInProgress) {
    // There is an existing invocation of this method which will execute all transitions in the queue.
    return;
  }

  if (transitionQueue.length > 0) {
    transitionInProgress = true;
    const targetScreen = transitionQueue.shift();
    const callback = function() {
      // When the current transition is complete, call this method again to invoke the next one in the queue.
      currentScreen = targetScreen;
      transitionInProgress = false;
      setTimeout(executeTransitionQueue, 100);  // A small delay is needed to avoid visual glitches.
    };

    if (targetScreen === currentScreen) {
      callback();
      return;
    }

    if (targetScreen === "sponsor") {
      initializeSponsorDisplay();
    }

    let transitions = transitionMap[currentScreen][targetScreen];
    if (transitions !== undefined) {
      transitions(callback);
    } else {
      // There is no direct transition defined; need to go to the blank screen first.
      transitionMap[currentScreen]["blank"](function() {
        transitionMap["blank"][targetScreen](callback);
      });
    }
  }
};

// Handles a websocket message to update the teams for the current match.
var handleMatchLoad = function(data) {
  currentMatch = data.Match;
  $("#" + redSide + "Team1").text(currentMatch.Red1);
  $("#" + redSide + "Team2").text(currentMatch.Red2);
  $("#" + redSide + "Team3").text(currentMatch.Red3);
  $("#" + redSide + "Team1Avatar").attr("src", getAvatarUrl(currentMatch.Red1));
  $("#" + redSide + "Team2Avatar").attr("src", getAvatarUrl(currentMatch.Red2));
  $("#" + redSide + "Team3Avatar").attr("src", getAvatarUrl(currentMatch.Red3));
  $("#" + blueSide + "Team1").text(currentMatch.Blue1);
  $("#" + blueSide + "Team2").text(currentMatch.Blue2);
  $("#" + blueSide + "Team3").text(currentMatch.Blue3);
  $("#" + blueSide + "Team1Avatar").attr("src", getAvatarUrl(currentMatch.Blue1));
  $("#" + blueSide + "Team2Avatar").attr("src", getAvatarUrl(currentMatch.Blue2));
  $("#" + blueSide + "Team3Avatar").attr("src", getAvatarUrl(currentMatch.Blue3));

  // Show alliance numbers if this is an elimination match.
  if (currentMatch.Type === "elimination") {
    $("#" + redSide + "ElimAlliance").text(currentMatch.ElimRedAlliance);
    $("#" + blueSide + "ElimAlliance").text(currentMatch.ElimBlueAlliance);
    $(".elim-alliance").show();

    // Show the series status if this playoff round isn't just a single match.
    if (data.Matchup.NumWinsToAdvance > 1) {
      $("#" + redSide + "ElimAllianceWins").text(data.Matchup.RedAllianceWins);
      $("#" + blueSide + "ElimAllianceWins").text(data.Matchup.BlueAllianceWins);
      $("#elimSeriesStatus").css("display", "flex");
    } else {
      $("#elimSeriesStatus").hide();
    }
  } else {
    $("#" + redSide + "ElimAlliance").text("");
    $("#" + blueSide + "ElimAlliance").text("");
    $(".elim-alliance").hide();
    $("#elimSeriesStatus").hide();
  }

  if (data.Match.Type === "test") {
    $("#matchName").text(currentMatch.DisplayName);
  } else {
    $("#matchName").text(data.MatchType + " " + currentMatch.DisplayName);
  }
};

// Handles a websocket message to update the match time countdown.
var handleMatchTime = function(data) {
  translateMatchTime(data, function(matchState, matchStateText, countdownSec) {
    var countdownString = String(countdownSec % 60);
    if (countdownString.length === 1) {
      countdownString = "0" + countdownString;
    }
    countdownString = Math.floor(countdownSec / 60) + ":" + countdownString;
    $("#matchTime").text(countdownString);
  });
};

// Handles a websocket message to update the match score.
var handleRealtimeScore = function(data) {
  $("#" + redSide + "ScoreNumber").text(data.Red.ScoreSummary.Score - data.Red.ScoreSummary.HangarPoints);
  $("#" + blueSide + "ScoreNumber").text(data.Blue.ScoreSummary.Score - data.Blue.ScoreSummary.HangarPoints);

  $("#" + redSide + "CargoNumerator").text(data.Red.ScoreSummary.CargoCount);
  $("#" + redSide + "CargoDenominator").text(data.Red.ScoreSummary.CargoGoal);
  $("#" + blueSide + "CargoNumerator").text(data.Blue.ScoreSummary.CargoCount);
  $("#" + blueSide + "CargoDenominator").text(data.Blue.ScoreSummary.CargoGoal);
  if (currentMatch.Type === "elimination") {
    $("#" + redSide + "CargoDenominator").hide();
    $("#" + blueSide + "CargoDenominator").hide();
    $(".cargo-splitter").hide();
  } else {
    $("#" + redSide + "CargoDenominator").show();
    $("#" + blueSide + "CargoDenominator").show();
    $(".cargo-splitter").show();
  }
};

// Handles a websocket message to populate the final score data.
var handleScorePosted = function(data) {
  $("#" + redSide + "FinalScore").text(data.RedScoreSummary.Score);
  $("#" + redSide + "FinalTeam1").html(getRankingText(data.Match.Red1, data.Rankings) + "" + data.Match.Red1);
  $("#" + redSide + "FinalTeam2").html(getRankingText(data.Match.Red2, data.Rankings) + "" + data.Match.Red2);
  $("#" + redSide + "FinalTeam3").html(getRankingText(data.Match.Red3, data.Rankings) + "" + data.Match.Red3);
  $("#" + redSide + "FinalTeam1Avatar").attr("src", getAvatarUrl(data.Match.Red1));
  $("#" + redSide + "FinalTeam2Avatar").attr("src", getAvatarUrl(data.Match.Red2));
  $("#" + redSide + "FinalTeam3Avatar").attr("src", getAvatarUrl(data.Match.Red3));
  $("#" + redSide + "FinalTaxiPoints").text(data.RedScoreSummary.TaxiPoints);
  $("#" + redSide + "FinalCargoPoints").text(data.RedScoreSummary.CargoPoints);
  $("#" + redSide + "FinalHangarPoints").text(data.RedScoreSummary.HangarPoints);
  $("#" + redSide + "FinalFoulPoints").text(data.RedScoreSummary.FoulPoints);
  $("#" + redSide + "FinalCargoBonusRankingPoint").html(data.RedScoreSummary.CargoBonusRankingPoint ? "&#x2714;" : "&#x2718;");
  $("#" + redSide + "FinalCargoBonusRankingPoint").attr("data-checked", data.RedScoreSummary.CargoBonusRankingPoint);
  $("#" + redSide + "FinalHangarBonusRankingPoint").html(data.RedScoreSummary.HangarBonusRankingPoint ? "&#x2714;" : "&#x2718;");
  $("#" + redSide + "FinalHangarBonusRankingPoint").attr("data-checked", data.RedScoreSummary.HangarBonusRankingPoint);
  $("#" + redSide + "FinalDoubleBonusRankingPoint").html(data.RedScoreSummary.DoubleBonusRankingPoint ? "&#x2714;" : "&#x2718;");
  $("#" + redSide + "FinalDoubleBonusRankingPoint").attr("data-checked", data.RedScoreSummary.DoubleBonusRankingPoint);
  $("#" + blueSide + "FinalScore").text(data.BlueScoreSummary.Score);
  $("#" + blueSide + "FinalTeam1").html(getRankingText(data.Match.Blue1, data.Rankings) + "" + data.Match.Blue1);
  $("#" + blueSide + "FinalTeam2").html(getRankingText(data.Match.Blue2, data.Rankings) + "" + data.Match.Blue2);
  $("#" + blueSide + "FinalTeam3").html(getRankingText(data.Match.Blue3, data.Rankings) + "" + data.Match.Blue3);
  $("#" + blueSide + "FinalTeam1Avatar").attr("src", getAvatarUrl(data.Match.Blue1));
  $("#" + blueSide + "FinalTeam2Avatar").attr("src", getAvatarUrl(data.Match.Blue2));
  $("#" + blueSide + "FinalTeam3Avatar").attr("src", getAvatarUrl(data.Match.Blue3));
  $("#" + blueSide + "FinalTaxiPoints").text(data.BlueScoreSummary.TaxiPoints);
  $("#" + blueSide + "FinalCargoPoints").text(data.BlueScoreSummary.CargoPoints);
  $("#" + blueSide + "FinalHangarPoints").text(data.BlueScoreSummary.HangarPoints);
  $("#" + blueSide + "FinalFoulPoints").text(data.BlueScoreSummary.FoulPoints);
  $("#" + blueSide + "FinalCargoBonusRankingPoint").html(data.BlueScoreSummary.CargoBonusRankingPoint ? "&#x2714;" : "&#x2718;");
  $("#" + blueSide + "FinalCargoBonusRankingPoint").attr("data-checked", data.BlueScoreSummary.CargoBonusRankingPoint);
  $("#" + blueSide + "FinalHangarBonusRankingPoint").html(data.BlueScoreSummary.HangarBonusRankingPoint ? "&#x2714;" : "&#x2718;");
  $("#" + blueSide + "FinalHangarBonusRankingPoint").attr("data-checked", data.BlueScoreSummary.HangarBonusRankingPoint);
  $("#" + blueSide + "FinalDoubleBonusRankingPoint").html(data.BlueScoreSummary.DoubleBonusRankingPoint ? "&#x2714;" : "&#x2718;");
  $("#" + blueSide + "FinalDoubleBonusRankingPoint").attr("data-checked", data.BlueScoreSummary.DoubleBonusRankingPoint);
  $("#finalSeriesStatus").text(data.SeriesStatus);
  $("#finalSeriesStatus").attr("data-leader", data.SeriesLeader);
  $("#finalMatchName").text(data.MatchType + " " + data.Match.DisplayName);

  // Reload the bracket to reflect any changes.
  $("#bracketSvg").attr("src", "/api/bracket/svg?activeMatch=saved&v=" + new Date().getTime());

  if (data.Match.Type === "elimination") {
    // Hide bonus ranking points.
    $(".playoffHiddenFields").hide();
  } else {
    $(".playoffHiddenFields").show();
  }
};

// Handles a websocket message to play a sound to signal match start/stop/etc.
var handlePlaySound = function(sound) {
  $("audio").each(function(k, v) {
    // Stop and reset any sounds that are still playing.
    v.pause();
    v.currentTime = 0;
  });
  $("#sound-" + sound)[0].play();
};

// Handles a websocket message to update the alliance selection screen.
var handleAllianceSelection = function(alliances) {
  if (alliances && alliances.length > 0) {
    var numColumns = alliances[0].TeamIds.length + 1;
    $.each(alliances, function(k, v) {
      v.Index = k + 1;
    });
    $("#allianceSelection").html(allianceSelectionTemplate({alliances: alliances, numColumns: numColumns}));
  }
};

// Handles a websocket message to populate and/or show/hide a lower third.
var handleLowerThird = function(data) {
  if (data.LowerThird !== null) {
    if (data.LowerThird.BottomText === "") {
      $("#lowerThirdTop").hide();
      $("#lowerThirdBottom").hide();
      $("#lowerThirdSingle").text(data.LowerThird.TopText);
      $("#lowerThirdSingle").show();
    } else {
      $("#lowerThirdSingle").hide();
      $("#lowerThirdTop").text(data.LowerThird.TopText);
      $("#lowerThirdBottom").text(data.LowerThird.BottomText);
      $("#lowerThirdTop").show();
      $("#lowerThirdBottom").show();
    }
  }

  var lowerThirdElement = $("#lowerThird");
  if (data.ShowLowerThird && !lowerThirdElement.is(":visible")) {
    lowerThirdElement.show();
    lowerThirdElement.transition({queue: false, left: "150px"}, 750, "ease");
  } else if (!data.ShowLowerThird && lowerThirdElement.is(":visible")) {
    lowerThirdElement.transition({queue: false, left: "-1000px"}, 1000, "ease", function () {
      lowerThirdElement.hide();
    });
  }
};

var transitionAllianceSelectionToBlank = function(callback) {
  $('#allianceSelectionCentering').transition({queue: false, right: "-60em"}, 500, "ease", callback);
};

var transitionBlankToAllianceSelection = function(callback) {
  $('#allianceSelectionCentering').css("right","-60em").show();
  $('#allianceSelectionCentering').transition({queue: false, right: "3em"}, 500, "ease", callback);
};

var transitionBlankToBracket = function(callback) {
  transitionBlankToLogo(function() {
    setTimeout(function() { transitionLogoToBracket(callback); }, 50);
  });
};

var transitionBlankToIntro = function(callback) {
  $("#overlayCentering").transition(overlayCenteringShowParams, 500, "ease", function() {
    $(".teams").css("display", "flex");
    $(".avatars").css("display", "flex");
    $(".avatars").css("opacity", 1);
    $(".score").transition({queue: false, width: scoreMid}, 500, "ease", function() {
      $("#eventMatchInfo").css("display", "flex");
      $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoDown}, 500, "ease", callback);
    });
  });
};

var transitionBlankToLogo = function(callback) {
  $(".blindsCenter.blank").css({rotateY: "0deg"});
  $(".blindsCenter.full").css({rotateY: "-180deg"});
  $(".blinds.right").transition({queue: false, right: 0}, 1000, "ease");
  $(".blinds.left").transition({queue: false, left: 0}, 1000, "ease", function() {
    $(".blinds.left").addClass("full");
    $(".blinds.right").hide();
    setTimeout(function() {
      $(".blindsCenter.blank").transition({queue: false, rotateY: "180deg"}, 500, "ease");
      $(".blindsCenter.full").transition({queue: false, rotateY: "0deg"}, 500, "ease", callback);
    }, 200);
  });
};

var transitionBlankToLogoLuma = function(callback) {
  $(".blindsCenter.blank").css({rotateY: "180deg"});
  $(".blindsCenter.full").transition({ queue: false, rotateY: "0deg" }, 1000, "ease", callback);
};

var transitionBlankToMatch = function(callback) {
  $("#overlayCentering").transition(overlayCenteringShowParams, 500, "ease", function() {
    $(".teams").css("display", "flex");
    $(".score-fields").css("display", "flex");
    $(".score-fields").transition({queue: false, width: scoreFieldsOut}, 500, "ease");
    $("#logo").transition({queue: false, top: logoUp}, 500, "ease");
    $(".score").transition({queue: false, width: scoreOut}, 500, "ease", function() {
      $("#eventMatchInfo").css("display", "flex");
      $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoDown}, 500, "ease", callback);
      $(".score-number").transition({queue: false, opacity: 1}, 750, "ease");
      $("#matchTime").transition({queue: false, opacity: 1}, 750, "ease");
      $(".score-fields").transition({queue: false, opacity: 1}, 750, "ease");
    });
  });
};

var transitionBlankToScore = function(callback) {
  transitionBlankToLogo(function() {
    setTimeout(function() { transitionLogoToScore(callback); }, 50);
  });
};

var transitionBlankToSponsor = function(callback) {
  $(".blindsCenter.blank").css({rotateY: "90deg"});
  $(".blinds.right").transition({queue: false, right: 0}, 1000, "ease");
  $(".blinds.left").transition({queue: false, left: 0}, 1000, "ease", function() {
    $(".blinds.left").addClass("full");
    $(".blinds.right").hide();
    setTimeout(function() {
      $("#sponsor").show();
      $("#sponsor").transition({queue: false, opacity: 1}, 1000, "ease", callback);
    }, 200);
  });
};

var transitionBlankToTimeout = function(callback) {
  $("#overlayCentering").transition(overlayCenteringShowParams, 500, "ease", function () {
    $("#logo").transition({queue: false, top: logoUp}, 500, "ease", function() {
      $("#matchTime").transition({queue: false, opacity: 1}, 750, "ease", callback);
    });
  });
};

var transitionBracketToBlank = function(callback) {
  transitionBracketToLogo(function() {
    transitionLogoToBlank(callback);
  });
};

var transitionBracketToLogo = function(callback) {
  $("#bracket").transition({queue: false, opacity: 0}, 500, "ease", function(){
    $("#bracket").hide();
  });
  $(".blindsCenter.full").transition({queue: false, top: 0, scale: 1}, 625, "ease", callback);
};

var transitionBracketToLogoLuma = function(callback) {
  transitionBracketToLogo(function() {
    transitionLogoToLogoLuma(callback);
  });
};

var transitionBracketToScore = function(callback) {
  $(".blindsCenter.full").transition({queue: false, top: scoreLogoTop, scale: 1}, 1000, "ease");
  $("#bracket").transition({queue: false, opacity: 0}, 1000, "ease", function(){
    $("#bracket").hide();
    $("#finalScore").show();
    $("#finalScore").transition({queue: false, opacity: 1}, 1000, "ease", callback);
  });
};

var transitionBracketToSponsor = function(callback) {
  transitionBracketToLogo(function() {
    transitionLogoToSponsor(callback);
  });
};

var transitionIntroToBlank = function(callback) {
  $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoUp}, 500, "ease", function() {
    $("#eventMatchInfo").hide();
    $(".score").transition({queue: false, width: scoreIn}, 500, "ease", function() {
      $(".avatars").css("opacity", 0);
      $(".avatars").hide();
      $(".teams").hide();
      $("#overlayCentering").transition(overlayCenteringHideParams, 1000, "ease", callback);
    });
  });
};

var transitionIntroToMatch = function(callback) {
  $(".avatars").transition({queue: false, opacity: 0}, 500, "ease", function() {
    $(".avatars").hide();
  });
  $(".score-fields").css("display", "flex");
  $(".score-fields").transition({queue: false, width: scoreFieldsOut}, 500, "ease");
  $("#logo").transition({queue: false, top: logoUp}, 500, "ease");
  $(".score").transition({queue: false, width: scoreOut}, 500, "ease", function() {
    $(".score-number").transition({queue: false, opacity: 1}, 750, "ease");
    $("#matchTime").transition({queue: false, opacity: 1}, 750, "ease", callback);
    $(".score-fields").transition({queue: false, opacity: 1}, 750, "ease");
  });
};

var transitionIntroToTimeout = function(callback) {
  $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoUp}, 500, "ease", function() {
    $("#eventMatchInfo").hide();
    $(".score").transition({queue: false, width: scoreIn}, 500, "ease", function() {
      $(".avatars").css("opacity", 0);
      $(".avatars").hide();
      $(".teams").hide();
      $("#logo").transition({queue: false, top: logoUp}, 500, "ease", function() {
        $("#matchTime").transition({queue: false, opacity: 1}, 750, "ease", callback);
      });
    });
  });
};

var transitionLogoToBlank = function(callback) {
  $(".blindsCenter.blank").transition({queue: false, rotateY: "360deg"}, 500, "ease");
  $(".blindsCenter.full").transition({queue: false, rotateY: "180deg"}, 500, "ease", function() {
    setTimeout(function() {
      $(".blinds.left").removeClass("full");
      $(".blinds.right").show();
      $(".blinds.right").transition({queue: false, right: "-50%"}, 1000, "ease");
      $(".blinds.left").transition({queue: false, left: "-50%"}, 1000, "ease", callback);
    }, 200);
  });
};

var transitionLogoToBracket = function(callback) {
  $(".blindsCenter.full").transition({queue: false, top: bracketLogoTop, scale: bracketLogoScale}, 625, "ease");
  $("#bracket").show();
  $("#bracket").transition({queue: false, opacity: 1}, 1000, "ease", callback);
};

var transitionLogoToLogoLuma = function(callback) {
  $(".blinds.left").removeClass("full");
  $(".blinds.right").show();
  $(".blinds.right").transition({queue: false, right: "-50%"}, 1000, "ease");
  $(".blinds.left").transition({queue: false, left: "-50%"}, 1000, "ease", function() {
    if (callback) {
      callback();
    }
  });
};

var transitionLogoToScore = function(callback) {
  $(".blindsCenter.full").transition({queue: false, top: scoreLogoTop}, 625, "ease");
  $("#finalScore").show();
  $("#finalScore").transition({queue: false, opacity: 1}, 1000, "ease", callback);
};

var transitionLogoToSponsor = function(callback) {
  $(".blindsCenter.full").transition({queue: false, rotateY: "90deg"}, 750, "ease", function () {
    $("#sponsor").show();
    $("#sponsor").transition({queue: false, opacity: 1}, 1000, "ease", callback);
  });
};

var transitionLogoLumaToBlank = function(callback) {
  $(".blindsCenter.full").transition({queue: false, rotateY: "180deg"}, 1000, "ease", callback);
};

var transitionLogoLumaToBracket = function(callback) {
  transitionLogoLumaToLogo(function() {
    transitionLogoToBracket(callback);
  });
};

var transitionLogoLumaToLogo = function(callback) {
  $(".blinds.right").transition({queue: false, right: 0}, 1000, "ease");
  $(".blinds.left").transition({queue: false, left: 0}, 1000, "ease", function() {
    $(".blinds.left").addClass("full");
    $(".blinds.right").hide();
    if (callback) {
      callback();
    }
  });
};

var transitionLogoLumaToScore = function(callback) {
  transitionLogoLumaToLogo(function() {
    transitionLogoToScore(callback);
  });
};

var transitionMatchToBlank = function(callback) {
  $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoUp}, 500, "ease");
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear");
  $(".score-fields").transition({queue: false, opacity: 0}, 300, "ease");
  $(".score-number").transition({queue: false, opacity: 0}, 300, "linear", function() {
    $("#eventMatchInfo").hide();
    $(".score-fields").transition({queue: false, width: 0}, 500, "ease");
    $("#logo").transition({queue: false, top: logoDown}, 500, "ease");
    $(".score").transition({queue: false, width: scoreIn}, 500, "ease", function() {
      $(".teams").hide();
      $(".score-fields").hide();
      $("#overlayCentering").transition(overlayCenteringHideParams, 1000, "ease", callback);
    });
  });
};

var transitionMatchToIntro = function(callback) {
  $(".score-number").transition({queue: false, opacity: 0}, 300, "linear");
  $(".score-fields").transition({queue: false, opacity: 0}, 300, "ease");
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear", function() {
    $(".score-fields").transition({queue: false, width: 0}, 500, "ease");
    $("#logo").transition({queue: false, top: logoDown}, 500, "ease");
    $(".score").transition({queue: false, width: scoreMid}, 500, "ease", function() {
      $(".score-fields").hide();
      $(".avatars").css("display", "flex");
      $(".avatars").transition({queue: false, opacity: 1}, 500, "ease", callback);
    });
  });
};

var transitionScoreToBlank = function(callback) {
  transitionScoreToLogo(function() {
    transitionLogoToBlank(callback);
  });
};

var transitionScoreToBracket = function(callback) {
  $(".blindsCenter.full").transition({queue: false, top: bracketLogoTop, scale: bracketLogoScale}, 1000, "ease");
  $("#finalScore").transition({queue: false, opacity: 0}, 1000, "ease", function(){
    $("#finalScore").hide();
    $("#bracket").show();
    $("#bracket").transition({queue: false, opacity: 1}, 1000, "ease", callback);
  });
};

var transitionScoreToLogo = function(callback) {
  $("#finalScore").transition({queue: false, opacity: 0}, 500, "ease", function(){
    $("#finalScore").hide();
  });
  $(".blindsCenter.full").transition({queue: false, top: 0}, 625, "ease", callback);
};

var transitionScoreToLogoLuma = function(callback) {
  transitionScoreToLogo(function() {
    transitionLogoToLogoLuma(callback);
  });
};

var transitionScoreToSponsor = function(callback) {
  transitionScoreToLogo(function() {
    transitionLogoToSponsor(callback);
  });
};

var transitionSponsorToBlank = function(callback) {
  $("#sponsor").transition({queue: false, opacity: 0}, 1000, "ease", function() {
    setTimeout(function() {
      $(".blinds.left").removeClass("full");
      $(".blinds.right").show();
      $(".blinds.right").transition({queue: false, right: "-50%"}, 1000, "ease");
      $(".blinds.left").transition({queue: false, left: "-50%"}, 1000, "ease", callback);
      $("#sponsor").hide();
    }, 200);
  });
};

var transitionSponsorToBracket = function(callback) {
  transitionSponsorToLogo(function() {
    transitionLogoToBracket(callback);
  });
};

var transitionSponsorToLogo = function(callback) {
  $("#sponsor").transition({queue: false, opacity: 0}, 1000, "ease", function() {
    $(".blindsCenter.full").transition({queue: false, rotateY: "0deg"}, 750, "ease", callback);
    $("#sponsor").hide();
  });
};

var transitionSponsorToScore = function(callback) {
  transitionSponsorToLogo(function() {
    transitionLogoToScore(callback);
  });
};

var transitionTimeoutToBlank = function(callback) {
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear", function() {
    $("#logo").transition({queue: false, top: logoDown}, 500, "ease", function() {
      $("#overlayCentering").transition(overlayCenteringHideParams, 1000, "ease", callback);
    });
  });
};

var transitionTimeoutToIntro = function(callback) {
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear", function() {
    $("#logo").transition({queue: false, top: logoDown}, 500, "ease", function() {
      $(".avatars").css("display", "flex");
      $(".avatars").css("opacity", 1);
      $(".teams").css("display", "flex");
      $(".score").transition({queue: false, width: scoreMid}, 500, "ease", function () {
        $("#eventMatchInfo").show();
        $("#eventMatchInfo").transition({queue: false, height: eventMatchInfoDown}, 500, "ease", callback);
      });
    });
  });
};

// Loads sponsor slide data and builds the slideshow HTML.
var initializeSponsorDisplay = function() {
  $.getJSON("/api/sponsor_slides", function(slides) {
    $("#sponsorContainer").empty();

    // Inject the HTML for each slide into the DOM.
    $.each(slides, function(index, slide) {
      slide.DisplayTimeMs = slide.DisplayTimeSec * 1000;
      slide.First = index === 0;

      var slideHtml;
      if (slide.Image) {
        slideHtml = sponsorImageTemplate(slide);
      } else {
        slideHtml = sponsorTextTemplate(slide);
      }
      $("#sponsorContainer").append(slideHtml);
    });

    // Start Carousel
    var t;
    var start = $('.carousel#sponsor').find('.active').attr('data-interval');
    t = setTimeout("$('.carousel#sponsor').carousel({interval: 1000});", start-1000);

    $('.carousel#sponsor').on('slid.bs.carousel', function () {   
         clearTimeout(t);  
         var duration = $(this).find('.active').attr('data-interval');

         $('.carousel#sponsor').carousel('pause');
         t = setTimeout("$('.carousel#sponsor').carousel();", duration-1000);
    });

    $('.carousel-control.right').on('click', function(){
        clearTimeout(t);   
    });

    $('.carousel-control.left').on('click', function(){
        clearTimeout(t);   
    });

  });
};

var getAvatarUrl = function(teamId) {
  return "/api/teams/" + teamId + "/avatar";
};

var getRankingText = function(teamId, rankings) {
  var ranking = rankings[teamId];
  if (ranking === undefined || ranking.Rank === 0) {
    return "<div class='rank-spacer'></div>";
  }

  if (ranking.Rank > ranking.PreviousRank && ranking.PreviousRank > 0) {
    return "<div class='rank-box rank-down'>" + ranking.Rank + "</div><div class='arrow-down'></div>";
  } else if (ranking.Rank < ranking.PreviousRank) {
    return "<div class='rank-box rank-up'>" + ranking.Rank + "</div><div class='arrow-up'></div>";
  }
  return "<div class='rank-box rank-same'>" + ranking.Rank + "</div>";
};

$(function() {
  // Read the configuration for this display from the URL query string.
  var urlParams = new URLSearchParams(window.location.search);
  document.body.style.backgroundColor = urlParams.get("background");
  var reversed = urlParams.get("reversed");
  if (reversed === "true") {
    redSide = "right";
    blueSide = "left";
  } else {
    redSide = "left";
    blueSide = "right";
  }
  $(".reversible-left").attr("data-reversed", reversed);
  $(".reversible-right").attr("data-reversed", reversed);
  if (urlParams.get("overlayLocation") === "top") {
    overlayCenteringHideParams = overlayCenteringTopHideParams;
    overlayCenteringShowParams = overlayCenteringTopShowParams;
    $("#overlayCentering").css("top", overlayCenteringTopUp);
  } else {
    overlayCenteringHideParams = overlayCenteringBottomHideParams;
    overlayCenteringShowParams = overlayCenteringBottomShowParams;
  }

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/audience/websocket", {
    allianceSelection: function(event) { handleAllianceSelection(event.data); },
    audienceDisplayMode: function(event) { handleAudienceDisplayMode(event.data); },
    lowerThird: function(event) { handleLowerThird(event.data); },
    matchLoad: function(event) { handleMatchLoad(event.data); },
    matchTime: function(event) { handleMatchTime(event.data); },
    matchTiming: function(event) { handleMatchTiming(event.data); },
    playSound: function(event) { handlePlaySound(event.data); },
    realtimeScore: function(event) { handleRealtimeScore(event.data); },
    scorePosted: function(event) { handleScorePosted(event.data); }
  });

  // Map how to transition from one screen to another. Missing links between screens indicate that first we
  // must transition to the blank screen and then to the target screen.
  transitionMap = {
    allianceSelection: {
      blank: transitionAllianceSelectionToBlank,
    },
    blank: {
      allianceSelection: transitionBlankToAllianceSelection,
      bracket: transitionBlankToBracket,
      intro: transitionBlankToIntro,
      logo: transitionBlankToLogo,
      logoLuma: transitionBlankToLogoLuma,
      match: transitionBlankToMatch,
      score: transitionBlankToScore,
      sponsor: transitionBlankToSponsor,
      timeout: transitionBlankToTimeout,
    },
    bracket: {
      blank: transitionBracketToBlank,
      logo: transitionBracketToLogo,
      logoLuma: transitionBracketToLogoLuma,
      score: transitionBracketToScore,
      sponsor: transitionBracketToSponsor,
    },
    intro: {
      blank: transitionIntroToBlank,
      match: transitionIntroToMatch,
      timeout: transitionIntroToTimeout,
    },
    logo: {
      blank: transitionLogoToBlank,
      bracket: transitionLogoToBracket,
      logoLuma: transitionLogoToLogoLuma,
      score: transitionLogoToScore,
      sponsor: transitionLogoToSponsor,
    },
    logoLuma: {
      blank: transitionLogoLumaToBlank,
      bracket: transitionLogoLumaToBracket,
      logo: transitionLogoLumaToLogo,
      score: transitionLogoLumaToScore,
    },
    match: {
      blank: transitionMatchToBlank,
      intro: transitionMatchToIntro,
    },
    score: {
      blank: transitionScoreToBlank,
      bracket: transitionScoreToBracket,
      logo: transitionScoreToLogo,
      logoLuma: transitionScoreToLogoLuma,
      sponsor: transitionScoreToSponsor,
    },
    sponsor: {
      blank: transitionSponsorToBlank,
      bracket: transitionSponsorToBracket,
      logo: transitionSponsorToLogo,
      score: transitionSponsorToScore,
    },
    timeout: {
      blank: transitionTimeoutToBlank,
      intro: transitionTimeoutToIntro,
    },
  }
});
