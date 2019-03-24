package main

import (
    "net/http"
    "fmt"
    "io"
    "io/ioutil"
    "encoding/json"
    "encoding/csv"
    "os"
    "strings"
    "sort"
    "strconv"
    "time"
)

var nrounds = 4
var push_cars = "-0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz"

func firebaseTimestamp(id string) int64 {
    id = id[0:8]
    var timestamp int64
    timestamp = 0
    for i := 0; i < len(id); i++ {
        timestamp = timestamp * 64 + int64(strings.Index(push_cars, string(id[i])))
    }
    return timestamp/1000
}

type sstring struct {
    t time.Time
    s []string
}

type Delegate struct {
    Name string `json:"name"`
    Phone string `json:"phone"`
    Email string `json:"email"`
}

type DelegateInfo map[string]Delegate

func (d Delegate) StringArray(key string) sstring {
    return sstring {
        time.Unix(firebaseTimestamp(key),0),
        []string {
            strings.TrimSpace(d.Name),
            d.Phone,
            strings.TrimSpace(d.Email) } }
}

type Preference struct {
    Committee string `json:"committee"`
    Country1 string `json:"country1"`
    Country2 string `json:"country2"`
}

type SingleDelegation struct {
    Name string `json:"name"`
    Institution string `json:"institution"`
    Phone string `json:"phone"`
    Email string `json:"email"`
    Experience string `json:"experience"`
    Ambassador string `json:"ambassador"`
    Referrer string `json:"referrer"`
    Preference1 Preference `json:"preference1"`
    Preference2 Preference `json:"preference2"`
}

type SingleDelegationInfo map[string](map[string]SingleDelegation)

func (d SingleDelegationInfo) StringArrayMap() map[string]([]sstring) {
    result := make(map[string]([]sstring))
    for pref1Committee := range d {
        array := make([]sstring, 0, len(d[pref1Committee]))
        for delegate_id := range d[pref1Committee] {
            array = append(array, d[pref1Committee][delegate_id].StringArray(delegate_id))
        }
        result[pref1Committee] = array
    }
    return result
}

func (d SingleDelegation) StringArray(key string) sstring {
    return sstring { time.Unix(firebaseTimestamp(key),0),
           []string { strings.TrimSpace(d.Name),
                      strings.TrimSpace(d.Institution),
                      strings.TrimSpace(d.Ambassador),
                      strings.TrimSpace(d.Referrer),
                      d.Phone,
                      strings.TrimSpace(d.Email),
                      strings.TrimSpace(d.Experience),
                      d.Preference1.Committee,
                      d.Preference1.Country1,
                      d.Preference1.Country2,
                      d.Preference2.Committee,
                      d.Preference2.Country1,
                      d.Preference2.Country2 } }
}

type DDDelegate struct {
    Name string `json:"name"`
    Institution string `json:"institution"`
    Phone string `json:"phone"`
    Email string `json:"email"`
    Experience string `json:"experience"`
    Ambassador string `json:"ambassador"`
    Referrer string `json:"referrer"`
}

type DoubleDelegation struct {
    Delegate1 DDDelegate `json:"delegate1"`
    Delegate2 DDDelegate `json:"delegate2"`
    Preference1 Preference `json:"preference1"`
    Preference2 Preference `json:"preference2"`
}

type DoubleDelegationInfo map[string](map[string]DoubleDelegation)

func (d DoubleDelegationInfo) StringArrayMap() map[string]([]sstring) {
    result := make(map[string]([]sstring))
    for pref1Committee := range d {
        array := make([]sstring, 0, len(d[pref1Committee]))
        for delegate_id := range d[pref1Committee] {
            array = append(array, d[pref1Committee][delegate_id].StringArray(delegate_id))
        }
        result[pref1Committee] = array
    }
    return result
}

