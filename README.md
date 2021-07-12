# What is this? #

This is a daemon that receives SNMP LinkUP/DOWN traps and strikes-back SNMP Get requests 
to find out host names, ifNames and ifAliases, storing it into MySQL.

# Why you should use it? #

- **Performance**. It handles traps asynchronously and will never loose one!
- **Speed**. Host names, ifNames and ifAliases are cached to prevent unnecessary SNMP Get requests.
- **FlapMyPort**. It does well with the <a href="http://flapmyport.com">FlapMyPort</a> project 
  and is developed to replace the outdated <a href="https://github.com/Pavel-Polyakov/trapharvester">TrapHandler</a>.

# What do you need to deploy it?
- A linux server with a MySQL database running
- Network devices supporting SNMPv2
- The devices should be reachable by SNMP protocol from the server

# Quick start in 3 steps #

## 1. Create a DB schema

```
mysql> create database handler;
# mysql handler < schema.sql
```

## 2. Create a config file

**handler.conf:**
```
listenAddress = "0.0.0.0"
listenPort = 162
dbHost = "localhost"
dbName = "handler"
dbUser = "root"
dbPassword = ""
community = "public"
logFilename = "/var/log/handler.log"
sendMail = false
mailList = ["user1@example.com", "user1@example.com"]
```

***Note:*** *Email notifications not implemented yet :(*

## 3. Run snmpflapd
```
> ./snmpflapd -f handler.conf
```
Check your log file for errors.

# How to build #

If you wish to make a build for a Linux 64-bit machine:

```
GOOS=linux GOARCH=amd64 go build -o handler
```

---
*And may a stable network be with you!*
