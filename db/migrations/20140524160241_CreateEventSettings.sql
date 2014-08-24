-- +goose Up
CREATE TABLE event_settings (
  id INTEGER PRIMARY KEY,
  name VARCHAR(255),
  code VARCHAR(16),
  displaybackgroundcolor VARCHAR(16),
  numelimalliances int,
  selectionround2order VARCHAR(1),
  selectionround3order VARCHAR(1),
  teaminfodownloadenabled bool,
  alliancedisplayhotgoals bool,
  redgoallightsaddress VARCHAR(255),
  bluegoallightsaddress VARCHAR(255),
  tbapublishingenabled bool,
  tbaeventcode VARCHAR(16),
  tbasecretid VARCHAR(255),
  tbasecret VARCHAR(255),
  networksecurityenabled bool,
  apaddress VARCHAR(255),
  apusername VARCHAR(255),
  appassword VARCHAR(255),
  switchaddress VARCHAR(255),
  switchpassword VARCHAR(255)
);

-- +goose Down
DROP TABLE event_settings;
