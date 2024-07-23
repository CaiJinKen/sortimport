
# sortimport

Sort/group imports in golang file

The order follows:

```go
import (
    sdk-packages   // eg: io/os ...

    inner-packages // self-define packages in current project

    outer-packages // eg: github.com/user/repo

)

```

## Installation

```sh
% go install github.com/CaiJinKen/sortimport@v0.1.0
```

## Usage

```sh
% sortimport -file=<filename> -writeback=true -std-out=false
or
% sortimport -file <filename> -writeback -std-out false
```

Flags:

```sh
-file string
    filename
-only-changed
    just print changed line, false will print all info
-std-out
    print info into stdout (default true)
-version string
    print fillstruct version
-writeback
    writeback to the file
```
