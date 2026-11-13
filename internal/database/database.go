package database

import (
	"context"
	"database/sql"
	"net"
	"time"

	"coffeepaste/internal/config"

	"github.com/go-sql-driver/mysql"
)

func Open(cfg *config.Config) (*sql.DB, error) {
	c := mysql.NewConfig()
	c.User = cfg.DBUser
	c.Passwd = cfg.DBPass
	c.Net = "tcp"
	c.Addr = net.JoinHostPort(cfg.DBHost, cfg.DBPort)
	c.DBName = cfg.DBName
	c.ParseTime = true
	c.ClientFoundRows = true
	c.Loc = time.Local
	// time_zone sesi disamakan dengan TZ aplikasi (WIB) agar NOW() di SQL dan waktu dari Go konsisten.
	c.Params = map[string]string{"charset": "utf8mb4", "time_zone": "'+07:00'"}

	db, err := sql.Open("mysql", c.FormatDSN())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
