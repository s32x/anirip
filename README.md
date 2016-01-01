# CrunchyRip
A Crunchyroll episode/subtitle scraper/dumper written in GO

## Dependencies
In order to run CrunchyRip the below binaries are required (not included)...
- FLVExtract as ```flvextract.exe``` in \temp\ dir
- MKVClean as ```mkvclean.exe``` in \temp\ dir
- MKVMerge as ```mkvmerge.exe``` in \temp\ dir
- RTMPDump as ```rtmpdump.exe``` in \temp\ dir

## How to build
If you're interested in experimenting with the application, are on a Windows machine, and agree to the below disclaimer, compile the core binary using ```builder.bat``` and create the three below directories next to ```CrunchyRip.exe```.

- \output\ (Will be used for moving completed files to in the future)
- \engine\ (Where the above dependencies must be held)
- \temp\ (Where temporary flv/ass files are held before merging)

Given that the required dependencies are in place the ```CrunchyRip.exe``` binary should execute without any issues.

## Disclaimer
This repo/project was written as an educational intro to web-scraping and network analysis. It is provided publicly as a an open source project for nothing other than educational purposes. I do not take responsibility for how you use this software nor do I recommend you use it in any way that may infringe on Crunchyroll as a business.

## Legal Warning
This application is not endorsed or affliated with any anime stream provider. The usage of this application enables episodes to be downloaded for offline convenience which may be forbidden by law in your country. Usage of this application may also cause a violation of the agreed Terms of Service between you and the stream provider. A tool is not responsible for your actions; please make an informed decision prior to using this application. Any Stream decryption is done by a third party program, in case of Crunchyroll by RTMPDump, in case of Daisuki by the akamai decryption flash library. Usage of this third party programs and/or libraries may be forbidden in your country without proper consent of the copyright holder. None of this programs/libraries are included in this release.
