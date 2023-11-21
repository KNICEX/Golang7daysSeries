package session

import (
	"geeorm/logger"
	"testing"
)

var (
	user1 = &User{"Tom", 18}
	user2 = &User{"Sam", 25}
	user3 = &User{"Jack", 25}
)

func (u *User) BeforeInsert(s *Session) error {
	logger.Info("before inert", u)
	u.Age += 1000
	return nil
}

func (u *User) AfterInsert(s *Session) error {
	logger.Info("after inert", u)
	return nil
}

func (u *User) AfterQuery(s *Session) error {
	logger.Info("after query", u)
	u.Name = "AfterQuery"
	return nil
}

func (u *User) BeforeUpdate(s *Session) error {
	logger.Info("before update", u)
	u.Name = "BeforeUpdate"
	return nil
}

func (u *User) AfterUpdate(s *Session) error {
	logger.Info("after update", u)
	u.Name = "AfterUpdate"
	return nil
}

func (u *User) BeforeDelete(s *Session) error {
	logger.Info("before delete", u)
	u.Name = "BeforeDelete"
	return nil
}

func (u *User) AfterDelete(s *Session) error {
	logger.Info("after delete", u)
	u.Name = "AfterDelete"
	return nil
}

func testRecordInit(t *testing.T) *Session {
	t.Helper()
	s := NewSession().Model(&User{})
	err1 := s.DropTable()
	err2 := s.CreateTable()
	_, err3 := s.Insert(user1, user2)
	if err1 != nil || err2 != nil || err3 != nil {
		t.Fatal("failed init test records")
	}
	return s
}

func TestSession_Find(t *testing.T) {
	s := testRecordInit(t)
	var users []User
	if err := s.Find(&users); err != nil || len(users) != 2 {
		t.Fatal("failed to query all")
	}
}

func TestSession_Limit(t *testing.T) {
	s := testRecordInit(t)
	var users []User
	err := s.Limit(1).Find(&users)
	if err != nil || len(users) != 1 {
		t.Fatal("failed to query with limit condition")
	}
}

func TestSession_Update(t *testing.T) {
	s := testRecordInit(t)
	affected, _ := s.Where("Name = ?", "Tom").Update("Age", 30)
	u := &User{}
	_ = s.OrderBy("Age DESC").First(u)

	if affected != 1 || u.Age != 30 {
		t.Fatal("failed to update")
	}
}

func TestSession_DeleteAndCount(t *testing.T) {
	s := testRecordInit(t)
	affected, _ := s.Where("Name = ?", "Tom").Delete()
	count, _ := s.Count()

	if affected != 1 || count != 1 {
		t.Fatal("failed to delete or count")
	}
}
