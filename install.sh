#!/usr/bin/env bash

set -e

if [[ "x$TOKEN" == "x" ]]; then
  echo "TOKEN is not set"
    exit 1
fi

cd ~

mkdir -p bin
curl --location -o gpc.zip --header "PRIVATE-TOKEN: $TOKEN" "https://gitlab.g-portal.se/api/v4/projects/gpcloud%2Fgpcloud-cli/jobs/artifacts/master/download?job=build"
unzip gpc.zip -d bin/
chmod +x bin/gpc

echo "Consider adding ~/bin to your PATH. Do you want to do that now? [y/N] "
read -r answer
if [[ "$answer" =~ ^([yY][eE][sS]|[yY])+$ ]]; then
  echo "export PATH=$PATH:~/bin" >> ~/.bashrc
  source ~/.bashrc
fi
