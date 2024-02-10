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

func main() {
	var workDir string
	var destRootDir string
	var merr error
	var err error
	var supportedMIMETypes []string
	var pathIgnoreSubstrings []string

	// TODO: make these configurable, not hardcoded
	modelTranslate := make(map[string]string)
	modelTranslate["Canon PowerShot"] = "CPS_"
	modelTranslate["CanonPowerShot"] = "CPS_"
	modelTranslate["EOS REBEL"] = ""
	modelTranslate["EOSREBEL"] = ""
	modelTranslate["motorola DROID3"] = "Droid3"
	modelTranslate["motorolaDROID3"] = "Droid3"
	modelTranslate["FC300S"] = "DJI-Phantom3Adv"
	modelTranslate["FC330"] = "DJI-Phantom4"
	modelTranslate["HG310Z"] = "DJI-OsmoPlus"

	spaceTranslate := make(map[string]string)
	spaceTranslate[" "] = ""

	// TODO: make ignore patterns not hardcoded
	// TODO: make ignore patterns regex-capable
	pathIgnoreSubstrings = append(pathIgnoreSubstrings, "/.sync/")

	supportedMIMETypes = append(supportedMIMETypes, "image", "video")

	// TODO: make log level configurable
	log.SetFormatter(new(logfmt.NonDebugFormatter))
	log.SetLevel(logrus.InfoLevel)

	startLog := log.WithFields(logrus.Fields{"verb": "startup:"})

	startLog.Infof("I AM %s PLEASE INSERT MEDIA", os.Args[0])
	for k, v := range os.Args {
		startLog.Tracef("arg[%d]: '%s'", k, v)
	}

	merr = nil
	err = nil

	// Hard Requirements
	// check for Exiftool
	exiftoolbin := which.Which("exiftool")
	if exiftoolbin == "" {
		err = errors.New("exiftool binary was not found")
		startLog.Debug((err))
		merr = multierr.Append(merr, err)
	} else {
		startLog.Infof("exiftool found at: %s", exiftoolbin)
	}

	// determine what paths we're working with
	workDir, destRootDir, err = paths.GetMediaPaths()
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
		startLog.Fatalf("failed to unmarshal JSON output from exiftool. reason: %s", err)
	}

	result := gjson.ParseBytes(output)

	fileCount := len(result.Array())
	startLog.Infof("Found %d files to process", fileCount)

