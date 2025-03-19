
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

Generate test case skeletons from existing files in the current directory:
```exiftool -r -json -dateFormat "%s%-3f" . | jq '[{casename:"", expected: {newPathSuffix:"", newFileName:"", fileExtension:"", err:""}, metadata: .[]}]'```