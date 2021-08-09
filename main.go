package main

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "time"
)

type tickData struct {
    Timestamp int64   `json:"timestamp"`
    Open      float64 `json:"open"`
    Close     float64 `json:"close"`
    High      float64 `json:"high"`
    Low       float64 `json:"low"`
    Turnover  float64 `json:"turnover"`
    Volume    float64 `json:"volume"`
}

var tickDatas = tickData{
    Close: 4987.777263535087, High: 4987.777263535087, Low: 4978.848433882813, Open: 4987.609622866156, Timestamp: 1628405520000, Turnover: 114441.62091504323, Volume: 22.954878888783895,
}

func getTickDatas(c *gin.Context) {
    now := time.Now()
    nanos := now.UnixNano()
    millis := nanos / 1000000
    tickDatas.Timestamp = millis
    c.IndentedJSON(http.StatusOK, tickDatas)
}

func main() {
    router := gin.Default()
    router.GET("/tickData", getTickDatas)

    router.Run("0.0.0.0:8000")
}
