package main

import (
  "bytes"
  "fmt"
  "log"
  "os/exec"
  "github.com/gin-gonic/gin"
)


func exec_shell(s string) string {
    cmd := exec.Command("/bin/bash", "-c", s)
    var out bytes.Buffer

    cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%s", out.String())
    return out.String()
}

func main() {
    r := gin.Default()
    r.GET("/info", func(c *gin.Context) {
        str := exec_shell("echo \"show info \" |socat /var/run/haproxy.sock stdio")
        c.JSON(200, gin.H{
		"message": str,
	})
    })
    r.GET("/table", func(c *gin.Context) {
        str := exec_shell("echo \"show table \" |socat /var/run/haproxy.sock stdio")
        c.JSON(200, gin.H{
		"message": str,
	})
    })
    r.POST("/table", func(c *gin.Context) {
        table := c.PostForm("table")
	ip := c.PostForm("ip")
	role := c.PostForm("role")
        command := fmt.Sprintf("set table %s key %s data.gpc0 %s",table,ip,role)
	fmt.Printf("%s", command)
	str := exec_shell("echo \""+command+"\" |socat /var/run/haproxy.sock stdio")
        c.JSON(200, gin.H{
		"message": str,
	})
    })
    r.Run() // listen and serve on 0.0.0.0:8080
}
