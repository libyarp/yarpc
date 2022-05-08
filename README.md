# yarpc

yarpc implements a compiler capable of generating source code from  YARP's IDL 
files.

## Installing

Currently, YARP is considered unstable. The only way to install `yarpc` is by
compiling it manually.

After installing Go, run:

```
go install github.com/libyarp/yarpc/cmd/yarpc@latest
```

## Usage

`yarpc`'s CLI interface is quite straightforward. Run `yarpc --help` for
available commands.
For information, for instance, on how to compile IDLs to Go, use `yarpc go --help`.
