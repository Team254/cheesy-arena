// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the Twitch stream display.

var websocket;

$(function () {
  // Read the configuration for this display from the URL query string.
  var urlParams = new URLSearchParams(window.location.search);

  // Embed the video stream.
  var embed = new Twitch.Embed("twitchEmbed", {
    channel: urlParams.get("channel"),
    width: window.innerWidth,
    height: window.innerHeight,
    layout: "video"
  });

  embed.addEventListener(Twitch.Embed.VIDEO_READY, function () {
    if (urlParams.get("captions") === "true") {
      embed.getPlayer().enableCaptions();
    } else {
      embed.getPlayer().disableCaptions();
    }
  });

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/twitch/websocket", {});
});
