package utils

import (
	"fmt"
	"testing"
)

func TestToChecksumAddress(t *testing.T) {
	usdcAddress := "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"

	//convert usdc address to checksum
	checksum, err := ToChecksumAddress(usdcAddress)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("USDC checksum is", checksum)

	//check if address is checksum
	isChecksum, err := IsChecksumAddress(checksum)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(checksum, "is checksum", isChecksum)
}
