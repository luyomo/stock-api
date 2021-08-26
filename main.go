package main

import (
    "flag"
//    "github.com/gin-gonic/gin"
    "time"
//    "math"
//    "math/rand"
    "fmt"
    "strconv"
    "database/sql"
    "log"
    _ "github.com/go-sql-driver/mysql"

     "net/http"
     "github.com/gorilla/websocket"
)

type tickData struct {
    SecurityName string  `json:"security_name"`
    Timestamp    int64   `json:"timestamp"`
    Open         int     `json:"open"`
    Close        int     `json:"close"`
    High         int     `json:"high"`
    Low          int     `json:"low"`
    Turnover     float64 `json:"turnover"`
    Volume       int `json:"volume"`
}

type dbConnInfoStruct struct {
    Host         string `json:"host"`
    Port         int    `json:"port"`
    User         string `json:"user"`
    Password     string `json:"pass"`
    Name         string `json:"name"`
}

var upgrader = websocket.Upgrader{} // use default options

var dbConnInfo dbConnInfoStruct

func getNow() int64 {
  now := time.Now()
  nanos := now.UnixNano()
  millis := nanos / 1000000
  return millis
}

//func fetchDataFromDB(c *gin.Context) {
func fetchDataFromDB() []tickData {
  dbConnStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbConnInfo.User, dbConnInfo.Password, dbConnInfo.Host, dbConnInfo.Port, dbConnInfo.Name)
  //fmt.Println("The connection string is ", dbConnStr)
  db, err := sql.Open("mysql", dbConnStr)
  if err != nil {
    panic(err)
  }
  defer db.Close()
    // See "Important settings" section.
  db.SetConnMaxLifetime(time.Minute * 3)
  db.SetMaxOpenConns(10)
  db.SetMaxIdleConns(10)

  theTickDataList := []tickData{}

  //rows, err := db.Query("SELECT current_timestamp as theTime") // 
  orderTime := getNow() - 3000
  //orderTime := getNow() - 1200000000
  //fmt.Println("The time is ", orderTime)
  query := fmt.Sprintf(`
  with tbl_open_close_time as (
    select security_name, (order_time-mod(order_time, 60000)) as time_minute
         , min(order_time) as open_time, max(order_time) as close_time
      from tickdata
     where order_time > %d
  group by security_name, (order_time-mod(order_time, 60000))
  ), tbl_summary as (
    select security_name
         , (order_time-mod(order_time, 60000)) as time_minute
         , max(price) as max_price, min(price) as min_price, sum(quantity) as volume
      from tickdata
     where order_time > %d
  group by security_name, (order_time-mod(order_time, 60000))
  )
  select t1.security_name, t1.time_minute, t2.price as open_price, t3.price as close_price
       , t4.max_price, t4.min_price, t4.volume
      from tbl_open_close_time t1
inner join tickdata t2
        on t1.open_time = t2.order_time
inner join tickdata t3
        on t1.close_time = t3.order_time
inner join tbl_summary t4
        on t1.time_minute = t4.time_minute` , orderTime, orderTime ) 
  rows, err := db.Query(query)
  if err != nil {
    panic(err.Error())
  }
  //fmt.Println(rows)

  columns, err := rows.Columns() // カラム名を取得
  if err != nil {
    panic(err.Error())
  }
  //fmt.Println(columns)

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
      }
      if columns[i] == "open_price" {
        iValue, err := strconv.Atoi(value)
        if err != nil {
           panic(err.Error())
        }
        theTickData.Open = iValue
      }
      if columns[i] == "close_price" {
        iValue, err := strconv.Atoi(value)
        if err != nil {
           panic(err.Error())
        }
        theTickData.Close = iValue
      }
      if columns[i] == "volume" {
        iValue, err := strconv.Atoi(value)
        if err != nil {
           panic(err.Error())
        }
        theTickData.Volume = iValue
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
  return theTickDataList
  //fmt.Println(theTickDataList)
  //c.IndentedJSON(http.StatusOK, theTickDataList)
}

func handleTickData(w http.ResponseWriter, r *http.Request) {
    c, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Print("upgrade:", err)
        return
    }
    defer c.Close()
    for {
        mt, message, err := c.ReadMessage()
        log.Println("Get message:", message, " and ", mt)
        if err != nil {
            log.Println("read:", err)
            break
        }
        for true {
            time.Sleep(2 * time.Second)
            data := fetchDataFromDB()
            log.Printf("recv: %s", data)
            //err = c.WriteMessage(mt, message)
            err = c.WriteJSON(data)
            if err != nil {
                log.Println("write:", err)
                break
            }
        }
    }
}

func main() {
    flag.StringVar(&dbConnInfo.Host    , "db-host", "127.0.0.1", "tidb db host")
    flag.IntVar   (&dbConnInfo.Port    , "db-port", 3306       , "tidb db port")
    flag.StringVar(&dbConnInfo.User    , "db-user", "root"     , "tidb db user")
    flag.StringVar(&dbConnInfo.Password, "db-pass", ""         , "tidb db pass")
    flag.StringVar(&dbConnInfo.Name    , "db-name", "tickdata" , "tidb db name")
    portPtr     := flag.Int("port", 8000, "Port the service is listening on")
    hostPtr     := flag.String("host","0.0.0.0", "The ip the service is listening on")
    flag.Parse()

    http.HandleFunc("/tickData", handleTickData)
    log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", *hostPtr, *portPtr), nil))
}
