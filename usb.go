package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/yne717/gousb/usb"
)

const MAX_SIZE = 64
const IR_FREQ = 38000
const IR_SEND_DATA_USB_SEND_MAX_LEN = 14

var (
	Device = flag.String("device", "22ea:0039", "select device. default \"22ea:0039\" ")
	Key    = flag.String("key", "none", "select key.")
	Number = flag.Int("number", 999999, "select number.")
	Debug  = flag.Int("debug", 3, "Debug level for libusb")
)

func main() {
	flag.Parse()

	ctx := usb.NewContext()
	defer func() {
		ctx.Close()
		fmt.Print("libusb_exit\n")
	}()

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
			fmt.Print("libusb_close\n")
		}

	}()

	if err != nil {
		log.Fatalf("usb.Open: %v", err)
	}

	if len(devs) == 0 {
		log.Fatal("not device.")
	}

	var data []byte
	if *Key != "none" {
		data = getDataByKey(*Key)
	} else {
		data = getDataByNumber(*Number)
	}

	// dev.DetachKernelDriver(0)

	// len, err := dev.Control(0x01, 0x0B, 0, 0, data)
	// if err != nil {
	// 	log.Fatalf("control device faild: %s", err)
	// }
	// fmt.Printf("wrote %vbyte\n", len)

	ep, err := devs[0].OpenEndpoint(uint8(1), uint8(0), uint8(0), uint8(1)|uint8(usb.ENDPOINT_DIR_OUT))
	if err != nil {
		log.Fatalf("open device faild: %s", err)
	}

	transfer(ep, data)

	// _ = transfer(devs[0], data)
}

func transfer(ep usb.Endpoint, ir_data []byte) {
	var (
		buf                                          []byte = make([]byte, MAX_SIZE, MAX_SIZE)
		send_bit_num, send_bit_pos, set_bit_size, fi int    = 0, 0, 0, 0
	)

	send_bit_num = len(ir_data) / 4

	for v := 0; ; {
		v++
		buf = make([]byte, MAX_SIZE, MAX_SIZE)

		for i := range buf {
			buf[i] = 0xFF
		}

		buf[0] = 0x34
		buf[1] = byte((send_bit_num >> 8) & 0xFF)
		buf[2] = byte(send_bit_num & 0xFF)
		buf[3] = byte((send_bit_pos >> 8) & 0xFF)
		buf[4] = byte(send_bit_pos & 0xFF)

		if send_bit_num > send_bit_pos {
			set_bit_size = send_bit_num - send_bit_pos
			if set_bit_size > IR_SEND_DATA_USB_SEND_MAX_LEN {
				set_bit_size = IR_SEND_DATA_USB_SEND_MAX_LEN
			}
		} else {
			set_bit_size = 0
		}

		buf[5] = byte(set_bit_size & 0xFF)

		if set_bit_size > 0 {
			fi = 0
			for fi = 0; fi < set_bit_size; fi++ {
				buf[6+(fi*4)] = ir_data[send_bit_pos*4]
				buf[6+(fi*4)+1] = ir_data[(send_bit_pos*4)+1]
				buf[6+(fi*4)+2] = ir_data[(send_bit_pos*4)+2]
				buf[6+(fi*4)+3] = ir_data[(send_bit_pos*4)+3]
				send_bit_pos++
			}

			len, err := ep.Write(buf)
			if err != nil {
				log.Fatalf("control faild: %v", err)
			}

			// len, err := ep.Control(0x01, 0x0B, 0, 0, buf)
			// if err != nil {
			// 	log.Fatalf("control faild: %v", err)
			// }

			fmt.Printf("wrote %vbyte\n", len)
		} else {
			break
		}
	}

	buf = make([]byte, MAX_SIZE, MAX_SIZE)

	for i := range buf {
		buf[i] = 0xFF
	}

	buf[0] = 0x35
	buf[1] = byte((IR_FREQ >> 8) & 0xFF)
	buf[2] = byte(IR_FREQ & 0xFF)
	buf[3] = byte((send_bit_num >> 8) & 0xFF)
	buf[4] = byte(send_bit_num & 0xFF)

	_, err := ep.Write(buf)
	if err != nil {
		log.Fatalf("control faild: %v", err)
	}

}

