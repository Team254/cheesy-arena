// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Author: nick@team254.com (Nick Eyre)
//
// Client-side methods for the audience display.

var websocket;
var transitionMap;
var currentScreen = "blank";
var allianceSelectionTemplate = Handlebars.compile($("#allianceSelectionTemplate").html());
var sponsorImageTemplate = Handlebars.compile($("#sponsorImageTemplate").html());
var sponsorTextTemplate = Handlebars.compile($("#sponsorTextTemplate").html());

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
  $("#redTeam1").text(data.Match.Red1);
  $("#redTeam2").text(data.Match.Red2);
  $("#redTeam3").text(data.Match.Red3);
  $("#blueTeam1").text(data.Match.Blue1);
  $("#blueTeam2").text(data.Match.Blue2);
  $("#blueTeam3").text(data.Match.Blue3);
  $("#matchName").text(data.MatchName + " " + data.Match.DisplayName);
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
  var redScoreBreakdown = data.Red.RealtimeScore.CurrentScore;
  $("#redScoreNumber").text(data.Red.Score);
  $("#redForceCubesIcon").attr("data-state", data.Red.ForceState);
  $("#redForceCubes").text(redScoreBreakdown.ForceCubes).attr("data-state", data.Red.ForceState);
  $("#redLevitateCubesIcon").attr("data-state", data.Red.LevitateState);
  $("#redLevitateCubes").text(redScoreBreakdown.LevitateCubes).attr("data-state", data.Red.LevitateState);
  $("#redBoostCubesIcon").attr("data-state", data.Red.BoostState);
  $("#redBoostCubes").text(redScoreBreakdown.BoostCubes).attr("data-state", data.Red.BoostState);

  var blueScoreBreakdown = data.Blue.RealtimeScore.CurrentScore;
  $("#blueScoreNumber").text(data.Blue.Score);
  $("#blueForceCubesIcon").attr("data-state", data.Blue.ForceState);
  $("#blueForceCubes").text(blueScoreBreakdown.ForceCubes).attr("data-state", data.Blue.ForceState);
  $("#blueLevitateCubesIcon").attr("data-state", data.Blue.LevitateState);
  $("#blueLevitateCubes").text(blueScoreBreakdown.LevitateCubes).attr("data-state", data.Blue.LevitateState);
  $("#blueBoostCubesIcon").attr("data-state", data.Blue.BoostState);
  $("#blueBoostCubes").text(blueScoreBreakdown.BoostCubes).attr("data-state", data.Blue.BoostState);

  // Switch/scale indicators.
  $("#scaleIndicator").attr("data-owned-by", data.ScaleOwnedBy);
  $("#redSwitchIndicator").attr("data-owned-by", data.Red.SwitchOwnedBy);
  $("#blueSwitchIndicator").attr("data-owned-by", data.Blue.SwitchOwnedBy);

  // Power up progress bars.
  if ((data.Red.ForceState === 2 || data.Red.BoostState === 2) && $("#redProgress").height() === 0) {
    $("#redProgress").height(85);
    $("#redProgress").transition({queue: false, height: 0}, 10000, "linear");
  }
  if ((data.Blue.ForceState === 2 || data.Blue.BoostState === 2) && $("#blueProgress").height() === 0) {
    $("#blueProgress").height(85);
    $("#blueProgress").transition({queue: false, height: 0}, 10000, "linear");
  }
};

