# mediafiler
**mediafiler** is a small tool written in go that files photos and videos captured from digital cameras into a directory structure based on their internal metadata. The metadata is retrieved using the amazing [exiftool](https://exiftool.org/) which does the heavy lifting of reading EXIF/XMP/other metadata from each file. Files are scanned recursively from a source directory and filed into a structure in the destination directory based on MIME types, timestamps, and camera model.

**mediafiler** will skip files where a timestamp can't be determined or aren't of a "image" or "video" MIME type.
```
# mediafiler src/ dest1/
14:17:21 I ::     startup: I AM mediafiler PLEASE INSERT MEDIA
14:17:21 I ::     startup: exiftool found at: /usr/bin/exiftool
14:17:21 I ::     startup: pre-flight checks passed.
14:17:21 I ::     startup: running exiftool command: /usr/bin/exiftool -r -json -dateFormat %s%-3f src/
14:18:07 I ::  processing: src/movies/2012-09-08 15.01.12.mp4 (3 of 1537)
14:18:07 I ::        skip: we did not find a timestamp
14:18:07 I ::  processing: src/camera/100CANON/IMG_8306.CR2.xmp (4 of 1537)
14:18:07 I ::        skip: The MIME type ('application') for this file is not supported
...
14:18:07 I ::  processing: src/tmp/IMG_20160725_110117.jpg (178 of 1537)
14:18:07 I ::        skip: we did not find a timestamp
14:18:07 I ::  processing: src/tmp/image/x-canon-cr3/2023/12/20231204T022319.860Z-CanonEOSR5.cr3 (179 of 1537)
14:18:07 I ::     renamed: >> dest1//image/x-canon-cr3/2023/12/20231204T022319.860Z-CanonR5.cr3
14:18:07 I ::  processing: src/tmp/image/x-canon-cr3/2024/02/20240208T222800.570Z-CanonEOSR5.cr3 (180 of 1537)
```

**mediafiler** will skip files that appear to be duplicates
```
14:22:51 I ::  processing: dest1/image/jpeg/2024/02/20240208T200531.000Z-DJI-Phantom4.jpg (1229 of 1361)
14:22:51 I ::   duplicate: sourceFile and destFile have the same size and sha256 sums
```
**mediafiler** will attempt to skip renaming files which the OS determines to be the same file, either from hard linking or running with the same source and destination directories.
```
# mediafiler  dest1/ dest1/
14:26:43 I ::     startup: I AM mediafiler PLEASE INSERT MEDIA
14:26:43 I ::     startup: exiftool found at: /usr/bin/exiftool
14:26:43 I ::     startup: pre-flight checks passed.
14:26:43 I ::     startup: running exiftool command: /usr/bin/exiftool -r -json -dateFormat %s%-3f dest1/
14:26:46 I ::     startup: Found 233 files to process
14:26:46 I ::  processing: dest1/image/x-adobe-dng/2018/05/20180512T141441.000Z-DJI-OsmoPlus.dng (1 of 233)
14:26:46 W ::        skip: the OS says that sourceFile and destFile are the same file
14:26:46 I ::  processing: dest1/image/x-adobe-dng/2018/05/20180512T141337.000Z-DJI-OsmoPlus.dng (2 of 233)
14:26:46 W ::        skip: the OS says that sourceFile and destFile are the same file
14:26:46 I ::  processing: dest1/image/x-adobe-dng/2018/05/20180512T141449.000Z-DJI-OsmoPlus.dng (3 of 233)
14:26:46 W ::        skip: the OS says that sourceFile and destFile are the same file
14:26:46 I ::  processing: dest1/image/x-adobe-dng/2018/05/20180512T140505.000Z-DJI-OsmoPlus.dng (4 of 233)
14:26:46 W ::        skip: the OS says that sourceFile and destFile are the same file

```

# Quickstart
- install Go (currently required to compile mediafiler)
- install ExifTool
- install mediafiler
```
go install github.com/d0ct0rvenkman/mediafiler@latest
```
- run mediafiler
```
~/go/bin/mediafiler  source_directory  destination_root_directory
```
- profit? probably not.


# Directory Structure
Files are renamed (moved) into the following structure, which is not currently configurable.
```
$DEST_ROOT_DIR/$MIME_TYPE/$MIME_SUBTYPE/$YEAR/$MONTH/$TIMESTAMP-MODEL.$EXTENSION
```
# File naming scheme
Files are renamed based on the timestamp they were created. The tool will attempt to use subsecond-resolution timestamps if they're present and falls back to less precise timestamps if necessary. The timestamp format used is a slightly shortened RFC3339 format with the special characters and timezone info removed, as all timestamps are rendered as UTC/GMT. Exiftool handles the conversion to UTC as part of its processing. If time zone data is present in the image, the file can be predictably renamed using UTC. If time zone information is not present in the file's metadata, Exiftool assumes the timestamp retrieved from the file metadata is in local time for the machine where mediafiler/exiftool is running, which is then converted to UTC. This can be less than predictable if you're processing on a machine in a different timezone from where the image was taken.

Camera models are currently renamed/shortened based on hard-coded patterns for cameras I've used over the years, but making this configurable is one of the first TODOs I plan to address.

Filename collisions are detected during processing. Source and destination files are checksummed to see if they're the duplicates of the same media. Duplicates are skipped without further processing. Non-duplicates are handled by appending a numeric index after the model and before the extension.
```
YYYYMMDDTHHMMSS.SSSZ-model[-NNN].extension
```

Not all cameras are great at storing their models in the file metadata (especially in video). If a camera model can't be determined, "unknown" is used in its place.

# Example
```
image/jpeg/2006/10/20061010T193922.000Z-E4300.jpg
image/jpeg/2008/03/20080308T044606.000Z-CPS_SD600.jpg
image/jpeg/2008/09/20080906T193317.000Z-CPS_SD600.jpg
image/jpeg/2012/07/20120703T192601.000Z-Droid.jpg
image/jpeg/2012/09/20120908T160050.000Z-CPS_SD600.jpg
image/jpeg/2013/05/20130515T022840.000Z-SCH-I535-001.jpg
image/jpeg/2013/05/20130515T022840.000Z-SCH-I535-002.jpg
image/jpeg/2013/05/20130515T022840.000Z-SCH-I535.jpg
image/jpeg/2014/09/20140927T204826.000Z-SPH-L720.jpg
image/jpeg/2015/01/20150125T201612.000Z-SCH-I535.jpg
image/jpeg/2015/08/20150809T001543.683Z-Nexus6.jpg
image/x-canon-cr2/2018/05/20180523T204222.570Z-Canon800D.cr2
image/x-canon-cr2/2023/08/20230826T212107.090Z-Canon800D.cr2
image/x-canon-cr3/2023/12/20231226T005750.460Z-CanonR5.cr3
video/3gpp/2009/11/20091126T235057.000Z-unknown.3gp
video/3gpp/2012/07/20120705T233313.000Z-unknown.3gp
video/mp4/2014/07/20140726T033851.000Z-unknown.mp4
video/mp4/2018/08/20180811T210151.000Z-DJI-Phantom4.mp4
video/mp4/2018/09/20180906T193749.000Z-DJI-OsmoPlus.mp4
video/mp4/2023/10/20231021T182248.000Z-unknown.mp4
video/quicktime/2004/04/20040419T171642.000Z-E3100.mov
video/quicktime/2004/06/20040624T014342.000Z-E3100.mov
video/quicktime/2016/05/20160530T165526.000Z-DJI-Phantom3Adv.mov
video/quicktime/2016/05/20160531T225438.000Z-DJI-Phantom3Adv.mov
video/x-msvideo/2007/04/20070406T214156.000Z-unknown.avi
video/x-msvideo/2007/06/20070625T052246.000Z-unknown.avi

```


# TODO
- [ ] Make camera model substitutions/translations configurable via config file
- [ ] Make camera model translations based on regex and not just substitution
- [ ] Don't assume unix-like path separators (Windows support)
- [ ] Make file/path ignores configurable via config file 
- [ ] Make file/path ignores based on regex
- [ ] Make logging level configurable at run time
- [ ] Make file/directory naming customizable via templates
- [ ] Dry-run/no-op option
- [ ] Refactor for tests
- [ ] Tests