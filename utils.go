package web3

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/crypto/sha3"
)

func ToChecksumAddress(address string) (string, error) {
	//check that the address is a valid Ethereum address
	re1 := regexp.MustCompile("^(0x)||(0X)?[0-9a-f]{40}$")
	if !re1.MatchString(address) {
		return "", fmt.Errorf("given address '%s' is not a valid Ethereum Address", address)
	}
	//convert the address to lowercase
	re2 := regexp.MustCompile("/^0x/i")
	address = re2.ReplaceAllString(address, "")
	address = strings.ToLower(address)

	//convert address to sha3 hash
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte(address))
	sum := hasher.Sum(nil)
	addressHash := fmt.Sprintf("%x", sum)
	addressHash = re2.ReplaceAllString(addressHash, "")

	//compile checksum address
	checksumAddress := ""

	for i := 0; i < len(address); i++ {
		indexedValue, err := (strconv.ParseInt(string(rune(addressHash[i])), 16, 32))
		if err != nil {
			fmt.Println("Error when parsing addressHash during checksum conversion", err)
			return "", err
		}
		if indexedValue > 7 {
			checksumAddress += strings.ToUpper(string(address[i]))
		} else {
			checksumAddress += string(address[i])
		}
	}
	return checksumAddress, nil
}

func IsChecksumAddress(address string) (bool, error) {
	checksumAddress, err := ToChecksumAddress(address)
	fmt.Println(checksumAddress)
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	if checksumAddress == address {
		return true, nil
	} else {
		return false, nil
	}
}