// Handles a websocket message to populate the final score data.
var handleScorePosted = function(data) {
  $("#redFinalScore").text(data.RedScoreSummary.Score);
  $("#redFinalTeam1").text(data.Match.Red1);
  $("#redFinalTeam2").text(data.Match.Red2);
  $("#redFinalTeam3").text(data.Match.Red3);
  $("#redFinalAutoRunPoints").text(data.RedScoreSummary.AutoRunPoints);
  $("#redFinalOwnershipPoints").text(data.RedScoreSummary.OwnershipPoints);
  $("#redFinalVaultPoints").text(data.RedScoreSummary.VaultPoints);
  $("#redFinalParkClimbPoints").text(data.RedScoreSummary.ParkClimbPoints);
  $("#redFinalFoulPoints").text(data.RedScoreSummary.FoulPoints);
  $("#redFinalAutoQuest").html(data.RedScoreSummary.AutoQuest ? "&#x2714;" : "&#x2718;");
  $("#redFinalAutoQuest").attr("data-checked", data.RedScoreSummary.AutoQuest);
  $("#redFinalFaceTheBoss").html(data.RedScoreSummary.FaceTheBoss ? "&#x2714;" : "&#x2718;");
  $("#redFinalFaceTheBoss").attr("data-checked", data.RedScoreSummary.FaceTheBoss);
  $("#blueFinalScore").text(data.BlueScoreSummary.Score);
  $("#blueFinalTeam1").text(data.Match.Blue1);
  $("#blueFinalTeam2").text(data.Match.Blue2);
  $("#blueFinalTeam3").text(data.Match.Blue3);
  $("#blueFinalAutoRunPoints").text(data.BlueScoreSummary.AutoRunPoints);
  $("#blueFinalOwnershipPoints").text(data.BlueScoreSummary.OwnershipPoints);
  $("#blueFinalVaultPoints").text(data.BlueScoreSummary.VaultPoints);
  $("#blueFinalParkClimbPoints").text(data.BlueScoreSummary.ParkClimbPoints);
  $("#blueFinalFoulPoints").text(data.BlueScoreSummary.FoulPoints);
  $("#blueFinalAutoQuest").html(data.BlueScoreSummary.AutoQuest ? "&#x2714;" : "&#x2718;");
  $("#blueFinalAutoQuest").attr("data-checked", data.BlueScoreSummary.AutoQuest);
  $("#blueFinalFaceTheBoss").html(data.BlueScoreSummary.FaceTheBoss ? "&#x2714;" : "&#x2718;");
  $("#blueFinalFaceTheBoss").attr("data-checked", data.BlueScoreSummary.FaceTheBoss);
  $("#finalMatchName").text(data.MatchName + " " + data.Match.DisplayName);
};

// Handles a websocket message to play a sound to signal match start/stop/etc.
var handlePlaySound = function(sound) {
  $("audio").each(function(k, v) {
    // Stop and reset any sounds that are still playing.
    v.pause();
    v.currentTime = 0;
  });
  $("#" + sound)[0].play();
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
  if (data.BottomText === "") {
    $("#lowerThirdTop").hide();
    $("#lowerThirdBottom").hide();
    $("#lowerThirdSingle").text(data.TopText);
    $("#lowerThirdSingle").show();
  } else {
    $("#lowerThirdSingle").hide();
    $("#lowerThirdTop").text(data.TopText);
    $("#lowerThirdBottom").text(data.BottomText);
    $("#lowerThirdTop").show();
    $("#lowerThirdBottom").show();
  }
};

var transitionBlankToIntro = function(callback) {
  $("#centering").transition({queue: false, bottom: "0px"}, 500, "ease", function() {
    $(".teams").transition({queue: false, width: "65px"}, 100, "linear", function() {
      $(".score").transition({queue: false, width: "120px"}, 500, "ease", function() {
        $("#eventMatchInfo").show();
        var height = -$("#eventMatchInfo").height();
        $("#eventMatchInfo").transition({queue: false, bottom: height + "px"}, 500, "ease", callback);
      });
    });
  });
};

var transitionIntroToInMatch = function(callback) {
  $("#logo").transition({queue: false, top: "10px"}, 500, "ease");
  $(".score").transition({queue: false, width: "275px"}, 500, "ease", function() {
    $(".score-number").transition({queue: false, opacity: 1}, 750, "ease");
    $(".score-fields").transition({queue: false, opacity: 1}, 750, "ease");
    $(".seesaw-indicator").transition({queue: false, opacity: 1}, 750, "ease");
    $("#matchTime").transition({queue: false, opacity: 1}, 750, "ease", callback);
  });
};

