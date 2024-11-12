package main

import (
	"fmt"
	tls "github.com/refraction-networking/utls"
	"os"
)

func main() {
	d, _ := os.ReadFile("./examples/mu/xxx")
	var spec tls.ClientHelloSpec

	err := spec.FromRaw(PrependRecordHeader2(d), true)
	fmt.Println(err)
}

func PrependRecordHeader2(hello []byte) []byte {
	l := len(hello)

	// utls.VersionTLS10 这个无所谓的， 最后会被 实际 raw 中的 扩展值修正
	minTLSVersion := tls.VersionTLS10
	recordType := 22
	//22 == recordTypeHandshake
	header := []byte{
		uint8(recordType),                                             // type
		uint8(minTLSVersion >> 8 & 0xff), uint8(minTLSVersion & 0xff), // record version is the minimum supported
		uint8(l >> 8 & 0xff), uint8(l & 0xff), // length
	}
	return append(header, hello...)
}
