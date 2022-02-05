package httphandlers

import (
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"go.uber.org/zap"
	"net/http"
)

type httpImpl struct {
	logger *zap.SugaredLogger
	db     sql.SQL
}

type HTTP interface {
	// user.go
	Login(w http.ResponseWriter, r *http.Request)
	NewUser(w http.ResponseWriter, r *http.Request)
	GetAllClasses(w http.ResponseWriter, r *http.Request)

	// testing.go
	GetSelfTestingTeacher(w http.ResponseWriter, r *http.Request)
	PatchSelfTesting(w http.ResponseWriter, r *http.Request)
	GetPDFSelfTestingReportStudent(w http.ResponseWriter, r *http.Request)
	GetTestingResults(w http.ResponseWriter, r *http.Request)

	// class.go
	NewClass(w http.ResponseWriter, r *http.Request)
	GetClasses(w http.ResponseWriter, r *http.Request)
	AssignUserToClass(w http.ResponseWriter, r *http.Request)
	RemoveUserFromClass(w http.ResponseWriter, r *http.Request)
	GetClass(w http.ResponseWriter, r *http.Request)
	DeleteClass(w http.ResponseWriter, r *http.Request)

	// admin.go
	GetAllUsers(w http.ResponseWriter, r *http.Request)
	ChangeRole(w http.ResponseWriter, r *http.Request)
	DeleteUser(w http.ResponseWriter, r *http.Request)
	GetTeachers(w http.ResponseWriter, r *http.Request)
}

func NewHTTPInterface(logger *zap.SugaredLogger, db sql.SQL) HTTP {
	return &httpImpl{
		logger: logger,
		db:     db,
	}
}
