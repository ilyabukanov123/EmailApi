# Email Api

### Заполнение конфига 
1. Для заполнение конфига нужно открыть файл config.json в пакете config и внести в него определенные изменения:
````
   {
   "storagePath": "/Users/ilabukanov/Documents/mail", // Путь до директории с файлами пользователей 
   "ttl": 30, // время жизни уникальной ссылки в секундах
   "addr": "localhost", // адрес сервера
   "port": 8080, // порт сервера
   "cleaningTime": 40, // интервал в секунда по очистке map от элементов с истекщим ttl
   "authorizationToken" :  // токен авторизации 
   "Az7u1fzEFuI/gaDAEJDu7tMA4eVf9AH-iDGnIy9AvCZ1hv5kO/N4f?4dc3!8AYrT!DBfAcBmvinUi1o25UIrGaW0DtwInndU04gnuoFXrytAF4jrlfnHcJ=FsSvm=CdZNl39voYyjxF0UsGtkIXAL5c1EmYq2o9hivoegx9FBQ7dIIBdHmaHeMO/jdErWqHpkjJJ3Hfzy-ywjmg2Szr8gjz1UXePYGHFahX!/CZOorCoev7y3gRvZg8=7OGYwbVH"
}
````
### Запуск приложения
Для запуска приложения необходимо вызвать функцию Start() из пакета /internal/http-server/app. В качестве параметра в данную функцию необходимо передать путь до папки с конфигом 
````
app.Start("/Users/ilabukanov/go/src/WB Work/api-mail/config/config.json")
````

### Папка с почтами пользователей 

В папке email находятся примеры с почтами пользователей. В дирректории с почтами пользователей должен находится файл readme.pdf 

### Запросы к API 

Запрос на получение уникальной ссылки
http://localhost:8080/*username/(http://localhost:8080/petrov.mikhail@wb.ru/)
Данный запрос требует статическую авторизацию. В headers необходимо передать ключ Authorization и токен из файла config.json
````
"authorizationToken" : "Az7u1fzEFuI/gaDAEJDu7tMA4eVf9AH-iDGnIy9AvCZ1hv5kO/N4f?4dc3!8AYrT!DBfAcBmvinUi1o25UIrGaW0DtwInndU04gnuoFXrytAF4jrlfnHcJ=FsSvm=CdZNl39voYyjxF0UsGtkIXAL5c1EmYq2o9hivoegx9FBQ7dIIBdHmaHeMO/jdErWqHpkjJJ3Hfzy-ywjmg2Szr8gjz1UXePYGHFahX!/CZOorCoev7y3gRvZg8=7OGYwbVH"
````
Если авторизация не будет пройдена вернется данное сообщение - You need to log in to get a unique link. Authorization failed


Запрос на получение архива по уникальной ссылке. Если ссылка указана не верно вернется статус StatusBadRequest и сообщение Invalid link
http://localhost:8080/get/*link(http://localhost:8080/get/9cff3d9e-f7e7-4a07-aa42-b73aaf3d32bf)
