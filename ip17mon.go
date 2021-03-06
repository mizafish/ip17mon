package ip17mon

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io/ioutil"
	"net"
	"strings"
)

const Null = "N/A"

var (
	ErrInvalidIp = errors.New("Invalid ip format")
	std          *Locator
)

func Init(dataFile string) (err error) {
	if std != nil {
		return
	}
	std, err = NewLocator(dataFile)
	return
}

func InitWithData(data []byte) {
	if std != nil {
		return
	}
	std = NewLocatorWithData(data)
	return
}

func Find(ipstr string) (*LocationInfo, error) {
	return std.Find(ipstr)
}

func FindByUint(ip uint32) *LocationInfo {
	return std.FindByUint(ip)
}

//-----------------------------------------------------------------------------

func NewLocator(dataFile string) (loc *Locator, err error) {
	data, err := ioutil.ReadFile(dataFile)
	if err != nil {
		return
	}
	loc = NewLocatorWithData(data)
	return
}

func NewLocatorWithData(data []byte) (loc *Locator) {
	loc = new(Locator)
	loc.init(data)
	return
}

type Locator struct {
	data       []byte
	indexData  []byte
	index      []uint32
	offset     uint32
	maxCompLen uint32
}

type LocationInfo struct {
	Country string
	Region  string
	City    string
	Isp     string
}

func (loc *Locator) Find(ipstr string) (info *LocationInfo, err error) {
	ip := net.ParseIP(ipstr)
	if ip == nil {
		err = ErrInvalidIp
		return
	}
	info = loc.FindByUint(binary.BigEndian.Uint32([]byte(ip.To4())))
	return
}

func (loc *Locator) FindByUint(ip uint32) (info *LocationInfo) {
	ioff := loc.findIndexOffset(ip, loc.index[ip>>24]<<3+1024)
	off := uint32(loc.indexData[ioff+4]) |
		uint32(loc.indexData[ioff+5])<<8 |
		uint32(loc.indexData[ioff+6])<<16
	off += loc.offset - 1024
	return newLocationInfo(loc.data[off : off+uint32(loc.indexData[ioff+7])])
}

// binary search
func (loc *Locator) findIndexOffset(ip, start uint32) uint32 {
	end := loc.maxCompLen
	for start < end {
		mid := (start/8 + end/8) / 2 * 8
		if ip > binary.BigEndian.Uint32(loc.indexData[mid:mid+4]) {
			start = mid + 8
		} else {
			end = mid
		}
	}

	if binary.BigEndian.Uint32(loc.indexData[end:end+4]) >= ip {
		return end
	}
	return start
}

func (loc *Locator) init(data []byte) {
	loc.data = data
	loc.offset = binary.BigEndian.Uint32(data[:4])
	loc.indexData = data[4:loc.offset]
	loc.maxCompLen = loc.offset - 1028
	loc.index = make([]uint32, 256)

	for i := 0; i < 256; i++ {
		loc.index[i] = binary.LittleEndian.Uint32(loc.indexData[i*4 : i*4+4])
	}
	return
}

func newLocationInfo(str []byte) *LocationInfo {
	fields := bytes.Split(str, []byte("\t"))
	if len(fields) != 5 {
		panic("unexpected ip info:" + string(str))
	}
	info := &LocationInfo{
		Country: string(fields[0]),
		Region:  string(fields[1]),
		City:    string(fields[2]),
		Isp:     string(fields[4]),
	}

	if len(info.Country) == 0 {
		info.Country = Null
	}
	if len(info.Region) == 0 {
		info.Region = Null
	}
	if len(info.City) == 0 {
		info.City = Null
	}
	if idx := strings.IndexAny(info.Isp, "/"); idx != -1 {
		info.Isp = info.Isp[:idx]
	}
	if len(info.Isp) == 0 {
		info.Isp = Null
	}
	return info
}
