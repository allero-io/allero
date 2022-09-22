#!/bin/bash

if [ $# -lt 1 ]
  then
    echo "Missing version"
    exit
fi

VERSION=$1

SHA256_MAC_INTEL=$(cat ./dist/checksums.txt | grep Darwin_x86_64 | cut -d" " -f1)
SHA256_MAC_ARM=$(cat ./dist/checksums.txt | grep Darwin_arm64 | cut -d" " -f1)
SHA256_LINUX_INTEL=$(cat ./dist/checksums.txt | grep Linux_x86_64 | cut -d" " -f1)
SHA256_LINUX_ARM=$(cat ./dist/checksums.txt | grep Linux_arm64 | cut -d" " -f1)

cat > homebrew-allero/allero.rb <<-EOF
# typed: false
# frozen_string_literal: true
class Allero < Formula
  desc ""
  homepage "https://allero.io/"
  version "$VERSION"
  if OS.mac? && Hardware::CPU.intel?
    url "https://github.com/allero-io/allero/releases/download/$VERSION/allero_${VERSION}_Darwin_x86_64.zip"
    sha256 "$SHA256_MAC_INTEL"
  end
  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/allero-io/allero/releases/download/$VERSION/allero_${VERSION}_Darwin_arm64.zip"
    sha256 "$SHA256_MAC_ARM"
  end
  if OS.linux? && Hardware::CPU.intel?
    url "https://github.com/allero-io/allero/releases/download/$VERSION/allero_${VERSION}_Linux_x86_64.zip"
    sha256 "$SHA256_LINUX_INTEL"
  end
  if OS.linux? && Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
    url "https://github.com/allero-io/allero/releases/download/$VERSION/allero_${VERSION}_Linux_arm64.zip"
    sha256 "$SHA256_LINUX_ARM"
  end
  def install
    bin.install "allero"
  end
  def caveats
    <<~EOS
      \033[32m[V] Downloaded Allero
      [V] Finished Installation
      \033[35m Usage: $ allero fetch github <owner|owner/repo ...>
    EOS
  end
end
EOF
