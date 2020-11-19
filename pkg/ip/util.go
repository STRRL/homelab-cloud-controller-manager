package ip

import (
	"encoding/binary"
	"errors"
	"net"
)

func Ip2long(ipAddr string) (uint32, error) {
	ip := net.ParseIP(ipAddr)
	if ip == nil {
		return 0, errors.New("wrong ipAddr format")
	}
	ip = ip.To4()
	return binary.BigEndian.Uint32(ip), nil
}

func Long2ip(ipLong uint32) string {
	ipByte := make([]byte, 4)
	binary.BigEndian.PutUint32(ipByte, ipLong)
	ip := net.IP(ipByte)
	return ip.String()
}

func SplitWithMask(ip string, mask uint) (netBits uint32, hostBits uint32, err error) {
	cidrMask := net.CIDRMask(int(mask), 32)
	parsedIp := net.ParseIP(ip)
	subnet := parsedIp.Mask(cidrMask)
	subnetLong, err := Ip2long(subnet.String())
	if err != nil {
		return 0, 0, err
	}
	ipLong, err := Ip2long(ip)
	if err != nil {
		return 0, 0, err
	}
	return subnetLong, subnetLong ^ ipLong, nil
}
