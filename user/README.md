# `os/user`

## Why?

Default `os/user` doesn't provide any interface to work with other systems. So
this package creates unified interface for different implementations of getting
POSIX users and groups (you can implement ACL logic even inside your web app!)
