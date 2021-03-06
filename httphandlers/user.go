package httphandlers

import (
	"encoding/json"
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"github.com/gorilla/mux"
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
	"net/http"
	"strconv"
	"time"
)

func (server *httpImpl) Login(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	pass := r.FormValue("pass")
	// Check if password is valid
	user, err := server.db.GetUserByEmail(email)

	if user.Role == "unverified" {
		WriteJSON(w, Response{Data: "You are unverified. You cannot login until the school administrator confirms you.", Success: false}, http.StatusForbidden)
		return
	}

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
	if server.config.BlockRegistrations {
		j := GetAuthorizationJWT(r)
		if j == "" {
			WriteForbiddenJWT(w)
			return
		}
		jwt, err := sql.CheckJWT(j)
		if err != nil {
			WriteForbiddenJWT(w)
			return
		}
		if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		} else {
			WriteForbiddenJWT(w)
			return
		}
	}
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

	var role = "unverified"

	isAdmin := !server.db.CheckIfAdminIsCreated()
	if isAdmin {
		role = "admin"
	}

	user := sql.User{
		ID:                     server.db.GetLastUserID(),
		Email:                  email,
		Password:               password,
		Role:                   role,
		Name:                   name,
		BirthCertificateNumber: "",
		Birthday:               "",
		CityOfBirth:            "",
		CountryOfBirth:         "",
		Users:                  "[]",
	}

	err = server.db.InsertUser(user)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to commit new user to database", Success: false}, http.StatusInternalServerError)
		return
	}

	WriteJSON(w, Response{Data: "Success", Success: true}, http.StatusCreated)
}

func (server *httpImpl) PatchUser(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		userId, err := strconv.Atoi(mux.Vars(r)["user_id"])
		if err != nil {
			WriteForbiddenJWT(w)
			return
		}

		user, err := server.db.GetUser(userId)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Data: "Failed to retrieve used from database", Success: false}, http.StatusInternalServerError)
			return
		}
		if r.FormValue("birthday") != "" {
			user.Birthday = r.FormValue("birthday")
		}
		if r.FormValue("country_of_birth") != "" {
			user.CountryOfBirth = r.FormValue("country_of_birth")
		}
		if r.FormValue("city_of_birth") != "" {
			user.CityOfBirth = r.FormValue("city_of_birth")
		}
		if r.FormValue("email") != "" {
			user.Email = r.FormValue("email")
		}
		if r.FormValue("birth_certificate_number") != "" {
			user.BirthCertificateNumber = r.FormValue("birth_certificate_number")
		}
		if r.FormValue("name") != "" {
			user.Name = r.FormValue("name")
		}
		if r.FormValue("is_passing") != "" {
			isPassing, err := strconv.ParseBool(r.FormValue("is_passing"))
			if err != nil {
				WriteBadRequest(w)
				return
			}
			user.IsPassing = isPassing
		}
		err = server.db.UpdateUser(user)
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Data: "Failed to update user", Success: false}, http.StatusInternalServerError)
			return
		}
		WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
		return
	}
}

func (server *httpImpl) HasClass(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if jwt["role"] == "admin" || jwt["role"] == "teacher" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
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
	} else {
		WriteForbiddenJWT(w)
		return
	}
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

	var birthCertNum = ""
	if jwt["role"] == "admin" {
		birthCertNum = user.BirthCertificateNumber
	}

	ujson := UserJSON{
		Name:                   user.Name,
		ID:                     user.ID,
		Email:                  user.Email,
		Role:                   user.Role,
		BirthCertificateNumber: birthCertNum,
		Birthday:               user.Birthday,
		CityOfBirth:            user.CityOfBirth,
		CountryOfBirth:         user.CountryOfBirth,
		IsPassing:              user.IsPassing,
	}
	WriteJSON(w, Response{Data: ujson, Success: true}, http.StatusOK)
}

