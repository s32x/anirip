cd anirip
go install
cd..
cd daisuki
go install
cd..
cd crunchyroll
go install
cd..
go generate
go build -o ANIRip.exe

echo "FINISHED ANIRIP BUILD SCRIPT"
pause
