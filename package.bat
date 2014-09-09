
del /Q db\backups\*

del /Q static\logs\*

mkdir static\logs

go clean

go build

zip -r -X cheesy-arena.zip LICENSE README.md ap_config.txt cheesy-arena.exe db font schedules static switch_config.txt templates