SOURCEFILE:
	for k, v := range result.Array() {
		sourceFile := v.Get("SourceFile").String()

		fileLogger := log.WithFields(logrus.Fields{
			"sourceFile": strings.Replace(sourceFile, workDir, "."+dirSep, 1),
			"fileIndex":  k + 1,
			"fileCount":  fileCount,
			"verb":       "  ",
		})

		log.WithFields(logrus.Fields{"verb": "processing:"}).Infof("%s (%d of %d)", sourceFile, k+1, fileCount)
		var timeObj time.Time
		var timeInput int64
		var timestampFound bool
		var sourceSum string
		var serr error

		timestampFound = false
		model := "unknown"
		cameraSerial := ""
		lensSerial := ""
		mimeType := ""
		mimeSubType := ""
		newPathSuffix := ""
		fileExtension := ""

		// ignore files from filtered paths
		for _, substr := range pathIgnoreSubstrings {
			if strings.Contains(sourceFile, substr) {
				fileLogger.WithFields(logrus.Fields{"verb": "skip:"}).Warnf("sourceFile matches an ignore path substring ('%s')", substr)
				continue SOURCEFILE
			}
		}

		sourceFileInfo, err := os.Stat(sourceFile)
		if err != nil {
			fileLogger.WithFields(logrus.Fields{"verb": "skip:"}).Error("could not Stat source file. interesting.")
			continue SOURCEFILE
		}

		if v.Get("FileTypeExtension").Exists() {
			fileExtension = strings.ToLower(v.Get("FileTypeExtension").String())
		} else {
			fileLogger.WithFields(logrus.Fields{"verb": "skip:"}).Infof("File doesn't have an extension? That's weird.")
			continue SOURCEFILE
		}

		if v.Get("MIMEType").Exists() {
			// using "mimeType, mimeSubType, ok := " seems to make the two mime variables local in scope?
			ok := false
			mimeType, mimeSubType, ok = strings.Cut(v.Get("MIMEType").String(), "/")
			if !ok {
				fileLogger.WithFields(logrus.Fields{"verb": "skip:"}).Infof("MIMEType string '%s' could not be cut.", v.Get("MIMEType").String())
				continue SOURCEFILE
			}

			if mimeType == "" || mimeSubType == "" {
				fileLogger.WithFields(logrus.Fields{"verb": "skip:"}).Infof("MIME Type ('%s') or Subtype ('%s') cannot be empty", mimeType, mimeSubType)
				continue SOURCEFILE
			}
		} else {
			fileLogger.WithFields(logrus.Fields{"verb": "skip:"}).Info("MIME type for this file was not found.")
			continue SOURCEFILE
		}

		if !slices.Contains(supportedMIMETypes, mimeType) {
			fileLogger.WithFields(logrus.Fields{"verb": "skip:"}).Infof("The MIME type ('%s') for this file is not supported", mimeType)
			continue SOURCEFILE
		}

		switch {
		case v.Get("SubSecDateTimeOriginal").Exists():
			timeInput, _ = strconv.ParseInt(v.Get("SubSecDateTimeOriginal").String(), 10, 64)
			fileLogger.Debugf("timeInput ('%d') pulled from 'SubSecDateTimeOriginal'", timeInput)
			timeObj = time.UnixMilli(timeInput)
			timestampFound = true

		case v.Get("DateTimeOriginal").Exists():
			timeInput, _ = strconv.ParseInt(v.Get("DateTimeOriginal").String(), 10, 64)
			fileLogger.Debugf("timeInput ('%d') pulled from 'DateTimeOriginal'", timeInput)
			timeObj = time.UnixMilli(timeInput)
			timestampFound = true

		case v.Get("CreateDate").Exists():
			timeInput, _ = strconv.ParseInt(v.Get("CreateDate").String(), 10, 64)
			fileLogger.Debugf("timeInput ('%d') pulled from 'CreateDate'", timeInput)
			timeObj = time.UnixMilli(timeInput)
			timestampFound = true

		case v.Get("ModifyDate").Exists(): // damnit, DROID3!
			timeInput, _ = strconv.ParseInt(v.Get("ModifyDate").String(), 10, 64)
			fileLogger.Debugf("timeInput ('%d') pulled from 'ModifyDate'", timeInput)
			timeObj = time.UnixMilli(timeInput)
			timestampFound = true
		}

		if !timestampFound {
			fileLogger.WithFields(logrus.Fields{"verb": "skip:"}).Info("we did not find a timestamp")
			continue SOURCEFILE
		}

		switch {
		case v.Get("Model").Exists():
			model = v.Get("Model").String()
		case v.Get("AndroidModel").Exists():
			model = v.Get("AndroidModel").String()
		}

		model = strmanip.Strtr(model, modelTranslate)
		model = strmanip.Strtr(model, spaceTranslate)

		if v.Get("SerialNumber").Exists() {
			cameraSerial = v.Get("SerialNumber").String()
			cameraSerial = strmanip.Strtr(cameraSerial, spaceTranslate)
		}

		if v.Get("LensSerialNumber").Exists() {
			lensSerial = v.Get("LensSerialNumber").String()
			lensSerial = strmanip.Strtr(lensSerial, spaceTranslate)
		}

		fileLogger.Debugf("model: %s", model)
		fileLogger.Debugf("cameraSerial: %s", cameraSerial)
		fileLogger.Debugf("lensSerial: %s", lensSerial)
		fileLogger.Debugf("timestamp: %s", timeObj.String())
		fileLogger.Debugf("MIME: %s / %s", mimeType, mimeSubType)
		fileLogger.Debugf("fileExtension: %s", fileExtension)

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

		// if cameraSerial != "" {
		// 	newFileName = fmt.Sprintf("%s_CS%s", newFileName, cameraSerial)
		// }

		// if lensSerial != "" {
		// 	newFileName = fmt.Sprintf("%s_LS%s", newFileName, lensSerial)
		// }

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
			sourceSum, serr = checksum.SHA256sum(sourceFile)
			if serr != nil {
				fileLogger.WithFields(logrus.Fields{"verb": "skip:"}).Error("couldn't checksum the source file.")
				continue SOURCEFILE
			}
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
		fileLogger.Debugf("creating target directory: %s", targetDir)
		err = os.MkdirAll(targetDir, 0755)
		if err != nil {
			fileLogger.Errorf("could not create destination directory! reason: %s", err)
		}

		err = os.Rename(sourceFile, destFile)
		if err != nil {
			fileLogger.WithFields(logrus.Fields{"verb": "error:"}).Errorf("could not rename file! reason: %s", err)
		} else {
			fileLogger.WithFields(logrus.Fields{"verb": "renamed:"}).Infof(">> %s", destFile)
		}

	} // ends: for k, v := range result.Array()
}