func (d DoubleDelegation) StringArray(key string) sstring {
    return sstring { time.Unix(firebaseTimestamp(key), 0),
           []string { strings.TrimSpace(d.Delegate1.Name),
                      strings.TrimSpace(d.Delegate1.Institution),
                      strings.TrimSpace(d.Delegate1.Ambassador),
                      strings.TrimSpace(d.Delegate1.Referrer),
                      d.Delegate1.Phone,
                      strings.TrimSpace(d.Delegate1.Email),
                      strings.TrimSpace(d.Delegate1.Experience),
                      strings.TrimSpace(d.Delegate2.Name),
                      strings.TrimSpace(d.Delegate2.Institution),
                      strings.TrimSpace(d.Delegate2.Ambassador),
                      strings.TrimSpace(d.Delegate2.Referrer),
                      d.Delegate2.Phone,
                      strings.TrimSpace(d.Delegate2.Email),
                      strings.TrimSpace(d.Delegate2.Experience),
                      d.Preference1.Committee,
                      d.Preference1.Country1,
                      d.Preference1.Country2,
                      d.Preference2.Committee,
                      d.Preference2.Country1,
                      d.Preference2.Country2 } }
}


type ByName []sstring
type ByTime []sstring

func (a ByName) Len() int { return len(a) }
func (a ByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].s[0] < a[j].s[0] }

func (a ByTime) Len() int { return len(a) }
func (a ByTime) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByTime) Less(i, j int) bool { return a[i].t.Before(a[j].t) }

func genericFetch(url string) []byte {
    resp, err := http.Get(url)
    if err != nil {
        fmt.Println("ERROR: Fetch failed - ", url)
        fmt.Println(err)
        return nil
    }
    defer resp.Body.Close()
    if resp.Status != "200 OK" {
        fmt.Println("ERROR: Fetch failed - ", url)
        fmt.Println(err)
        return nil
    }
    bodyBytes, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println("ERROR: Failed reading response body", url)
        fmt.Println(err)
        return nil
    }
    return bodyBytes
}

func martialDelegates(body []byte) []sstring {
    var delegateInfo DelegateInfo
    err := json.Unmarshal(body, &delegateInfo)
    if err != nil {
        fmt.Println("ERROR: Parsing JSON failed")
        fmt.Println(err)
        return nil
    }
    delegates := make([]sstring, 0, len(delegateInfo))
    for k := range delegateInfo {
        delegates = append(delegates, delegateInfo[k].StringArray(k))
    }
    return delegates
}

func martialSingleDelegation(body []byte) map[string]([]sstring) {
    var singleDelegateInfo SingleDelegationInfo
    err := json.Unmarshal(body, &singleDelegateInfo)
    if err != nil {
        fmt.Println("ERROR: Parsing JSON failed")
        fmt.Println(err)
        return nil
    }
    return singleDelegateInfo.StringArrayMap()
}

func martialDoubleDelegation(body []byte) map[string]([]sstring) {
    var doubleDelegateInfo DoubleDelegationInfo
    err := json.Unmarshal(body, &doubleDelegateInfo)
    if err != nil {
        fmt.Println("ERROR: Parsing JSON failed")
        fmt.Println(err)
        return nil
    }
    return doubleDelegateInfo.StringArrayMap()
}

func generateCsv(w io.Writer, data []sstring) {
    csvWriter := csv.NewWriter(w)
    if data == nil {
        fmt.Fprint(w, "Error. Please contact Vivek")
        return
    }
    for rec := range data {
        err := csvWriter.Write(data[rec].s)
        if err != nil {
            fmt.Println("ERROR: Error writing to CSV", err)
            fmt.Fprint(w, "Error. Please contact Vivek")
            return
        }
    }
    csvWriter.Flush()
}

func flatMap(info map[string]([]sstring)) []sstring {
    totalLength := 0
    for committee := range info {
        totalLength += len(info[committee])
    }
    var data = make([]sstring, 0, totalLength)
    for committee := range info {
        data = append(data, info[committee]...)
    }
    return data
}