func (server *httpImpl) GetAbsencesUser(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
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
				WriteJSON(w, Response{Data: "Could not fetch classes", Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
			var valid = false
			for i := 0; i < len(classes); i++ {
				class := classes[i]
				var users []int
				err := json.Unmarshal([]byte(class.Students), &users)
				if err != nil {
					WriteJSON(w, Response{Data: "Could not unmarshal students", Error: err.Error(), Success: false}, http.StatusInternalServerError)
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
		} else if jwt["role"] == "parent" {
			if !server.config.ParentViewAbsences {
				WriteForbiddenJWT(w)
				return
			}
			parent, err := server.db.GetUser(teacherId)
			if err != nil {
				WriteJSON(w, Response{Data: "Could not fetch parent", Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
			var students []int
			err = json.Unmarshal([]byte(parent.Users), &students)
			if err != nil {
				WriteJSON(w, Response{Data: "Could not unmarshal students", Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
			if !contains(students, studentId) {
				WriteForbiddenJWT(w)
				return
			}
		}
	}
	var absenceJson = make([]Absence, 0)
	absences, err := server.db.GetAbsencesForUser(studentId)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			WriteJSON(w, Response{Data: absenceJson, Error: err.Error(), Success: true}, http.StatusOK)
			return
		}
		WriteJSON(w, Response{Data: "Could not fetch absences", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}
	for i := 0; i < len(absences); i++ {
		absence := absences[i]
		teacher, err := server.db.GetUser(absence.TeacherID)
		if err != nil {
			WriteJSON(w, Response{Data: "Could not fetch teacher", Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		user, err := server.db.GetUser(absence.UserID)
		if err != nil {
			WriteJSON(w, Response{Data: "Could not fetch user", Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		meeting, err := server.db.GetMeeting(absence.MeetingID)
		if err != nil {
			WriteJSON(w, Response{Data: "Could not fetch meeting", Error: err.Error(), Success: false}, http.StatusInternalServerError)
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

	var userId = make([]int, 0)
	var isTeacher = false
	if jwt["role"] == "admin" || jwt["role"] == "teacher" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		uid := r.URL.Query().Get("id")
		if uid == "" {
			u, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
			userId = append(userId, u)
			isTeacher = true
		} else {
			u, err := strconv.Atoi(uid)
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
			userId = append(userId, u)
		}
	} else if jwt["role"] == "parent" {
		u, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		user, err := server.db.GetUser(u)
		if err != nil {
			return
		}
		err = json.Unmarshal([]byte(user.Users), &userId)
		if err != nil {
			return
		}
	} else {
		u, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		userId = append(userId, u)
	}

	classes, err := server.db.GetClasses()
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	var myclasses = make([]sql.Class, 0)
	var myClassesInt = make([]int, 0)

	for i := 0; i < len(classes); i++ {
		class := classes[i]
		if isTeacher {
			for n := 0; n < len(userId); n++ {
				if class.Teacher == userId[n] {
					myclasses = append(myclasses, class)
				}
			}
		} else {
			var students []int
			err := json.Unmarshal([]byte(class.Students), &students)
			if err != nil {
				WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
				return
			}
			server.logger.Debug(students, userId)
			for n := 0; n < len(students); n++ {
				for l := 0; l < len(userId); l++ {
					if students[n] == userId[l] && !contains(myClassesInt, students[n]) {
						user, err := server.db.GetUser(students[n])
						if err != nil {
							return
						}
						var className = class.Name
						if jwt["role"] == "parent" {
							class.Name = fmt.Sprintf("%s - %s", class.Name, user.Name)
						}
						myclasses = append(myclasses, class)
						myClassesInt = append(myClassesInt, students[n])
						if jwt["role"] == "parent" {
							class.Name = className
						}
						break
					}
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
	if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" {
		students, err := server.db.GetStudents()
		if err != nil {
			return
		}
		var studentsJson = make([]UserJSON, 0)
		for i := 0; i < len(students); i++ {
			student := students[i]
			studentsJson = append(studentsJson, UserJSON{
				Name:                   student.Name,
				ID:                     student.ID,
				Email:                  student.Email,
				Role:                   student.Role,
				BirthCertificateNumber: student.BirthCertificateNumber,
				Birthday:               student.Birthday,
				CityOfBirth:            student.CityOfBirth,
				CountryOfBirth:         student.CountryOfBirth,
			})
		}
		WriteJSON(w, Response{Data: studentsJson, Success: true}, http.StatusOK)
	} else {
		WriteForbiddenJWT(w)
	}
}

func (server *httpImpl) HasBirthday(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	userId, err := strconv.Atoi(fmt.Sprint(jwt["user_id"]))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	user, err := server.db.GetUser(userId)
	if err != nil {
		return
	}
	currentTime := time.Now()
	birthday, err := time.Parse("2006-01-02", user.Birthday)
	if err != nil {
		WriteJSON(w, Response{Error: err.Error(), Data: "Failed to parse date", Success: false}, http.StatusInternalServerError)
		return
	}
	if currentTime.Before(birthday) {
		WriteJSON(w, Response{Data: "Invalid birthday", Success: false}, http.StatusConflict)
		return
	}
	_, tm, td := currentTime.Date()
	_, bm, bd := birthday.Date()
	WriteJSON(w, Response{Data: tm-bm == 0 && td-bd == 0, Success: true}, http.StatusOK)
}

func (server *httpImpl) CertificateOfSchooling(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}

	if jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant" || jwt["role"] == "school psychologist" {
		userId, err := strconv.Atoi(mux.Vars(r)["user_id"])
		if err != nil {
			WriteBadRequest(w)
			return
		}

		student, err := server.db.GetUser(userId)
		if err != nil {
			return
		}
		if student.Role != "student" {
			WriteForbiddenJWT(w)
			return
		}

		classes, err := server.db.GetClasses()
		if err != nil {
			return
		}

		var classId = -1

		for i := 0; i < len(classes); i++ {
			class := classes[i]
			var students []int
			err := json.Unmarshal([]byte(class.Students), &students)
			if err != nil {
				return
			}
			if contains(students, userId) {
				classId = class.ID
				break
			}
		}

		if classId == -1 {
			return
		}

		class, err := server.db.GetClass(classId)
		if err != nil {
			return
		}

		var students []int
		err = json.Unmarshal([]byte(class.Students), &students)
		if err != nil {
			return
		}
		if !contains(students, userId) {
			WriteForbiddenJWT(w)
			return
		}

		m := pdf.NewMaroto(consts.Portrait, consts.A4)

		m.AddUTF8Font("OpenSans", consts.Normal, "fonts/opensans.ttf")
		m.SetDefaultFontFamily("OpenSans")

		m.Row(40, func() {

			m.Col(3, func() {
				_ = m.FileImage("icons/school_logo.png", props.Rect{
					Center:  true,
					Percent: 80,
				})
			})

			m.ColSpace(1)

			m.Col(4, func() {
				m.Text("Potrdilo o ??olanju", props.Text{
					Top:         12,
					Size:        25,
					Extrapolate: true,
				})
				m.Text("MeetPlan sistem", props.Text{
					Top:         23,
					Size:        13,
					Extrapolate: true,
				})
			})
			m.ColSpace(1)

			m.Col(3, func() {
				_ = m.FileImage("icons/country_coat_of_arms_black.png", props.Rect{
					Center:  true,
					Percent: 80,
				})
			})
		})

		m.Line(10)

		m.Row(40, func() {
			m.Text(fmt.Sprintf(
				"U??enec %s, rojen %s, %s, %s, v ??olskem letu %s",
				student.Name, student.Birthday, student.CityOfBirth,
				student.CountryOfBirth, class.ClassYear,
			), props.Text{
				Top:         12,
				Size:        11,
				Extrapolate: true,
			})
			m.Text(fmt.Sprintf("obiskuje %s razred ??ole %s.",
				class.Name, server.config.SchoolName,
			), props.Text{
				Top:         16,
				Size:        11,
				Extrapolate: true,
			})
		})

		principal, err := server.db.GetPrincipal()
		if err != nil {
			return
		}

		m.Row(40, func() {
			m.ColSpace(1)
			m.Col(6, func() {
				m.Text("_________________________", props.Text{
					Top:  14,
					Size: 15,
				})
				m.Text(principal.Name, props.Text{
					Top:  14,
					Size: 15,
				})
				m.Text("digitalni podpis ravnatelja", props.Text{Top: 20, Size: 9})
			})
			m.Col(3, func() {
				m.Text("_________________________", props.Text{
					Top:  14,
					Size: 15,
				})
				m.Text("podpis ravnatelja", props.Text{Top: 20, Size: 9})
			})
		})

		output, err := m.Output()
		if err != nil {
			WriteJSON(w, Response{Error: err.Error(), Success: false}, http.StatusInternalServerError)
			return
		}
		w.Write(output.Bytes())
	} else {
		WriteForbiddenJWT(w)
		return
	}
}
