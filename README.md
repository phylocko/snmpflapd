# What is this? #

This is a daemon that receives SNMP LinkUP/DOWN traps and strikes-back SNMP Get requests 
to find out host names, ifNames and ifAliases, storing it into MySQL. 

✉️ The system also sends combined emails

# Why you should use it? #

- **Performance**. It handles traps asynchronously and will never loose one!
- **Speed**. Host names, ifNames and ifAliases are cached to prevent unnecessary SNMP Get requests.
- **FlapMyPort**. It doing well with the <a href="http://flapmyport.com">FlapMyPort</a> project 
  and is developed to replace it's outdated <a href="https://github.com/Pavel-Polyakov/trapharvester">TrapHandler</a>.

# Quick start #

```
# mysql -u root
mysql> create database handler;
# mysql -u root handler < schema.sql
# handler -f handler.conf
```

**Example config file**

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

# How to build #

If you wish to make a build for a Linux 64-bit machine:

```
GOOS=linux GOARCH=amd64 go build -o handler
```

---
*And may a stable network be with you!*
