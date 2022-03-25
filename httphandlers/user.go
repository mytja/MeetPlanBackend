package httphandlers

import (
	"encoding/json"
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (server *httpImpl) Login(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	pass := r.FormValue("pass")
	// Check if password is valid
	user, err := server.db.GetUserByEmail(email)
	hashCorrect := sql.CheckHash(pass, user.Password)
	if !hashCorrect {
		WriteJSON(w, Response{Data: "Hashes don't match...", Success: false}, http.StatusForbidden)
		return
	}

	// Extract JWT
	jwt, err := sql.GetJWTFromUserPass(email, user.Role, user.ID)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, Response{Data: jwt, Success: true}, http.StatusOK)
}

func (server *httpImpl) NewUser(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	pass := r.FormValue("pass")
	name := r.FormValue("name")
	if email == "" || pass == "" || name == "" {
		WriteJSON(w, Response{Data: "Bad Request. A parameter isn't provided", Success: false}, http.StatusBadRequest)
		return
	}
	// Check if user is already in DB
	var userCreated = true
	_, err := server.db.GetUserByEmail(email)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			userCreated = false
		} else {
			WriteJSON(w, Response{Error: err.Error(), Data: "Could not retrieve user from database", Success: false}, http.StatusInternalServerError)
			return
		}
	}
	if userCreated == true {
		WriteJSON(w, Response{Data: "User is already in database", Success: false}, http.StatusUnprocessableEntity)
		return
	}

	password, err := sql.HashPassword(pass)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to hash your password", Success: false}, http.StatusInternalServerError)
		return
	}

	var role = "student"

	isAdmin := !server.db.CheckIfAdminIsCreated()
	if isAdmin {
		role = "admin"
	}

	user := sql.User{ID: server.db.GetLastUserID(), Email: email, Password: password, Role: role, Name: name}

	err = server.db.InsertUser(user)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to commit new user to database", Success: false}, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, Response{Data: "Success", Success: true}, http.StatusCreated)
}

func (server *httpImpl) HasClass(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "student" {
		WriteForbiddenJWT(w)
		return
	}
	userId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	classes, err := server.db.GetClasses()
	if err != nil {
		return
	}
	var hasClass = false
	for i := 0; i < len(classes); i++ {
		if classes[i].Teacher == userId {
			hasClass = true
			break
		}
	}
	WriteJSON(w, Response{Data: hasClass, Success: true}, http.StatusOK)
}

func (server *httpImpl) GetUserData(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		return
	}
	var userId int
	if jwt["role"] == "student" {
		userId, err = strconv.Atoi(fmt.Sprint(jwt["user_id"]))
		if err != nil {
			WriteBadRequest(w)
			return
		}
	} else {
		userId, err = strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			WriteBadRequest(w)
			return
		}
	}
	user, err := server.db.GetUser(userId)
	if err != nil {
		return
	}
	ujson := UserJSON{
		Name:  user.Name,
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	}
	WriteJSON(w, Response{Data: ujson, Success: true}, http.StatusOK)
}

func (server *httpImpl) GetAbsencesUser(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		return
	}
	var studentId int
	if jwt["role"] == "student" {
		studentId, err = strconv.Atoi(fmt.Sprint(jwt["user_id"]))
		if err != nil {
			WriteBadRequest(w)
			return
		}
	} else {
		studentId, err = strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			WriteBadRequest(w)
			return
		}
		teacherId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
		if err != nil {
			WriteBadRequest(w)
			return
		}
		if jwt["role"] == "teacher" {
			classes, err := server.db.GetClasses()
			if err != nil {
				return
			}
			var valid = false
			for i := 0; i < len(classes); i++ {
				class := classes[i]
				var users []int
				err := json.Unmarshal([]byte(class.Students), &users)
				if err != nil {
					return
				}
				for j := 0; j < len(users); j++ {
					if users[j] == studentId && class.Teacher == teacherId {
						valid = true
						break
					}
				}
				if valid {
					break
				}
			}
			if !valid {
				WriteForbiddenJWT(w)
				return
			}
		}
	}
	absences, err := server.db.GetAbsencesForUser(studentId)
	if err != nil {
		return
	}
	var absenceJson = make([]Absence, 0)
	for i := 0; i < len(absences); i++ {
		absence := absences[i]
		teacher, err := server.db.GetUser(absence.TeacherID)
		if err != nil {
			return
		}
		user, err := server.db.GetUser(absence.UserID)
		if err != nil {
			return
		}
		meeting, err := server.db.GetMeeting(absence.MeetingID)
		if err != nil {
			return
		}
		if absence.AbsenceType == "ABSENT" || absence.AbsenceType == "LATE" {
			absenceJson = append(absenceJson, Absence{
				Absence:     absence,
				TeacherName: teacher.Name,
				UserName:    user.Name,
				MeetingName: meeting.MeetingName,
			})
		}
	}
	WriteJSON(w, Response{Data: absenceJson, Success: true}, http.StatusOK)
}

func (server *httpImpl) GetAllClasses(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		return
	}

	var userId int
	var isTeacher = false
	if jwt["role"] == "admin" || jwt["role"] == "teacher" {
		uid := r.URL.Query().Get("id")
		if uid == "" {
			userId, err = strconv.Atoi(fmt.Sprint(jwt["user_id"]))
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
			isTeacher = true
		} else {
			userId, err = strconv.Atoi(uid)
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
		}
	} else {
		userId, err = strconv.Atoi(fmt.Sprint(jwt["user_id"]))
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
	}

	classes, err := server.db.GetClasses()
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	var myclasses = make([]sql.Class, 0)

	for i := 0; i < len(classes); i++ {
		class := classes[i]
		if isTeacher {
			if class.Teacher == userId {
				myclasses = append(myclasses, class)
			}
		} else {
			var students []int
			err := json.Unmarshal([]byte(class.Students), &students)
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
			for n := 0; n < len(students); n++ {
				if students[n] == userId {
					myclasses = append(myclasses, class)
					break
				}
			}
		}
	}
	WriteJSON(w, Response{Data: myclasses, Success: true}, http.StatusOK)
}

func (server *httpImpl) GetStudents(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "admin" {
		students, err := server.db.GetStudents()
		if err != nil {
			return
		}
		WriteJSON(w, Response{Data: students, Success: true}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
	}
}
