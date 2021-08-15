package main

import (
//    "net/http"
    "runtime"

    "flag"
    "fmt"
//    "github.com/gin-gonic/gin"
    "time"
    "math/rand"
    "math"
//    "sort"
    "database/sql"
    "strconv"
    "sync"
    _ "github.com/go-sql-driver/mysql"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
var count int

func RandStringRunes(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letterRunes[rand.Intn(len(letterRunes))]
    }
    return string(b)
}

type dbConnInfoStruct struct {
    Host         string `json:"host"`
    Port         int    `json:"port"`
    User         string `json:"user"`
    Password     string `json:"pass"`
    Name         string `json:"name"`
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

var dbConnInfo dbConnInfoStruct

func getNow() int64 {
  now := time.Now()
  nanos := now.UnixNano()
  millis := nanos / 1000000
  return millis
}

func generateTickData(basePrice float64, dataSize int, wg *sync.WaitGroup) {
  defer wg.Done()

  dbConnStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbConnInfo.User, dbConnInfo.Password, dbConnInfo.Host, dbConnInfo.Port, dbConnInfo.Name)
  db, err := sql.Open("mysql", dbConnStr)

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

  //fmt.Println("Hell world ", basePrice)
  r := rand.New(rand.NewSource(99))
  //fmt.Println("Hello world ", r.Float64())
  baseValue := basePrice
  fmt.Println("The count is ", count)
  for iCount :=1; iCount < (count + 1); iCount++ {
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
      //fmt.Println(iLen)
      tickData.Comment = RandStringRunes(iLen)
      //fmt.Println(tickData)

      _, err = tx.Exec("INSERT INTO tickdata (order_time, security_name, security_code, price, quantity, buy_sell, is_maker, is_margin, comment) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", tickData.Timestamp, tickData.SecurityName, tickData.SecurityCode, tickData.Price, tickData.Quantity, tickData.BuySell, tickData.IsMaker, tickData.IsMargin, tickData.Comment)
      if err != nil {
        return
      }
    }
    tx.Commit()
    if iCount % 100 == 0 {
      db.Close()
      db, err = sql.Open("mysql", dbConnStr)

      if err != nil {
        panic(err)
      }
    }

    tx, err = db.Begin()
    if err != nil {
      return
    }
  }
}

func tableExist() int {
  dbConnStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbConnInfo.User, dbConnInfo.Password, dbConnInfo.Host, dbConnInfo.Port, dbConnInfo.Name)
  db, err := sql.Open("mysql", dbConnStr)
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
  //fmt.Println(rows)

  columns, err := rows.Columns() // カラム名を取得
  if err != nil {
    panic(err.Error())
  }
  //fmt.Println(columns)

  values := make([]sql.RawBytes, len(columns))
  //fmt.Println(values)

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
    for _, col := range values {
      // Here we can check if the value is nil (NULL value)
      if col == nil {
        value = "NULL"
      } else {
        value = string(col)
      }
      //fmt.Println(columns[i], ": ", value)
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
  dbConnStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbConnInfo.User, dbConnInfo.Password, dbConnInfo.Host, dbConnInfo.Port, dbConnInfo.Name)
  db, err := sql.Open("mysql", dbConnStr)
  if err != nil {
    panic(err)
  }
  defer db.Close()
    // See "Important settings" section.
  db.SetConnMaxLifetime(time.Minute * 3)
  db.SetMaxOpenConns(10)
  db.SetMaxIdleConns(10)

  db.Exec(`create table tickdata( 
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
  //fmt.Println(test)
}

func main() {
  //generatedKLineDataList(5000, 10)
  //dbHostPtr := flag.String("db-host"    , "127.0.0.1", "tidb db host"     )
  fmt.Printf("GOMAXPROCS is %d\n", runtime.GOMAXPROCS(0)  )
  startTime := getNow()
  flag.StringVar(&dbConnInfo.Host    , "db-host", "127.0.0.1", "tidb db host")
  flag.IntVar   (&dbConnInfo.Port    , "db-port", 3306       , "tidb db port")
  flag.StringVar(&dbConnInfo.User    , "db-user", "root"     , "tidb db user")
  flag.StringVar(&dbConnInfo.Password, "db-pass", ""         , "tidb db pass")
  flag.StringVar(&dbConnInfo.Name    , "db-name", "tickdata" , "tidb db name")
  var threads   int
  var baseValue float64
  var rows      int
  flag.IntVar     (&threads  , "threads"   , 2       , "multiple thread to import the data ")
  flag.Float64Var (&baseValue, "base-value", 5000    , "Base value of the price"            )
  flag.IntVar     (&rows     , "rows"      , 50      , "rows per commit"                    )
  flag.IntVar     (&count    , "count"     , 1000    , "Count to insert the data"           )

  flag.Parse()

  fmt.Println("db host:  -> ", dbConnInfo.Host    )
  fmt.Println("db port:  -> ", dbConnInfo.Port    )
  fmt.Println("db user:  -> ", dbConnInfo.User    )
  fmt.Println("db pass:  -> ", dbConnInfo.Password)
  fmt.Println("db name:  -> ", dbConnInfo.Name    )

  fmt.Println("The multiupl thread is ", threads)

  tableExists := tableExist()
  fmt.Println("The table exists here is ", tableExists)
  if tableExists == 0 {
    createTable()
  }

  var wg sync.WaitGroup

  for i := 1; i < threads + 1; i++ {
    wg.Add(1)
    go generateTickData(baseValue, rows, &wg)
  }
  wg.Wait()
  takenTime := getNow() - startTime
  fmt.Println("The data insert took ", takenTime, " mills seconds")
//    router := gin.Default()
//    router.GET("/tickData", getTickDatas)
//
//    router.Run("0.0.0.0:8000")
}

// select from_unixtime(floor((order_time-order_time%60000)/1000)) as time_minute, max(price) as max_price, min(price) as min_price from tickdata group by from_unixtime(floor((order_time-order_time%60000)/1000));
