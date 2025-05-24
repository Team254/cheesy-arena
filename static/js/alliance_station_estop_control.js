let alliance;

$(function() {
  alliance = window.location.href.split("/").slice(-1)[0];
  $("#alliance").attr("data-alliance", alliance);
});


function triggerEStop(stationNumber, state) {
    const url = "/api/freezy/eStopState"; // Relative URL, works dynamically
    if (alliance === "blue") {
        stationNumber = stationNumber+6
    }

    const payload = [
        { channel: stationNumber, state: state }
    ];

    fetch(url, {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(payload)
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`Request failed with status ${response.status}`);
        }
        return response.text();
    })
    .then(data => {
        console.log("Request successful!");
        console.log("Response:", data);
        console.log("Payload sent:", payload);
    })
    .catch(error => {
        console.error("An error occurred:", error);
    });
  }