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
# note only shorthand 'packages' is implemented currently
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

A `[[user]]` block manages users on the system.

```toml
[[user]]
name = "carlos"
home = "/home/carlos"
groups = ["wheel"]
ssh_public_key = "bkt://key.pub"
ssh_private_key = "bkt://key"
shell = "/bin/bash"
absent = false
```

The following attributes are supported:

**name** - The name of the user

**home** - The home directory to create for the user. No home is created if this
attribute is blank.

**groups** - An array of groups the user should be added to. Groups must exist 
or be previously created with a `[[group]]` block.

**ssh_public_key** - A valid resolver path (such as a bucket) where a public key
can be fetched as bytes and written to disk at `$HOME/.ssh/id_rsa.pub` for the user.

**ssh_private_key** - A valid resolver path (such as a bucket) where a private
key can be fetched as bytes and written to disk at `$HOME/.ssh/id_rsa` for the 
user.

**shell** - The login shell for the user. 

**absent** - A boolean that indicates whether a user should be present or absent
on a system. Note that if `false` is applied, a user will be deleted, but some
filesystem artifacts (like the home directory) will remain.


## group

The `[[group]]` block manages groups on a system.

```toml
[[group]]
name = "devs"
gid = "1099"
passwordless_sudo = true
```

The following attributes are supported:

**name** - the name of the group

**gid** - a GID for the group. If none is provided, the system will assign one. 
A provided GID must not be in use.

**passwordless_sudo** - if set, hpt will modify **/etc/sudoers** such that group
members can use sudo without a password.

**absent** - a boolean that indicates whether a group should be on the system.
If true, hpt will attempt to remove the group. 


## clone

The `[[clone]]` block clones a git repository to disk. Only cloning public repos
is currently supported. Authenticated repos are a TODO.

```toml
[[clone]]
url = "https://git.foo.com/project"
dest = "/opt/project"
```

The following attributes are supported:

**url** - a valid git repository url

**dest** - a directory on disk to clone into

## exec

The `[[exec]]` block supports arbitrary command execution. 

```toml
[[exec]]
script = """
mv foo.txt /tmp
"""
user = "someone"
```

The following attributes are supported:

**script** - a multi-line string that will be interpreted as a script. The 
string is written to a temporary location and executed.

**user** - the user to run the script as.

**pwd** - the directory to run the exec block from. Defaults to **/tmp**.

## file

The file block manages files and directories. 

```toml
[[file]]
path = "/opt/dest"
source = "bkt://some/file"
perms = 0600
dir = false
owner = "someone"
group = "someone"
```

The following attributes are supported:

**path** - the path on disk the file block represents

**source** - a resolver path to the file

**perms** - permissions for the file, in octal. To specify octal you must have a
leading 0.

**dir** - if true, the file block will be interpreted as a directory.

**owner, group** - user and group ownership for the file 

**absent** - if true, the file(s) on disk will be removed.

## service

The `[[service]]` block is a wrapper around `systemctl`.

```toml
[[service]]
status = "started"
enabled = true
```

The following attributes are supported:

**status** - one of: started, restarted, stopped

**enabled** - if true, the service will start on boot


