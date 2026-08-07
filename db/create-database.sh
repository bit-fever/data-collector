#!/bin/bash

dbName=collector
dbUser=admin
dbPass=admin
dbPort=3306

mysql --user=root --password=root --host=127.0.0.1 --port=$dbPort << EOF
drop database if exists ${dbName}; 
create database ${dbName} character set = 'utf8' collate = 'utf8_bin';

grant references,alter,create,drop,index,select,update,delete,insert on ${dbName}.* to '${dbUser}'@'localhost' identified by '${dbPass}';
grant references,alter,create,drop,index,select,update,delete,insert on ${dbName}.* to '${dbUser}'@'%'         identified by '${dbPass}';
EOF

mysql --user=$dbUser --password=$dbPass --host=127.0.0.1 --port=$dbPort $dbName < $dbName-schema.sql
