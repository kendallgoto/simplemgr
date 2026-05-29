# dfu

Contains MCUboot header information for DFU package binaries

## Regenerate
```sh
kaitai-struct-compiler ./dfu/dfu.ksy --no-auto-read -t go --read-pos --outdir . --go-package dfu
```
