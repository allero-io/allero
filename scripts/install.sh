#!/bin/bash
set -e

osName=$(uname -s)

osArchitecture=$(uname -m)

if [[ $osArchitecture == *'aarch'* || $osArchitecture == *'arm'* ]]; then
	osArchitecture='arm64'
fi

if ! [[ -d /usr/local/bin/ ]]
then
	mkdir -p "/usr/local/bin/" 2> /dev/null || sudo mkdir -p "/usr/local/bin/"
fi

DOWNLOAD_URL=$(curl --silent "https://api.github.com/repos/allero-io/allero/releases/latest" | grep -o "browser_download_url.*\_${osName}_${osArchitecture}.zip")
DOWNLOAD_URL=${DOWNLOAD_URL//\"}
DOWNLOAD_URL=${DOWNLOAD_URL/browser_download_url: /}


OUTPUT_BASENAME=allero-latest
OUTPUT_BASENAME_WITH_POSTFIX=$OUTPUT_BASENAME.zip

echo "Installing Allero..."
echo

curl -sL $DOWNLOAD_URL -o $OUTPUT_BASENAME_WITH_POSTFIX
echo -e "\033[32m[V] Downloaded Allero\033[0m"

if ! unzip >/dev/null 2>&1;then
    echo -e "\033[31;1m error: unzip command not found \033[0m"
    echo -e "\033[33;1m install unzip command in your system \033[0m"
    exit 1
fi

unzip -qq $OUTPUT_BASENAME_WITH_POSTFIX -d $OUTPUT_BASENAME

rm -f /usr/local/bin/allero 2> /dev/null || sudo rm -f /usr/local/bin/allero
cp $OUTPUT_BASENAME/allero /usr/local/bin 2> /dev/null || sudo cp $OUTPUT_BASENAME/allero /usr/local/bin

rm $OUTPUT_BASENAME_WITH_POSTFIX
rm -rf $OUTPUT_BASENAME

echo -e "\033[32m[V] Finished Installation\033[0m"

echo
