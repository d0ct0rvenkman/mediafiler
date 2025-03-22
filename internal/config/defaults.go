package config

var defaultConfigYAML = []byte(`
debug: false
dry-run: false
exiftool-binary: /usr/bin/exiftool
model-replace-rules:
- replace_type: "string"
  find_pattern: "FooBarMatic"
  replace_with: "FBM"
- replace_type: "regex"
  find_pattern: '\s+'
  replace_with: ""
path-ignore-patterns:
- type: "string"
  pattern: '.git/'
- type: "string"
  pattern: '.git\'
- type: "regex"
  pattern: '^.*[Ii][Cc][Oo]$'
  
`)
