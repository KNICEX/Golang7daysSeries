package main

import (
	"geeorm"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func main() {
	engine, _ := geeorm.NewEngine("sqlite3", "test.db")
	defer engine.Close()
	s := engine.NewSession()
	_, _ = s.Raw("DROP TABLE IF EXISTS user;").Exec()
	_, _ = s.Raw("CREATE TABLE user(name text, age integer);").Exec()
	_, _ = s.Raw("CREATE TABLE user(name text, age integer);").Exec()
	result, err := s.Raw(
		"INSERT INTO user('name', 'age') values (?, ?), (?, ?)",
		"Chtholly", 17, "Arcueid", 300).Exec()
	if err == nil {
		affected, _ := result.RowsAffected()
		log.Printf("insert %v rows", affected)
	}

}
