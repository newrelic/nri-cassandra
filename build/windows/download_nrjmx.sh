#!/usr/bin/env bash
set -e

# Downloads nrjmx v2.10.1 MSI installer

echo "Downloading nrjmx v2.10.1 MSI installer"
if [[ -z $NRJMX_URL ]]; then
    NRJMX_VERSION="2.10.1"
    echo "Using nrjmx version $NRJMX_VERSION."
    NRJMX_URL=https://github.com/newrelic/nrjmx/releases/download/v$NRJMX_VERSION/nrjmx-amd64.$NRJMX_VERSION.msi
fi

curl -LSs --fail "$NRJMX_URL" -o "build/package/windows/bundle/nrjmx-amd64.msi"