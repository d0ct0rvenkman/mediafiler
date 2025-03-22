package main

/*
TODO:
- don't assume unix-like path separators

*/

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/codingsince1985/checksum"
	"github.com/d0ct0rvenkman/mediafiler/internal/config"
	"github.com/d0ct0rvenkman/mediafiler/internal/fileops"
	"github.com/d0ct0rvenkman/mediafiler/internal/logfmt"
	"github.com/d0ct0rvenkman/mediafiler/internal/paths"
	"github.com/d0ct0rvenkman/mediafiler/internal/strmanip"
	which "github.com/hairyhenderson/go-which"
	multierr "github.com/hashicorp/go-multierror"
	logrus "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

const (
	dirSep string = "/" // TODO: find a way to determine this programmatically
)

var log = logrus.New()
var confLoaded bool
var confErr error

func main() {
	var workDir string
	var destRootDir string
	var merr error
	var err error
	var supportedMIMETypes []string

	dryrun := false

	log.SetFormatter(new(logfmt.NonDebugFormatter))
	log.SetLevel(logrus.InfoLevel)
	log.SetOutput(os.Stdout)

	startLog := log.WithFields(logrus.Fields{"verb": "startup:"})

	config.Initialize(os.Args[1:])
	config.ProcessFatalFlags()
	config.UseDefaultConfigPaths()
	confLoaded, confErr = config.ReadConfiguration()

	// startLog.Infof("conf: %v   %v", confLoaded, confErr)

	if confLoaded {
		err = config.ProcessConfiguration()
		if err != nil {
			startLog.Fatalf("an error occured while processing loaded configuration: %s", err)
		}
	}

	if !confLoaded {
		startLog.Fatalf("configuration could not be loaded. reason: %s", confErr)
	}

	if config.Config.GetBool("debug") {
		log.SetLevel(logrus.TraceLevel)
	}

	//var modelReplacer strmanip.Replacer
	var specialReplacer strmanip.Replacer

	specialReplacer.AddRule(strmanip.ReplacerRule{Type: "string", Find: `/`, ReplaceWith: "_"})
	specialReplacer.AddRule(strmanip.ReplacerRule{Type: "string", Find: `\`, ReplaceWith: "_"})

	supportedMIMETypes = append(supportedMIMETypes, "image", "video")

	startLog.Infof("I AM %s PLEASE INSERT MEDIA", os.Args[0])

	for k, v := range os.Args {
		startLog.Tracef("arg[%d]: '%s'", k, v)
	}

	if confLoaded && (confErr != nil) {
		if confErr.Error() == config.DEFAULT_CONFIG_USED {
			startLog.Warn("falling back to default configuration")
		}
	}

	if config.Config.GetBool("dry-run") {
		dryrun = true
		startLog.Info("dry-run mode enabled")
	}

	merr = nil
	err = nil

	// Hard Requirements
	exiftoolbin := ""
	if exiftoolbin = config.Config.GetString("exiftool-binary"); exiftoolbin != "" {
		startLog.Infof("using user-specified exiftool binary: %s", exiftoolbin)
	} else {
		// check for Exiftool
		exiftoolbin = which.Which("exiftool")
		if exiftoolbin == "" {
			err = errors.New("exiftool binary was not found")
			startLog.Debug((err))
			merr = multierr.Append(merr, err)
		} else {
			startLog.Infof("exiftool found at: %s", exiftoolbin)
		}
	}
	// determine what paths we're working with
	workDir, destRootDir, err = paths.GetMediaPaths(config.FS.Args())
	if err != nil {
		merr = multierr.Append(merr, err)
		err = fmt.Errorf("error determining paths. %s", err)
		startLog.Debug(err)

	}

	if merr != nil {
		startLog.Fatalf("basic requirements not satisfied. %s", merr)
	}

	merr = nil

	// softer checks
	// validate our paths
	err = paths.ValidateFileOrDirectory(workDir)
	if err != nil {
		err = fmt.Errorf("working directory is not valid for use. %s", err)
		merr = multierr.Append(merr, err)
	}

	err = paths.ValidateDirectory(destRootDir)
	if err != nil {
		err = fmt.Errorf("destination directory is not valid for use. %s", err)
		merr = multierr.Append(merr, err)
	}

	if merr != nil {
		startLog.Fatalf("paths provided are not usable. %s", merr)
	}

	startLog.Info("pre-flight checks passed.")

	// "2006-01-02T15:04:05.999999999Z07:00"
	dateFormat := "%s%-3f"

	cmd := exec.Command(exiftoolbin, "-r", "-json", "-dateFormat", dateFormat, workDir)
	startLog.Infof("running exiftool command: %s", cmd.String())
	output, err := cmd.Output()
	if err != nil {
		startLog.Warnf("exiftool reported an error. %s", err)
	}

	if len(output) == 0 {
		startLog.Info("exiftool output was empty. exiting.")
		os.Exit(0)
	}

	if !gjson.ValidBytes(output) {
		startLog.Fatalf("failed to unmarshal JSON output from exiftool")
	}

	result := gjson.ParseBytes(output)

	fileCount := len(result.Array())
	startLog.Infof("Found %d files to process", fileCount)

SOURCEFILE:
	for k, v := range result.Array() {
		var sourceSum string

		sourceFile := v.Get("SourceFile").String()

		fileLogger := log.WithFields(logrus.Fields{
			"sourceFile": strings.Replace(sourceFile, workDir, "."+dirSep, 1),
			"fileIndex":  k + 1,
			"fileCount":  fileCount,
			"verb":       "  ",
		})

		log.WithFields(logrus.Fields{"verb": "processing:"}).Infof("%s (%d of %d)", sourceFile, k+1, fileCount)

		ignore, err := config.PathIgnorer.IsPathFiltered(sourceFile)
		if err != nil {
			fileLogger.WithFields(logrus.Fields{"verb": "skip:"}).Fatalf("Path Ignore Filter execution failed: reason ('%s')", err)
		}
		if ignore {
			fileLogger.WithFields(logrus.Fields{"verb": "skip:"}).Warnf("sourceFile matches an ignore path pattern")
			continue SOURCEFILE
		}

		sourceFileInfo, err := os.Stat(sourceFile)
		if err != nil {
			fileLogger.WithFields(logrus.Fields{"verb": "skip:"}).Error("could not Stat source file. interesting.")
			continue SOURCEFILE
		}

		newPathSuffix, newFileName, fileExtension, err := generateFilenameBase(v, supportedMIMETypes, config.ModelReplacer, specialReplacer)
		if err != nil {
			fileLogger.WithFields(logrus.Fields{"verb": "skip:"}).Infof("generateFilenameBase: %s", err)
			continue SOURCEFILE
		}

		fileLogger.Debugf("destRootDir: %s", destRootDir)
		fileLogger.Debugf("newPathSuffix: %s", newPathSuffix)
		fileLogger.Debugf("newFileName: %s", newFileName)

		// we've got the stuff we need to rename the file, now lets see if the destination file already exists.
		destFile := fmt.Sprintf("%s%s%s%s%s.%s", destRootDir, dirSep, newPathSuffix, dirSep, newFileName, fileExtension)
		suffixIndex := 0
		pathAvailable, pathInfo, pathErr := paths.IsPathAvailable(destFile)
		if !pathAvailable {
			// grab some characteristics about the source file. only need to do it once.
			fileLogger.Debug("initial destFile isn't available")

		}

	TESTPATH:
		for !pathAvailable && (suffixIndex < 1000) {
			testLogger := fileLogger.WithFields(logrus.Fields{
				"destFile":    destFile,
				"suffixIndex": suffixIndex,
				"verb":        "    ",
			})

			// path isn't available, lets figure out if we should try again with an updated suffix
			switch pathErr.Error() {
			case paths.E_AVAIL_PATH_EXISTS:
				// see if the file is a duplicate. if not, try a new path.

				if os.SameFile(sourceFileInfo, pathInfo) {
					testLogger.WithFields(logrus.Fields{"verb": "skip:"}).Warn("the OS says that sourceFile and destFile are the same file")
					continue SOURCEFILE
				}

				if suffixIndex == 0 {
					sourceSum, err = checksum.SHA256sum(sourceFile)
					if err != nil {
						fileLogger.WithFields(logrus.Fields{"verb": "skip:"}).Error("couldn't checksum the source file.")
						continue SOURCEFILE
					}
				}

				destSum, derr := checksum.SHA256sum(destFile)

				if derr != nil {
					testLogger.Warn("couldn't checksum the File at destFile. try another destFile")
					continue TESTPATH
				} else if sourceFileInfo.Size() == pathInfo.Size() && sourceSum == destSum {
					testLogger.WithFields(logrus.Fields{"verb": "duplicate:"}).Info("sourceFile and destFile have the same size and sha256 sums")
					continue SOURCEFILE
				} else {
					testLogger.Debug("doesn't look like a duplicate. try another destFile")
				}

			case paths.E_AVAIL_PERMS:
				testLogger.Error("permission was denied while testing if path was available")
			case paths.E_AVAIL_UNKNOWN:
				testLogger.Error("got an unknown error passed dowm from IsPathAvailable()")
			default:
				testLogger.Error("got an unknown error from IsPathAvailable()")
			}

			suffixIndex++
			destFile = fmt.Sprintf("%s%s%s%s%s-%03d.%s", destRootDir, dirSep, newPathSuffix, dirSep, newFileName, suffixIndex, fileExtension)
			pathAvailable, pathInfo, pathErr = paths.IsPathAvailable(destFile)
		}

		fileLogger.Debugf("destination file: %s", destFile)

		targetDir := destRootDir + dirSep + newPathSuffix

		if !dryrun {
			fileLogger.Debugf("creating target directory: %s", targetDir)
			err = os.MkdirAll(targetDir, 0755)
			if err != nil {
				fileLogger.Errorf("could not create destination directory! reason: %s", err)
			}

			err = fileops.Move(sourceFile, destFile)
			if err != nil {
				fileLogger.WithFields(logrus.Fields{"verb": "error:"}).Errorf("could not rename file! reason: %s", err)
			} else {
				fileLogger.WithFields(logrus.Fields{"verb": "renamed:"}).Infof(">> %s", destFile)
			}
		} else {
			fileLogger.WithFields(logrus.Fields{"verb": "dry-run:"}).Infof(">> %s", destFile)
		}
	} // ends: for k, v := range result.Array()
}

func generateFilenameBase(meta gjson.Result, supportedMIMETypes []string, modelReplacer strmanip.Replacer, specialReplacer strmanip.Replacer) (string, string, string, error) {
	var timeObj time.Time
	var timeInput int64
	var timestampFound bool
	var serr error

	gfbLogger := log.WithFields(logrus.Fields{
		"sourceFile": meta.Get("SourceFile").String(),
	})

	timestampFound = false
	model := "unknown"
	cameraSerial := ""
	lensSerial := ""
	mimeType := ""
	mimeSubType := ""
	newPathSuffix := ""
	fileExtension := ""

	if meta.Get("FileTypeExtension").Exists() {
		fileExtension = strings.ToLower(meta.Get("FileTypeExtension").String())
	} else {
		serr = errors.New("file metadata doesn't contain an extension")
		return "", "", "", serr
	}

	if meta.Get("MIMEType").Exists() {
		// using "mimeType, mimeSubType, ok := " seems to make the two mime variables local in scope?
		ok := false
		mimeType, mimeSubType, ok = strings.Cut(meta.Get("MIMEType").String(), "/")
		if !ok {
			serr = fmt.Errorf("MIMEType string '%s' could not be cut", meta.Get("MIMEType").String())
			return "", "", "", serr
		}

		if mimeType == "" || mimeSubType == "" {
			serr = fmt.Errorf("MIME Type ('%s') or Subtype ('%s') cannot be empty", mimeType, mimeSubType)
			return "", "", "", serr
		}
	} else {
		serr = fmt.Errorf("MIME type for this file was not found")
		return "", "", "", serr
	}

	if !slices.Contains(supportedMIMETypes, mimeType) {
		serr = fmt.Errorf("the MIME type ('%s') for this file is not supported", mimeType)
		return "", "", "", serr
	}

	switch {
	case meta.Get("SubSecDateTimeOriginal").Exists():
		timeInput, _ = strconv.ParseInt(meta.Get("SubSecDateTimeOriginal").String(), 10, 64)
		gfbLogger.Debugf("timeInput ('%d') pulled from 'SubSecDateTimeOriginal'", timeInput)
		timeObj = time.UnixMilli(timeInput)
		timestampFound = true

	case meta.Get("DateTimeOriginal").Exists():
		timeInput, _ = strconv.ParseInt(meta.Get("DateTimeOriginal").String(), 10, 64)
		gfbLogger.Debugf("timeInput ('%d') pulled from 'DateTimeOriginal'", timeInput)
		timeObj = time.UnixMilli(timeInput)
		timestampFound = true

	case meta.Get("CreateDate").Exists():
		timeInput, _ = strconv.ParseInt(meta.Get("CreateDate").String(), 10, 64)
		gfbLogger.Debugf("timeInput ('%d') pulled from 'CreateDate'", timeInput)
		timeObj = time.UnixMilli(timeInput)
		timestampFound = true

	case meta.Get("ModifyDate").Exists(): // damnit, DROID3!
		timeInput, _ = strconv.ParseInt(meta.Get("ModifyDate").String(), 10, 64)
		gfbLogger.Debugf("timeInput ('%d') pulled from 'ModifyDate'", timeInput)
		timeObj = time.UnixMilli(timeInput)
		timestampFound = true

	case meta.Get("GPSDateTime").Exists(): // damnit, Nexus6!
		timeInput, _ = strconv.ParseInt(meta.Get("GPSDateTime").String(), 10, 64)
		gfbLogger.Debugf("timeInput ('%d') pulled from 'GPSDateTime'", timeInput)
		gfbLogger.Info("fell back to using 'GPSDateTime' for image timestamp, which is not necessarily accurate")
		timeObj = time.UnixMilli(timeInput)
		timestampFound = true
	}

	if !timestampFound {
		serr = fmt.Errorf("we did not find a timestamp")
		return "", "", "", serr
	}

	switch {
	case meta.Get("Model").Exists():
		model = meta.Get("Model").String()
	case meta.Get("AndroidModel").Exists():
		model = meta.Get("AndroidModel").String()
	}

	model, _ = modelReplacer.Replace(model)
	model, _ = specialReplacer.Replace(model)

	if meta.Get("SerialNumber").Exists() {
		cameraSerial = meta.Get("SerialNumber").String()
		cameraSerial, _ = specialReplacer.Replace(cameraSerial)
	}

	if meta.Get("LensSerialNumber").Exists() {
		lensSerial = meta.Get("LensSerialNumber").String()
		lensSerial, _ = specialReplacer.Replace(lensSerial)
	}

	gfbLogger.Debugf("model: %s", model)
	gfbLogger.Debugf("cameraSerial: %s", cameraSerial)
	gfbLogger.Debugf("lensSerial: %s", lensSerial)
	gfbLogger.Debugf("timestamp: %s", timeObj.String())
	gfbLogger.Debugf("MIME: %s / %s", mimeType, mimeSubType)
	gfbLogger.Debugf("fileExtension: %s", fileExtension)

	// TODO: make this something that can be made into a template
	newPathSuffix = fmt.Sprintf("%s%s%s%s%04d%s%02d",
		mimeType, dirSep, mimeSubType, dirSep, timeObj.UTC().Year(), dirSep, timeObj.UTC().Month())

	newFileName := fmt.Sprintf("%04d%02d%02dT%02d%02d%02d.%03dZ-%s",
		timeObj.UTC().Year(),
		timeObj.UTC().Month(),
		timeObj.UTC().Day(),
		timeObj.UTC().Hour(),
		timeObj.UTC().Minute(),
		timeObj.UTC().Second(),
		timeObj.UTC().Round(time.Microsecond).Nanosecond()/1e6,
		model,
	)

	// TODO: Should make this something the user can enable/disable
	// if cameraSerial != "" {
	// 	newFileName = fmt.Sprintf("%s_CS%s", newFileName, cameraSerial)
	// }

	// if lensSerial != "" {
	// 	newFileName = fmt.Sprintf("%s_LS%s", newFileName, lensSerial)
	// }

	return newPathSuffix, newFileName, fileExtension, nil
}
