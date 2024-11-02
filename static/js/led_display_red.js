
var websocket;

function setAmplified(res) {
    const messageDiv = $("#amplified");
    if (res) {
        messageDiv.html("&#10003;");
        messageDiv.removeClass("cross");
        messageDiv.addClass("checkmark");
    } else {
        messageDiv.html("&#10006;");
        messageDiv.removeClass("checkmark");
        messageDiv.addClass("cross");
    }
}

function setCoop(res) {
    const messageDiv = $("#coop");
    if (res) {
        messageDiv.html("&#10003;");
        messageDiv.removeClass("cross");
        messageDiv.addClass("checkmark");
    } else {
        messageDiv.html("&#10006;");
        messageDiv.removeClass("checkmark");
        messageDiv.addClass("cross");
    }
}

function setAmpTime(time) {
    const messageDiv = $("#time");
    messageDiv.text(time.toString() + "s")
}

function setBanked(num) {
    const messageDiv = $("#banked");
    messageDiv.text(num.toString())
}

// Handles a websocket message to update the realtime scoring fields.
const handleRealtimeScore = function(data) {
    let realtimeScore;
    realtimeScore = data.Red;
    const score = realtimeScore.Score;
  
    setAmplified(realtimeScore.AmplifiedTimeRemainingSec != 0);
    console.log(realtimeScore);
    setCoop(score.AmpSpeaker.CoopActivated);
    setAmpTime(realtimeScore.AmplifiedTimeRemainingSec);
    setBanked(score.AmpSpeaker.BankedAmpNotes);
};

$(function() {
  // Read the configuration for this display from the URL query string.
  const urlParams = new URLSearchParams(window.location.search);
  const message = urlParams.get("message");
  const messageDiv = $("#message");
  messageDiv.text(message);
  messageDiv.toggle(message !== "");

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/leds/red/websocket", {
    realtimeScore: function(event) { handleRealtimeScore(event.data); },
  });
});
