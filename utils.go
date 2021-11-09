package web3

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/crypto/sha3"
)

//Convert hex address to checksum address
func ToChecksumAddress(address string) string {
	//check that the address is a valid Ethereum address
	re1 := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	if !re1.MatchString(address) {
		fmt.Printf("given address '%s' is not a valid Ethereum Address\n", address)
		return ""
	}
	//convert the address to lowercase
	re2 := regexp.MustCompile("^0x")
	address = re2.ReplaceAllString(address, "")
	address = strings.ToLower(address)

	//convert address to sha3 hash
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte(address))
	sum := hasher.Sum(nil)
	addressHash := fmt.Sprintf("%x", sum)
	addressHash = re2.ReplaceAllString(addressHash, "")

	//compile checksum address
	checksumAddress := "0x"

	for i := 0; i < len(address); i++ {
		indexedValue, err := strconv.ParseInt(string(rune(addressHash[i])), 16, 32)
		if err != nil {
			fmt.Println("Error when parsing addressHash during checksum conversion", err)
			return ""
		}
		if indexedValue > 7 {
			checksumAddress += strings.ToUpper(string(address[i]))
		} else {
			checksumAddress += string(address[i])
		}
	}
	return checksumAddress
}

//Check if hex address is a checksum address
func IsChecksumAddress(address string) (bool, error) {
	checksumAddress := ToChecksumAddress(address)
	if checksumAddress == address {
		return true, nil
	} else {
		return false, nil
	}
}
