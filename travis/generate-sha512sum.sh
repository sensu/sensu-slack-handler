#!/bin/bash


if [ -d dist ]; then
  files=( dist/*sha256-checksums.txt )  
  file=`basename ${files[0]}`
  IFS=\_ read -r package prefix x <<< $file
  if [ -n "$prefix" ]; then
    echo "Generating sha512sum for ${package}_${prefix}"
    cd dist
    sha512sum *.tar.gz > "${package}_${prefix}_sha512-checksum.txt"
    cat "${package}_${prefix}_sha512-checksum.txt"
  fi
fi
