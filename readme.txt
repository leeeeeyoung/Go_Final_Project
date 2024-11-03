install bellow package first:

go get github.com/gorilla/mux
go get github.com/dgrijalva/jwt-go
go get golang.org/x/crypto/bcrypt
go get -u gorm.io/gorm
go get -u gorm.io/driver/sqlite
go get -u gorm.io/driver/mysql


run bellow at the MySQL server:

CREATE DATABASE memo_app;
CREATE USER 'username'@'localhost' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON memo_app.* TO 'username'@'localhost';
FLUSH PRIVILEGES;
