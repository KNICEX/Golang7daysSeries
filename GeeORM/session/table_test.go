package session

import (
	"database/sql"
	"geeorm/dialect"
	_ "github.com/mattn/go-sqlite3"
	"testing"
)

type User struct {
	Name string `geeorm:"primary key"`
	Age  int
}

func TestCreateTable(t *testing.T) {
	dial, _ := dialect.GetDialect("sqlite3")
	db, _ := sql.Open("sqlite3", "test.db")
	s := New(db, dial)
	s.Model(User{})
	_ = s.DropTable()
	_ = s.CreateTable()
	if !s.HasTable() {
		t.Fatal("Failed to create table")
	}
}
