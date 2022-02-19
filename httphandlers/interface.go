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
	GetStudents(w http.ResponseWriter, r *http.Request)

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

	// meetings.go
	GetTimetable(w http.ResponseWriter, r *http.Request)
	NewMeeting(w http.ResponseWriter, r *http.Request)
	PatchMeeting(w http.ResponseWriter, r *http.Request)
	DeleteMeeting(w http.ResponseWriter, r *http.Request)
	GetMeeting(w http.ResponseWriter, r *http.Request)
	GetAbsencesTeacher(w http.ResponseWriter, r *http.Request)
	PatchAbsence(w http.ResponseWriter, r *http.Request)

	// subjects.go
	GetSubjects(w http.ResponseWriter, r *http.Request)
	NewSubject(w http.ResponseWriter, r *http.Request)
	GetSubject(w http.ResponseWriter, r *http.Request)
	AssignUserToSubject(w http.ResponseWriter, r *http.Request)
	RemoveUserFromSubject(w http.ResponseWriter, r *http.Request)
	DeleteSubject(w http.ResponseWriter, r *http.Request)

	// grades.go
	GetGradesForMeeting(w http.ResponseWriter, r *http.Request)
	NewGrade(w http.ResponseWriter, r *http.Request)
	PatchGrade(w http.ResponseWriter, r *http.Request)
	DeleteGrade(w http.ResponseWriter, r *http.Request)

	// homework.go
	NewHomework(w http.ResponseWriter, r *http.Request)
	GetAllHomeworksForSpecificSubject(w http.ResponseWriter, r *http.Request)
	PatchHomeworkForStudent(w http.ResponseWriter, r *http.Request)
}

func NewHTTPInterface(logger *zap.SugaredLogger, db sql.SQL) HTTP {
	return &httpImpl{
		logger: logger,
		db:     db,
	}
}
