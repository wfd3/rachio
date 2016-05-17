package main

import (
	"fmt"
	"errors"
	"time"
	"os"
	"io/ioutil"
	"encoding/json"
	"net/http"
	"net/url"
	 "flag"
        "github.com/rakyll/globalconf"
)

const accessToken string = "Bearer 9dbe74f4-f982-4c44-8923-29614f3335fa"
const apiURL string = "https://api.rach.io/1"
var debug bool = true


type personID struct {
	Id string `json:"id"`
}
var p personID
func (p personID) String() (string) {
	return p.Id
}

type personInfo struct {
	ID string `json:"id"`
	Username string `json:"username"`
	FullName string `json:"fullName"`
	Email string `json:"email"`
	Devices []struct {
		ID string `json:"id"`
		Status string `json:"status"`
		Zones []struct {
			ID string `json:"id"`
			ZoneNumber float64 `json:"zoneNumber"`
			Name string `json:"name"`
			Enabled bool `json:"enabled"`
			CustomNozzle struct {
				Name string `json:"name"`
				ImageURL string `json:"imageUrl"`
				Category string `json:"category"`
				InchesPerHour float64 `json:"inchesPerHour"`
			} `json:"customNozzle"`
			AvailableWater float64 `json:"availableWater"`
			RootZoneDepth float64 `json:"rootZoneDepth"`
			ManagementAllowedDepletion float64 `json:"managementAllowedDepletion"`
			Efficiency float64 `json:"efficiency"`
			YardAreaSquareFeet float64 `json:"yardAreaSquareFeet"`
			IrrigationAmount float64 `json:"irrigationAmount"`
			DepthOfWater float64 `json:"depthOfWater"`
			Runtime float64 `json:"runtime"`
		} `json:"zones"`
		TimeZone string `json:"timeZone"`
		Latitude float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Zip string `json:"zip"`
		Name string `json:"name"`
		ScheduleRules []struct {
			ID string `json:"id"`
			Name string `json:"name"`
			ExternalName string `json:"externalName"`
			SerialNumber string `json:"serialNumber"`
			RainDelayExpirationDate float64 `json:"rainDelayExpirationDate"`
			RainDelayStartDate float64 `json:"rainDelayStartDate"`
			MacAddress string `json:"macAddress"`
			Elevation float64 `json:"elevation"`
			Webhooks []interface{} `json:"webhooks"`
			Paused bool `json:"paused"`
			On bool `json:"on"`
			FlexScheduleRules []interface{} `json:"flexScheduleRules"`
			UtcOffset int `json:"utcOffset"`
		} `json:"scheduleRules"`
		Enabled bool `json:"enabled"`
	} `json:"devices"`
}
var pinfo personInfo

type deviceInfo struct {
	ID string `json:"id"`
	Status string `json:"status"`
	Zones []struct {
		ID string `json:"id"`
		ZoneNumber int `json:"zoneNumber"`
		Name string `json:"name"`
		Enabled bool `json:"enabled"`
		CustomNozzle struct {
			Name string `json:"name"`
			ImageURL string `json:"imageUrl"`
			Category string `json:"category"`
			InchesPerHour float64 `json:"inchesPerHour"`
		} `json:"customNozzle"`
		AvailableWater float64 `json:"availableWater"`
		RootZoneDepth float64 `json:"rootZoneDepth"`
		ManagementAllowedDepletion float64 `json:"managementAllowedDepletion"`
		Efficiency float64 `json:"efficiency"`
		YardAreaSquareFeet float64 `json:"yardAreaSquareFeet"`
		IrrigationAmount float64 `json:"irrigationAmount"`
		DepthOfWater float64 `json:"depthOfWater"`
		Runtime int `json:"runtime"`
	} `json:"zones"`
	TimeZone string `json:"timeZone"`
	Latitude float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Zip string `json:"zip"`
	Name string `json:"name"`
	ScheduleRules []struct {
		ID string `json:"id"`
		Name string `json:"name"`
		ExternalName string `json:"externalName"`
	} `json:"scheduleRules"`
	SerialNumber string `json:"serialNumber"`
	RainDelayExpirationDate int64 `json:"rainDelayExpirationDate"`
	RainDelayStartDate int64 `json:"rainDelayStartDate"`
	MacAddress string `json:"macAddress"`
	Elevation float64 `json:"elevation"`
	Webhooks []interface{} `json:"webhooks"`
	Paused bool `json:"paused"`
	On bool `json:"on"`
	FlexScheduleRules []interface{} `json:"flexScheduleRules"`
	UtcOffset int `json:"utcOffset"`
}
var dinfo deviceInfo

