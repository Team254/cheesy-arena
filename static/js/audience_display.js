// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Author: nick@team254.com (Nick Eyre)
//
// Client-side methods for the audience display.

var websocket;
var transitionMap;
var currentScreen = "blank";
var redSide;
var blueSide;
var overlayCenteringHideParams;
var overlayCenteringShowParams;
var allianceSelectionTemplate = Handlebars.compile($("#allianceSelectionTemplate").html());
var sponsorImageTemplate = Handlebars.compile($("#sponsorImageTemplate").html());
var sponsorTextTemplate = Handlebars.compile($("#sponsorTextTemplate").html());

// Constants for overlay positioning. The CSS is the source of truth for the values that represent initial state.
var overlayCenteringTopUp = "-130px";
var overlayCenteringBottomHideParams = {queue: false, bottom: $("#overlayCentering").css("bottom")};
var overlayCenteringBottomShowParams = {queue: false, bottom: "0px"};
var overlayCenteringTopHideParams = {queue: false, top: overlayCenteringTopUp};
var overlayCenteringTopShowParams = {queue: false, top: "50px"};
var eventMatchInfoDown = "30px";
var eventMatchInfoUp = $("#eventMatchInfo").css("height");
var logoUp = "10px";
var logoDown = $("#logo").css("top");
var scoreIn = $(".score").css("width");
var scoreMid = "135px";
var scoreOut = "255px";
var scoreFieldsOut = "40px";

// Handles a websocket message to change which screen is displayed.
var handleAudienceDisplayMode = function(targetScreen) {
  if (targetScreen === currentScreen) {
    return;
  }

  if (targetScreen === "sponsor") {
    initializeSponsorDisplay();
  }

  transitions = transitionMap[currentScreen][targetScreen];
  if (transitions == null) {
    // There is no direct transition defined; need to go to the blank screen first.
    transitions = function() {
      transitionMap[currentScreen]["blank"](transitionMap["blank"][targetScreen]);
    };
  }
  transitions();

  currentScreen = targetScreen;
};

// Handles a websocket message to update the teams for the current match.
var handleMatchLoad = function(data) {
  $("#" + redSide + "Team1").text(data.Match.Red1);
  $("#" + redSide + "Team2").text(data.Match.Red2);
  $("#" + redSide + "Team3").text(data.Match.Red3);
  $("#" + redSide + "Team1Avatar").attr("src", getAvatarUrl(data.Match.Red1));
  $("#" + redSide + "Team2Avatar").attr("src", getAvatarUrl(data.Match.Red2));
  $("#" + redSide + "Team3Avatar").attr("src", getAvatarUrl(data.Match.Red3));
  $("#" + blueSide + "Team1").text(data.Match.Blue1);
  $("#" + blueSide + "Team2").text(data.Match.Blue2);
  $("#" + blueSide + "Team3").text(data.Match.Blue3);
  $("#" + blueSide + "Team1Avatar").attr("src", getAvatarUrl(data.Match.Blue1));
  $("#" + blueSide + "Team2Avatar").attr("src", getAvatarUrl(data.Match.Blue2));
  $("#" + blueSide + "Team3Avatar").attr("src", getAvatarUrl(data.Match.Blue3));
  $("#matchName").text(data.MatchType + " " + data.Match.DisplayName);
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

    // Set opacity of auxiliary score fields based on whether the match is in auto or teleop period.
    var autoOpacity = 1;
    var teleopOpacity = 0.5;
    if (matchStates[data.MatchState] === "TELEOP_PERIOD" || matchStates[data.MatchState] === "POST_MATCH") {
      autoOpacity = 0.5;
      teleopOpacity = 1;
    }
    $("#" + redSide + "AutoCargoRemaining").css("opacity", autoOpacity);
    $("#" + redSide + "TeleopCargoRemaining").css("opacity", teleopOpacity);
    $("#" + blueSide + "AutoCargoRemaining").css("opacity", autoOpacity);
    $("#" + blueSide + "TeleopCargoRemaining").css("opacity", teleopOpacity);
  });
};

