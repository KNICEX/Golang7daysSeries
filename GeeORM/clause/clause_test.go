package clause

import "testing"

func TestClause_Build(t *testing.T) {
	var clause Clause
	clause.Set(LIMIT, 3)
	clause.Set(SELECT, "user", []string{"name", "age"})
	clause.Set(WHERE, "name = ?", "Chtholly")
	clause.Set(ORDERBY, "age asc")
	sql, vars := clause.Build(SELECT, WHERE, ORDERBY, LIMIT)
	t.Log(sql, vars)
	if sql != "SELECT name,age FROM user WHERE name = ? ORDER BY age asc LIMIT ?" {
		t.Fatal("sql incorrect")
	}
}