type deviceSchedule []struct {
	StartDate int64 `json:"startDate"`
	Zones []struct {
		ZoneID string `json:"zoneId"`
		ZoneNumber int `json:"zoneNumber"`
		Duration int `json:"duration"`
		SortOrder int `json:"sortOrder"`
	} `json:"zones"`
	ScheduleRuleID string `json:"scheduleRuleId"`
	CycleSoak bool `json:"cycleSoak"`
	TotalDuration int `json:"totalDuration"`
}
var dsched deviceSchedule

type scheduleRules struct {
	ID string `json:"id"`
	Zones []struct {
		ZoneID string `json:"zoneId"`
		ZoneNumber int `json:"zoneNumber"`
		Duration int `json:"duration"`
		SortOrder int `json:"sortOrder"`
	} `json:"zones"`
	ScheduleJobTypes []string `json:"scheduleJobTypes"`
	Summary string `json:"summary"`
	RainDelay bool `json:"rainDelay"`
	WaterBudget bool `json:"waterBudget"`
	CycleSoakStatus string `json:"cycleSoakStatus"`
	StartDate int64 `json:"startDate"`
	Name string `json:"name"`
	Enabled bool `json:"enabled"`
	TotalDuration int `json:"totalDuration"`
	WeatherIntelligenceSensitivity float64 `json:"weatherIntelligenceSensitivity"`
	SeasonalAdjustment int `json:"seasonalAdjustment"`
	TotalDurationNoCycle int `json:"totalDurationNoCycle"`
	Cycles int `json:"cycles"`
	ExternalName string `json:"externalName"`
	CycleSoak bool `json:"cycleSoak"`
}
var schedrules scheduleRules

func debugf(format string, a ...interface{}) (n int, err error) {
	if debug {
		format = "# " + format
		return fmt.Fprintf(os.Stderr, format, a...)
	}
	return 0, nil
}

func doHTTP(method string, rawurl string, v url.Values, s interface{}) (err error) {
	var r []byte
	var res *http.Response
	var c http.Client
	
	c.Timeout = 60 * time.Second

	u, _ := url.Parse(rawurl)
	u.RawQuery = v.Encode()
	debugf("doHTTP: %s, %s\n", method, u)
	
	req, err := http.NewRequest(method, u.String(), nil);
	if err != nil {
		debugf("http.NewRequest error: %s\n", err)
		os.Exit(1)
	}
	req.Header.Add("Authorization", accessToken)

	fmt.Printf("%+v\n", req)

	t := time.Now()
	if res, err = c.Do(req); err != nil {
		debugf("Get() failed: %s\n", err)
		return err
	}
	d := time.Since(t)
	debugf("   HTTP Response: %s, in %s\n", res.Status, d)
	if res.StatusCode != http.StatusOK {
		return errors.New(res.Status)
	}
	r, err = ioutil.ReadAll(res.Body)
	res.Body.Close()

	if err != nil { 
		debugf("ReadAll() failed: %s\n", err)
		return err
	}

//	fmt.Printf("%+v\n", r)

	err =  json.Unmarshal(r, s)
	if err != nil {
		fmt.Printf("unmarshal: %+v\n", err)
		jerr := err.(*json.UnmarshalTypeError)
		fmt.Printf("Unexpected value: %s\n", (*jerr).Value)
		fmt.Printf("Unexpected type: %v\n", (*jerr).Type)
		fmt.Printf("Offset: %d\n", (*jerr).Offset)
		os.Exit(1)
	}
	return err
}

func doGet(rawurl string, v url.Values, s interface{}) (err error) {
	return doHTTP("GET", rawurl, v, s)
}

