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

const apiURL string = "https://api.rach.io/1"
var accessToken string
var debug bool = false
var http_runs int = 0

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

type scheduleItem []struct { // /public/device/:id/scheduleitem
	Date int64 `json:"date"`
	StartHour int `json:"startHour"`
	StartMinute int `json:"startMinute"`
	Zones []struct {
		ZoneID string `json:"zoneId"`
		ZoneNumber int `json:"zoneNumber"`
		Duration int `json:"duration"`
		SortOrder int `json:"sortOrder"`
	} `json:"zones"`
	ScheduleRuleID string `json:"scheduleRuleId"`
	TotalDuration int `json:"totalDuration"`
	ScheduleType string `json:"scheduleType"`
	AbsoluteStartDate int64 `json:"absoluteStartDate"`
	Iso8601Date time.Time `json:"iso8601Date"`
}

type scheduleRules struct { // /public/device/schedulerules/:id
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
	SeasonalAdjustment float64 `json:"seasonalAdjustment"`
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
	debugf("doHTTP() request: %s, %s\n", method, u)
	
	req, err := http.NewRequest(method, u.String(), nil);
	if err != nil {
		debugf("http.NewRequest error: %s\n", err)
		os.Exit(1)
	}
	req.Header.Add("Authorization", "Bearer " + accessToken)

	debugf("doHTTP() response: %+v\n", req)

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
/*
	http_runs++
	str := "o/out" + strconv.Itoa(http_runs)
	ioutil.WriteFile(str, r, 0666)
*/

	err =  json.Unmarshal(r, s)
	if err != nil {
		fmt.Printf("unmarshal: %+v\n", err)
		jerr := err.(*json.UnmarshalTypeError)
		fmt.Printf("Unexpected value: %s\n", (*jerr).Value)
		fmt.Printf("Unexpected type: %v\n", (*jerr).Type)
		fmt.Printf("Offset: %d\n", (*jerr).Offset)
		ioutil.WriteFile("HTTP_ERROR", r, 0666)
		os.Exit(1)
	}
	return err
}

func doGet(rawurl string, v url.Values, s interface{}) (err error) {
	return doHTTP("GET", rawurl, v, s)
}

