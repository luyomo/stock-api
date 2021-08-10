package main

import (
    "net/http"
    "flag"
    "github.com/gin-gonic/gin"
    "time"
    "math"
    "math/rand"
    "fmt"
    "strconv"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

type tickData struct {
    SecurityName string  `json:"security_name"`
    Timestamp    int64   `json:"timestamp"`
    Open         int     `json:"open"`
    Close        int     `json:"close"`
    High         int     `json:"high"`
    Low          int     `json:"low"`
    Turnover     float64 `json:"turnover"`
    Volume       float64 `json:"volume"`
}

type dbConnInfoStruct struct {
    Host         string `json:"host"`
    Port         int    `json:"port"`
    User         string `json:"user"`
    Password     string `json:"pass"`
    Name         string `json:"name"`
}

var dbConnInfo dbConnInfoStruct

func getNow() int64 {
  now := time.Now()
  nanos := now.UnixNano()
  millis := nanos / 1000000
  return millis
}

var tickDatas = tickData{
    Close: 4987, High: 4987, Low: 4978, Open: 4987, Timestamp: 1628405520000, Turnover: 114441.62091504323, Volume: 22.954878888783895,
    // 1628521665170
}


func getTickDatas(c *gin.Context) {
    now := time.Now()
    nanos := now.UnixNano()
    millis := nanos / 1000000
    tickDatas.Timestamp = millis
    c.IndentedJSON(http.StatusOK, tickDatas)
    //c.IndentedJSON(http.StatusOK, theTickDataList)
}

func fetchDataFromDB(c *gin.Context) {
  dbConnStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbConnInfo.User, dbConnInfo.Password, dbConnInfo.Host, dbConnInfo.Port, dbConnInfo.Name)
  fmt.Println("The connection string is ", dbConnStr)
  db, err := sql.Open("mysql", dbConnStr)
  if err != nil {
    panic(err)
  }
  defer db.Close()
    // See "Important settings" section.
  db.SetConnMaxLifetime(time.Minute * 3)
  db.SetMaxOpenConns(10)
  db.SetMaxIdleConns(10)

  var theTickDataList []tickData

  //rows, err := db.Query("SELECT current_timestamp as theTime") // 
  //orderTime := getNow() - 1200000
  orderTime := getNow() - 1200000000
  fmt.Println("The time is ", orderTime)
  //query := fmt.Sprintf("select security_name, floor((order_time-mod(order_time, 60000))/1000) as time_minute, max(price) as max_price, min(price) as min_price from tickdata where order_time > %d group by security_name, floor((order_time-mod(order_time, 60000))/1000) desc limit 1" , orderTime ) 
  query := fmt.Sprintf("select security_name, (order_time-mod(order_time, 1000)) as time_minute, max(price) as max_price, min(price) as min_price from tickdata where order_time > %d group by security_name, (order_time-mod(order_time, 1000)) desc limit 2" , orderTime ) 
  rows, err := db.Query(query)
  if err != nil {
    panic(err.Error())
  }
  fmt.Println(rows)

  columns, err := rows.Columns() // カラム名を取得
  if err != nil {
    panic(err.Error())
  }
  fmt.Println(columns)

  values := make([]sql.RawBytes, len(columns))
  ////fmt.Println(values)

  scanArgs := make([]interface{}, len(values))
  for i := range values {
    scanArgs[i] = &values[i]
  }

  for rows.Next() {
    var theTickData tickData
    err = rows.Scan(scanArgs...)
    if err != nil {
      panic(err.Error())
    }

    var value string
    for i, col := range values {
      // Here we can check if the value is nil (NULL value)
      if col == nil {
        value = "NULL"
      } else {
        value = string(col)
      }
      //fmt.Println(columns[i], ": ", value)
      if columns[i] == "security_name" {
        theTickData.SecurityName = value
      }
      if columns[i] == "max_price" {
        iValue, err := strconv.Atoi(value)
        if err != nil {
           panic(err.Error())
        }
        theTickData.High = iValue
      }

      if columns[i] == "min_price" {
        iValue, err := strconv.Atoi(value)
        if err != nil {
           panic(err.Error())
        }
        theTickData.Low = iValue
        theTickData.Open = theTickData.Low + int(math.Floor(rand.Float64() * float64(theTickData.High - theTickData.Low)))
        theTickData.Close = theTickData.Low + int(math.Floor(rand.Float64() * float64(theTickData.High - theTickData.Low)))
      }

      if columns[i] == "time_minute" {
        iValue, err := strconv.ParseInt(value, 10, 64)
        if err != nil {
           panic(err.Error())
        }
        theTickData.Timestamp = iValue
      }

    }
    theTickDataList = append(theTickDataList, theTickData)
  }
  fmt.Println(theTickDataList)
  c.IndentedJSON(http.StatusOK, theTickDataList)
}

func main() {
    flag.StringVar(&dbConnInfo.Host    , "db-host", "127.0.0.1", "tidb db host")
    flag.IntVar   (&dbConnInfo.Port    , "db-port", 3306       , "tidb db port")
    flag.StringVar(&dbConnInfo.User    , "db-user", "root"     , "tidb db user")
    flag.StringVar(&dbConnInfo.Password, "db-pass", ""         , "tidb db pass")
    flag.StringVar(&dbConnInfo.Name    , "db-name", "tickdata" , "tidb db name")
    flag.Parse()
    //fetchDataFromDB()

    router := gin.Default()
    //router.GET("/tickData", getTickDatas)
    router.GET("/tickData", fetchDataFromDB)

    router.Run("0.0.0.0:8000")
}
