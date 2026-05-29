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
