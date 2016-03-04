# ANIRip
A Crunchyroll & Daisuki episode/subtitle ripper written in GO

## Dependencies
In order to run ANIRip the below windows binaries are required (not included)...
- PHP(5.2.0) in \engine\php\ dir
- AdobeHDS as ```AdobeHDS.php``` in \engine\ dir
- RTMPDump as ```rtmpdump.exe``` in \engine\ dir
- FFMpeg as ```ffmpeg.exe``` in \engine\ dir
- FFProbe as ```ffprobe.exe``` in \engine\ dir
- MKVClean as ```mkvclean.exe``` in \engine\ dir

## How to build
If you're interested in experimenting with the application, are on a Windows machine, have Go 1.5 installed, and agree to the below disclaimer, compile the core binary using ```builder.bat``` and create the two below directories in the same directory as ```ANIRip.exe```.

- \engine\ (Where the above dependencies must be held)
- \temp\ (Where temporary flv/fragment/mkv/ass files are held before finalization)

Given that the required dependencies are in place the ```ANIRip.exe``` binary should execute and operate without any issues.

## Disclaimer
This repo/project was written as an educational intro to web-scraping and network analysis. It is provided publicly as a an open source project for nothing other than educational purposes. I do not take responsibility for how you use this software nor do I recommend you use it in any way that may infringe on Crunchyroll or Daisuki as a business.

## Legal Warning
This application is not endorsed or affiliated with any anime stream provider. The usage of this application enables episodes to be downloaded for offline convenience which may be forbidden by law in your country. Usage of this application may also cause a violation of the agreed Terms of Service between you and the stream provider. A tool is not responsible for your actions; please make an informed decision prior to using this application. Any Stream decryption is done by a third party program, in the case of Crunchyroll by RTMPDump, in the case of Daisuki by the akamai decryption flash library. Usage of this third party programs and/or libraries may be forbidden in your country without proper consent of the copyright holder. None of these programs/libraries are included in this release.
