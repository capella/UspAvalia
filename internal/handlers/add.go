package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
	"uspavalia/internal/models"
	"uspavalia/internal/services"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const addYearsBack = 3

// addPageForm is what the add page renders: the current selection (so the
// form survives a validation error) plus the semester choices.
type addPageForm struct {
	UnitID          uint
	UnitLabel       string
	DisciplineID    uint
	DisciplineLabel string
	DisciplineFixed bool
	ProfessorID     uint
	ProfessorLabel  string
	ProfessorFixed  bool
	Year            int
	Half            int
	Years           []int
	Error           string
}

// handleAddPage serves GET /adicionar: the form to record that an existing
// professor taught an existing discipline in a given semester. Discipline and
// professor are chosen from the catalog through typeaheads; nothing is
// created from free text. ?disciplina_id= or ?professor_id= pre-select and
// lock a field, the way the old add2 page did.
func (s *Server) handleAddPage(w http.ResponseWriter, r *http.Request) {
	form := s.newAddForm()

	if id, err := strconv.ParseUint(r.URL.Query().Get("disciplina_id"), 10, 32); err == nil {
		var d models.Discipline
		if s.db.First(&d, id).Error == nil {
			form.DisciplineID, form.DisciplineLabel, form.DisciplineFixed = d.ID, d.Code+" - "+d.Name, true
			form.UnitID = d.UnitID
		}
	}
	if id, err := strconv.ParseUint(r.URL.Query().Get("professor_id"), 10, 32); err == nil {
		var p models.Professor
		if s.db.First(&p, id).Error == nil {
			form.ProfessorID, form.ProfessorLabel, form.ProfessorFixed = p.ID, p.Name, true
			if form.UnitID == 0 {
				form.UnitID = p.UnitID
			}
		}
	}
	if form.UnitID > 0 {
		var u models.Unit
		if s.db.First(&u, form.UnitID).Error == nil {
			form.UnitLabel = u.Name
		}
	}

	s.renderAddPage(w, r, form)
}

// handleAddSubmit serves POST /adicionar. It accepts only IDs of existing
// rows, records the discipline/professor pair and the semester's offering,
// then redirects to the pair's page.
func (s *Server) handleAddSubmit(w http.ResponseWriter, r *http.Request) {
	form := s.newAddForm()
	form.UnitID = parseID(r.FormValue("unit_id"))
	form.UnitLabel = r.FormValue("unit_label")
	form.DisciplineID = parseID(r.FormValue("discipline_id"))
	form.DisciplineLabel = r.FormValue("discipline_label")
	form.ProfessorID = parseID(r.FormValue("professor_id"))
	form.ProfessorLabel = r.FormValue("professor_label")
	if y, err := strconv.Atoi(r.FormValue("year")); err == nil {
		form.Year = y
	}
	if h, err := strconv.Atoi(r.FormValue("half")); err == nil {
		form.Half = h
	}

	var discipline models.Discipline
	var professor models.Professor
	if form.DisciplineID == 0 || s.db.First(&discipline, form.DisciplineID).Error != nil {
		form.Error = "Selecione uma disciplina da lista de sugestões."
		form.DisciplineID, form.DisciplineLabel = 0, ""
		s.renderAddPage(w, r, form)
		return
	}
	if form.ProfessorID == 0 || s.db.First(&professor, form.ProfessorID).Error != nil {
		form.Error = "Selecione um professor da lista de sugestões."
		form.ProfessorID, form.ProfessorLabel = 0, ""
		s.renderAddPage(w, r, form)
		return
	}

	sem := services.Semester{Year: form.Year, Half: form.Half}
	if !sem.Valid(time.Now()) {
		form.Error = "Selecione um semestre válido."
		s.renderAddPage(w, r, form)
		return
	}

	var classProfessor models.ClassProfessor
	var createdPair, createdOffering bool
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var err error
		classProfessor, createdPair, err = services.GetOrCreateClassProfessor(tx, discipline.ID, professor.ID)
		if err != nil {
			return err
		}
		_, createdOffering, err = services.EnsureUserOffering(tx, discipline, professor, sem)
		return err
	})
	if err != nil {
		logrus.Printf("Error adding class professor: %v", err)
		form.Error = "Erro ao salvar. Tente novamente."
		s.renderAddPage(w, r, form)
		return
	}

	if user := s.getCurrentUser(r); user != nil {
		logrus.Printf("User %d added %s / %s for %s (new pair: %t, new offering: %t)",
			user.ID, discipline.Code, professor.Name, sem, createdPair, createdOffering)
	}

	target := fmt.Sprintf("/ver/%d", classProfessor.ID)
	if !createdPair && !createdOffering {
		target += "?existente=1"
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

func (s *Server) newAddForm() addPageForm {
	current := services.CurrentSemester(time.Now())
	years := make([]int, 0, addYearsBack+1)
	for y := current.Year; y >= current.Year-addYearsBack; y-- {
		years = append(years, y)
	}
	return addPageForm{Year: current.Year, Half: current.Half, Years: years}
}

func (s *Server) renderAddPage(w http.ResponseWriter, r *http.Request, form addPageForm) {
	if form.Error != "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
	s.renderTemplate(w, r, "adicionar", PageData{
		User: s.getCurrentUser(r),
		Data: form,
	})
}

func parseID(value string) uint {
	id, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0
	}
	return uint(id)
}
