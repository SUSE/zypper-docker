# zypper-docker documentation

This directory contains the needed files to generate the man pages for
`zypper-docker`. In order to generate the man pages, execute the following:

```
$ godep go run generate.go
```

This should create a directory named `man1`, with all the man pages in there.
Alternatively, you can use the task on the `Makefile` file in the root
directory. Thus, an equivalent to the above command would be:

```
$ make man
```

Once you have installed the generated man pages, you can read them by
performing commands like:

```
$ man zypper-docker
$ man zypper-docker images
$ man zypper-docker-images # identical to the previous one.
```
