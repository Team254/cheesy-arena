// Copyright 2015 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the lower thirds management interface.

var websocket;

// Sends a websocket message to save the text for the given lower third.
var saveLowerThird = function(button) {
  websocket.send("saveLowerThird", constructLowerThird(button));
};

// Sends a websocket message to delete the given lower third.
var deleteLowerThird = function(button) {
  websocket.send("deleteLowerThird", constructLowerThird(button));
};

// Sends a websocket message to show the given lower third.
var showLowerThird = function(button) {
  websocket.send("showLowerThird", constructLowerThird(button));
};

// Sends a websocket message to hide the lower third.
var hideLowerThird = function(button) {
  websocket.send("hideLowerThird", constructLowerThird(button));
};

// Sends a websocket message to reorder the given the lower third.
var reorderLowerThird = function(button, moveUp) {
  websocket.send("reorderLowerThird", { Id: parseInt(button.form.id.value), MoveUp: moveUp })
};

// Gathers the lower third info and constructs a JSON object.
var constructLowerThird = function(button) {
  return { Id: parseInt(button.form.id.value), TopText: button.form.topText.value,
      BottomText: button.form.bottomText.value, DisplayOrder: parseInt(button.form.displayOrder.value) }
};

$(function() {
  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/setup/lower_thirds/websocket", {});
});
