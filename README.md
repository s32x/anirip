# CrunchyRip
A Crunchyroll episode/subtitle scraper/dumper written in GO

## Dependencies
In order to run CrunchyRip the below binaries are required (not included)...
- FFMpeg as ```ffmpeg.exe``` in \temp\ dir
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
