package simplemgr_test

import (
	"context"
	"fmt"

	"github.com/kendallgoto/simplemgr"
	"go.bug.st/serial"
)

func Example() {
	serialPort, _ := serial.Open("/dev/ttyACM0", &serial.Mode{
		BaudRate: 115200,
	})

	mgr := simplemgr.New(serialPort)
	response, _ := mgr.GetImageState(context.Background())
	for i, img := range response.Images {
		fmt.Printf("%d: ver %s\n", i, img.Version)
	}
}
