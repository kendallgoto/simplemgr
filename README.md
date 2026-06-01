# simplemgr

MCUmgr style SMP interface for serial MCUboot recovery

## Install

```sh
go get github.com/kendallgoto/simplemgr
```

## Example
```go
serialPort, _ := serial.Open("/dev/ttyACM0", &serial.Mode{
	BaudRate: 115200,
})

mgr := simplemgr.New(serialPort)
response, _ := mgr.GetImageState(context.Background())
for i, img := range response.Images {
	fmt.Printf("%d: ver %s\n", i, img.Version)
}
// 0: ver 1.0.0
// 1: ver 1.0.0
```

## CLI
A CLI tool is also included to talk to serial-connected devices.
```sh
go install github.com/kendallgoto/simplemgr/cmd/simplemgr@latest

simplemgr version
# simplemgr v1.0.0
simplemgr --port /dev/ttyACM0 image list
# {
#   "images": [
#     {
#       "image": 0,
#       "slot": 0,
#       "version": "1.0.0",
#       "hash": null
#     },
#     {
#       "image": 1,
#       "slot": 0,
#       "version": "0.0.0",
#       "hash": null
#     }
#   ]
# }
simplemgr --port /dev/ttyACM0 image upload zephyr.bin
# [########################] 100% 266.1KiB/266.1KiB 13.7KiB/s eta 0s
./bin/simplemgr --port /dev/ttyACM0 image list
# {
#   "images": [
#     {
#       "image": 0,
#       "slot": 0,
#       "version": "1.2.3",
#       "hash": null
#     },
#     {
#       "image": 1,
#       "slot": 0,
#       "version": "0.0.0",
#       "hash": null
#     }
#   ]
# }
```
