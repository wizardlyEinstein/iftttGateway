package main

import (
	"github.com/jsgoecke/tesla"
  "time"
  "fmt"
  "strings"
  "os"
  "github.com/gorilla/handlers"
  "encoding/json"
  "net/http"
  "log"
  "github.com/gorilla/mux"
  //"github.com/hashicorp/mdns"
)

// Secret to prove you are worthy.
var Secret = os.Getenv("SECRET") 

type request struct{
    Secret string `json:"secret"`
    State string `json:"state"`
}

func secretOk(request request) bool {
    if request.Secret == Secret {
        return true
    } 
    return false
}

func root(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello World!\n"))
}

func party(w http.ResponseWriter, r *http.Request) {
    var request request
    _ = json.NewDecoder(r.Body).Decode(&request)
    if !secretOk(request) {
        w.WriteHeader(http.StatusForbidden)
        w.Write([]byte("nope\n"))
        return
    }
    if strings.ToLower(request.State) == "on" {
        go func() { 
            _, _ = http.Get("http://192.168.1.247/rotate")
        }()
    } 
    if strings.ToLower(request.State) == "off" {
        go func() { 
            _, _ = http.Get("http://192.168.1.247/setcolor?red=255&green=255&blue=255")
        }()
    }
    w.Write([]byte("ok\n"))
}

func white(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("ok\n"))
    go func() {
        _, _ = http.Get("http://192.168.1.247/setcolor?red=255&green=255&blue=255")
    }()
}

func hotwater(w http.ResponseWriter, r *http.Request) {
    var request request
    _ = json.NewDecoder(r.Body).Decode(&request)
    if !secretOk(request) {
        w.WriteHeader(http.StatusForbidden)
        w.Write([]byte("nope\n"))
        return
    }
    if strings.ToLower(request.State) == "on" {
        go func() {
            _, _ = http.Get("http://192.168.1.247/setcolor?red=255&green=0&blue=0")
            _, _ = http.Get("http://192.168.1.150:8080/HW/on?for=900")
            time.Sleep(1 * time.Second)
            _, _ = http.Get("http://192.168.1.247/setcolor?red=255&green=255&blue=255")
        }()
    }
    if strings.ToLower(request.State) == "off" {
        go func() {
            _, _ = http.Get("http://192.168.1.247/setcolor?red=0&green=0&blue=255")
            _, _ = http.Get("http://192.168.1.150:8080/HW/off")
            time.Sleep(1 * time.Second)
            _, _ = http.Get("http://192.168.1.247/setcolor?red=255&green=255&blue=255")
        }()
    }
    w.Write([]byte("ok\n"))
}

func conditionTesla(w http.ResponseWriter, r *http.Request) {
  var request request
  _ = json.NewDecoder(r.Body).Decode(&request)
  if !secretOk(request) {
      w.WriteHeader(http.StatusForbidden)
      w.Write([]byte("nope\n"))
      return
  }
  go func() {
    client, err := tesla.NewClient(
      &tesla.Auth{
        ClientID:     os.Getenv("TESLA_CLIENT_ID"),
        ClientSecret: os.Getenv("TESLA_CLIENT_SECRET"),
        Email:        os.Getenv("TESLA_USERNAME"),
        Password:     os.Getenv("TESLA_PASSWORD"),
      })
    if err != nil {
      panic(err)
    }
    vehicles, err := client.Vehicles()
    if err != nil {
      panic(err)
    }

    vehicle := vehicles[0]
    fmt.Println(vehicle.State)

    state, _ := vehicle.Wakeup()
    for state.State != "online" {
      fmt.Println(state.State)
      time.Sleep(1 * time.Second)
      state, _ = vehicle.Wakeup()
    }
    fmt.Println(state.State)

    if strings.ToLower(request.State) == "on" {
      _ = vehicle.StartAirConditioning()
    }
    if strings.ToLower(request.State) == "off" {
      _ = vehicle.StopAirConditioning()
    }
    //_, _ = http.Get("http://192.168.1.247/setcolor?red=255&green=255&blue=255")
    _ = vehicle.FlashLights()
  }()

}
/*
requests.get("http://192.168.1.247/setpixelcolor?pixel="+str(i)+"&red=0&green=0&blue=255")
requests.get("http://192.168.1.247/setpixelcolor?pixel=1&red=255&green=255&blue=255")
requests.get("http://192.168.1.247/setcolor?red=255&green=255&blue=255")
*/

func main() {
  /*
  // Make a channel for results and start listening
  entriesCh := make(chan *mdns.ServiceEntry, 4)
  go func() {
      for entry := range entriesCh {
          fmt.Printf("Got new entry: %v\n", entry)
      }
  }()

  // Start the lookup
  mdns.Lookup("kitchen.local", entriesCh)
  close(entriesCh)
  */

    r := mux.NewRouter()
    r.HandleFunc("/", root)
    r.HandleFunc("/kitchen/party", party)
    r.HandleFunc("/white", white)
    r.HandleFunc("/hotwater", hotwater)
    r.HandleFunc("/tesla/condition", conditionTesla)

    loggedRouter := handlers.LoggingHandler(os.Stdout, r)
    log.Fatal(http.ListenAndServe(":8000", loggedRouter))
}