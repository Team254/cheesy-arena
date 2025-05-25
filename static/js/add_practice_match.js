function populateEditModal(id, red1, red2, red3, blue1, blue2, blue3, name) {
    document.getElementById("editMatchId").value = id;
    document.getElementById("editMatchLabel").textContent = id;
    document.getElementById("editRed1").value = red1;
    document.getElementById("editRed2").value = red2;
    document.getElementById("editRed3").value = red3;
    document.getElementById("editBlue1").value = blue1;
    document.getElementById("editBlue2").value = blue2;
    document.getElementById("editBlue3").value = blue3;
}
