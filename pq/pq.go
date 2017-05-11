// Package pq wraps the lib/pq package to instrument all the things.
package pq

import (
	"database/sql"
	"database/sql/driver"
	"net"
	"time"

	"github.com/lib/pq"
	"github.com/remind101/pkg/metrics"
)

func init() {
	sql.Register("postgresx", &drv{})
}

// We were seeing queries hang, and we think it's because of NAT gateway timeouts:
// http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/vpc-nat-gateway.html#nat-gateway-troubleshooting-timeout
// Setting a keepalive should help.
const defaultKeepAlive = 3 * time.Minute

type customDialer struct{}

func (d customDialer) Dial(ntw, addr string) (net.Conn, error) {
	dialer := d.dialer()
	return dialer.Dial(ntw, addr)
}

func (d customDialer) DialTimeout(ntw, addr string, timeout time.Duration) (net.Conn, error) {
	dialer := d.dialer()
	dialer.Timeout = timeout
	return dialer.Dial(ntw, addr)
}

func (d customDialer) dialer() net.Dialer {
	var dialer net.Dialer
	dialer.KeepAlive = defaultKeepAlive
	return dialer
}

type drv struct{}

func (d *drv) Open(name string) (driver.Conn, error) {
	t := metrics.Time("db.Conn.Open", nil, 1.0)
	defer t.Done()

	c, err := pq.DialOpen(customDialer{}, name)
	return conn{c}, err
}

type conn struct {
	driver.Conn
}

func (c conn) Close() error {
	t := metrics.Time("db.Conn.Close", nil, 1.0)
	defer t.Done()

	err := c.Conn.Close()
	return err
}

func (c conn) Query(query string, args []driver.Value) (driver.Rows, error) {
	t := metrics.Time("db.Conn.Query", nil, 1.0)
	defer t.Done()

	return c.Conn.(driver.Queryer).Query(query, args)
}

func (c conn) Exec(query string, args []driver.Value) (driver.Result, error) {
	t := metrics.Time("db.Conn.Exec", nil, 1.0)
	defer t.Done()

	return c.Conn.(driver.Execer).Exec(query, args)
}
