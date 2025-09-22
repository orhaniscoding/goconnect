package main
import (
  "flag"
  "fmt"
  "net/http"
  "github.com/gin-gonic/gin"
)
var (
  version = "dev"
  commit  = "none"
  date    = "unknown"
  builtBy = "orhaniscoding"
)
func main(){
  showVersion := flag.Bool("version", false, "print version and exit")
  flag.Parse()
  if *showVersion {
    fmt.Printf("goconnect-server %s (commit %s, build %s) built by %s
", version, commit, date, builtBy)
    return
  }
  r:=gin.Default()
  r.GET("/health", func(c *gin.Context){ c.JSON(200, gin.H{"ok":true,"service":"goconnect-server"}) })
  r.POST("/v1/auth/register", func(c *gin.Context){ c.JSON(200, gin.H{"data": gin.H{"user_id":"ulid"}})})
  r.POST("/v1/auth/login", func(c *gin.Context){ c.JSON(200, gin.H{"data": gin.H{"access_token":"dev","refresh_token":"dev"}})})
  _=http.ListenAndServe(":8080", r)
}
