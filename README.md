# ANIRip
A Crunchyroll & Daisuki episode/subtitle ripper written in GO

![alt text](/images/anirip130.jpg "ANIRip v1.3.0 Screenshot")

## Usage
```
anirip -u dankUsername69 -p str8c4s497 -t daisuki http://www.daisuki.net/us/en/anime/detail.ONEPUNCHMAN.html
```
```
anirip help
```
### Setup Guide
**1)** Install [`ffmpeg`](https://ffmpeg.org/download.html) and [`mkvtoolnix`](https://mkvtoolnix.download/downloads.html) if they are not already installed on your system. We will used these tools primarily for trimming and editing video metadata.

**2)** Install [`mkclean`](https://www.matroska.org/downloads/mkclean.html). We use this tool to clean up our final mkv files.

**3) (for Daisuki support)** Install and correctly configure [`PHP`](http://php.net/get/php-5.6.19.tar.bz2/from/a/mirror) (5.6.xx). Specifically, make sure to follow [this guide](https://github.com/K-S-V/Scripts/wiki#installing-php-for-dummies-windows-only) and use the ```php.ini``` file provided in the guide. Also make sure to store the [`AdobeHDS.php`](https://raw.githubusercontent.com/K-S-V/Scripts/master/AdobeHDS.php) script in the same directory as our ANIRip.exe.

**4) (for Crunchyroll support)** Install [`rtmpdump`](https://github.com/K-S-V/Scripts/releases/tag/v2.4). To use ANIRip on Crunchyroll you will need a patched rtmpdump that supports Handshake 10. 

**5)** Clone the `ANIRip` repository or [download the latest release](https://github.com/sdwolfe32/ANIRip/releases).

**6)** `cd` into the `ANIRip` repository directory and execute the following commands:
```
$ go get
$ go generate
$ go build -o ANIRip.exe
```

**7)** Feel free to move `ANIRip` wherever you'd like, just be sure to leave it in the same directory as ```AdobeHDS.php``` (for Daisuki users), and there should be no issues.

**8)** Execute the command shown at the top of this `README` document with your credentials, and you should be up and running!

Note : When I say "Install", I mean you need to set these executables up in your path so that ANIRip can call them directly from the command line.

## Disclaimer
This repo/project was written as an educational intro to web-scraping and network analysis. It is provided publicly as a an open source project for nothing other than educational purposes. I do not take responsibility for how you use this software nor do I recommend you use it in any way that may infringe on Crunchyroll or Daisuki as a business.

## Legal Warning
This application is not endorsed or affiliated with any anime stream provider. The usage of this application enables episodes to be downloaded for offline convenience which may be forbidden by law in your country. Usage of this application may also cause a violation of the agreed Terms of Service between you and the stream provider. A tool is not responsible for your actions; please make an informed decision prior to using this application. Any Stream decryption is done by a third party program, in the case of Crunchyroll by RTMPDump, in the case of Daisuki by the akamai decryption flash library. Usage of this third party programs and/or libraries may be forbidden in your country without proper consent of the copyright holder. None of these programs/libraries are included in this release.
