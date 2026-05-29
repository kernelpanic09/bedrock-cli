# Installation

## Homebrew (macOS / Linux)

```sh
brew install kernelpanic09/tap/bedrock-cli
```

The tap will be published at `github.com/kernelpanic09/homebrew-tap` once the first release is cut.

## go install

If you have Go 1.22+ installed:

```sh
go install github.com/kernelpanic09/bedrock-cli/cmd/bedrock-cli@latest
```

This puts the binary in `$GOPATH/bin` (usually `~/go/bin`). Make sure that's on your PATH.

## Prebuilt binaries

Download the latest release from the [releases page](https://github.com/kernelpanic09/bedrock-cli/releases).
Binaries are available for:

- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

Checksums are published alongside each release.

```sh
# Example: Linux amd64
curl -LO https://github.com/kernelpanic09/bedrock-cli/releases/latest/download/bedrock-cli_Linux_x86_64.tar.gz
tar -xzf bedrock-cli_Linux_x86_64.tar.gz
chmod +x bedrock-cli
sudo mv bedrock-cli /usr/local/bin/
```

## Build from source

```sh
git clone https://github.com/kernelpanic09/bedrock-cli
cd bedrock-cli
make install
```

## Verify the install

```sh
bedrock-cli version
```
