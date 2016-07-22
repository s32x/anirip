cd anirip
go get
go install
cd..
cd daisuki
go get
go install
cd..
cd crunchyroll
go get
go install
cd..
go get
go install
go generate
go build -o anirip.exe

echo "FINISHED ANIRIP BUILD SCRIPT"
