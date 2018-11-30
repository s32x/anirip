# anirip
A Crunchyroll & Daisuki episode/subtitle ripper written in GO

![alt text](/images/anirip130.jpg "anirip v1.3.0 Screenshot")

## Usage
To download shows from Crunchyroll :
```
anirip myUsername0123 myPassword5535 http://www.crunchyroll.com/strike-the-blood
```
## Setup Guide (EASY)

**1)** Download the [latest release](https://github.com/s32x/anirip/releases).

**2)** Extract and cd into the release directory.

**5)** Follow the usage instructions above.

## Setup Guide (HARD)

**1)** Install [`ffmpeg`](https://ffmpeg.org/download.html) if it isn't already installed on your system. We will using this tool primarily for dumping episode content and editing video metadata.

**2)** Install [`mkclean`](https://sourceforge.net/projects/matroska/files/mkclean/mkclean-win32.v0.8.7.zip) if it isnt' already installed on your system. We use this in order to clean up metadata after each episode has been dumped.

**3)** Clone the `anirip` repository.

**4)** `cd` into the `anirip` repository directory and execute the following commands:
```
$ go get
$ go install
```

**5)** Follow the usage instructions above.

Note : When I say "Install", I mean you need to set these executables up in your PATH OR relatively next to your anirip binary so that anirip can access them directly from the command line.

## Disclaimer
This repo/project was written as an educational intro to web-scraping and network analysis. It is provided publicly as a an open source project for nothing other than educational purposes. I do not take responsibility for how you use this software nor do I recommend you use it in any way that may infringe on Crunchyroll as a business.

## Legal Warning
This application is not endorsed or affiliated with any anime stream provider. The usage of this application enables episodes to be downloaded for offline convenience which may be forbidden by law in your country. Usage of this application may also cause a violation of the agreed Terms of Service between you and the stream provider. A tool is not responsible for your actions; please make an informed decision prior to using this application. Any Stream decryption is done by a third party program, in the case of Crunchyroll by ffmpeg. Usage of third party programs and/or libraries may be forbidden in your country without proper consent of the copyright holder.

The MIT License (MIT)
=====================

Copyright © 2017 Steven Wolfe

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