func getDataByKey(key string) []byte {
	list := getKeyList()
	return list[key]
}

func getDataByNumber(number int) []byte {
	return []byte{0xff, 0xff, 0xff, 0xff}
}

func getKeyList() map[string][]byte {
	return map[string][]byte{
		"restart":      {0x01, 0x56, 0x00, 0xAA, 0x00, 0x18, 0x00, 0x40, 0x00, 0x17, 0x00, 0x15, 0x00, 0x18, 0x00, 0x15, 0x00, 0x17, 0x00, 0x15, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x15, 0x00, 0x17, 0x00, 0x40, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x15, 0x00, 0x17, 0x00, 0x40, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x15, 0x00, 0x17, 0x00, 0x40, 0x00, 0x17, 0x00, 0x16, 0x00, 0x17, 0x00, 0x15, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x15, 0x00, 0x17, 0x00, 0x15, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x15, 0x00, 0x17, 0x00, 0x15, 0x00, 0x18, 0x00, 0x15, 0x00, 0x17, 0x00, 0x15, 0x00, 0x18, 0x00, 0x15, 0x00, 0x17, 0x00, 0x40, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x40, 0x00, 0x17, 0x00, 0x40, 0x00, 0x17, 0x00, 0x40, 0x00, 0x17, 0x1E, 0x0D},
		"fast_back":    {0x01, 0x58, 0x00, 0xA9, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x14, 0x00, 0x17, 0x00, 0x14, 0x00, 0x19, 0x00, 0x3E, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x19, 0x00, 0x3E, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x14, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x19, 0x00, 0x3E, 0x00, 0x19, 0x00, 0x3F, 0x00, 0x18, 0x1E, 0x0D},
		"tmp_stop":     {0x01, 0x57, 0x00, 0xAA, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x14, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x19, 0x00, 0x13, 0x00, 0x19, 0x00, 0x3E, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x19, 0x00, 0x3E, 0x00, 0x19, 0x1E, 0x0D},
		"fast_forward": {0x01, 0x58, 0x00, 0xA9, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x19, 0x00, 0x3E, 0x00, 0x19, 0x00, 0x14, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x14, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x19, 0x00, 0x3E, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x14, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x19, 0x00, 0x3E, 0x00, 0x19, 0x00, 0x3F, 0x00, 0x18, 0x1E, 0x0D},
		"key_original": {0x01, 0x57, 0x00, 0xAA, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x14, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x3E, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x3E, 0x00, 0x19, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x3E, 0x00, 0x19, 0x1E, 0x0D},
		"tempo_up":     {0x01, 0x57, 0x00, 0xAA, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x15, 0x00, 0x17, 0x00, 0x40, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x3E, 0x00, 0x19, 0x00, 0x3E, 0x00, 0x19, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x14, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x19, 0x00, 0x3E, 0x00, 0x19, 0x00, 0x14, 0x00, 0x19, 0x00, 0x3E, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x14, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x3E, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x1E, 0x0D},
		"tempo_down":   {0x01, 0x57, 0x00, 0xAA, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x14, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x19, 0x00, 0x3E, 0x00, 0x19, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x14, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x19, 0x00, 0x3E, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x3E, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x14, 0x00, 0x17, 0x00, 0x40, 0x00, 0x18, 0x00, 0x14, 0x00, 0x19, 0x00, 0x14, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x1E, 0x0D},
		"key_up":       {0x01, 0x56, 0x00, 0xAB, 0x00, 0x17, 0x00, 0x40, 0x00, 0x17, 0x00, 0x16, 0x00, 0x17, 0x00, 0x15, 0x00, 0x18, 0x00, 0x15, 0x00, 0x17, 0x00, 0x40, 0x00, 0x17, 0x00, 0x16, 0x00, 0x17, 0x00, 0x40, 0x00, 0x17, 0x00, 0x40, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x17, 0x00, 0x16, 0x00, 0x17, 0x00, 0x40, 0x00, 0x17, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x14, 0x00, 0x18, 0x00, 0x40, 0x00, 0x17, 0x00, 0x15, 0x00, 0x17, 0x00, 0x16, 0x00, 0x17, 0x00, 0x15, 0x00, 0x18, 0x00, 0x15, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x17, 0x00, 0x16, 0x00, 0x17, 0x00, 0x15, 0x00, 0x17, 0x00, 0x16, 0x00, 0x17, 0x00, 0x15, 0x00, 0x18, 0x00, 0x15, 0x00, 0x17, 0x00, 0x40, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x17, 0x00, 0x16, 0x00, 0x17, 0x00, 0x40, 0x00, 0x17, 0x00, 0x40, 0x00, 0x17, 0x00, 0x40, 0x00, 0x17, 0x00, 0x40, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x1E, 0x0D},
		"key_down":     {0x01, 0x56, 0x00, 0xAB, 0x00, 0x16, 0x00, 0x41, 0x00, 0x16, 0x00, 0x16, 0x00, 0x17, 0x00, 0x16, 0x00, 0x16, 0x00, 0x16, 0x00, 0x17, 0x00, 0x41, 0x00, 0x17, 0x00, 0x15, 0x00, 0x16, 0x00, 0x41, 0x00, 0x17, 0x00, 0x40, 0x00, 0x17, 0x00, 0x41, 0x00, 0x16, 0x00, 0x16, 0x00, 0x16, 0x00, 0x41, 0x00, 0x18, 0x00, 0x3F, 0x00, 0x18, 0x00, 0x15, 0x00, 0x16, 0x00, 0x41, 0x00, 0x16, 0x00, 0x17, 0x00, 0x16, 0x00, 0x16, 0x00, 0x17, 0x00, 0x40, 0x00, 0x17, 0x00, 0x16, 0x00, 0x16, 0x00, 0x41, 0x00, 0x17, 0x00, 0x16, 0x00, 0x16, 0x00, 0x16, 0x00, 0x17, 0x00, 0x16, 0x00, 0x16, 0x00, 0x16, 0x00, 0x17, 0x00, 0x16, 0x00, 0x16, 0x00, 0x17, 0x00, 0x16, 0x00, 0x41, 0x00, 0x16, 0x00, 0x18, 0x00, 0x16, 0x00, 0x40, 0x00, 0x16, 0x00, 0x41, 0x00, 0x16, 0x00, 0x41, 0x00, 0x15, 0x00, 0x42, 0x00, 0x16, 0x00, 0x41, 0x00, 0x16, 0x1E, 0x0D},
		"stop":         {0x01, 0x55, 0x00, 0xAB, 0x00, 0x18, 0x00, 0x41, 0x00, 0x14, 0x00, 0x17, 0x00, 0x15, 0x00, 0x18, 0x00, 0x16, 0x00, 0x16, 0x00, 0x17, 0x00, 0x40, 0x00, 0x17, 0x00, 0x16, 0x00, 0x17, 0x00, 0x40, 0x00, 0x16, 0x00, 0x41, 0x00, 0x16, 0x00, 0x41, 0x00, 0x17, 0x00, 0x16, 0x00, 0x15, 0x00, 0x42, 0x00, 0x16, 0x00, 0x41, 0x00, 0x15, 0x00, 0x18, 0x00, 0x15, 0x00, 0x42, 0x00, 0x15, 0x00, 0x18, 0x00, 0x15, 0x00, 0x17, 0x00, 0x15, 0x00, 0x18, 0x00, 0x16, 0x00, 0x16, 0x00, 0x15, 0x00, 0x18, 0x00, 0x15, 0x00, 0x17, 0x00, 0x17, 0x00, 0x16, 0x00, 0x15, 0x00, 0x18, 0x00, 0x15, 0x00, 0x17, 0x00, 0x15, 0x00, 0x18, 0x00, 0x15, 0x00, 0x42, 0x00, 0x15, 0x00, 0x42, 0x00, 0x16, 0x00, 0x41, 0x00, 0x15, 0x00, 0x42, 0x00, 0x15, 0x00, 0x42, 0x00, 0x16, 0x00, 0x42, 0x00, 0x15, 0x00, 0x42, 0x00, 0x15, 0x00, 0x42, 0x00, 0x15, 0x1E, 0x0D},
	}
}
