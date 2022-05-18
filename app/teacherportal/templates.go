package teacherportal

import (
	"html/template"
)

var rootTemplate *template.Template

func ImportTemplates() (err error) {
	//studentsPath, _ := filepath.Abs("app/teacherportal/students.gohtml")
	//studentPath, _ := filepath.Abs("app/teacherportal/student.gohtml")
	rootTemplate, err = template.ParseFiles(
		"./students.gohtml",
		"./student.gohtml",

		//"student.gohtml",
	)

	if err != nil {
		return err
	}
	return nil
}
