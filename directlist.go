package main

import (
	"net"
	"os"
	"strings"
	"sync"

	"github.com/cyfdecyf/bufio"
)

type DomainList struct {
	Domain map[string]DomainType
	sync.RWMutex
}

type DomainType byte

const (
	domainTypeUnknown DomainType = iota
	domainTypeDirect
	domainTypeProxy
	domainTypeReject
)

func newDomainList() *DomainList {
	return &DomainList{
		Domain: map[string]DomainType{},
	}
}

func (domainList *DomainList) judge(url *URL) (domainType DomainType) {
	debug.Printf("judging host: %s", url.Host)
	domainList.Lock()
	hostDomainType := domainList.Domain[url.Host]
	domainDomainType := domainList.Domain[url.Domain]
	domainList.Unlock()
	if hostDomainType == domainTypeReject || domainDomainType == domainTypeReject {
		debug.Printf("host or domain should reject")
		return domainTypeReject
	}
	if parentProxy.empty() { // no way to retry, so always visit directly
		return domainTypeDirect
	}
	if url.Domain == "" { // simple host or private ip
		return domainTypeDirect
	}
	if hostDomainType == domainTypeDirect || domainDomainType == domainTypeDirect {
		debug.Printf("host or domain should direct")
		return domainTypeDirect
	}
	if hostDomainType == domainTypeProxy || domainDomainType == domainTypeProxy {
		debug.Printf("host or domain should using proxy")
		return domainTypeProxy
	}

	if !config.JudgeByIP {
		return domainTypeProxy
	}
	debug.Printf("judging by ip")
	var ip string
	isIP, isPrivate := hostIsIP(url.Host)
	if isIP {
		if isPrivate {
			domainList.add(url.Host, domainTypeDirect)
			return domainTypeDirect
		}
		ip = url.Host
	} else {
		hostIPs, err := net.LookupIP(url.Host)
		if err != nil {
			errl.Printf("error looking up host ip %s, err %s", url.Host, err)
			return domainTypeProxy
		}
		ip = hostIPs[0].String()
	}

	if ipShouldDirect(ip) {
		domainList.add(url.Host, domainTypeDirect)
		debug.Printf("host or domain should direct")
		return domainTypeDirect
	} else {
		domainList.add(url.Host, domainTypeProxy)
		debug.Printf("host or domain should using proxy")
		return domainTypeProxy
	}
}

func (domainList *DomainList) add(host string, domainType DomainType) {
	domainList.Lock()
	defer domainList.Unlock()
	domainList.Domain[host] = domainType
}

func (domainList *DomainList) GetDomainList() []string {
	lst := make([]string, 0)
	for site, domainType := range domainList.Domain {
		if domainType == domainTypeDirect {
			lst = append(lst, site)
		}
	}
	return lst
}

var domainList = newDomainList()

func initDomainList(domainListFile string, domainType DomainType) {
	var err error
	if err = isFileExists(domainListFile); err != nil {
		return
	}
	f, err := os.Open(domainListFile)
	if err != nil {
		errl.Println("Error opening domain list:", err)
		return
	}
	defer f.Close()

	domainList.Lock()
	defer domainList.Unlock()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		domain := strings.TrimSpace(scanner.Text())
		if domain == "" {
			continue
		}
		debug.Printf("Loaded domain %s as type %v", domain, domainType)
		domainList.Domain[domain] = domainType
	}
	if scanner.Err() != nil {
		errl.Printf("Error reading domain list %s: %v\n", domainListFile, scanner.Err())
	}
}