func genericDelegateRound(w io.Writer, json string, round string, parse func([]byte) []sstring) {
    fmt.Println("LOG: " + json + " Round "+round)
    url := "https://jumun2019-9c834.firebaseio.com/"
    if round == "1" {
        url += json
    } else if round == "all" {
        var finalResponse []sstring
        for r := 1; r <= nrounds; r++ {
            u := url
            if r == 1 {
                u += json
            } else {
                u += "round" + strconv.Itoa(r) + "/" + json
            }
            response := genericFetch(url)
            delegates := parse(response)
            sort.Sort(ByTime(delegates))
            finalResponse = append(finalResponse, delegates...)
        }
        generateCsv(w, finalResponse)
        return
    } else {
        url += "round" + round + "/" + json
    }
    response := genericFetch(url)
    delegates := parse(response)
    sort.Sort(ByTime(delegates))
    generateCsv(w, delegates)
}

func parseDelegates(response []byte) []sstring {
    return martialDelegates(response)
}

func parseSingleDelegates(response []byte) []sstring {
    info := martialSingleDelegation(response)
    return flatMap(info)
}

func parseDoubleDelegates(response []byte) []sstring {
    info := martialDoubleDelegation(response)
    return flatMap(info)
}

func getAccom(w io.Writer, round string) {
    genericDelegateRound(w, "accom.json", round, parseDelegates)
}

func getNonVeg(w io.Writer, round string) {
    genericDelegateRound(w, "nonveg.json", round, parseDelegates)
}

func getVeg(w io.Writer, round string) {
    genericDelegateRound(w, "veg.json", round, parseDelegates)
}

func getMerch(w io.Writer, round string) {
    genericDelegateRound(w, "merch.json", round, parseDelegates)
}

func getSingleDeleg(w io.Writer, round string) {
    genericDelegateRound(w, "single_deleg.json", round, parseSingleDelegates)
}

func getDoubleDeleg(w io.Writer, round string) {
    genericDelegateRound(w, "double_deleg.json", round, parseDoubleDelegates)
}

func parseRoundAndCall(w http.ResponseWriter, req *http.Request, filename string, fn func(io.Writer,string)) {
    round, err := req.URL.Query()["round"]
    if !err {
        fmt.Println("ERROR: Error reading URL param round", filename, " '", round, "'")
        fmt.Fprint(w, "Error. Please contact Vivek")
        return
    }
    w.Header().Set("Content-Type", "application/csv")
    w.Header().Set("Content-Disposition", `inline; filename="` + filename + "_round_" + round[0] + ".csv" + `"`)
    fn(w, round[0])
}

func getDoubleDelegHandler(w http.ResponseWriter, req *http.Request) {
    parseRoundAndCall(w,req,"double-delegation",getDoubleDeleg)
}

func getSingleDelegHandler(w http.ResponseWriter, req *http.Request) {
    parseRoundAndCall(w,req,"single-delegation",getSingleDeleg)
}

func getAccomHandler(w http.ResponseWriter, req *http.Request) {
    parseRoundAndCall(w,req,"accommodation",getAccom)
}

func getMerchHandler(w http.ResponseWriter, req *http.Request) {
    parseRoundAndCall(w,req,"merchandise",getMerch)
}

func getVegHandler(w http.ResponseWriter, req *http.Request) {
    parseRoundAndCall(w,req,"veg",getVeg)
}

func getNonVegHandler(w http.ResponseWriter, req *http.Request) {
    parseRoundAndCall(w,req,"nonveg",getNonVeg)
}

