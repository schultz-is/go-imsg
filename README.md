# go-imsg

![Tests](https://github.com/schultz-is/go-imsg/workflows/Tests/badge.svg)

This package provides tooling around OpenBSD's
[imsg functions](https://man.openbsd.org/imsg_init.3).These imsg functions are
used for inter-process communication, generally over unix sockets, and generally
where the communicating processes have different privileges. Examples of imsg
usage can be found in several utilities including [OpenBGP](http://openbgp.org/)
and [tmux](https://github.com/tmux/tmux).


## Usage

### Message Creation

The easiest way to create an imsg using this library is to call the
`ComposeIMsg()` package method. Doing so will automatically populate the `PID`
by calling `os.Getpid()` as a convenience. For example:

```go
package main

import (
  "log"

  "github.com/schultz-is/go-imsg"
)

const MsgTypeSaySomething = 1234

func main() {
  im, err := imsg.ComposeIMsg(
    MsgTypeSaySomething,      // Type
    0,                        // Peer ID
    []byte("Hello, world!"),  // Data
  )
  if err != nil {
    log.Fatal(err)
  }

  log.Printf("%#v", im)
}
```

```console
$ go run main.go
2009/11/10 23:00:00 &imsg.IMsg{Type:0x4d2, PeerID:0x0, PID:0x3, Data:[]uint8{0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21}, flags:0x0}
```

[Open in go playground](https://play.golang.org/p/dZPkuEIlHDf)

Creating an imsg can also be done manually:

```go
package main

import (
  "log"

  "github.com/schultz-is/go-imsg"
)

const MsgTypeSaySomething = 1234

func main() {
  im := &imsg.IMsg{
    Type: MsgTypeSaySomething,
    Data: []byte("Hello, world!"),
  }

  log.Printf("%#v", im)
}
```

```console
$ go run main.go
2009/11/10 23:00:00 &imsg.IMsg{Type:0x4d2, PeerID:0x0, PID:0x0, Data:[]uint8{0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21}, flags:0x0}
```

[Open in go playground](https://play.golang.org/p/KtS6eBVlpYi)

### Message Serialization

To prepare an imsg for transmission across the wire, simply marshal it into a
binary representation:

```go
package main

import (
  "log"

  "github.com/schultz-is/go-imsg"
)

const MsgTypeSaySomething = 1234

func main() {
  im := &imsg.IMsg{
    Type: MsgTypeSaySomething,
    Data: []byte("Hello, world!"),
  }

  bs, err := im.MarshalBinary()
  if err != nil {
    log.Fatal(err)
  }

  log.Printf("% x", bs)
}
```

```console
$ go run main.go
2009/11/10 23:00:00 d2 04 00 00 1d 00 00 00 00 00 00 00 00 00 00 00 48 65 6c 6c 6f 2c 20 77 6f 72 6c 64 21
```

[Open in go playground](https://play.golang.org/p/fedGWpZrRoj)

### Message Parsing

Reading and parsing an imsg directly from a socket is as simple as providing the
socket (or any `io.Reader`,) to the `ReadIMsg()` package method:

```go
package main

import (
  "bytes"
  "log"

  "github.com/schultz-is/go-imsg"
)

func main() {
  buf := bytes.NewReader(
    []byte{
      0xaa, 0xaa, 0xaa, 0xaa, // type
      0x1d, 0x00,             // len
      0x00, 0x00,             // flags
      0xbb, 0xbb, 0xbb, 0xbb, // peerid
      0xcc, 0xcc, 0xcc, 0xcc, // pid

      // Hello, world!
      0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21,
    },
  )

  im, err := imsg.ReadIMsg(buf)
  if err != nil {
    log.Fatal(err)
  }

  log.Printf("%#v", im)
}
```

```console
$ go run main.go
2009/11/10 23:00:00 &imsg.IMsg{Type:0xaaaaaaaa, PeerID:0xbbbbbbbb, PID:0xcccccccc, Data:[]uint8{0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21}, flags:0x0}
```

[Open in go playground](https://play.golang.org/p/awU33secF8G)


## Data Layout

An imsg is constructed of a 16 byte header followed by accompanying data of a
length provided by the `len` field in the header. An imsg must be at least 16
bytes (this occurs when the message contains no ancillary data,) and can be no
larger than the currently defined maximum of 16384 bytes. When written to disk
(or across a socket,) the data is laid out as follows:

| Field      | Type       | Description                                    |
|------------|------------|------------------------------------------------|
| **type**   | `uint32_t` | Describes the intent of the message            |
| **len**    | `uint16_t` | Total length of the message (including header) |
| **flags**  | `uint16_t` | _(Internal use only)_                          |
| **peerid** | `uint32_t` | Free for use by caller                         |
| **pid**    | `uint32_t` | Free for use by caller                         |
| **data**   | `void *`   | Ancillary data accompanying the message        |

For a very simple example, on an `amd64` system, the hexidecimal values of an
imsg with small ancillary data (the word "hi") might look like so:

```
aa aa aa aa 12 00 00 00 bb bb bb bb cc cc cc cc 68 69
```

When grouped by field, that data looks like so:

```
aa aa aa aa  // type
12 00        // len (0x0012 / 18 bytes)
00 00        // flags
bb bb bb bb  // peerid
cc cc cc cc  // pid
68 69        // data (0x6869 / "hi")
```


## Documentation

Documentation for this package is available at
[pkg.go.dev](https://pkg.go.dev/github.com/schultz-is/go-imsg?tab=doc).
