# hpt

Really simple host provisioning.

## Overview

This project is aiming to be a simple and general provisioning tool for modern
linux servers. Ultimately we want hpt to do the kinds of things that Ansible and
SaltStack do, but with TOML config and the joy of Go.

## Roadmap

The hope is that by focusing on robust local execution first, we can design a 
tool that feels similar whether it's being run locally or over a network.

* [x] local execution - run hpt during a packer build or straight up "on the box" 
* [ ] remote execution - SSH? gRPC? It's TBD.

## Usage: on the host 

Put `hpt` on the PATH and provide config(s) as arguments.

```
sudo hpt config.toml
```

The config will be applied, and the results of the apply will be printed to the
console. If no changes were required, none will be applied.



