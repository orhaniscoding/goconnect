package main
import (
  "flag"
  "fmt"
  "log"
  "math/rand"
  "net/http"
)
var (
  version = "dev"
  commit  = "none"
  date    = "2025-09-22"
  builtBy = "orhaniscoding"
)
func main(){
  showVersion := flag.Bool("version", false, "print version and exit")
  flag.Parse()
  if *showVersion {
    fmt.Printf("goconnect-daemon %s (commit %s, build %s) built by %s\n", version, commit, date, builtBy)
    return
  }
  port := 12000 + rand.Intn(1000)
  http.HandleFunc("/status", func(w http.ResponseWriter,_ *http.Request){
    w.Header().Set("Content-Type","application/json")
    w.Write([]byte(`{"running":true,"wg":{"active":false}}`))
  })
  addr := fmt.Sprintf("127.0.0.1:%d", port)
  log.Printf("daemon bridge at http://%s", addr)
  log.Fatal(http.ListenAndServe(addr,nil))
}
