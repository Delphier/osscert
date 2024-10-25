set CGO_ENABLED=0

set GOOS=linux
go build -o bin/osscert-fc fc/main.go
bin\build-fc-zip.exe -output bin\osscert-fc.zip bin\osscert-fc

set GOOS=windows
go build -o bin/osscert.exe exe/main.go