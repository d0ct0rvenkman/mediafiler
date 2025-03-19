
Test Case format
```
[
  {
    "casename": "",
    "expected": {
      "newPathSuffix": "",
      "newFileName": "",
      "fileExtension": "",
      "err": ""
    },
    "metadata": { /* single-file exiftool output */ }
  }
]
```

Generate test case skeletons from existing files with this helper function:
```
function mediafiler_testcase {
  exiftool -r -json -dateFormat "%s%-3f" $1 | jq '[{casename:"", expected: {newPathSuffix:"", newFileName:"", fileExtension:"", err:""}, metadata: .[]}]' | sed -r 's/"GPS(Latitude|Longitude|Position|Coordinates|Coordinates-err)": "(.*?)"([,]?)$/"GPS\1": "redacted"\3/g'
}
```


