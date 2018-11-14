#!/bin/sh

CURL=$(which curl)

if [ "$?" = "1" ]; then
    echo "You need curl to use this script."
    exit 1
fi

VERSION=$(curl -sI https://github.com/nervo/manala/releases/latest | grep Location | awk -F"/" '{ printf "%s", $NF }' | tr -d '\r')

if [ ! ${VERSION} ]; then
    echo "Failed while attempting to install manala. Please manually install:"
    echo ""
    echo "1. Open your web browser and go to https://github.com/nervo/manala/releases"
    echo "2. Download the latest release for your platform. Call it 'manala'."
    echo "3. chmod +x ./manala"
    echo "4. mv ./manala /usr/local/bin"
    exit 1
fi

HAS=$(which manala)

if [ "$?" = "0" ]; then
    echo
    echo "You already have manala!"
    export N=3
    echo "Overwriting in ${N} seconds.. Press Control+C to cancel."
    echo
    sleep ${N}
fi

case $(uname) in
    "Darwin")
        OS="darwin"
        ARCH="amd64"
    ;;
    "Linux")
        OS="linux"
        ARCH="amd64"
    ;;
esac

URL=https://github.com/nervo/manala/releases/download/${VERSION}/manala_${OS}_${ARCH}
DST=/usr/local/bin/manala

echo "Downloading package ${URL} as ${DST}"

curl -sSL ${URL} --output ${DST}

chmod +x ${DST}

echo "Download complete, happy manaling!"
