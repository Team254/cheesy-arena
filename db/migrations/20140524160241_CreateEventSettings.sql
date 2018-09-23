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
  switchaddress VARCHAR(255),
  switchpassword VARCHAR(255),
  plcaddress VARCHAR(255),
  tbadownloadenabled bool,
  adminpassword VARCHAR(255),
  readerpassword VARCHAR(255),
  scaleledaddress VARCHAR(255),
  redswitchledaddress VARCHAR(255),
  blueswitchledaddress VARCHAR(255),
  redvaultledaddress VARCHAR(255),
  bluevaultledaddress VARCHAR(255)
);

-- +goose Down
DROP TABLE event_settings;
