# smp
Contains MCUmgr SMP protocol command/response payloads

## Regenerate
```sh
kaitai-struct-compiler ./smp/smp.ksy --no-auto-read -t go --read-pos --outdir . --go-package smp
```
