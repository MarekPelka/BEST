# BEST

Usage
-----
First you need to start database, here we use docker-compose as glorified docker run:
    docker-compose up

Next we can use the application to generate rainbowtable:
```
    cd app/
    go run hash.go -g <FILENAME>
```
And use it to break passwords:
```
    cd app/
    go run hash.go -f <PasswordHash>
    EXAMPLE:
    go run hash.go -f f82a7d02e8f0a728b7c3e958c278745cb224d3d7b2e3b84c0ecafc5511fdbdb7 #should return password!
```
Todo:
1. In **func reduction(h string) string** have to implement some sort of reduction function - currently it just a hash value - not very common password
2. Do some performance benchmarks of breaking passwords - maybe CLUSTERED COLUMNSTORE?
3. Dockerize hash.go

Tools:
To connect to DB manually first install Microsoft's ODBC driver with *scripts/installODCB.sh* and the run it with *scripts/sqlplus.sh*
```
     scripts/sqlplus.sh
     > Use HashDB
     > GO

     > SELECT * FROM RainbowSchema.rainbow
     > GO

     > DELETE FROM RainbowSchema.rainbow
     > GO
```