func handleRoot(w http.ResponseWriter, req *http.Request) {
    url := "https://jumun2019-9c834.firebaseio.com/"
    var dump map[string]bool
    for round := 1; round <= nrounds; round++ {
        u := url
        if round == 1 {
            u += "data_dump.json?shallow=true"
        } else {
            u += "round" + strconv.Itoa(round) + "/data_dump.json?shallow=true"
        }
        response := genericFetch(u)
        err := json.Unmarshal(response, &dump)
        if err != nil {
            fmt.Println("ERROR: Parsing JSON failed")
            fmt.Println(err)
            fmt.Fprint(w, "Error. Please contact Vivek")
            return
        }
    }
    fmt.Fprint(w, `
        <!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="UTF-8">
            <meta content="width=device-width,initial-scale=1,shrink-to-fit=no" name="viewport" />
            <title>JUMUN Data</title>
        </head>
        <style>
            * {
                font-family: BlinkMacSystemFont,-apple-system,"Segoe UI","Fira Sans",Roboto,Ubuntu,Oxygen-Sans,Cantarell,"Helvetica Neue",Arial,sans-serif;
                font-weight: 400;
            }
            body {
                padding: 1rem;
                max-width: 800px;
                margin: auto;
            }
            a {
                color: inherit;
            }
            li {
                padding: 0.5em;
            }
        </style>
        <body>
            <h1>JUMUN 2019 Delegate Information</h1>
            <h2 style="font-weight: bold;">Total applications: `, len(dump), `</h2>
            <hr>
            <h3>Round 1</h3>
            <ul>
                <li><a href="/single-deleg?round=1">Single Delegations</a></li>
                <li><a href="/double-deleg?round=1">Double Delegations</a></li>
                <li><a href="/merch?round=1">Merchandise Requests</a></li>
                <li><a href="/accom?round=1">Accommodation Requests</a></li>
                <li><a href="/veg?round=1">Veg Food Requests</a></li>
                <li><a href="/nonveg?round=1">Non-Veg Food Requests</a></li>
            </ul>
            <h3>Round 2</h3>
            <ul>
                <li><a href="/single-deleg?round=2">Single Delegations</a></li>
                <li><a href="/double-deleg?round=2">Double Delegations</a></li>
                <li><a href="/merch?round=2">Merchandise Requests</a></li>
                <li><a href="/accom?round=2">Accommodation Requests</a></li>
                <li><a href="/veg?round=2">Veg Food Requests</a></li>
                <li><a href="/nonveg?round=2">Non-Veg Food Requests</a></li>
            </ul>
            <h3>Round 3</h3>
            <ul>
                <li><a href="/single-deleg?round=3">Single Delegations</a></li>
                <li><a href="/double-deleg?round=3">Double Delegations</a></li>
                <li><a href="/merch?round=3">Merchandise Requests</a></li>
                <li><a href="/accom?round=3">Accommodation Requests</a></li>
                <li><a href="/veg?round=3">Veg Food Requests</a></li>
                <li><a href="/nonveg?round=3">Non-Veg Food Requests</a></li>
            </ul>
            <h3>Rolling Round</h3>
            <ul>
                <li><a href="/single-deleg?round=4">Single Delegations</a></li>
                <li><a href="/double-deleg?round=4">Double Delegations</a></li>
                <li><a href="/merch?round=4">Merchandise Requests</a></li>
                <li><a href="/accom?round=4">Accommodation Requests</a></li>
                <li><a href="/veg?round=4">Veg Food Requests</a></li>
                <li><a href="/nonveg?round=4">Non-Veg Food Requests</a></li>
            </ul>
            <h3>All Delegations</h3>
            <ul>
                <li><a href="/single-deleg?round=all">Single Delegations</a></li>
                <li><a href="/double-deleg?round=all">Double Delegations</a></li>
                <li><a href="/merch?round=all">Merchandise Requests</a></li>
                <li><a href="/accom?round=all">Accommodation Requests</a></li>
                <li><a href="/veg?round=all">Veg Food Requests</a></li>
                <li><a href="/nonveg?round=all">Non-Veg Food Requests</a></li>
            </ul>
        </body>
        </html>
    `)
}

func main() {
    port := os.Getenv("PORT")
    http.HandleFunc("/", handleRoot)
    http.HandleFunc("/single-deleg", getSingleDelegHandler)
    http.HandleFunc("/double-deleg", getDoubleDelegHandler)
    http.HandleFunc("/merch", getMerchHandler)
    http.HandleFunc("/accom", getAccomHandler)
    http.HandleFunc("/veg", getVegHandler)
    http.HandleFunc("/nonveg", getNonVegHandler)
    fmt.Println("Starting server at port " + port + "...")
    if err := http.ListenAndServe(":"+port, nil); err != nil {
        fmt.Println("ERROR: Serving failed", err)
        return
    }
}
