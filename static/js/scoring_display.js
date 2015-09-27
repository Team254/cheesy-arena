// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the scoring interface.

var websocket;
var selectedStack = 0;
var numStacks = 10;
var stacks;
var stackScoreChanged = false;
var scoreCommitted = false;

function Stack() {
  this.Totes = 0;
  this.Container = false;
  this.Litter = false;
}

// Handles a websocket message to update the realtime scoring fields.
var handleScore = function(data) {
  // Update autonomous period values.
  var score = data.CurrentScore;
  $("#autoRobotSet").text(score.AutoRobotSet ? "Yes" : "No");
  $("#autoRobotSet").attr("data-value", score.AutoRobotSet);
  $("#autoContainerSet").text(score.AutoContainerSet ? "Yes" : "No");
  $("#autoContainerSet").attr("data-value", score.AutoContainerSet);
  $("#autoToteSet").text(score.AutoToteSet ? "Yes" : "No");
  $("#autoToteSet").attr("data-value", score.AutoToteSet);
  $("#autoStackedToteSet").text(score.AutoStackedToteSet ? "Yes" : "No");
  $("#autoStackedToteSet").attr("data-value", score.AutoStackedToteSet);

  // Update teleoperated period values.
  $("#coopertitionSet").text(score.CoopertitionSet ? "Yes" : "No");
  $("#coopertitionSet").attr("data-value", score.CoopertitionSet);
  $("#coopertitionStack").text(score.CoopertitionStack ? "Yes" : "No");
  $("#coopertitionStack").attr("data-value", score.CoopertitionStack);

  // Don't stomp on pending changes to the stack score.
  if (stackScoreChanged == false) {
    if (score.Stacks == null) {
      stacks = new Array();
      for (i = 0; i < numStacks; i++) {
        stacks.push(new Stack());
      }
    } else {
      stacks = score.Stacks;
    }
    for (i = 0; i < numStacks; i++) {
      updateStackView(i);
    }

    // Reset indications that the stack score is uncommitted.
    $("#teleopMessage").css("opacity", 0);
    $(".stack-grid").attr("data-changed", false);
  }

  // Update component visibility.
  if (!data.AutoCommitted) {
    $("#autoCommands").show();
    $("#autoScore").show();
    $("#teleopCommands").hide();
    $("#teleopScore").hide();
    $("#waitingMessage").hide();
    scoreCommitted = false;
  } else if (!data.TeleopCommitted) {
    $("#autoCommands").hide();
    $("#autoScore").hide();
    $("#teleopCommands").show();
    $("#teleopScore").show();
    $("#waitingMessage").hide();
    scoreCommitted = false;
  } else {
    $("#autoCommands").hide();
    $("#autoScore").hide();
    $("#teleopCommands").hide();
    $("#teleopScore").hide();
    $("#commitMatchScore").hide();
    $("#waitingMessage").show();
    scoreCommitted = true;
  }
};

// Handles a keyboard event and sends the appropriate websocket message.
var handleKeyPress = function(event) {
  var key = String.fromCharCode(event.keyCode);
  switch(key) {
    case "r":
      websocket.send("robotSet");
      break;
    case "c":
      if ($("#autoCommands").is(":visible")) {
        websocket.send("containerSet");
      } else {
        stacks[selectedStack].Container = !stacks[selectedStack].Container;
        if (!stacks[selectedStack].Container) {
          stacks[selectedStack].Litter = false;
        }
        updateStackView(selectedStack);
        invalidateStackScore();
      }
      break;
    case "t":
      websocket.send("toteSet");
      break;
    case "s":
      websocket.send("stackedToteSet");
      break;
    case "j":
      if (selectedStack > 0) {
        selectedStack--;
        updateSelectedStack();
      }
      break;
    case "l":
      if (selectedStack < numStacks - 1) {
        selectedStack++;
        updateSelectedStack();
      }
      break;
    case "i":
      if (stacks[selectedStack].Totes < 6) {
        stacks[selectedStack].Totes++;
        updateStackView(selectedStack);
        invalidateStackScore();
      }
      break;
    case "k":
      if (stacks[selectedStack].Totes > 0) {
        stacks[selectedStack].Totes--;
        updateStackView(selectedStack);
        invalidateStackScore();
      }
      break;
    case "n":
      if (stacks[selectedStack].Container) {
        stacks[selectedStack].Litter = !stacks[selectedStack].Litter;
        updateStackView(selectedStack);
        invalidateStackScore();
      }
      break;
    case "\r":
      websocket.send("commit", stacks);
      stackScoreChanged = false;
      break;
    case "a":
      websocket.send("uncommitAuto");
      break;
  }
};

// Handles a websocket message to update the match status.
var handleMatchTime = function(data) {
  if (matchStates[data.MatchState] == "POST_MATCH" && !scoreCommitted) {
    $("#commitMatchScore").show();
  } else {
    $("#commitMatchScore").hide();
  }
};

// Updates the stack grid to highlight only the active stack.
var updateSelectedStack = function() {
  for (i = 0; i < numStacks; i++) {
    $("#stack" + i).attr("data-selected", i == selectedStack);
  }
};

// Updates the appearance of the given stack in the grid to match the scoring data.
var updateStackView = function(stackIndex) {
  stack = stacks[stackIndex];
  $("#stack" + stackIndex + " .stack-tote-count").text(stack.Totes);
  $("#stack" + stackIndex + " .stack-container").toggle(stack.Container);
  $("#stack" + stackIndex + " .stack-litter").toggle(stack.Litter);
};

// Shows message indicating that the stack score has been changed but not yet sent to the server.
var invalidateStackScore = function() {
  $("#teleopMessage").css("opacity", 1);
  $(".stack-grid").attr("data-changed", true);
  stackScoreChanged = true;
};

// Sends a websocket message to indicate that the score for this alliance is ready.
var commitMatchScore = function() {
  websocket.send("commit", stacks);
  websocket.send("commitMatch");
};

$(function() {
  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/scoring/" + alliance + "/websocket", {
    score: function(event) { handleScore(event.data); },
    matchTime: function(event) { handleMatchTime(event.data); }
  });

  updateSelectedStack();

  $(document).keypress(handleKeyPress);
});