func Person () {
	var v url.Values 

	debugf("**** Person()\n")
	err := doGet(apiURL+"/public/person/info", v, &p)
	if err != nil {
		debugf("doGet: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("\nid = %s\n", p.Id)
}

func PersonInfo() {
	var v url.Values
	
	debugf("**** PersonInfo()\n")	
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

	debugf("**** getDevice()\n")
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

	debugf("**** getScheduleRules\n")
	err := doGet(apiURL+"/public/schedulerule/"+id, v, &d)
	if err != nil {
		debugf("doGet: %s\n", err)
		return d, err
	}
	return d, err
}

func getScheduleItem(id string) (scheduleItem, error) {
	var d scheduleItem
	var v url.Values

	debugf("**** getScheduleItem\n")
	err := doGet(apiURL+"/public/device/"+id+"/scheduleitem", v, &d)
	if err != nil {
		debugf("doGet: %s\n", err)
		return d, err
	}
	return d, err
}

func jobTypes(jobTypes []string) (string) {
	var r string
	for _, s := range jobTypes {
		switch s {
		case "DAY_OF_WEEK_0": r += "Sunday"
		case "DAY_OF_WEEK_1": r += "Monday"
		case "DAY_OF_WEEK_2": r += "Tuesday"
		case "DAY_OF_WEEK_3": r += "Wednesday"
		case "DAY_OF_WEEK_4": r += "Thursday"
		case "DAY_OF_WEEK_5": r += "Friday"
		case "DAY_OF_WEEK_6": r += "Saturday"
		case "ODD": r += "Odd Days"
		case "EVEN": r += "Even"
		case "INTERVAL_1": r += "Every 1 day"
		case "INTERVAL_2": r += "Every 2 days"
		case "INTERVAL_3": r += "Every 3 days"
		case "INTERVAL_4": r += "Every 4 days"
		case "INTERVAL_5": r += "Every 5 days"
		case "INTERVAL_6": r += "Every 6 days"
		case "INTERVAL_7": r += "Every 7 days"
		case "INTERVAL_8": r += "Every 8 days"
		case "INTERVAL_9": r += "Every 9 days"
		case "INTERVAL_10": r += "Every 10 days"
		case "INTERVAL_11": r += "Every 11 days"
		case "INTERVAL_12": r += "Every 12 days"
		case "INTERVAL_13": r += "Every 13 days"
		case "INTERVAL_14": r += "Every 14 days"
		case "INTERVAL_15": r += "Every 15 days"
		case "INTERVAL_16": r += "Every 16 days"
		case "INTERVAL_17": r += "Every 17 days"
		case "INTERVAL_18": r += "Every 18 days"
		case "INTERVAL_19": r += "Every 19 days"
		case "INTERVAL_20": r += "Every 20 days"
		case "INTERVAL_21": r += "Every 21 days"
		case "ANY": r += "Any day of the week"
		default: r += "UNKNOWN (" + s + ")"
		}
		r += ", "
	}
	return r
}

func showSched() {
	var err error

	seen := make(map[string]bool)

	// for each device, find 
	for  i,d := range pinfo.Devices {
		fmt.Printf("Device %d:\n", i)
		fmt.Printf("Device ID: %s\nName: %s\nStatus: %s\nEnabled: %t\n", d.ID, d.Name,  d.Status, d.Enabled)
		si, _ := getScheduleItem(d.ID)
		if err != nil {
			fmt.Printf("getSchedule: %s\n", err)
			os.Exit(1)
		}
		for _, isi := range si {
			if !seen[isi.ScheduleRuleID] {
				var dur int64
				seen[isi.ScheduleRuleID] = true
				sr, _ := getScheduleRules(isi.ScheduleRuleID)
				fmt.Printf("\t\tSchedule name: %s, ScheduleRuleID: %s, enabled: %t\n", sr.Name, sr.ID, sr.Enabled)
				fmt.Printf("\t\t\tStart time: %02d:%02d\n", isi.StartHour, isi.StartMinute)
				fmt.Printf("\t\t\tISODate: %s\n", isi.Iso8601Date.Local())
				t := time.Unix(isi.AbsoluteStartDate / int64(time.Microsecond), 0)
				fmt.Printf("\t\t\tAbsolute Start Date: %s\n", t.Local().Format(time.UnixDate))
				fmt.Printf("\t\t\tScheduleJobTypes: %s\n", jobTypes(sr.ScheduleJobTypes));
				fmt.Printf("\t\t\tTotal duration %s, Schedule Type %s\n",
					time.Duration(isi.TotalDuration) * time.Second,	isi.ScheduleType)
				for _, z := range isi.Zones {
					fmt.Printf("\t\t\t\tZone %d: duration %s, sort order %d\n", z.ZoneNumber,
						time.Duration(z.Duration) * time.Second, z.SortOrder)
					dur += int64(z.Duration)
				}
				fmt.Printf("\t\t\tComputed duration: %s\n", time.Duration(dur) * time.Second)
			}
		}
		fmt.Println("---------------------------------------------------------")
	}
}

func main() {

	var conf *globalconf.GlobalConf
	var err error
	

	flag.StringVar(&accessToken, "accessToken", "", "OAuth2 token")
	flag.BoolVar(&debug, "D", debug, "Debugging enabled")
	flag.Parse()

	// read confg
        if conf, err = globalconf.New("rachio"); err != nil {
                fmt.Printf("Error: %s\n", err)
                os.Exit(1)
        }
        conf.ParseAll()

	debugf("accessToken = %s\n", accessToken)
	if accessToken == "" {
		fmt.Println("Error: no accessToken provided")
		os.Exit(1)
	}
	
	Person()
	PersonInfo()
	showSched()

/*
	var s int64
	s = 1463382000000 / int64(time.Microsecond) // 5am
	t := time.Unix(s,0)
	fmt.Println(t.Format(time.UnixDate))
	d := time.Duration(s) * time.Microsecond
	fmt.Println(d)

	tz, err := time.LoadLocation("US/Mountain")
	if err != nil {
		fmt.Printf("LoadLocation(): %s\n", err)
		os.Exit(1)
	}
	fmt.Println(tz)
	t = t.In(tz)
	fmt.Println(t.Format(time.UnixDate))	
*/
}
