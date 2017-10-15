# hpt

Really simple host provisioning.


Goals:

* add users
* put files on hosts
* install services
* install packages
* arbitrary execution
* declarative configs 
* really easy dsl
* mergable configs for code reuse

Non-goals:

* launch servers (use terraform instead)
* build images (use packer instead)


## Usage: on the host 

Put `hpt` on the path and provide config(s)

```
sudo hpt config.toml
```

This is easy during packer builds.