var transitionIntroToBlank = function(callback) {
  $("#eventMatchInfo").transition({queue: false, bottom: "0px"}, 500, "ease", function() {
    $("#eventMatchInfo").hide();
    $(".score").transition({queue: false, width: "0px"}, 500, "ease");
    $(".teams").transition({queue: false, width: "40px"}, 500, "ease", function() {
      $("#centering").transition({queue: false, bottom: "-340px"}, 1000, "ease", callback);
    });
  });
};

var transitionBlankToInMatch = function(callback) {
  $("#centering").transition({queue: false, bottom: "0px"}, 500, "ease", function() {
    $(".teams").transition({queue: false, width: "65px"}, 100, "linear", function() {
      $("#logo").transition({queue: false, top: "10px"}, 500, "ease");
      $(".score").transition({queue: false, width: "275px"}, 500, "ease", function() {
        $("#eventMatchInfo").show();
        $(".score-number").transition({queue: false, opacity: 1}, 750, "ease");
        $(".score-fields").transition({queue: false, opacity: 1}, 750, "ease");
        $(".seesaw-indicator").transition({queue: false, opacity: 1}, 750, "ease");
        $("#matchTime").transition({queue: false, opacity: 1}, 750, "ease", callback);
        var height = -$("#eventMatchInfo").height();
        $("#eventMatchInfo").transition({queue: false, bottom: height + "px"}, 500, "ease", callback);
      });
    });
  });
};

var transitionInMatchToIntro = function(callback) {
  $(".score-number").transition({queue: false, opacity: 0}, 300, "linear");
  $(".score-fields").transition({queue: false, opacity: 0}, 300, "linear");
  $(".seesaw-indicator").transition({queue: false, opacity: 0}, 300, "linear");
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear", function() {
    $("#logo").transition({queue: false, top: "35px"}, 500, "ease");
    $(".score").transition({queue: false, width: "120px"}, 500, "ease");
    $(".teams").transition({queue: false, width: "65px"}, 500, "ease", callback);
  });
};

var transitionInMatchToBlank = function(callback) {
  $("#eventMatchInfo").transition({queue: false, bottom: "0px"}, 500, "ease");
  $("#matchTime").transition({queue: false, opacity: 0}, 300, "linear");
  $(".score-fields").transition({queue: false, opacity: 0}, 300, "linear");
  $(".seesaw-indicator").transition({queue: false, opacity: 0}, 300, "linear");
  $(".score-number").transition({queue: false, opacity: 0}, 300, "linear", function() {
    $("#eventMatchInfo").hide();
    $("#logo").transition({queue: false, top: "35px"}, 500, "ease");
    $(".score").transition({queue: false, width: "0px"}, 500, "ease");
    $(".teams").transition({queue: false, width: "40px"}, 500, "ease", function() {
      $("#centering").transition({queue: false, bottom: "-340px"}, 1000, "ease", callback);
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

var transitionBlankToLowerThird = function(callback) {
  $("#lowerThird").show();
  $("#lowerThird").transition({queue: false, left: "150px"}, 750, "ease", callback);
};

var transitionLowerThirdToBlank = function(callback) {
  $("#lowerThird").transition({queue: false, left: "-1000px"}, 1000, "ease", function() {
    $("#lowerThird").hide();
    if (callback) {
      callback();
    }
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

$(function() {
  // Read the configuration for this display from the URL query string.
  var urlParams = new URLSearchParams(window.location.search);
  document.body.style.backgroundColor = urlParams.get("background");

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
      lowerThird: transitionBlankToLowerThird
    },
    intro: {
      blank: transitionIntroToBlank,
      match: transitionIntroToInMatch
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
    lowerThird: {
      blank: transitionLowerThirdToBlank
    }
  }
});
