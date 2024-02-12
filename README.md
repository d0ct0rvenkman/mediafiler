# mediafiler
*mediafiler* is a small tool written in go that files photos and videos captured from digital cameras into a directory structure based on their internal metadata. The metadata is retrieved using the amazing [exiftool](https://exiftool.org/) which does the heavy lifting of reading EXIF/XMP/other metadata from each file. Files are scanned recursively from a source directory and filed into a structure in the destination directory based on MIME types, timestamps, and camera model.

# Directory Structure
Files are renamed (moved) into the following structure, which is not (currently) customizable.
```
$DEST_ROOT_DIR/$MIME_TYPE/$MIME_SUBTYPE/$YEAR/$MONTH/$TIMESTAMP-MODEL.$EXTENSION
```
# File naming scheme
Files are renamed based on the timestamp they were created. The tool will attempt to use subsecond-resolution timestamps if they're present and falls back to less precise timestamps if necessary. The timestamp format used is a slightly shortened RFC3339 format with the special characters and timezone info removed, as all timestamps are rendered from UTC/GMT. Camera models are currently renamed/shortened based on hard-coded patterns for cameras I've used over the years, but making this configurable is one of the first TODOs I plan to address.

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
- [ ] Refactor for tests
- [ ] Tests