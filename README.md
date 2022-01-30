# gink-go
Ginkgo is a p2sp file transfer tool designed for ML dataset.

## Build

```bash
go build cmd/ginkgo.go
```

## Usage

On server side, just start server with the following command:

```bash
ginkgo 
```

On client side, just run ginkgo as a scp command, both directory and file are supported.

```bash
ginkgo SrcHost:/path/to/src /path/to/dest
```

- The `/path/to/src` can be either absolute or relative.
- The src and dest path handling behavior is just compatible with GNU scp command

## How

- Client will get other clients list from server side and broadcast itself to them.
- Client chooses the block to download according to consistent hashing result of its serving `host:port`.
- Other clients also choose which peer to download the block from by the hash.
- The seed.Seed structure records the blocks and files metadata.


Golang version of [gingko](https://github.com/auxten/gingko)

> `ginkgo` is an alternative form of `gingko`. `Gink-go` is the golang version `Gingko`.


## Todo

- [x] HTTP range downloader
- [x] HTTP file server
- [x] Directory support
- [x] Consistent hashing locator
- [ ] Client rate limit
- [ ] Server rate limit
