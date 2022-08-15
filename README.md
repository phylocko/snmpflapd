# What is this? #

This is a daemon that receives SNMP LinkUP/DOWN traps and strikes-back SNMP Get requests 
to find out host names, ifNames and ifAliases, storing it into MySQL.

# Why you should use it? #

- **Performance**. It handles traps asynchronously and will never loose one!
- **Speed**. Host names, ifNames and ifAliases are cached to prevent unnecessary SNMP Get requests
- **FlapMyPort**. It does well with the <a href="http://flapmyport.com">FlapMyPort</a> monitoring system

# What do you need to deploy it?
- A linux server with a MySQL database running
- Network devices supporting SNMPv2
- The devices should be reachable by SNMP protocol from the server

# Quick start in 3 steps #

## 1. Create a DB schema

```
mysql> create database snmpflapd;
# mysql snmpflapd < schema.sql
```

## 2. Create a config file

**settings.conf:**
```
listenAddress = "0.0.0.0"
listenPort = 162
dbHost = "localhost"
dbName = "snmpflapd"
dbUser = "root"
dbPassword = ""
community = "public"
logFilename = "/var/log/snmpflapd.log"
```

> settings.conf is optional. You may use environment variables instaed
> Available environment variables are
> LISTEN_ADDRESS, LISTEN_PORT, DBHOST, DBNAME, DBUSER, DBPASSWORD, COMMUNITY, LOGFILE

## 3. Run snmpflapd
```
> ./snmpflapd -f settings.conf
```
Check your log file for errors.

# How to build #

Use `build.sh` instead of `go build`!

If you wish to make a build for a Linux 64-bit machine:

```
GOOS=linux GOARCH=amd64 build.sh
```

---
*And may a stable network be with you!*
