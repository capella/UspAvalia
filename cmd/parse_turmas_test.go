package cmd

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// Newer Jupiter pages leave the professor column of the schedule empty and
// list professors in an "Atividades Didáticas" table instead. That table
// contains the words "Horário Variável", which used to make the parser treat
// it as a second schedule table and wipe the real schedule. Fixture: the live
// ENO0302 page (Escola de Enfermagem), captured 2026-09-05.
func TestParseTurmasReadsProfessorsFromAtividadesDidaticas(t *testing.T) {
	f, err := os.Open("testdata/turma_atividades_eno0302.html")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		t.Fatal(err)
	}

	turmas, err := parseTurmas(doc)
	if err != nil {
		t.Fatal(err)
	}
	if len(turmas) != 1 {
		t.Fatalf("expected 1 turma, got %d", len(turmas))
	}
	turma := turmas[0]

	if turma.Codigo != "2026201" || turma.Inicio != "10/08/2026" || turma.Fim != "07/12/2026" ||
		turma.Tipo != "Teórica" {
		t.Fatalf("unexpected turma header: %+v", turma)
	}

	want := []string{"Marcelo José dos Santos (R)", "Simone Graziele Silva Cunha (R)"}
	if !reflect.DeepEqual(turma.Professores, want) {
		t.Fatalf("professores = %q, want %q", turma.Professores, want)
	}

	if len(turma.Horario) != 1 {
		t.Fatalf("expected the Monday slot to survive, got %+v", turma.Horario)
	}
	slot := turma.Horario[0]
	if slot.Dia != "seg" || slot.Inicio != "14:00" || slot.Fim != "17:00" {
		t.Fatalf("unexpected slot: %+v", slot)
	}
	if !reflect.DeepEqual(slot.Professores, want) {
		t.Fatalf("slot professores = %q, want %q", slot.Professores, want)
	}

	if v, ok := turma.Vagas["Obrigatória"]; !ok || v.Vagas != 80 || v.Matriculados != 71 {
		t.Fatalf("unexpected vagas: %+v", turma.Vagas)
	}
}

// Older pages keep the professor in the fourth schedule column and have no
// "Atividades Didáticas" table; they must parse exactly as before.
func TestParseTurmasKeepsScheduleProfessorsOnOldLayout(t *testing.T) {
	page := `<html><body>
<table><tr><td>Código da Turma:</td><td>2026101</td></tr>
<tr><td>Início:</td><td>23/02/2026</td></tr><tr><td>Fim:</td><td>04/07/2026</td></tr>
<tr><td>Tipo da Turma:</td><td>Teórica</td></tr></table>
<table><tr><td>Horário</td><td></td><td></td><td>Prof(a).</td></tr>
<tr><td>seg</td><td>08:00</td><td>09:40</td><td>(R) Alair Pereira do Lago</td></tr>
<tr><td>qua</td><td>08:00</td><td>09:40</td><td>(R) Alair Pereira do Lago</td></tr></table>
<table><tr><td></td><td>Vagas</td><td>Inscritos</td><td>Pendentes</td><td>Matriculados</td></tr>
<tr><td>Obrigatória</td><td>50</td><td>48</td><td>0</td><td>48</td></tr></table>
</body></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(page))
	if err != nil {
		t.Fatal(err)
	}

	turmas, err := parseTurmas(doc)
	if err != nil {
		t.Fatal(err)
	}
	if len(turmas) != 1 {
		t.Fatalf("expected 1 turma, got %d", len(turmas))
	}
	turma := turmas[0]
	if turma.Professores != nil {
		t.Fatalf("no atividades table, but professores = %q", turma.Professores)
	}
	if len(turma.Horario) != 2 {
		t.Fatalf("expected 2 slots, got %+v", turma.Horario)
	}
	for _, slot := range turma.Horario {
		if !reflect.DeepEqual(slot.Professores, []string{"(R) Alair Pereira do Lago"}) {
			t.Fatalf("unexpected slot professors: %+v", slot)
		}
	}
}

func TestParseAtividadesDedupesAndSkipsHeaders(t *testing.T) {
	page := `<table>
<tr><td>Atividades Didáticas</td><td></td><td></td></tr>
<tr><td>Prof(a).</td><td>Tipo de Atividade</td><td>Carga horária (total)</td></tr>
<tr><td> Ana Souza (R) </td><td>Aulas Teóricas</td><td>45</td></tr>
<tr><td> Ana Souza (R) </td><td>Coordenação</td><td>9</td></tr>
<tr><td> Bruno Lima (R) </td><td>Aulas Práticas</td><td>30</td></tr>
<tr><td>Obrigatória</td><td>80</td><td>not-a-number</td></tr>
</table>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(page))
	if err != nil {
		t.Fatal(err)
	}
	got := parseAtividades(doc.Find("table"))
	want := []string{"Ana Souza (R)", "Bruno Lima (R)"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}
