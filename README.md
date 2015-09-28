todo
====

A simple REST API based in-memory storage for TODO application.

Usage
-----

### Create

`curl -i -X POST -H "Content-Type: application/json" -d '{"title":"new"}' http://localhost:8080/task/`

### Read

`curl -i -X GET -H "Accept: application/json" http://localhost:8080/task/0`

### Read all

`curl -i -X GET -H "Accept: application/json" http://localhost:8080/task/`

### Update

`curl -i -X PUT -H "Content-Type: application/json" -d '{"id":0,"title":"update"}' http://localhost:8080/task/0`

### Delete

`curl -i -X DELETE http://localhost:8080/task/0`