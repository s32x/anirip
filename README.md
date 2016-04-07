# ANIRip
A Crunchyroll & Daisuki episode/subtitle ripper written in GO

![alt text](/images/anirip130.jpg "ANIRip v1.3.0 Screenshot")

## Usage
To download videos from CrunchyRoll (Note: You only need to login once):
```
anirip login --user dankUsername69 --pass str8c4s497 crunchyroll
anirip http://www.crunchyroll.com/bladedance-of-elementalers
```
To download videos from Daisuki (Note: You only need to login once):
```
anirip login -u dai5uk1ISLyf -p Yes2325235 daisuki
anirip --trim daisuki http://www.daisuki.net/us/en/anime/detail.ONEPUNCHMAN.html
```
To clear all temporary ANIRip files on the system:
```
anirip clear
```
To get a list of CLI commands:
```
anirip help
```
### Setup Guide
**1)** Install [`ffmpeg`](https://ffmpeg.org/download.html) and [`mkvtoolnix`](https://mkvtoolnix.download/downloads.html) if they are not already installed on your system. We will used these tools primarily for trimming and editing video metadata.

**2)** Install [`mkclean`](https://www.matroska.org/downloads/mkclean.html). We use this tool to clean up and optimize our final mkv files.

**3) (for Daisuki support)** Install and correctly configure [`PHP`](http://php.net/get/php-5.6.19.tar.bz2/from/a/mirror) (5.6.xx). Specifically, make sure to follow [this guide](https://github.com/K-S-V/Scripts/wiki#installing-php-for-dummies-windows-only) and use the ```php.ini``` file provided in the guide.

**4) (for Crunchyroll support)** Install [`rtmpdump`](https://github.com/K-S-V/Scripts/releases). To use ANIRip on Crunchyroll you will need a patched rtmpdump that supports Handshake 10.

**5)** Clone the `ANIRip` repository or [download the latest release](https://github.com/sdwolfe32/ANIRip/releases).

**6)** `cd` into the `ANIRip` repository directory and execute the following commands:
```
$ go get
$ go generate
$ go build -o ANIRip.exe
```

**7)** Feel free to move `ANIRip` wherever you'd like, as it is a CLI you should be able to call it as such as long as it's in a relative directory/in your path.

**8)** Try a few of the usage commands listed above...

Note : When I say "Install", I mean you need to set these executables up in your path so that ANIRip can call them directly from the command line.

## Disclaimer
This repo/project was written as an educational intro to web-scraping and network analysis. It is provided publicly as a an open source project for nothing other than educational purposes. I do not take responsibility for how you use this software nor do I recommend you use it in any way that may infringe on Crunchyroll or Daisuki as a business.

## Legal Warning
This application is not endorsed or affiliated with any anime stream provider. The usage of this application enables episodes to be downloaded for offline convenience which may be forbidden by law in your country. Usage of this application may also cause a violation of the agreed Terms of Service between you and the stream provider. A tool is not responsible for your actions; please make an informed decision prior to using this application. Any Stream decryption is done by a third party program, in the case of Crunchyroll by RTMPDump, in the case of Daisuki by the akamai decryption flash library. Usage of this third party programs and/or libraries may be forbidden in your country without proper consent of the copyright holder. None of these programs/libraries are included in this release.
