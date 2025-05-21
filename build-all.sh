#!/bin/bash

cd $(dirname "$0")/

cd execs
rm -f k9s*
cd -
env GOOS=darwin GOARCH=arm64 make build
cd execs
zip k9s_darwin_arm64.zip k9s*
cd -


cd execs
rm -f k9s*
cd -
env GOOS=darwin GOARCH=amd64 make build
cd execs
zip k9s_darwin_amd64.zip k9s*
cd -


cd execs
rm -f k9s*
cd -
env GOOS=linux GOARCH=amd64 make build
cd execs
zip k9s_linux_amd64.zip k9s*
cd -


cd execs
rm -f k9s*
cd -
env GOOS=windows GOARCH=amd64 make build
cd execs
zip k9s_windows_amd64.zip k9s*
cd -
