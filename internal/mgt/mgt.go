package mgt

import (
	"crypto/sha1"
	"fmt"

	"github.com/rriverak/gogo/internal/utils"
	"github.com/sirupsen/logrus"
)

//Logger for mgt
var Logger *logrus.Logger

//Repository Interface for SQL
type Repository interface {
	Insert(v interface{}) error
	Update(v interface{}) error
	Delete(v interface{}) error
	List() ([]interface{}, error)
	ByID(id int64) (interface{}, error)
}

//newSHA1 Generate a new ID
func newSHA1() string {
	randString := utils.RandSeq(5)

	hash := sha1.New()
	hash.Write([]byte(randString))
	bs := hash.Sum(nil)

	return fmt.Sprintf("%x", bs)
}
