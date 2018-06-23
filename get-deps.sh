#!/bin/bash

while read package version; do
  go get -u "$package"
  ( cd "$GOPATH/src/$package" && git checkout "$version" )
done
