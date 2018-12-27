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
    // "archive/zip"
)

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
                      d.Delegate1.Phone,
                      strings.TrimSpace(d.Delegate1.Email),
                      strings.TrimSpace(d.Delegate1.Experience),
                      strings.TrimSpace(d.Delegate2.Name),
                      strings.TrimSpace(d.Delegate2.Institution),
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
    if err := csvWriter.WriteAll(data); err != nil {
        fmt.Println("ERROR: Error writing to CSV", err)
        return
    }
}

func generateDoubleDelegCsv(w io.Writer, info map[string]([][]string)) {
    csvWriter := csv.NewWriter(os.Stdout)
    for committee := range info {
        if err := csvWriter.WriteAll(info[committee]); err != nil {
            fmt.Println("ERROR: Error writing to CSV", err)
            return
        }
    }
}

func generateSingleDelegCsv(w io.Writer, info map[string]([][]string)) {
    csvWriter := csv.NewWriter(os.Stdout)
    for committee := range info {
        if err := csvWriter.WriteAll(info[committee]); err != nil {
            fmt.Println("ERROR: Error writing to CSV", err)
            return
        }
    }
}

func getAccom(w io.Writer) {
    url := "https://jumun2019-9c834.firebaseio.com/accom.json"
    response := genericFetch(url)
    delegates := martialDelegates(response)
    sort.Sort(ByName(delegates))
    generateCsv(w, delegates)
}

func getNonVeg(w io.Writer) {
    url := "https://jumun2019-9c834.firebaseio.com/nonveg.json"
    response := genericFetch(url)
    delegates := martialDelegates(response)
    sort.Sort(ByName(delegates))
    generateCsv(w, delegates)
}

func getVeg(w io.Writer) {
    url := "https://jumun2019-9c834.firebaseio.com/veg.json"
    response := genericFetch(url)
    delegates := martialDelegates(response)
    sort.Sort(ByName(delegates))
    generateCsv(w, delegates)
}

func getMerch(w io.Writer) {
    url := "https://jumun2019-9c834.firebaseio.com/merch.json"
    response := genericFetch(url)
    delegates := martialDelegates(response)
    sort.Sort(ByName(delegates))
    generateCsv(w, delegates)
}

func getSingleDeleg(w io.Writer, flat bool) {
    url := "https://jumun2019-9c834.firebaseio.com/single_deleg.json"
    response := genericFetch(url)
    info := martialSingleDelegation(response)
    if flat {
        totalLength := 0
        for committee := range info {
            totalLength += len(info[committee])
        }
        var data = make([][]string, 0, totalLength)
        for committee := range info {
            data = append(data, info[committee]...)
        }
        sort.Sort(ByName(data))
        generateCsv(w, data)
    } else {
        for committee := range info {
            sort.Sort(ByName(info[committee]))
        }
        generateSingleDelegCsv(w, info)
    }
}

func getDoubleDeleg(w io.Writer, flat bool) {
    url := "https://jumun2019-9c834.firebaseio.com/double_deleg.json"
    response := genericFetch(url)
    info := martialDoubleDelegation(response)
    if flat {
        totalLength := 0
        for committee := range info {
            totalLength += len(info[committee])
        }
        var data = make([][]string, 0, totalLength)
        for committee := range info {
            data = append(data, info[committee]...)
        }
        sort.Sort(ByName(data))
        generateCsv(w, data)
    } else {
        for committee := range info {
            sort.Sort(ByName(info[committee]))
        }
        generateDoubleDelegCsv(w, info)
    }
}

func getDoubleDelegHandler(w http.ResponseWriter, req *http.Request) {
    w.Header().Set("Content-Type", "application/csv")
    w.Header().Set("Content-Disposition", `inline; filename="double-delegation.csv"`)
    getDoubleDeleg(w, true)
}

func getDoubleDelegZipHandler(w http.ResponseWriter, req *http.Request) {
    w.Header().Set("Content-Type", "application/zip")
    w.Header().Set("Content-Disposition", `inline; filename="double-delegation.zip"`)
    getDoubleDeleg(w, false)
}

func getSingleDelegHandler(w http.ResponseWriter, req *http.Request) {
    w.Header().Set("Content-Type", "application/csv")
    w.Header().Set("Content-Disposition", `inline; filename="single-delegation.csv"`)
    getSingleDeleg(w, true)
}

func getSingleDelegZipHandler(w http.ResponseWriter, req *http.Request) {
    w.Header().Set("Content-Type", "application/zip")
    w.Header().Set("Content-Disposition", `inline; filename="single-delegation.zip"`)
    getSingleDeleg(w, false)
}

func handleRoot(w http.ResponseWriter, req *http.Request) {
    fmt.Fprint(w, "Hello World")
}

func main() {
    http.HandleFunc("/", handleRoot)
    http.HandleFunc("/single-deleg", getSingleDelegHandler)
    http.HandleFunc("/single-deleg-zip", getSingleDelegZipHandler)
    http.HandleFunc("/double-deleg", getDoubleDelegHandler)
    http.HandleFunc("/double-deleg-zip", getDoubleDelegZipHandler)
    fmt.Println("Starting server at port 1234...")
    if err := http.ListenAndServe(":1234", nil); err != nil {
        fmt.Println("ERROR: Serving failed", err)
        return
    }
}
