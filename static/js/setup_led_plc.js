// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the field setup page.

var websocket;

// Handles a websocket message to update the PLC IO status.
var handlePlcIoChange = function(data) {
  $.each(data.Inputs, function(index, input) {
    $("#input" + index).text(input)
    $("#input" + index).attr("data-plc-value", input);
  });

  $.each(data.Registers, function(index, register) {
    $("#register" + index).text(register)
  });

  $.each(data.Coils, function(index, coil) {
    $("#coil" + index).text(coil)
    $("#coil" + index).attr("data-plc-value", coil);
  });
};

$(function() {
  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/setup/led_plc/websocket", {
    plcIoChange: function(event) { handlePlcIoChange(event.data); }
  });
});