func Person () {
	var v url.Values 

	err := doGet(apiURL+"/public/person/info", v, &p)
	if err != nil {
		debugf("doGet: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("\nid = %s\n", p.Id)
}

func PersonInfo() {
	var v url.Values
	
	err := doGet(apiURL+"/public/person/"+p.String(), v, &pinfo)
	if err != nil {
		debugf("doGet: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Name: %s\nUsername: %s\nEmail: %s\nID: %s\n", pinfo.FullName, pinfo.Username, pinfo.Email, pinfo.ID)
}

func getDevice(id string) (deviceInfo, error) {
	var d deviceInfo
	var v url.Values

	err := doGet(apiURL+"/public/device/"+id, v, &d)
	if err != nil {
		debugf("doGet: %s\n", err)
		return d, err
	}
	return d, err
}


func getScheduleRules(id string) (scheduleRules, error) {
	var d scheduleRules
	var v url.Values

	err := doGet(apiURL+"/public/schedulerule/"+id, v, &d)
	if err != nil {
		debugf("doGet: %s\n", err)
		return d, err
	}
	return d, err
}

func Devices() {
	var dinfo deviceInfo
	var err error
	var sr scheduleRules
	
	for  i,d := range pinfo.Devices {
		fmt.Printf("Device %d:\n", i)
		fmt.Printf("\tDevice ID: %s\n\tName: %s\n\tStatus: %s\n\tEnabled: %t\n", d.ID, d.Name,  d.Status, d.Enabled)

		dinfo, err = getDevice(d.ID)
		if err != nil {
			debugf("\t\tDevices: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("\t\tID: %s\n", dinfo.ID)

		for j,s := range dinfo.ScheduleRules {
			fmt.Printf("ScheduleRule %d: %s id %s\n", j, s.Name, s.ID)
			sr, err = getScheduleRules(s.ID)
			fmt.Printf("\tSchedule '%s': start %d, total duration %d, enabled %t\n", sr.Name, sr.StartDate, sr.TotalDuration, sr.Enabled)
			for _, r := range sr.Zones {
				fmt.Printf("\t\tZone %d: %d\n", r.ZoneNumber, r.Duration)
			}
		}
	}
}

func Devices_OLD() {
	var dinfo deviceInfo
	var dsched deviceSchedule
	var err error
	
	for  i,d := range pinfo.Devices {
		fmt.Printf("Device %d:\n", i)
		fmt.Printf("\tDevice ID: %s\n\tName: %s\n\tStatus: %s\n\tEnabled: %t\n", d.ID, d.Name,  d.Status, d.Enabled)

		dinfo, err = getDevice(d.ID)
		if err != nil {
			debugf("\t\tDevices: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("\t\tID: %s\n", dinfo.ID)
		for j,s := range dinfo.ScheduleRules {
			fmt.Printf("ScheduleRule %d: %s id %s\n", j, s.Name, s.ID)
		}
		for _,z := range dinfo.Zones {
			fmt.Printf("\t\tZone %d, Name %s, ID %s\n", z.ZoneNumber, z.Name, z.ID)
		}


		if err != nil {
			debugf("getSchedule: %s\n", err)
			os.Exit(1)
		}

		for _, s := range dsched {
			fmt.Printf("\t\tStart: %d, duration %d\n", s.StartDate, s.TotalDuration/60)
			for _, z := range s.Zones {
				fmt.Printf("\t\t\tZone %d, %d minutes\n", z.ZoneNumber, z.Duration/60)
			}
			
		}
	}
}

func getSchedule(id string) (deviceSchedule, error) {
	var d deviceSchedule
	var v url.Values

	err := doGet(apiURL+"/public/device/"+id+"/scheduleitem", v, &d)
	if err != nil {
		debugf("doGet: %s\n", err)
		return d, err
	}
	return d, err
}

func main() {
	var authtoken string
	var config *globalconf.GlobalConf
	var err error

	flag.StringVar(&authtoken, "authtoken", "", "OAuth2 token")
	flag.Parse()
	// read confg
        if config, err = globalconf.New("rachio"); err != nil {
                fmt.Printf("Error: %s\n", err)
                os.Exit(1)
        }
        config.ParseAll()
	
	Person()
	PersonInfo()
	Devices()
}
