-- +goose Up
CREATE TABLE event_settings (
  id INTEGER PRIMARY KEY,
  name VARCHAR(255),
  numelimalliances int,
  selectionround2order VARCHAR(1),
  selectionround3order VARCHAR(1),
  teaminfodownloadenabled bool,
  tbapublishingenabled bool,
  tbaeventcode VARCHAR(16),
  tbasecretid VARCHAR(255),
  tbasecret VARCHAR(255),
  networksecurityenabled bool,
  apaddress VARCHAR(255),
  apusername VARCHAR(255),
  appassword VARCHAR(255),
  apteamchannel int,
  apadminchannel int,
  apadminwpakey VARCHAR(255),
  ap2address VARCHAR(255),
  ap2username VARCHAR(255),
  ap2password VARCHAR(255),
  ap2teamchannel int,
  switchaddress VARCHAR(255),
  switchpassword VARCHAR(255),
  plcaddress VARCHAR(255),
  tbadownloadenabled bool,
  adminpassword VARCHAR(255),
  habdockingthreshold int
);

-- +goose Down
DROP TABLE event_settings;
