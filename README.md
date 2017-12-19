# hpt

A host provisioning tool.

## Overview

This project is aiming to be a simple and general provisioning tool for modern
linux servers. Ultimately we want hpt to do the kinds of things that Ansible and
SaltStack do, but with TOML config and the joy of Go.

## Usage: on the host 

Put `hpt` on the PATH and provide config(s) as arguments.

```
sudo hpt config.toml
```

The config will be applied, and the results of the apply will be printed to the
console. If no changes were required, none will be applied.

## Development status

hpt is only tested on Centos-based machines for now, but since most management
tasks are shelled-out calls to systemd and other linux utilities, adding support
for other distros is feasible.

