# Goly, The Go URL Shortener

This is the repo for the video found [here](https://youtu.be/bTLQT7W12dQ)

---
## PostgreSQL Docker Image
I used postgres:14 and can run it
```bash
$ docker run --name name-of-container -e POSTGRES_USER=admin -e POSTGRES_PASSWORD=test -d postgres:14
```
You can name the container anything you want, or not name it all. `--name` flag is really used to run docker commands easier, rather than using the randomly generated UUID.

---
:zap: Happy Coding!

it is good to consider using go doc [packageName] in order to get a better understanding of what we want to use in our software.
For example:
go doc fmt.Prinln
or simply:
go doc fmt


This is how we can declare Variables, var variableName varType
Example:
  var stationName string
  var nextTrainTime int8
  var isUptownTrain bool

