#!/bin/bash


if [ -d dist ]; then
  files=( dist/*sha256-checksums.txt )  
  file=`basename ${files[0]}`
  IFS=\_ read -r package prefix x <<< $file
  if [ -n "$prefix" ]; then
    echo "Generating sha512sum for ${package}_${prefix}"
    cd dist
    sha512_file="${package}_${prefix}_sha512-checksums.txt"
    echo "${sha512_file}" > sha512_file
    echo "sha512_file: $(cat sha512_file)"
    sha512sum *.tar.gz > "${sha512_file}"
    echo ""
    cat "${sha512_file}"
  fi
else
  echo "error"
fi
