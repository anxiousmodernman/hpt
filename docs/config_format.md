# Provisioning Config File

Here we document the TOML format for files that provision servers. If executed
on the server, hpt takes a config file as an argument.

```
hpt example.toml
```

The name **example.toml** is not important, only that it is a valid TOML file.
The following configuration blocks can be configured.


## package

A `[[package]]` block installs or uninstalls a package using the standard system
package manager. This is `apt` on Debian systems and `yum` on Centos systems.

```toml
[[package]]
name = "nginx"
state = "installed"
```

A `packages` shorthand is also available for the common "installed" case. This 
will install 3 packages on the target machine if it appears in a config file 
before any other block.

```toml
packages =  ["nginx", "libvirt", "python-dev"]
```

The following attributes are available:

**name** - This should exactly match the identifier that works in the system
package repositories.

**state** - one of: installed, absent


## user

## clone

## exec

## file

## service

