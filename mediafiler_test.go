package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/d0ct0rvenkman/mediafiler/internal/strmanip"
	"github.com/tidwall/gjson"
)

/*
Test_generateFilenameBase will load in test metadata sampled from real cameras
to simulate various permutations of metadata presence/absence. Given that even the same
camera can populate that metadata differently depending on environmental circumstances,
we'll likely need a lot of test cases.

The tester will load in JSON files to populate individual test cases. Each file should contain
metadata for one or more simulated input files, along with test case data describing the test
case name and expected results.

The test will not do any replace actions on any of the file name components aside from removing
forward and back slashes, whish shouldn't be present anyway.

Check out README.md in ./test/main/generateFilenameBase for more info on test file structure.
*/
func Test_generateFilenameBase(t *testing.T) {
	casesProcessed := 0

	testDataPath := "test/main/generateFilenameBase/"

	var supportedMIMETypes []string
	supportedMIMETypes = append(supportedMIMETypes, "image", "video")

	var modelReplacer strmanip.Replacer
	var spaceReplacer strmanip.Replacer

	spaceReplacer.AddRule(strmanip.ReplacerRule{Type: "string", Find: `/`, ReplaceWith: "_"})
	spaceReplacer.AddRule(strmanip.ReplacerRule{Type: "string", Find: `\`, ReplaceWith: "_"})

	testFileList := make([]string, 0)
	e := filepath.Walk(testDataPath, func(path string, f os.FileInfo, err error) error {
		//t.Logf("checking %s\n", path)
		if err != nil {
			fmt.Println(err)
			return err
		}
		if (path[len(path)-5:] == ".json") && (f.Mode().IsRegular()) {
			//t.Logf("path %s is a file with the correct extension\n", path)
			testFileList = append(testFileList, path)
		}

		return err
	})

	if e != nil {
		t.Errorf("could not walk test data directory. reason: '%s'", e)
	}

	if len(testFileList) == 0 {
		t.Fatal("the test won't work without test data")
	}

	for k, v := range testFileList {
		t.Logf("processing test file '%s' (%d of %d)\n", v, k+1, len(testFileList))

		fileContents, err := os.ReadFile(v)
		if err != nil {
			t.Errorf("could not read test data from '%s'\n", v)
		}

		if len(fileContents) == 0 {
			t.Fatalf("test case was empty")
		}

		if !gjson.ValidBytes(fileContents) {
			t.Fatalf("failed to unmarshal JSON output from test case")
		}

		result := gjson.ParseBytes(fileContents)

		fileCount := len(result.Array())
		t.Logf("- Found %d simulated files to process", fileCount)

		for casenum, testcase := range result.Array() {
			casesProcessed++

			var casename string
			var fqcasename string
			var exp_newPathSuffix string
			var exp_newFileName string
			var exp_fileExtension string
			var exp_err string
			var tmpjson gjson.Result

			tmpjson = testcase.Get("casename")
			if !tmpjson.Exists() {
				t.Fatalf("test case name for simulated file %d in %s is missing", casenum, v)
			} else {
				casename = tmpjson.String()
			}

			if len(casename) == 0 {
				t.Fatalf("test case name for simulated file %d in %s is empty", casenum, v)
			}

			fqcasename, _ = spaceReplacer.Replace(v[len(testDataPath):] + "-" + fmt.Sprintf("%d", casenum) + "-" + casename)

			t.Run(fqcasename, func(t *testing.T) {

				tmpjson = testcase.Get("expected.newPathSuffix")
				if !tmpjson.Exists() {
					t.Fatalf("expected newPathSuffix for simulated file %d in %s is missing", casenum, v)
				} else {
					exp_newPathSuffix = tmpjson.String()
				}

				tmpjson = testcase.Get("expected.newFileName")
				if !tmpjson.Exists() {
					t.Fatalf("expected newFileName for simulated file %d in %s is missing", casenum, v)
				} else {
					exp_newFileName = tmpjson.String()
				}

				tmpjson = testcase.Get("expected.fileExtension")
				if !tmpjson.Exists() {
					t.Fatalf("expected fileExtension for simulated file %d in %s is missing", casenum, v)
				} else {
					exp_fileExtension = tmpjson.String()
				}

				tmpjson = testcase.Get("expected.err")
				if !tmpjson.Exists() {
					t.Fatalf("expected err for simulated file %d in %s is missing", casenum, v)
				} else {
					exp_err = tmpjson.String()
				}

				tmpjson = testcase.Get("metadata")
				if !tmpjson.Exists() {
					t.Fatalf("metadata for simulated file %d in %s is missing", casenum, v)
				}

				if len(casename) == 0 {
					t.Fatalf("test case name for simulated file %d in %s is empty", casenum, v)
				}

				newPathSuffix, newFileName, fileExtension, err := generateFilenameBase(tmpjson, supportedMIMETypes, modelReplacer, spaceReplacer)
				if (err != nil) && (err.Error() != exp_err) {
					t.Errorf("generateFilenameBase() err = %v, exp_err %v", err, exp_err)
					return
				}
				if newPathSuffix != exp_newPathSuffix {
					t.Errorf("generateFilenameBase() newPathSuffix = %v, want %v", newPathSuffix, exp_newPathSuffix)
				}
				if newFileName != exp_newFileName {
					t.Errorf("generateFilenameBase() newFileName = %v, want %v", newFileName, exp_newFileName)
				}
				if fileExtension != exp_fileExtension {
					t.Errorf("generateFilenameBase() fileExtension = %v, want %v", fileExtension, exp_fileExtension)
				}
			})

		}

	}

	if casesProcessed == 0 {
		t.Fatalf("the test didn't process any test cases")
	} else {
		t.Logf("the test processed %d simulated media files", casesProcessed)
	}

}
