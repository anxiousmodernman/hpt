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

The TOML provisioner config will be applied, and the results of the apply will 
be printed to the console. If no changes were required, none should be applied.

## Usage: grpc (experimental)

Create a target keystore per the instructions in **docs/architecture.md**, then
run hpt as a server on the target machine.

```
hpt serve --keystore /path/to/keystore.db
```

From the client that generated the (named) target's keystore, provide both 
`--target` and `--ip` to remotely connect and provision the target.

```
hpt --target=foo --ip=<ip addr> config.toml
```

## Development status

hpt is only tested on Centos-based machines for now, but since most management
tasks are shelled-out calls to systemd and other linux utilities, adding support
for other distros is feasible.

## CurveTLS fork

To avoid the pain of generating x509 certs, we're using a somewhat experimental
transport layer based on a fork of **Rudd-O/curvetls**. There are some 
issues around [high CPU usage](#17), so regular TLS support will probably
be added in the future.

