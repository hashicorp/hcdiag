# Installation

There are several ways to install hcdiag, depending on your preferences.

## Official Release Binaries

You can always manually download and install a release binary from our [releases page](https://releases.hashicorp.com/hcdiag/).

## Package Managers

The easiest way to download the latest stable version of hcdiag is through a package manager. We provide an `hcdiag` package via the Hashicorp repository on Ubuntu, Debian, and Red Hat Linux, along with Mac OS via `homebrew`.

The process is generally the same as for any other package: add the hashicorp repository, and then download the `hcdiag` package.

The advantage of installing via a package manager is that hcdiag will automatically receive any available updates when you update/upgrade your packages.

### Example: Ubuntu

On the latest LTS version of Ubuntu 22.04, you'd run:

```
# Add the hashicorp repository
wget -O- https://apt.releases.hashicorp.com/gpg | gpg --dearmor > /usr/share/keyrings/hashicorp-archive-keyring.gpg && echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com jammy main" | tee /etc/apt/sources.list.d/hashicorp.list

# update repository information
apt-get update

# install hcdiag
apt-get install -y hcdiag
```

### Example: homebrew

Using `homebrew` on Mac OS:
```
brew install hashicorp/tap/hcdiag
```
