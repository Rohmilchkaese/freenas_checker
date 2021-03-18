// (C) Cevin Freitag 2021
// X-Tention
// Freenas V2 API Checker
package main

import (
    "flag"
    "fmt"
    "io/ioutil"
    "net/http"
    "encoding/json"
    "strings"
)

//Define Variables (globally)
var poolObject Pool

//Define Strucs (globally)
type Pool[]struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	GUID       string `json:"guid"`
	Encrypt    int    `json:"encrypt"`
	Encryptkey string `json:"encryptkey"`
	Path       string `json:"path"`
	Status     string `json:"status"`
	Topology struct {
		Data []struct {
			Type   string `json:"type"`
			Path   string `json:"path"`
			GUID   string `json:"guid"`
			Status string `json:"status"`
			Stats  struct {
				ReadErrors       int           `json:"read_errors"`
				WriteErrors      int           `json:"write_errors"`
				ChecksumErrors   int           `json:"checksum_errors"`
				Size             int64         `json:"size"`
				Allocated        int64         `json:"allocated"`
				Fragmentation    int           `json:"fragmentation"`
				SelfHealed       int           `json:"self_healed"`
			} `json:"stats"`
			Device      string        `json:"device"`
			Disk        string        `json:"disk"`
		} `json:"data"`
	Healthy      bool        `json:"healthy"`
	StatusDetail interface{} `json:"status_detail"`
	IsDecrypted    bool        `json:"is_decrypted"`
    } `json:"topology"`
}


//END definition Strucs


func convert_disk_size(size int64) string{
    var out string
    KB := ((float64(size))/1024)
    MB := (KB/1024)
    GB := (MB/1024)
    TB := (GB/1024)


    if size <= 1024{
        out = fmt.Sprint(" Byte ", size)
    } else if (KB <= 1024) && MB <= 1 {
        out = fmt.Sprint(" KB ", KB)
    } else if (MB <= 1024) && GB <= 1 {
        out = fmt.Sprint(" MB ", MB)
    } else if (GB <= 1024) && TB <= 1 {
        out = fmt.Sprint(" GB ", GB)
    } else {
        out = fmt.Sprint(" TB ", TB)
    } 
    return out
}

func pooldata(poolObject Pool, warning int, critical int){
    var i int 
    var sum int64
    var alctd int64
    var free int64
    for i = 0; i < len(poolObject); i++ {
    fmt.Printf("Pool Name: %v\nMount Point: %v\nStatus: %v\n", poolObject[i].Name, poolObject[i].Path, poolObject[i].Status)
    fmt.Printf("Consists of %v Disks:\n", len(poolObject[i].Topology.Data))
        for i2 := 0; i2 < len(poolObject[i].Topology.Data); i2++ {
         fmt.Printf("%v Path:%v Status:%v Size:%v Allocated:%v\n", poolObject[i].Topology.Data[i2].Disk, poolObject[i].Topology.Data[i2].Path, poolObject[i].Topology.Data[i2].Status, convert_disk_size(poolObject[i].Topology.Data[i2].Stats.Size), convert_disk_size(poolObject[i].Topology.Data[i2].Stats.Allocated))
        }
        for i2 := 0; i2 < len(poolObject[i].Topology.Data); i2++ {
            sum += poolObject[i].Topology.Data[i2].Stats.Size
            alctd += poolObject[i].Topology.Data[i2].Stats.Allocated
            free = sum - alctd
        }
        fmt.Printf("Sum:%v Allocated:%v Free:%v\n", convert_disk_size(sum), convert_disk_size(alctd), convert_disk_size(free))
        if alctd >= ((sum*int64(warning))/100) {
            fmt.Printf("Warning: More Than %v %% used\n", warning)
        } else if alctd >= ((sum*int64(critical))/100) {
            fmt.Printf("Crital: More Than %v %% used\n", critical)
        } else if alctd <= ((sum*int64(warning))/100) {
            free = ((alctd*100)/sum)
            fmt.Printf("Disk space used %% %v\n", free)
        }    
    }
}



func apicall(query string, user string, password string, command string, warning int, critical int) {
//Making real API request
//    fmt.Println("Calling API...")
    client := &http.Client{}
     req, err := http.NewRequest("GET", query, nil)
     if err != nil {
      fmt.Print(err.Error())
     }
     req.Header.Add("Accept", "application/json")
     req.Header.Add("Content-Type", "application/json")
     req.SetBasicAuth(user, password)
     resp, err := client.Do(req)
     if err != nil {
      fmt.Print(err.Error())
     }
    defer resp.Body.Close()
     bodyBytes, err := ioutil.ReadAll(resp.Body)
     if err != nil {
     fmt.Print(err.Error())
     }

    //Print body Bytes out
    //fmt.Println(string(bodyBytes))

    if strings.Compare(command, "disk") == 0 { 
        json.Unmarshal(bodyBytes, &poolObject)
        pooldata(poolObject, warning, critical)
    } else if strings.Compare(command, "alerts") == 0 {
        fmt.Println((string(bodyBytes)))
    }
}

func gencall(command string, host string, user string, password string, warning int, critical int) {

    // Declare Variables 
    const api = "/api/v2.0/" 
    const pre = "http://"
    var apicmd string

    if strings.Compare(command, "disk") == 0 {
        apicmd = "pool"
    } else if strings.Compare(command, "alerts") == 0 {
        apicmd = "alert/list"
    } 

    query := fmt.Sprint(pre, host, api, apicmd)
    apicall(query, user, password, command, warning, critical)
}

func main() {

// Declare Flags 
    host := flag.String("h", "", "Host (IP Adress)")
    command := flag.String("c", "", "Command to perform")
    user := flag.String("u", "", "Username")
    password := flag.String("p", "", "Password")
    warning := flag.Int("warning", 80, "Warning regards free disk space %")
    critical := flag.Int("critical", 90, "Crital regards free disk space %")

    flag.Parse()

    //Print info entered over Flags for Debugging
    //fmt.Println("IP Adress:", *host)
    //fmt.Println("Command to perform:", *command)
    //fmt.Println("Username:", *user)
    //fmt.Println("Password:", *password)

    gencall(*command, *host, *user, *password, *warning, *critical)
}
