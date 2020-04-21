package mgt

import (
	"github.com/jinzhu/gorm"
	"github.com/rriverak/gogo/internal/config"
	"github.com/rriverak/gogo/internal/utils"
)

//User in SQL
type User struct {
	gorm.Model
	UserName     string
	PasswordHash string
	IsAdmin      bool
}

//ChangePassword of User
func (u *User) ChangePassword(password string) {
	u.PasswordHash = utils.GenerateHash(password)
}

//CheckPassword of User
func (u *User) CheckPassword(password string) (bool, error) {
	return utils.CompareHash(u.PasswordHash, password)
}

//NewUserRepository returns a SQL Repostitory
func NewUserRepository(cfg *config.Config, db *gorm.DB) Repository {
	db.AutoMigrate(&User{})
	repo := userRepository{config: cfg, db: db}
	admin := NewUser("admin", "admin")
	repo.Insert(&admin)
	return &repo
}

//NewUser Instance
func NewUser(userName string, password string) User {
	pwHash := utils.GenerateHash(password)
	return User{UserName: userName, PasswordHash: pwHash, IsAdmin: true}
}

type userRepository struct {
	config *config.Config
	db     *gorm.DB
}

//Insert User
func (ur *userRepository) Insert(v interface{}) error {
	return ur.db.Create(v).Error
}

//Update User
func (ur *userRepository) Update(v interface{}) error {
	return ur.db.Save(v).Error
}

//Delete User
func (ur *userRepository) Delete(v interface{}) error {
	return ur.db.Delete(v).Error
}

//List Users
func (ur *userRepository) List() ([]interface{}, error) {
	users := []User{}
	err := ur.db.Find(&users).Error
	result := make([]interface{}, len(users))
	for i := range users {
		result[i] = users[i]
	}
	return result, err
}

//ByID User
func (ur *userRepository) ByID(id int64) (interface{}, error) {
	user := User{}
	err := ur.db.Where("id = ?", id).First(&user).Error
	return user, err
}
