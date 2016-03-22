go get github.com/PuerkitoBio/goquery
go get github.com/fatih/color
go get github.com/josephspurrier/goversioninfo/cmd/goversioninfo
go get github.com/go-ini/ini
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
