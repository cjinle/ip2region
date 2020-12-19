package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	indexBlockLength  = 12
	totalHeaderLength = 8192
)

// IP2Region  struct ...
type IP2Region struct {
	dbFileHandler *os.File
	headerSip     []int64
	headerPtr     []int64
	headerLen     int64

	firstIndexPtr int64
	lastIndexPtr  int64
	totalBlocks   int64

	dbBinStr []byte
	dbFile   string
}

// IPInfo struct ...
type IPInfo struct {
	CityID   int64  `json:"cityid"`
	Country  string `json:"country"`
	Region   string `json:"region"`
	Province string `json:"province"`
	City     string `json:"city"`
	ISP      string `json:"isp"`
}

// New object
func New(path string) (*IP2Region, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return &IP2Region{
		dbFile:        path,
		dbFileHandler: file,
	}, nil
}

// Close function
func (i2r *IP2Region) Close() {
	i2r.dbFileHandler.Close()
}

// MemorySearch find ip to region
func (i2r *IP2Region) MemorySearch(ipStr string) (*IPInfo, error) {
	ipInfo := &IPInfo{}

	if i2r.totalBlocks == 0 {
		i2r.dbBinStr, err = ioutil.ReadFile(i2r.dbFile)
		if err != nil {

			return ipInfo, err
		}

		i2r.firstIndexPtr = getLong(i2r.dbBinStr, 0)
		i2r.lastIndexPtr = getLong(i2r.dbBinStr, 4)
		i2r.totalBlocks = (i2r.lastIndexPtr-i2r.firstIndexPtr)/indexBlockLength + 1
	}

	ip, err := ip2long(ipStr)
	if err != nil {
		return ipInfo, err
	}

	h := i2r.totalBlocks
	var dataPtr, l int64
	for l <= h {

		m := (l + h) >> 1
		p := i2r.firstIndexPtr + m*indexBlockLength
		sip := getLong(i2r.dbBinStr, p)
		if ip < sip {
			h = m - 1
		} else {
			eip := getLong(i2r.dbBinStr, p+4)
			if ip > eip {
				l = m + 1
			} else {
				dataPtr = getLong(i2r.dbBinStr, p+8)
				break
			}
		}
	}
	if dataPtr == 0 {
		return ipInfo, errors.New("not found")
	}

	dataLen := ((dataPtr >> 24) & 0xFF)
	dataPtr = (dataPtr & 0x00FFFFFF)
	ipInfo = getIPInfo(getLong(i2r.dbBinStr, dataPtr), i2r.dbBinStr[(dataPtr)+4:dataPtr+dataLen])
	return ipInfo, nil
}

func getLong(b []byte, offset int64) int64 {
	val := (int64(b[offset]) |
		int64(b[offset+1])<<8 |
		int64(b[offset+2])<<16 |
		int64(b[offset+3])<<24)

	return val
}

func ip2long(ipStr string) (int64, error) {
	bits := strings.Split(ipStr, ".")
	if len(bits) != 4 {
		return 0, errors.New("ip format error")
	}

	var sum int64
	for i, n := range bits {
		bit, _ := strconv.ParseInt(n, 10, 64)
		sum += bit << uint(24-8*i)
	}

	return sum, nil
}

func getIPInfo(cityID int64, line []byte) *IPInfo {
	lineSlice := strings.Split(string(line), "|")
	ipInfo := &IPInfo{}
	length := len(lineSlice)
	ipInfo.CityID = cityID
	if length < 5 {
		for i := 0; i <= 5-length; i++ {
			lineSlice = append(lineSlice, "")
		}
	}

	ipInfo.Country = lineSlice[0]
	ipInfo.Region = lineSlice[1]
	ipInfo.Province = lineSlice[2]
	ipInfo.City = lineSlice[3]
	ipInfo.ISP = lineSlice[4]
	return ipInfo
}

// IPHander find ip region
func IPHander(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ret := []byte("[]")
	ip := r.URL.Query().Get("ip")
	if ip != "" {
		ipInfo, err := region.MemorySearch(ip)
		if err != nil {
			log.Println(ip, err)
		} else {
			ret, _ = json.Marshal(&ipInfo)
		}
	}
	w.Write(ret)
}

var err error
var region *IP2Region

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	db := "ip2region.db"
	_, err = os.Stat(db)
	if os.IsNotExist(err) {
		panic("not found db " + db)
	}

	region, err = New(db)
	if err != nil {
		panic(err)
	}
	defer region.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", IPHander)

	srv := &http.Server{
		Handler:      mux,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
