# Architecture

## Agent-based

hpt requires an hpt binary to be sitting on a target machine. This binary is
passed a provisioning config, either directly at the command line or over gRPC.

## gRPC

To provision targets remotely, we use gRPC. This communication channel is 
encrypted with (a fork of) curvetls, a simple library that encrypts tcp 
transports with asymmetric public-key encryption.

During remote provisioning the machine where `hpt` is invoked is the **client**,
and the **target** machine's instance running `hpt serve` is the server, waiting
for connections. Both must authenticate each other with the other's public key. 
A server will only accept connections from clients whose public key is in its 
keystore. A client will only connect to targets if it knows the target/server's 
public key. This raises the practical problem of delivering a keystore to the 
target before it can be provisioned.

The client also has a keystore. This is a database of 1) it's own keypair and
2) all the known target keypairs, identified by their unique name. Passing 
`--target foo` at the command line will select a keypair from the database by
that name. Given that `--target` and `--ip` are fundamentally decoupled, it is
possible to easy reuse a target keystore on many hosts. This is suitable for 
testing, but really you should create unique keys for every host you provision.

### Generating a Keystore for a Target

Generating keystores happens client-side. This is a scenario similar
to SaltStack, where the is a "master" server and many "minions" the server
talks to and provisions. Here our client is playing the master role.

Generate a new keypair for target "foo" on the client by running

```
hpt gen-target-keys --target foo
```

Two things will happen.

1. The client's local keystore will be modified. A new keypair entry for target
   "foo" will be added, with a populated public key and a missing private key.
2. A completely new boltdb database file will be generated. This is the
   keystore we need to deliver to the target. Inside, the target's public and 
   private key are at a well-known location, and, importantly, the client's public
   key is also available. The server uses the client's public key to authenticate
   any gRPC connection.

A target's name is _arbitrary_. Unlike x509 certificates, curvetls keypairs 
don't have any metadata internal to them. The target name is simply a way for
the client to pick a keypair when connecting remotely during provisioning.

```
hpt --target=foo --ip=1.2.3.4 provision_foo.toml
```

The client's keystore path can be set with the env var `HPT_KEYSTORE` or the 
`--keystore` flag.

Print the known targets and their public keys.

```
HPT_KEYSTORE=~/.config/hpt/keystore.db hpt list-target-keys 
```



