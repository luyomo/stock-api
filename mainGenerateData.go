package main

import (
    "net/http"

    "fmt"
    "github.com/gin-gonic/gin"
    "time"
    "math/rand"
    "math"
    "sort"
    "database/sql"
    "strconv"
    _ "github.com/go-sql-driver/mysql"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func RandStringRunes(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letterRunes[rand.Intn(len(letterRunes))]
    }
    return string(b)
}

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

type tickDataRecord struct {
    Timestamp          int64  `json:"timestamp"`
    SecurityName       string `json:"security_name"`
    SecurityCode       int64  `json:"security_code"`
    Price              int64  `json:"price"`
    Quantity           int64  `json:"quantity"`
    BuySell            int    `json:"buy_sell"`
    IsMaker            int    `json:"is_maker"`
    IsMargin           int    `json:"is_margin"`
    Comment            string `json:"comment"`
}

func getTickDatas(c *gin.Context) {
    now := time.Now()
    nanos := now.UnixNano()
    millis := nanos / 1000000
    tickDatas.Timestamp = millis
    c.IndentedJSON(http.StatusOK, tickDatas)
}

func getNow() int64 {
  now := time.Now()
  nanos := now.UnixNano()
  millis := nanos / 1000000
  return millis
}

func generatedKLineDataList(basePrice float64, dataSize int) {
  var tickDatas []tickData
  //fmt.Println("Hell world ", basePrice)
  r := rand.New(rand.NewSource(99))
  //fmt.Println("Hello world ", r.Float64())
  baseValue := basePrice
  for i :=1; i < dataSize; i++ {
    //fmt.Println("The data numbering is ", i, baseValue)
    var prices []float64
    for i := 1; i < 5; i++ {
      baseValue = baseValue + r.Float64() * 20 - 10
      prices = append(prices, math.Round((r.Float64() - 0.5) * 12 + basePrice ))
    }
    sort.Float64s(prices)
    //fmt.Println(prices)
    openIdx  := int(math.Round(r.Float64() * 3))
    closeIdx := int(math.Round(r.Float64() * 2))
    if openIdx == closeIdx {
        closeIdx++
    }
    //fmt.Println("The open and close index is ", openIdx, closeIdx, getNow())
    var theTickData tickData
    theTickData.Open  = prices[openIdx]
    theTickData.Close = prices[closeIdx]
    theTickData.High  = prices[3]
    theTickData.Low   = prices[0]
    theTickData.Volume = math.Round(r.Float64() * 50 + 10)
    theTickData.Timestamp = getNow()
    fmt.Println(theTickData)
    tickDatas = append(tickDatas, theTickData) 
  }
}

func generateTickData(basePrice float64, dataSize int) {
  db, err := sql.Open("mysql", "tickuser:tickuser@tcp(192.168.1.105:3306)/tickdata")
  if err != nil {
    panic(err)
  }
  defer db.Close()

  tx, err := db.Begin()
  if err != nil {
    return
  }

  defer func() {
    switch err {
      case nil:
        err = tx.Commit()
      default:
        tx.Rollback()
    }
  }()

  //var tickDatas []tickData
  //fmt.Println("Hell world ", basePrice)
  r := rand.New(rand.NewSource(99))
  //fmt.Println("Hello world ", r.Float64())
  baseValue := basePrice
  for i :=1; i < dataSize; i++ {
    baseValue = baseValue + r.Float64() * 20 - 10
    price := int64(math.Round((r.Float64() - 0.5) * 12 + basePrice ))

    var tickData tickDataRecord
    tickData.SecurityName = "Dummy Secutiry"
    tickData.SecurityCode = 1111
    tickData.Price = price
    tickData.Quantity = int64(math.Round(r.Float64() * 50 + 10))
    tickData.BuySell = 1
    tickData.IsMaker = 1
    tickData.IsMargin = 1
    tickData.Timestamp = getNow()
    iLen := int(r.Float64()*64) + 64
    fmt.Println(iLen)
    tickData.Comment = RandStringRunes(iLen)
    fmt.Println(tickData)

    _, err = tx.Exec("INSERT INTO tickdata (order_time, security_name, security_code, price, quantity, buy_sell, is_maker, is_margin, comment) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", tickData.Timestamp, tickData.SecurityName, tickData.SecurityCode, tickData.Price, tickData.Quantity, tickData.BuySell, tickData.IsMaker, tickData.IsMargin, tickData.Comment)
    if err != nil {
      return
	}
  }
}

func tableExist() int {
  db, err := sql.Open("mysql", "tickuser:tickuser@tcp(192.168.1.105:3306)/tickdata")
  if err != nil {
    panic(err)
  }
  defer db.Close()
    // See "Important settings" section.
  db.SetConnMaxLifetime(time.Minute * 3)
  db.SetMaxOpenConns(10)
  db.SetMaxIdleConns(10)

  //rows, err := db.Query("SELECT current_timestamp as theTime") // 
  rows, err := db.Query("SELECT count(*) as cnt from information_schema.tables where table_name = 'tickdata'") // 
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
  fmt.Println(values)

  scanArgs := make([]interface{}, len(values))
  for i := range values {
    scanArgs[i] = &values[i]
  }

  for rows.Next() {
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
      fmt.Println(columns[i], ": ", value)
      i, err := strconv.Atoi(value)
      if err != nil {
        panic(err.Error())
      }
      return i
    }
  }
  return 0
}

func createTable() {
  db, err := sql.Open("mysql", "tickuser:tickuser@tcp(192.168.1.105:3306)/tickdata")
  if err != nil {
    panic(err)
  }
  defer db.Close()
    // See "Important settings" section.
  db.SetConnMaxLifetime(time.Minute * 3)
  db.SetMaxOpenConns(10)
  db.SetMaxIdleConns(10)

  test, err := db.Exec(`create table tickdata( 
      id bigint AUTO_INCREMENT
    , order_time         bigint not null
    , security_name varchar(64) not null
    , security_code      bigint not null
    , price              bigint not null
    , quantity           bigint not null
    , buy_sell           int    not null
    , is_maker           int
    , is_margin          int
    , comment            varchar(256)
    , primary key(id))`) 
  if err != nil {
    panic(err.Error())
  }
  fmt.Println(test)
}

func main() {
  //generatedKLineDataList(5000, 10)
  generateTickData(5000, 50)
  tableExists := tableExist()
  fmt.Println("The table exists here is ", tableExists)
  if tableExists == 0 {
    createTable()
  }
//    router := gin.Default()
//    router.GET("/tickData", getTickDatas)
//
//    router.Run("0.0.0.0:8000")
}
