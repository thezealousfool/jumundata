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
)

var nrounds = 2

type Delegate struct {
    Name string `json:"name"`
    Phone string `json:"phone"`
    Email string `json:"email"`
}

type DelegateInfo map[string]Delegate

func (d Delegate) StringArray() []string {
    return []string {strings.TrimSpace(d.Name), d.Phone, strings.TrimSpace(d.Email)}
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

func (d SingleDelegationInfo) StringArrayMap() map[string]([][]string) {
    result := make(map[string]([][]string))
    for pref1Committee := range d {
        array := make([][]string, 0, len(d[pref1Committee]))
        for delegate_id := range d[pref1Committee] {
            array = append(array, d[pref1Committee][delegate_id].StringArray())
        }
        result[pref1Committee] = array
    }
    return result
}

func (d SingleDelegation) StringArray() []string {
    return []string { strings.TrimSpace(d.Name),
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
                      d.Preference2.Country2 }
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

func (d DoubleDelegationInfo) StringArrayMap() map[string]([][]string) {
    result := make(map[string]([][]string))
    for pref1Committee := range d {
        array := make([][]string, 0, len(d[pref1Committee]))
        for delegate_id := range d[pref1Committee] {
            array = append(array, d[pref1Committee][delegate_id].StringArray())
        }
        result[pref1Committee] = array
    }
    return result
}

func (d DoubleDelegation) StringArray() []string {
    return []string { strings.TrimSpace(d.Delegate1.Name),
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
                      d.Preference2.Country2 }
}


type ByName [][]string

func (a ByName) Len() int { return len(a) }
func (a ByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i][0] < a[j][0] }

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

func martialDelegates(body []byte) [][]string {
    var delegateInfo DelegateInfo
    err := json.Unmarshal(body, &delegateInfo)
    if err != nil {
        fmt.Println("ERROR: Parsing JSON failed")
        fmt.Println(err)
        return nil
    }
    delegates := make([][]string, 0, len(delegateInfo))
    for k := range delegateInfo {
        delegates = append(delegates, delegateInfo[k].StringArray())
    }
    return delegates
}

func martialSingleDelegation(body []byte) map[string]([][]string) {
    var singleDelegateInfo SingleDelegationInfo
    err := json.Unmarshal(body, &singleDelegateInfo)
    if err != nil {
        fmt.Println("ERROR: Parsing JSON failed")
        fmt.Println(err)
        return nil
    }
    return singleDelegateInfo.StringArrayMap()
}

func martialDoubleDelegation(body []byte) map[string]([][]string) {
    var doubleDelegateInfo DoubleDelegationInfo
    err := json.Unmarshal(body, &doubleDelegateInfo)
    if err != nil {
        fmt.Println("ERROR: Parsing JSON failed")
        fmt.Println(err)
        return nil
    }
    return doubleDelegateInfo.StringArrayMap()
}

func generateCsv(w io.Writer, data [][]string) {
    csvWriter := csv.NewWriter(w)
    if data == nil {
        fmt.Fprint(w, "Error. Please contact Vivek")
        return
    }
    if err := csvWriter.WriteAll(data); err != nil {
        fmt.Println("ERROR: Error writing to CSV", err)
        fmt.Fprint(w, "Error. Please contact Vivek")
        return
    }
}

func flatMap(info map[string]([][]string)) [][]string {
    totalLength := 0
    for committee := range info {
        totalLength += len(info[committee])
    }
    var data = make([][]string, 0, totalLength)
    for committee := range info {
        data = append(data, info[committee]...)
    }
    return data
}

func genericDelegateRound(w io.Writer, json string, round string, parse func([]byte) [][]string) {
    fmt.Println("LOG: " + json + " Round "+round)
    url := "https://jumun2019-9c834.firebaseio.com/"
    if round == "1" {
        url += json
    } else if round == "all" {
        var finalResponse [][]string
        for r := 1; r <= nrounds; r++ {
            u := url
            if r == 1 {
                u += json
            } else {
                u += "round" + strconv.Itoa(r) + "/" + json
            }
            response := genericFetch(url)
            delegates := parse(response)
            sort.Sort(ByName(delegates))
            finalResponse = append(finalResponse, delegates...)
        }
        generateCsv(w, finalResponse)
        return
    } else {
        url += "round" + round + "/" + json
    }
    response := genericFetch(url)
    delegates := parse(response)
    sort.Sort(ByName(delegates))
    generateCsv(w, delegates)
}

func parseDelegates(response []byte) [][]string {
    return martialDelegates(response)
}

func parseSingleDelegates(response []byte) [][]string {
    info := martialSingleDelegation(response)
    return flatMap(info)
}

func parseDoubleDelegates(response []byte) [][]string {
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
    w.Header().Set("Content-Type", "application/csv")
    w.Header().Set("Content-Disposition", `inline; filename="` + filename + `"`)
    round, err := req.URL.Query()["round"]
    if !err {
        fmt.Println("ERROR: Error reading URL param round", filename, " '", round, "'")
        fmt.Fprint(w, "Error. Please contact Vivek")
        return
    }
    fn(w, round[0])
}

func getDoubleDelegHandler(w http.ResponseWriter, req *http.Request) {
    parseRoundAndCall(w,req,"double-delegation.csv",getDoubleDeleg)
}

func getSingleDelegHandler(w http.ResponseWriter, req *http.Request) {
    parseRoundAndCall(w,req,"single-delegation.csv",getSingleDeleg)
}

func getAccomHandler(w http.ResponseWriter, req *http.Request) {
    parseRoundAndCall(w,req,"accommodation.csv",getAccom)
}

func getMerchHandler(w http.ResponseWriter, req *http.Request) {
    parseRoundAndCall(w,req,"merchandise.csv",getMerch)
}

func getVegHandler(w http.ResponseWriter, req *http.Request) {
    parseRoundAndCall(w,req,"veg.csv",getVeg)
}

func getNonVegHandler(w http.ResponseWriter, req *http.Request) {
    parseRoundAndCall(w,req,"nonveg.csv",getNonVeg)
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
