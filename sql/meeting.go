package sql

type Meeting struct {
	ID             int    `db:"id"`
	MeetingName    string `db:"meeting_name"`
	TeacherID      int    `db:"teacher_id"`
	SubjectID      int    `db:"subject_id"`
	Hour           int    `db:"hour"`
	Date           string `db:"date"`
	IsMandatory    bool   `db:"is_mandatory"`
	URL            string `db:"url"`
	Details        string `db:"details"`
	IsSubstitution bool   `db:"is_substitution"`
	// Ocenjevanje
	IsGrading           bool `db:"is_grading"`
	IsWrittenAssessment bool `db:"is_written_assessment"`
	// Preverjanje znanja
	IsTest bool `db:"is_test"`
}

func (db *sqlImpl) GetMeeting(id int) (meeting Meeting, err error) {
	err = db.db.Get(&meeting, "SELECT * FROM meetings WHERE id=$1", id)
	return meeting, err
}

func (db *sqlImpl) GetMeetingsOnSpecificTime(date string, hour int) (meetings []Meeting, err error) {
	err = db.db.Select(&meetings, "SELECT * FROM meetings WHERE date=$1 AND hour=$2 ORDER BY id ASC", date, hour)
	return meetings, err
}

func (db *sqlImpl) GetMeetingsOnSpecificDate(date string) (meetings []Meeting, err error) {
	err = db.db.Select(&meetings, "SELECT * FROM meetings WHERE date=$1 ORDER BY id ASC", date)
	return meetings, err
}

func (db *sqlImpl) GetMeetingsForTeacherOnSpecificDate(teacherId int, date string) (meetings []Meeting, err error) {
	err = db.db.Select(&meetings, "SELECT * FROM meetings WHERE date=$1 AND teacher_id=$2 ORDER BY id ASC", date, teacherId)
	return meetings, err
}

func (db *sqlImpl) GetMeetingsForSubject(subjectId int) (meetings []Meeting, err error) {
	err = db.db.Select(&meetings, "SELECT * FROM meetings WHERE subject_id=$1 ORDER BY id ASC", subjectId)
	return meetings, err
}

func (db *sqlImpl) InsertMeeting(meeting Meeting) (err error) {
	i := `
	INSERT INTO meetings (id, meeting_name, teacher_id, subject_id, hour, date, is_mandatory, url, details, is_grading, is_written_assessment, is_test, is_substitution)
		VALUES (:id, :meeting_name, :teacher_id, :subject_id, :hour, :date, :is_mandatory, :url, :details, :is_grading, :is_written_assessment, :is_test, :is_substitution)
	`
	_, err = db.db.NamedExec(
		i,
		meeting)
	return err
}

func (db *sqlImpl) UpdateMeeting(meeting Meeting) error {
	i := `
	UPDATE meetings SET meeting_name=:meeting_name, teacher_id=:teacher_id,
	                    subject_id=:subject_id, hour=:hour, date=:date,
	                    is_mandatory=:is_mandatory, url=:url, details=:details,
	                    is_grading=:is_grading, is_written_assessment=:is_written_assessment,
	                    is_test=:is_test, is_substitution=:is_substitution WHERE id=:id
	`
	_, err := db.db.NamedExec(
		i,
		meeting)
	return err
}

func (db *sqlImpl) GetLastMeetingID() (id int) {
	err := db.db.Get(&id, "SELECT id FROM meetings WHERE id = (SELECT MAX(id) FROM meetings)")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0
		}
		db.logger.Info(err)
		return -1
	}
	return id + 1
}

func (db *sqlImpl) GetMeetings() (meetings []Meeting, err error) {
	err = db.db.Select(&meetings, "SELECT * FROM meetings ORDER BY id ASC")
	return meetings, err
}

func (db *sqlImpl) GetMeetingsForSubjectWithIDLower(id int, subjectId int) (meetings []Meeting, err error) {
	err = db.db.Select(&meetings, "SELECT * FROM meetings WHERE id<=$1 AND subject_id=$2", id, subjectId)
	return meetings, err
}

func (db *sqlImpl) DeleteMeeting(ID int) error {
	_, err := db.db.Exec("DELETE FROM meetings WHERE id=$1", ID)
	return err
}

func (db *sqlImpl) DeleteMeetingsForTeacher(ID int) error {
	_, err := db.db.Exec("DELETE FROM meetings WHERE teacher_id=$1", ID)
	return err
}

func (db *sqlImpl) DeleteMeetingsForSubject(ID int) error {
	_, err := db.db.Exec("DELETE FROM meetings WHERE subject_id=$1", ID)
	return err
}