// Handles a websocket message to update the match score.
var handleRealtimeScore = function(data) {
  $("#" + redSide + "ScoreNumber").text(data.Red.ScoreSummary.Score - data.Red.ScoreSummary.HangarPoints);
  $("#" + blueSide + "ScoreNumber").text(data.Blue.ScoreSummary.Score - data.Blue.ScoreSummary.HangarPoints);

  $("#" + redSide + "AutoCargoRemaining").text(data.Red.ScoreSummary.AutoCargoRemaining);
  $("#" + redSide + "TeleopCargoRemaining").text(data.Red.ScoreSummary.TeleopCargoRemaining);
  $("#" + blueSide + "AutoCargoRemaining").text(data.Blue.ScoreSummary.AutoCargoRemaining);
  $("#" + blueSide + "TeleopCargoRemaining").text(data.Blue.ScoreSummary.TeleopCargoRemaining);
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
  $("#finalSeriesStatus").text(data.SeriesStatus);
  $("#finalSeriesStatus").attr("data-leader", data.SeriesLeader);
  $("#finalMatchName").text(data.MatchType + " " + data.Match.DisplayName);
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
    var numColumns = alliances[0].length + 1;
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

var transitionIntroToInMatch = function(callback) {
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

var transitionBlankToInMatch = function(callback) {
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

var transitionInMatchToIntro = function(callback) {
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

var transitionInMatchToBlank = function(callback) {
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

var transitionBlankToLogoLuma = function(callback) {
  $(".blindsCenter.full").transition({ queue: false, rotateY: "0deg" }, 1000, "ease", callback);
};

var transitionLogoLumaToBlank = function(callback) {
  $(".blindsCenter.full").transition({queue: false, rotateY: "180deg"}, 1000, "ease", callback);
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

var transitionLogoToScore = function(callback) {
  $(".blindsCenter.full").transition({queue: false, top: "-350px"}, 625, "ease");
  $("#finalScore").show();
  $("#finalScore").transition({queue: false, opacity: 1}, 1000, "ease", callback);
};

var transitionBlankToScore = function(callback) {
  transitionBlankToLogo(function() {
    setTimeout(function() { transitionLogoToScore(callback); }, 50);
  });
};

var transitionScoreToLogo = function(callback) {
  $("#finalScore").transition({queue: false, opacity: 0}, 500, "ease", function(){
    $("#finalScore").hide();
  });
  $(".blindsCenter.full").transition({queue: false, top: 0}, 625, "ease", callback);
};

var transitionScoreToBlank = function(callback) {
  transitionScoreToLogo(function() {
    transitionLogoToBlank(callback);
  });
};

var transitionBlankToAllianceSelection = function(callback) {
  $('#allianceSelectionCentering').css("right","-60em").show();
  $('#allianceSelectionCentering').transition({queue: false, right: "3em"}, 500, "ease", callback);
};

var transitionAllianceSelectionToBlank = function(callback) {
  $('#allianceSelectionCentering').transition({queue: false, right: "-60em"}, 500, "ease", callback);
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

var transitionLogoToSponsor = function(callback) {
  $(".blindsCenter.full").transition({queue: false, rotateY: "90deg"}, 750, "ease", function () {
    $("#sponsor").show();
    $("#sponsor").transition({queue: false, opacity: 1}, 1000, "ease", callback);
  });
};

var transitionSponsorToLogo = function(callback) {
  $("#sponsor").transition({queue: false, opacity: 0}, 1000, "ease", function() {
    $(".blindsCenter.full").transition({queue: false, rotateY: "0deg"}, 750, "ease", callback);
    $("#sponsor").hide();
  });
};

var transitionScoreToSponsor = function(callback) {
  transitionScoreToLogo(function() {
    transitionLogoToSponsor(callback);
  });
};

var transitionSponsorToScore = function(callback) {
  transitionSponsorToLogo(function() {
    transitionLogoToScore(callback);
  });
};

var transitionBlankToTimeout = function(callback) {
  $("#overlayCentering").transition(overlayCenteringShowParams, 500, "ease", function () {
    $("#logo").transition({queue: false, top: logoUp}, 500, "ease", function() {
      $("#matchTime").transition({queue: false, opacity: 1}, 750, "ease", callback);
    });
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
    blank: {
      intro: transitionBlankToIntro,
      match: transitionBlankToInMatch,
      score: transitionBlankToScore,
      logo: transitionBlankToLogo,
      sponsor: transitionBlankToSponsor,
      allianceSelection: transitionBlankToAllianceSelection,
      timeout: transitionBlankToTimeout,
      logoluma: transitionBlankToLogoLuma
    },
    intro: {
      blank: transitionIntroToBlank,
      match: transitionIntroToInMatch,
      timeout: transitionIntroToTimeout
    },
    match: {
      blank: transitionInMatchToBlank,
      intro: transitionInMatchToIntro
    },
    score: {
      blank: transitionScoreToBlank,
      logo: transitionScoreToLogo,
      sponsor: transitionScoreToSponsor
    },
    logo: {
      blank: transitionLogoToBlank,
      score: transitionLogoToScore,
      sponsor: transitionLogoToSponsor
    },
    sponsor: {
      blank: transitionSponsorToBlank,
      logo: transitionSponsorToLogo,
      score: transitionSponsorToScore
    },
    allianceSelection: {
      blank: transitionAllianceSelectionToBlank
    },
    timeout: {
      blank: transitionTimeoutToBlank,
      intro: transitionTimeoutToIntro
    },
    logoluma: {
      blank: transitionLogoLumaToBlank
    }
  }
});
