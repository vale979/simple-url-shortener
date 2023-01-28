package main

import (
  "github.com/go-redis/redis"
  "github.com/go-chi/chi"
  "os"
  "log"
  "fmt"
  "math/rand"
  "path"
  "time"
  "strings"
  "regexp"
  "net/http"
  "html/template"
)

var redis_address = "localhost:6379"
var redis_pwd = os.Getenv("REDIS_PWD")
//var redis_pwd = ""

var client = redis.NewClient(&redis.Options{
  Addr     : redis_address,
  Password : redis_pwd,
  DB       : 10,
})
 


func main(){
    // Test db connection
  pong, err := client.Ping().Result()
  if err != nil {
    log.Fatal(err)
    return
  }
  if pong != "" {
    fmt.Println("Redis say pong!")
  }
  
  r := chi.NewRouter()
  r.Get("/", homeHandler)
  r.Get("/{linkId}", linkHandler)

  fmt.Println("Server started at port 8080")
  http.ListenAndServe(":8080", r)
}

func validateUrl(url string) bool {
  var regex, err = regexp.Compile(`https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]{2,4}\b([-a-zA-Z0-9@:%_\+.~#?&//=]*)`)

  if err != nil {
    fmt.Println(err.Error())
  }

  var isMatch = regex.MatchString(url)
  return isMatch
}

func homeHandler(w http.ResponseWriter, r *http.Request){
  var filepath = path.Join("template", "index.html")
  var tmpl, err = template.ParseFiles(filepath)
  
  if err != nil{
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  
  // check first if url parameter exists

  url := r.FormValue("url")

  var short_key = ""

  if validateUrl(url) {
    // give shortened url
    for true {
      short_key = get_string(5,13)
      _, err := client.Get(short_key).Result()
      if err == redis.Nil{
        // If the key does not already exist, use the key
        break
      }
    }
    err := client.Set(short_key, url, 0).Err()
    if err != nil{
      http.Error(w, err.Error(), http.StatusInternalServerError)
    } 
  }
  var data_render = map[string]interface{}{
      "url": short_key,
  }
  err = tmpl.Execute(w, data_render)
  if err != nil{
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func linkHandler(w http.ResponseWriter, r *http.Request){
  link := chi.URLParam(r, "linkId")
  true_link, err := client.Get(link).Result()
  if err == redis.Nil {
    http.Error(w, "URL not found", http.StatusNotFound)
  } else if err != nil{
    http.Error(w, err.Error(), http.StatusInternalServerError)
  } else {
    http.Redirect(w, r, true_link, http.StatusFound)
  }
}


func get_string(minLength, maxLength int) string {
  const seedBytes = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNOPQRSTUVWXYZ023579_-"
  var src = rand.NewSource(time.Now().UnixNano())
  const (
    letterIdxBits = 6
    letterIdxMask = 1 << letterIdxBits - 1
    letterIdxMax = 63/letterIdxBits
  )
    if minLength >= maxLength{
      log.Fatal("Invalid length")
    }

    if minLength < 1{
      log.Fatal("Invalid minlength")
    }
    // https://stackoverflow.com/a/31832326
    n := (src.Int63() % int64(maxLength - minLength)) + int64(minLength)
    sb := strings.Builder{}

    for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0;{
      if remain == 0{
        cache, remain = src.Int63(), letterIdxMax
      }

      if idx := int(cache & letterIdxMask); idx < len(seedBytes){
        sb.WriteByte(seedBytes[idx])
        i--
      }

      cache >>= letterIdxBits
      remain--
    }
    return sb.String()
}
