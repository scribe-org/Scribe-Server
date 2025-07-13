# Creating a Scribe-Server Database

The following commands can be ran within a local MariaDB instance to create the database for Scribe-Server:

```sql
CREATE DATABASE scribe_server_v1;
SHOW DATABASES;  -- check that it's been created
ALTER USER 'root'@'localhost' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON scribe_server_v1.* TO 'root'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```
