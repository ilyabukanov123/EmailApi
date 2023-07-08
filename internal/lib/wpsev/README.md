# WPSEV - обертка для работы с net/http
+ Имеет удобный роутер, с указанием http методов.
+ Есть возможность задавать динамические url
+ Мидалвары можно указывать списком, а не заворачивать один в другой
+ Поддержка мультипаттерна
+ Есть возможность выбора HTTP протокола
+ Сервер настраивается стандартными средствами net/http

### Пример простого сервера свыше перечисленными возможностями.

```go

func main() {

middleware := func(w http.ResponseWriter, r *http.Request) {
w.Write([]byte("this middleware\n"))
}

hs := &http.Server{}

myServer := wpsev.NewServer(hs, wpsev.HTTP3)

myServer.AddRouter(http.MethodGet, "/", middleware, Home)
myServer.AddRouter(http.MethodPost, "/", middleware, Home)
myServer.AddRouter(http.MethodGet, "/person/:name/:age", middleware, Person)
myServer.AddRouter(http.MethodGet, "/*file", middleware, File)
myServer.AddRouter(http.MethodGet, "/upload/:id/*file", middleware, Upload)

go func() {
err := myServer.Start("localhost", 6565)
if err != nil {
panic(err)
}
}()

go func() {
err := myServer.StartTLS("localhost", 6566, "wb.ru.crt", "wb.ru.key")
if err != nil {
panic(err)
}
}()

osSignalsCh := make(chan os.Signal, 1)
signal.Notify(osSignalsCh, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

<-osSignalsCh

err := myServer.Stop()
if err != nil {
panic(err)
}

}

func Upload(w http.ResponseWriter, r *http.Request) {
pattern := wpsev.GetParam(r, "pattern")
file := wpsev.GetParam(r, "file")
id := wpsev.GetParam(r, "id")
w.Write([]byte(fmt.Sprintf("pattern:%s\nid:%s\nfile:%s", pattern, id, file)))
}

func Person(w http.ResponseWriter, r *http.Request) {
w.Write([]byte("Pattern: " + wpsev.GetParam(r, "pattern") + "\n"))
w.Write([]byte("Name: " + wpsev.GetParam(r, "name") + "\n"))
w.Write([]byte("Age: " + wpsev.GetParam(r, "age") + "\n"))
}

func File(w http.ResponseWriter, r *http.Request) {
w.Write([]byte("Pattern: " + wpsev.GetParam(r, "pattern") + "\n"))
w.Write([]byte("File: " + wpsev.GetParam(r, "file")))
}

func Home(w http.ResponseWriter, r *http.Request) {
w.Write([]byte("Home " + r.Method))
}
```
