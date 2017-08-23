// +build generate
// go run chinaip_gen.go

package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// use china ip list database by ipip.net
const (
	chinaIPListFile = "https://github.com/17mon/china_ip_list/raw/master/china_ip_list.txt"
)

func main() {
	resp, err := http.Get(chinaIPListFile)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != 200 {
		panic(fmt.Errorf("Unexpected status %d", resp.StatusCode))
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)

	startList := []string{}
	countList := []string{}

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		parts := strings.Split(line, "/")
		if len(parts) != 2 {
			panic(errors.New("Invalid CIDR"))
		}
		ip := parts[0]
		mask := parts[1]
		count, err := cidrCalc(mask)
		if err != nil {
			panic(err)
		}

		ipLong, err := ipToUint32(ip)
		if err != nil {
			panic(err)
		}
		startList = append(startList, strconv.FormatUint(uint64(ipLong), 10))
		countList = append(countList, strconv.FormatUint(uint64(count), 10))
	}

	file, err := os.OpenFile("chinaip_data.go", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("Failed to generate chinaip_data.go: %v", err)
	}
	defer file.Close()

	fmt.Fprintln(file, "package main")
	fmt.Fprint(file, "var CNIPDataStart = []uint32 {\n	")
	fmt.Fprint(file, strings.Join(startList, ",\n	"))
	fmt.Fprintln(file, ",\n	}")

	fmt.Fprint(file, "var CNIPDataNum = []uint{\n	")
	fmt.Fprint(file, strings.Join(countList, ",\n	"))
	fmt.Fprintln(file, ",\n	}")
}

func cidrCalc(mask string) (uint, error) {
	i, err := strconv.Atoi(mask)
	if err != nil || i > 32 {
		return 0, errors.New("Invalid Mask")
	}
	p := 32 - i
	res := uint(intPow2(p))
	return res, nil
}

func intPow2(p int) int {
	r := 1
	for i := 0; i < p; i++ {
		r *= 2
	}
	return r
}

func ipToUint32(ipstr string) (uint32, error) {
	ip := net.ParseIP(ipstr)
	if ip == nil {
		return 0, errors.New("Invalid IP")
	}
	ip = ip.To4()
	if ip == nil {
		return 0, errors.New("Not IPv4")
	}
	return binary.BigEndian.Uint32(ip), nil
}
