# anirip
A Crunchyroll & Daisuki episode/subtitle ripper written in GO

![alt text](/images/anirip130.jpg "anirip v1.3.0 Screenshot")

## Usage
To login to providers (Note: You only need to login once):
```
anirip login --user dankUsername69 --pass str8c4s497 crunchyroll
anirip login -u dai5uk1ISLyf -p Yes2325235 daisuki
```
To download videos from Daisuki and CrunchyRoll (Note: You only need to login once):
```
anirip http://www.crunchyroll.com/strike-the-blood
anirip --trim daisuki http://www.daisuki.net/us/en/anime/detail.ONEPUNCHMAN.html
```
To download multiple shows just add more urls:
```
anirip http://www.crunchyroll.com/strike-the-blood http://www.crunchyroll.com/god-eater http://www.crunchyroll.com/attack-on-titan
```
To clear all temporary anirip files on the system:
```
anirip clear
```
To get a list of CLI commands:
```
anirip help
```
### Setup Guide
**1)** Install [`ffmpeg`](https://ffmpeg.org/download.html) and [`mkvtoolnix`](https://mkvtoolnix.download/downloads.html) if they are not already installed on your system. We will used these tools primarily for trimming and editing video metadata. You will also need [`mkclean`](https://sourceforge.net/projects/matroska/files/mkclean/mkclean-win32.v0.8.7.zip). We use this in order to clean up metadata after the file has been dumped.

**2) (for Daisuki support)** Install and correctly configure [`PHP`](http://windows.php.net/download/) (5.6.xx). Specifically, make sure to follow [this guide](https://github.com/K-S-V/Scripts/wiki#installing-php-for-dummies-windows-only) and use the ```php.ini``` file provided in the guide.

**3) (for Crunchyroll support)** Install [`rtmpdump`](https://github.com/K-S-V/Scripts/releases). To use anirip on Crunchyroll you will need a patched rtmpdump that supports Handshake 10.

**4)** Clone the `anirip` repository or [download the latest release](https://github.com/sdwolfe32/anirip/releases).

**5)** `cd` into the `anirip` repository directory and execute the following commands:
```
$ go get
$ go generate
$ go build -o anirip.exe
```

**6)** Feel free to move `anirip` wherever you'd like, as it is a CLI you should be able to call it as such as long as it's in a relative directory/in your path.

**7)** Try a few of the usage commands listed above...

Note : When I say "Install", I mean you need to set these executables up in your PATH OR relatively next to anirip.exe so that anirip can access them directly from the command line.

## Disclaimer
This repo/project was written as an educational intro to web-scraping and network analysis. It is provided publicly as a an open source project for nothing other than educational purposes. I do not take responsibility for how you use this software nor do I recommend you use it in any way that may infringe on Crunchyroll or Daisuki as a business.

## Legal Warning
This application is not endorsed or affiliated with any anime stream provider. The usage of this application enables episodes to be downloaded for offline convenience which may be forbidden by law in your country. Usage of this application may also cause a violation of the agreed Terms of Service between you and the stream provider. A tool is not responsible for your actions; please make an informed decision prior to using this application. Any Stream decryption is done by a third party program, in the case of Crunchyroll by RTMPDump, in the case of Daisuki by the akamai decryption flash library. Usage of this third party programs and/or libraries may be forbidden in your country without proper consent of the copyright holder. None of these programs/libraries are included in this release.

The MIT License (MIT)
=====================

Copyright © 2016 Steven Wolfe

Permission is hereby granted, free of charge, to any person
obtaining a copy of this software and associated documentation
files (the “Software”), to deal in the Software without
restriction, including without limitation the rights to use,
copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following
conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
