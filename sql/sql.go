package sql

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

type sqlImpl struct {
	db     *sqlx.DB
	logger *zap.SugaredLogger
}

func (db *sqlImpl) Init() {
	db.db.MustExec(schema)
}

type SQL interface {
	Init()

	UpdateTestingResult(testing Testing) error
	InsertTestingResult(testing Testing) error
	GetTestingResults(date string, classId int) ([]Testing, error)
	GetTestingResult(date string, id int) (Testing, error)
	GetLastTestingID() int

	GetUser(id int) (message User, err error)
	InsertUser(user User) (err error)
	GetLastUserID() (id int)
	GetUserByEmail(email string) (message User, err error)
}

func NewSQL(driver string, drivername string, logger *zap.SugaredLogger) (SQL, error) {
	db, err := sqlx.Connect(driver, drivername)
	return &sqlImpl{
		db:     db,
		logger: logger,
	}, err
}