const row = document.querySelector('#buttonBottomRow');

const nexusButton = document.createElement('button');
nexusButton.classList.add('btn', 'btn-dark', 'btn-lg', 'nexus');
nexusButton.textContent = 'Nexus'
row.prepend(nexusButton);

nexusButton.addEventListener('click', loadLineups)

async function loadLineups() {
  let matchKey = document.querySelector('#matchName').textContent;
  matchKey = matchKey.replace('Practice ', 'p');
  matchKey = matchKey.replace('Match ', 'sf');
  matchKey = matchKey.replace('Final ', 'f1m');

  try {
    let lineups = await(await fetch(`https://api.frc.nexus/v1/2023cc/${matchKey}/lineup`)).json();

    let changed = false;
    for(let i = 0; i < lineups.blue?.length; i++) {
      let el = document.querySelector(`#statusB${i + 1} input.team-number`);
      if(el.value === lineups.blue[i]) continue;
      el.value = lineups.blue[i];
      changed = true;
    }
    for(let i = 0; i < lineups.red?.length; i++) {
      let el = document.querySelector(`#statusR${i + 1} input.team-number`);
      if(el.value === lineups.red[i]) continue;
      el.value = lineups.red[i];
      changed = true;
    }
    if(changed) {
      document.querySelector('#substituteTeams').disabled = false;
    }
  } catch (e) {
    console.warn(e);
  }
}