//go:generate go run chinaip_gen.go

package main

import (
	"encoding/binary"
	"errors"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/cyfdecyf/bufio"
)

// data range by first byte
var CNIPDataRange [256]struct {
	start int
	end   int
}

func initCNIPData() {
	importCNIPFile()

	n := len(CNIPDataStart)
	var curr uint32
	var preFirstByte uint32
	for i := 0; i < n; i++ {
		firstByte := CNIPDataStart[i] >> 24
		if curr != firstByte {
			curr = firstByte
			if preFirstByte != 0 {
				CNIPDataRange[preFirstByte].end = i - 1
			}
			CNIPDataRange[firstByte].start = i
			preFirstByte = firstByte
		}
	}
	CNIPDataRange[preFirstByte].end = n - 1
}

func importCNIPFile() {
	if err := isFileExists(config.CNIPFile); err != nil {
		return
	}
	f, err := os.Open(config.CNIPFile)
	if err != nil {
		errl.Println("Error opening china ip list:", err)
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	CNIPDataStart = []uint32{}
	CNIPDataNum = []uint{}

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		parts := strings.Split(line, "/")
		if len(parts) != 2 {
			panic(errors.New("Invalid CIDR Format"))
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
		CNIPDataStart = append(CNIPDataStart, ipLong)
		CNIPDataNum = append(CNIPDataNum, count)
	}
	debug.Printf("Load china ip list")
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
