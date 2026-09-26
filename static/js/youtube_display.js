// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the YouTube stream display.

var websocket;

function onYouTubeIframeAPIReady() {
  var urlParams = new URLSearchParams(window.location.search);
  if (!urlParams.get("videoId")) {
    return;
  }

  var playerVars = {origin: window.location.origin, playsinline: 1};
  // YouTube can force captions on, but otherwise uses the viewer's preference; it has no supported force-off option.
  if (urlParams.get("captions") === "true") {
    playerVars.cc_load_policy = 1;
  } else {
    playerVars.cc_load_policy = 0;
  }
  new YT.Player("youtubeEmbed", {
    videoId: urlParams.get("videoId"),
    width: "100%",
    height: "100%",
    playerVars: playerVars,
    events: {
      onReady: function (event) {
        // Muted playback allows autoplay in browsers that block it with audio.
        event.target.mute();
        event.target.playVideo();
      }
    }
  });
}

$(function () {
  if (!new URLSearchParams(window.location.search).get("videoId")) {
    $("#youtubeEmbed").text("Set videoId to the YouTube live stream's video ID in Display Configuration.");
  }

  // Keep remote configuration and reloads available even if the player API fails to load.
  websocket = new CheesyWebsocket("/displays/youtube/websocket", {});
});
