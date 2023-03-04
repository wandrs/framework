// +build linux

package xid

import "io/ioutil"

func readPlatformMachineID() (string, error) {
	b, err := os.ReadFile("/sys/class/dmi/id/product_uuid")
	return string(b), err
}
