package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/yne717/gousb/usb"
)

var (
	Device = flag.String("device", "22ea:0039", "select device. default \"0403:6001\" ")
	Power  = flag.String("power", "on", "amp power. on or off")
	Music  = flag.Int("music", -20, "music volume. -62 ~ 0")
	Mic    = flag.Int("mic", -20, "mic volume. -62 ~ 0")
	Echo   = flag.Int("echo", 20, "echo volume. 0 ~ 63")
	Debug  = flag.Int("debug", 3, "Debug level for libusb")
)

func main() {
	flag.Parse()

	ctx := usb.NewContext()
	defer ctx.Close()

	ctx.Debug(*Debug)

	devs, err := ctx.ListDevices(func(desc *usb.Descriptor) bool {
		if fmt.Sprintf("%s:%s", desc.Vendor, desc.Product) != *Device {
			return false
		}

		return true
	})

	defer func() {
		for _, dev := range devs {
			dev.Close()
		}
	}()

	if err != nil {
		log.Fatalf("usb.Open: %v", err)
	}

	if len(devs) == 0 {
		log.Fatal("not device.")
	}

	dev := devs[0]

	powerData := getPowerData()
	musicMicData := getMusicMicData()
	echoData := getEchoData()

	data := []byte{
		getStx(),
		getTextTop(),
		powerData[*Power],
		musicMicData[*Music],
		musicMicData[*Mic],
		echoData[*Echo],
		getEtx(),
	}

	data = append(data, getXor(data))

	ep, err := dev.OpenEndpoint(uint8(1), uint8(0), uint8(0), uint8(2)|uint8(usb.ENDPOINT_DIR_OUT))
	if err != nil {
		log.Fatalf("open device faild: %s", err)
	}

	len, err := ep.Write(data)
	if err != nil {
		log.Fatalf("control faild: %v", err)
	}

	fmt.Printf("wrote %vbyte\n", len)
}
