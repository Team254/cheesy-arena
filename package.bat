del /Q db\backups\*

del /Q static\logs\*

mkdir static\logs

go clean

go build

zip -r -X cheesy-arena.zip LICENSE README.md access_point_config.tar.gz cheesy-arena.exe db font schedules static switch_config.txt templates
