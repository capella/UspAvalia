package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"uspavalia/internal/config"
	"uspavalia/internal/database"
	"uspavalia/internal/models"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var fetchDisciplinesCMD = &cobra.Command{
	Use:   "fetch-disciplines [output-directory]",
	Short: "Fetch USP discipline information from Jupiter Web",
	Long:  `Fetches detailed discipline information from USP Jupiter Web including schedules, professors, and enrollment data.`,
	Args:  cobra.MaximumNArgs(1),
	Run:   runFetchDisciplines,
}

type DisciplineInfo struct {
	Codigo           string      `json:"codigo"`
	Nome             string      `json:"nome"`
	Unidade          string      `json:"unidade"`
	Departamento     string      `json:"departamento"`
	Campus           string      `json:"campus"`
	Objetivos        string      `json:"objetivos"`
	ProgramaResumido string      `json:"programa_resumido"`
	CreditosAula     int         `json:"creditos_aula"`
	CreditosTrabalho int         `json:"creditos_trabalho"`
	Turmas           []TurmaInfo `json:"turmas"`
}

type TurmaInfo struct {
	Codigo        string               `json:"codigo"`
	CodigoTeorica string               `json:"codigo_teorica"`
	Inicio        string               `json:"inicio"`
	Fim           string               `json:"fim"`
	Tipo          string               `json:"tipo"`
	Observacoes   string               `json:"observacoes"`
	Horario       []HorarioInfo        `json:"horario"`
	Vagas         map[string]VagasInfo `json:"vagas"`
	// Professores lists the people in the "Atividades Didáticas" table, which
	// newer Jupiter pages use instead of the professor column of the schedule.
	Professores []string `json:"professores,omitempty"`
}

type HorarioInfo struct {
	Dia         string   `json:"dia"`
	Inicio      string   `json:"inicio"`
	Fim         string   `json:"fim"`
	Professores []string `json:"professores"`
}

type VagasInfo struct {
	Vagas        int                  `json:"vagas"`
	Inscritos    int                  `json:"inscritos"`
	Pendentes    int                  `json:"pendentes"`
	Matriculados int                  `json:"matriculados"`
	Grupos       map[string]VagasInfo `json:"grupos"`
}

var (
	fetchDisciplinesUnits       []string
	fetchDisciplinesTimeout     int
	fetchDisciplinesOut         string
	fetchDisciplinesConcurrency int
	fetchDisciplinesNoGzip      bool
	fetchDisciplinesStore       bool
)

func init() {
	rootCmd.AddCommand(fetchDisciplinesCMD)

	fetchDisciplinesCMD.Flags().
		StringSliceVarP(&fetchDisciplinesUnits, "units", "u", nil, "Fetch only these unit codes")
	fetchDisciplinesCMD.Flags().
		IntVarP(&fetchDisciplinesTimeout, "timeout", "t", 120, "HTTP request timeout in seconds")
	fetchDisciplinesCMD.Flags().
		StringVarP(&fetchDisciplinesOut, "output", "o", "db.json", "Output filename")
	fetchDisciplinesCMD.Flags().
		IntVarP(&fetchDisciplinesConcurrency, "concurrency", "c", 100, "Number of concurrent requests")
	fetchDisciplinesCMD.Flags().
		BoolVar(&fetchDisciplinesNoGzip, "no-gzip", false, "Don't create gzipped output")
	fetchDisciplinesCMD.Flags().
		BoolVar(&fetchDisciplinesStore, "store", false, "Store disciplines in database")
}

func runFetchDisciplines(cmd *cobra.Command, args []string) {
	startTime := time.Now()

	// When storing, open the database first so the run is recorded even if
	// Jupiter Web is unreachable. Preview runs leave the database alone.
	var db *gorm.DB
	var recorder *scrapeRecorder
	if fetchDisciplinesStore {
		cfg := config.Load()
		var err error
		db, err = database.Initialize(cfg)
		if err != nil {
			fmt.Printf("Error: Failed to initialize database: %v\n", err)
			os.Exit(1)
		}
		recorder, err = startScrapeRun(db, "fetch-disciplines")
		if err != nil {
			fmt.Printf("Warning: %v\n", err)
		}
	}

	fail := func(format string, args ...interface{}) {
		err := fmt.Errorf(format, args...)
		fmt.Println("Error:", err)
		recorder.finish(stats, err)
		os.Exit(1)
	}

	fmt.Println("- Obtaining list of all teaching units -")

	// Get teaching units (reuse from parse_courses.go)
	units, err := getTeachingUnits()
	if err != nil {
		fail("getting teaching units: %v", err)
	}
	if len(units) == 0 {
		fail("no teaching units found on Jupiter Web (page layout changed?)")
	}

	fmt.Printf("- %d teaching units found -\n", len(units))

	// Filter units if specified
	var targetUnits []string
	if len(fetchDisciplinesUnits) > 0 {
		targetUnits = fetchDisciplinesUnits
	} else {
		for _, unit := range units {
			targetUnits = append(targetUnits, unit.Code)
		}
	}

	fmt.Println("- Starting discipline processing -")

	// Process disciplines concurrently
	disciplines, err := processDisciplinesConcurrently(targetUnits, units)
	if err != nil {
		fail("processing disciplines: %v", err)
	}

	if fetchDisciplinesStore {
		// Store in database
		fmt.Println("\nStoring disciplines in database...")

		storedDisciplines := 0
		storedProfessors := 0
		storedAssociations := 0

		for _, disc := range disciplines {
			// Get or create unit by name
			unit, err := getOrCreateUnit(db, disc.Unidade)
			if err != nil {
				fmt.Printf("Warning: Failed to get or create unit %s: %v\n", disc.Unidade, err)
				stats.storeErrors.Add(1)
				continue
			}

			// Create or update discipline by code
			dbDiscipline, err := getOrCreateDiscipline(db, models.Discipline{
				Code:   disc.Codigo,
				Name:   disc.Nome,
				UnitID: unit.ID,
			})
			if err != nil {
				fmt.Printf("Warning: Failed to store discipline %s: %v\n", disc.Codigo, err)
				stats.storeErrors.Add(1)
				continue
			}
			storedDisciplines++
			stats.disciplinesStored.Add(1)

			// Update discipline with MatrUSP fields
			if err := db.Model(&dbDiscipline).Updates(map[string]interface{}{
				"Department":   disc.Departamento,
				"Campus":       disc.Campus,
				"CreditsClass": disc.CreditosAula,
				"CreditsWork":  disc.CreditosTrabalho,
				"Objectives":   disc.Objetivos,
				"Summary":      disc.ProgramaResumido,
			}).Error; err != nil {
				fmt.Printf("Warning: Failed to update discipline %s: %v\n", disc.Codigo, err)
				stats.storeErrors.Add(1)
			}

			// Store class offerings (turmas)
			for _, turma := range disc.Turmas {
				// Serialize schedules and vacancies to JSON
				schedulesJSON, _ := json.Marshal(turma.Horario)
				vacanciesJSON, _ := json.Marshal(turma.Vagas)

				offering := models.ClassOffering{
					DisciplineID:    dbDiscipline.ID,
					Code:            turma.Codigo,
					TheoreticalCode: turma.CodigoTeorica,
					StartDate:       turma.Inicio,
					EndDate:         turma.Fim,
					Type:            turma.Tipo,
					Notes:           turma.Observacoes,
					Schedules:       string(schedulesJSON),
					Vacancies:       string(vacanciesJSON),
				}

				// Check if offering exists
				var existingOffering models.ClassOffering
				result := db.Where(
					"discipline_id = ? AND code = ? AND start_date = ?",
					dbDiscipline.ID, turma.Codigo, turma.Inicio,
				).First(&existingOffering)

				if result.Error != nil {
					// Offering doesn't exist, create it
					if err := db.Create(&offering).Error; err != nil {
						fmt.Printf(
							"Warning: Failed to create class offering %s-%s: %v\n",
							disc.Codigo,
							turma.Codigo,
							err,
						)
						stats.storeErrors.Add(1)
					} else {
						stats.offeringsStored.Add(1)
					}
				} else {
					// Offering exists, update it
					if err := db.Model(&existingOffering).Updates(offering).Error; err != nil {
						fmt.Printf(
							"Warning: Failed to update class offering %s-%s: %v\n",
							disc.Codigo,
							turma.Codigo,
							err,
						)
						stats.storeErrors.Add(1)
					} else {
						stats.offeringsStored.Add(1)
					}
				}
			}

			// Extract professors from turmas
			professorNames := make(map[string]bool)
			for _, turma := range disc.Turmas {
				// Names from the schedule column plus the "Atividades Didáticas"
				// table, so classes without a fixed schedule still get their
				// professors.
				rawNames := append([]string(nil), turma.Professores...)
				for _, horario := range turma.Horario {
					rawNames = append(rawNames, horario.Professores...)
				}
				for _, professorRaw := range rawNames {
					// Remove content in parentheses
					professorName := regexp.MustCompile(`\(.*?\)`).
						ReplaceAllString(professorRaw, "")
					professorName = strings.TrimSpace(professorName)
					if professorName != "" && len(professorName) > 3 &&
						len(professorName) < 50 &&
						regexp.MustCompile(`[a-zA-ZÀ-ÿ]`).MatchString(professorName) &&
						strings.Contains(professorName, " ") &&
						!strings.Contains(professorName, "Júpiter") &&
						professorName != "Instituto Oceanográfico" {
						professorNames[professorName] = true
					}
				}
			}

			// Store professors and create associations
			for professorName := range professorNames {
				professor, created, err := getOrCreateProfessor(db, professorName, unit.ID)
				if err != nil {
					fmt.Printf(
						"Warning: Failed to get or create professor %s: %v\n",
						professorName,
						err,
					)
					stats.storeErrors.Add(1)
					continue
				}
				if created {
					storedProfessors++
					stats.professorsCreated.Add(1)
				}

				_, created, err = getOrCreateClassProfessor(db, dbDiscipline.ID, professor.ID)
				if err != nil {
					fmt.Printf(
						"Warning: Failed to create association for %s-%s: %v\n",
						disc.Codigo,
						professorName,
						err,
					)
					stats.storeErrors.Add(1)
					continue
				}
				if created {
					storedAssociations++
				}
			}
		}

		fmt.Printf("✓ Successfully stored:\n")
		fmt.Printf("  - %d disciplines\n", storedDisciplines)
		fmt.Printf("  - %d professors\n", storedProfessors)
		fmt.Printf("  - %d class-professor associations\n", storedAssociations)
	} else {
		// Just display the data
		fmt.Printf("\n- Found %d disciplines -\n\n", len(disciplines))
		for i, disc := range disciplines {
			if i >= 10 {
				fmt.Printf("... and %d more disciplines\n", len(disciplines)-10)
				break
			}
			fmt.Printf("%-10s | %-50s | %s\n", disc.Codigo, disc.Nome, disc.Unidade)
		}
	}

	recorder.finish(stats, nil)

	fmt.Println("- DONE! -")
	fmt.Printf("- Execution time: %v seconds -\n", time.Since(startTime).Seconds())
}

func processDisciplinesConcurrently(
	targetUnits []string,
	allUnits []UnitInfo,
) ([]DisciplineInfo, error) {
	// Create unit name lookup
	unitNames := make(map[string]string)
	for _, unit := range allUnits {
		unitNames[unit.Code] = unit.Name
	}

	// First, get all discipline codes from all units
	fmt.Println("- Getting discipline list from all units -")
	allDisciplineCodes := make(chan []disciplineBasicInfo, len(targetUnits))
	errorsChan := make(chan error, len(targetUnits))

	var wg sync.WaitGroup
	for _, unitCode := range targetUnits {
		wg.Add(1)
		go func(code string) {
			defer wg.Done()
			disciplines, err := getDisciplinesFromUnit(code)
			if err != nil {
				errorsChan <- err
				return
			}
			allDisciplineCodes <- disciplines
		}(unitCode)
	}

	go func() {
		wg.Wait()
		close(allDisciplineCodes)
		close(errorsChan)
	}()

	// Collect discipline codes
	var allDisciplines []disciplineBasicInfo
	for disciplines := range allDisciplineCodes {
		allDisciplines = append(allDisciplines, disciplines...)
	}

	// A unit whose discipline list failed silently drops every discipline in
	// it, so count those failures rather than discarding them.
	for err := range errorsChan {
		fmt.Printf("Warning: Failed to list disciplines for a unit: %v\n", err)
		stats.unitListErrors.Add(1)
	}
	stats.disciplinesListed.Store(int64(len(allDisciplines)))

	fmt.Printf("- %d disciplines found -\n", len(allDisciplines))
	fmt.Println("- Starting detailed discipline processing -")

	// Process each discipline in detail
	resultsChan := make(chan *DisciplineInfo, len(allDisciplines))
	processErrorsChan := make(chan error, len(allDisciplines))

	// Semaphore to limit concurrency
	sem := make(chan struct{}, fetchDisciplinesConcurrency)
	var processWg sync.WaitGroup

	for _, disc := range allDisciplines {
		processWg.Add(1)
		go func(discipline disciplineBasicInfo) {
			defer processWg.Done()
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			disciplineInfo, err := processDisciplineDetailed(discipline)
			if err != nil {
				processErrorsChan <- err
				return
			}
			if disciplineInfo != nil {
				resultsChan <- disciplineInfo
			}
		}(disc)
	}

	go func() {
		processWg.Wait()
		close(resultsChan)
		close(processErrorsChan)
	}()

	// Collect results
	var processedDisciplines []DisciplineInfo
	var errors []error

	for {
		select {
		case discipline, ok := <-resultsChan:
			if !ok {
				resultsChan = nil
			} else if discipline != nil {
				processedDisciplines = append(processedDisciplines, *discipline)
			}
		case err, ok := <-processErrorsChan:
			if !ok {
				processErrorsChan = nil
			} else if err != nil {
				errors = append(errors, err)
			}
		}

		if resultsChan == nil && processErrorsChan == nil {
			break
		}
	}

	fmt.Printf("- %d disciplines processed -\n", len(processedDisciplines))
	stats.disciplinesProcessed.Store(int64(len(processedDisciplines)))

	if len(errors) > 0 {
		fmt.Printf("Warning: Encountered %d errors during processing\n", len(errors))
	}

	return processedDisciplines, nil
}

type disciplineBasicInfo struct {
	Code string
	Name string
}

func getDisciplinesFromUnit(unitCode string) ([]disciplineBasicInfo, error) {
	url := fmt.Sprintf(
		"https://uspdigital.usp.br/jupiterweb/jupDisciplinaLista?letra=A-Z&tipo=T&codcg=%s",
		unitCode,
	)

	fmt.Printf("- Getting disciplines from unit %s -\n", unitCode)

	doc, err := httpGetWithCharset(url, 120*time.Second)
	if err != nil {
		return nil, err
	}

	var disciplines []disciplineBasicInfo
	doc.Find("a[href*='obterTurma']").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		name := s.Text()

		re := regexp.MustCompile(`sgldis=([A-Z0-9\s]{7})`)
		matches := re.FindStringSubmatch(href)
		if len(matches) > 1 {
			disciplines = append(disciplines, disciplineBasicInfo{
				Code: strings.TrimSpace(matches[1]),
				Name: name,
			})
		}
	})

	fmt.Printf("- %d disciplines found in unit %s -\n", len(disciplines), unitCode)
	return disciplines, nil
}

func processDisciplineDetailed(
	discipline disciplineBasicInfo,
) (*DisciplineInfo, error) {
	fmt.Printf("- Processing %s - %s -\n", discipline.Code, discipline.Name)

	// Get class information (schedules, professors, enrollment)
	turmasURL := fmt.Sprintf(
		"https://uspdigital.usp.br/jupiterweb/obterTurma?print=true&sgldis=%s",
		discipline.Code,
	)

	turmasDoc, statusCode, err := httpGetWithCharsetAndStatus(
		turmasURL,
		time.Duration(fetchDisciplinesTimeout)*time.Second,
	)
	if err != nil {
		return nil, err
	}

	if statusCode != 200 {
		fmt.Printf(
			"Warning: Could not get class info for %s (HTTP %d)\n",
			discipline.Code,
			statusCode,
		)
		return nil, nil
	}

	turmas, err := parseTurmas(turmasDoc)
	if err != nil {
		return nil, err
	}

	if len(turmas) == 0 {
		fmt.Printf(
			"Warning: Discipline %s has no valid classes registered. Skipping...\n",
			discipline.Code,
		)
		stats.disciplinesSkipped.Add(1)
		return nil, nil
	}

	// Get discipline information (description, objectives, credits)
	infoURL := fmt.Sprintf(
		"https://uspdigital.usp.br/jupiterweb/obterDisciplina?print=true&sgldis=%s",
		discipline.Code,
	)

	infoDoc, infoStatusCode, err := httpGetWithCharsetAndStatus(
		infoURL,
		time.Duration(fetchDisciplinesTimeout)*time.Second,
	)
	if err != nil {
		return nil, err
	}

	if infoStatusCode != 200 {
		fmt.Printf(
			"Warning: Could not get discipline info for %s (HTTP %d)\n",
			discipline.Code,
			infoStatusCode,
		)
		return nil, nil
	}

	disciplineInfo, err := parseDisciplineInfo(infoDoc)
	if err != nil {
		return nil, err
	}

	if disciplineInfo == nil {
		// The class page listed offerings, so the info page should have had a
		// header: a missing one means the HTML changed or the page was empty.
		fmt.Printf(
			"Warning: Discipline %s has no information registered. Skipping...\n",
			discipline.Code,
		)
		stats.parseErrors.Add(1)
		return nil, nil
	}

	// Add class information
	disciplineInfo.Turmas = turmas

	return disciplineInfo, nil
}

func parseTurmas(doc *goquery.Document) ([]TurmaInfo, error) {
	var turmas []TurmaInfo
	var currentTurma *TurmaInfo
	var currentHorario []HorarioInfo
	var currentVagas map[string]VagasInfo
	var currentAtividades []string

	doc.Find("table").Each(func(i int, table *goquery.Selection) {
		tableText := strings.Join(strings.Fields(table.Text()), " ")
		hasAtividades := strings.Contains(tableText, "Atividades Didáticas")

		// Check for class code table. It must be small (< 1000 chars) and hold
		// no nested tables, otherwise the wrapper around a class block matches
		// too and yields a duplicate class.
		if strings.Contains(tableText, "Código da Turma") && len(tableText) < 1000 &&
			table.Find("table").Length() == 0 {
			// Save previous turma if exists
			if currentTurma != nil {
				// Save turma with whatever data we have (even if incomplete)
				turmas = append(turmas, finishTurma(
					currentTurma, currentHorario, currentVagas, currentAtividades,
				))
			}

			// Parse new turma info
			currentTurma = parseTurmaInfo(table)
			currentHorario = nil
			currentVagas = nil
			currentAtividades = nil
		}

		// The "Atividades Didáticas" table mentions "Horário Variável", so it
		// must not be mistaken for the schedule table: it has three columns
		// and would wipe the schedule parsed just before it.
		if strings.Contains(tableText, "Horário") && !hasAtividades {
			// Parse schedule
			currentHorario = parseHorario(table)
		}

		if hasAtividades {
			if names := parseAtividades(table); len(names) > 0 {
				currentAtividades = names
			}
		}

		if strings.Contains(tableText, "Vagas") {
			// Parse enrollment data
			currentVagas = parseVagas(table)
		}
	})

	// Don't forget the last turma
	if currentTurma != nil {
		turmas = append(turmas, finishTurma(
			currentTurma, currentHorario, currentVagas, currentAtividades,
		))
	}

	return turmas, nil
}

// finishTurma assembles a class from the tables parsed for it. Professors
// listed in the "Atividades Didáticas" table are attached to every schedule
// slot that has no professor of its own, so MatrUSP and the professor
// extraction see them in the same place as on older pages.
func finishTurma(
	turma *TurmaInfo,
	horario []HorarioInfo,
	vagas map[string]VagasInfo,
	atividades []string,
) TurmaInfo {
	if len(atividades) > 0 {
		for i := range horario {
			if !hasProfessor(horario[i].Professores) {
				horario[i].Professores = append([]string(nil), atividades...)
			}
		}
	}
	turma.Horario = horario
	turma.Vagas = vagas
	turma.Professores = atividades

	if horario == nil {
		fmt.Printf("Warning: Class %s has no schedule registered\n", turma.Codigo)
	}
	if vagas == nil {
		fmt.Printf("Warning: Class %s has no enrollment data registered\n", turma.Codigo)
	}
	return *turma
}

func hasProfessor(names []string) bool {
	for _, n := range names {
		if strings.TrimSpace(n) != "" {
			return true
		}
	}
	return false
}

// parseAtividades reads professor names from an "Atividades Didáticas" table.
// Each data row is "name | activity type | hours", and a professor appears once
// per activity, so names are de-duplicated in page order. Rows are accepted
// only when the last cell is a number, which keeps header rows and unrelated
// three-cell rows of enclosing wrapper tables out.
func parseAtividades(table *goquery.Selection) []string {
	var names []string
	seen := make(map[string]bool)

	table.Find("tr").Each(func(i int, tr *goquery.Selection) {
		tds := tr.Find("td")
		if tds.Length() != 3 {
			return
		}
		name := strings.Join(strings.Fields(tds.Eq(0).Text()), " ")
		hours := strings.TrimSpace(tds.Eq(2).Text())
		if name == "" || name == "Prof(a)." {
			return
		}
		if _, err := strconv.Atoi(hours); err != nil {
			return
		}
		if !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	})

	return names
}

func parseTurmaInfo(table *goquery.Selection) *TurmaInfo {
	turma := &TurmaInfo{}

	table.Find("tr").Each(func(i int, tr *goquery.Selection) {
		tds := tr.Find("td")
		if tds.Length() < 2 {
			return
		}

		label := strings.TrimSpace(tds.First().Text())
		value := strings.TrimSpace(tds.Eq(1).Text())

		// Normalize label by replacing newlines and multiple spaces with single space
		label = strings.Join(strings.Fields(label), " ")

		switch {
		case strings.Contains(label, "Código da Turma Teórica"):
			turma.CodigoTeorica = value
		case strings.Contains(label, "Código da Turma") && !strings.Contains(label, "Teórica"):
			// Extract just the code part (first word)
			if parts := strings.Fields(value); len(parts) > 0 {
				turma.Codigo = parts[0]
			}
		case strings.Contains(label, "Início"):
			turma.Inicio = parseDate(value)
		case strings.Contains(label, "Fim"):
			turma.Fim = parseDate(value)
		case strings.Contains(label, "Tipo da Turma"):
			turma.Tipo = value
		case strings.Contains(label, "Observações"):
			turma.Observacoes = value
		}
	})

	return turma
}

func parseDate(dateStr string) string {
	// Simple date parsing - could be improved
	return dateStr
}

func parseHorario(table *goquery.Selection) []HorarioInfo {
	var horario []HorarioInfo
	var currentSlot *HorarioInfo

	table.Find("tr").Each(func(i int, tr *goquery.Selection) {
		tds := tr.Find("td")
		if tds.Length() < 4 {
			return
		}

		dia := strings.TrimSpace(tds.Eq(0).Text())
		inicio := strings.TrimSpace(tds.Eq(1).Text())
		fim := strings.TrimSpace(tds.Eq(2).Text())
		professor := strings.TrimSpace(tds.Eq(3).Text())

		if dia == "Horário" {
			// Header row
			return
		}

		if dia != "" {
			// New time slot
			if currentSlot != nil {
				horario = append(horario, *currentSlot)
			}
			currentSlot = &HorarioInfo{
				Dia:         dia,
				Inicio:      inicio,
				Fim:         fim,
				Professores: []string{professor},
			}
		} else if currentSlot != nil {
			// Additional professor or extended time
			if inicio == "" && fim != "" {
				// Extended end time
				if fim > currentSlot.Fim {
					currentSlot.Fim = fim
				}
				if professor != "" {
					currentSlot.Professores = append(currentSlot.Professores, professor)
				}
			} else if inicio != "" {
				// Another time slot on same day
				horario = append(horario, *currentSlot)
				currentSlot = &HorarioInfo{
					Dia:         currentSlot.Dia,
					Inicio:      inicio,
					Fim:         fim,
					Professores: []string{professor},
				}
			}
		}
	})

	if currentSlot != nil {
		horario = append(horario, *currentSlot)
	}

	return horario
}

func parseVagas(table *goquery.Selection) map[string]VagasInfo {
	vagas := make(map[string]VagasInfo)
	var currentType string
	var currentVaga VagasInfo

	table.Find("tr").Each(func(i int, tr *goquery.Selection) {
		tds := tr.Find("td")
		if tds.Length() < 5 {
			return
		}

		cells := make([]string, tds.Length())
		tds.Each(func(j int, td *goquery.Selection) {
			cells[j] = strings.TrimSpace(td.Text())
		})

		if len(cells) == 5 && cells[0] == "" {
			// Header
			return
		} else if len(cells) == 5 && cells[0] != "" {
			// New enrollment type
			if currentType != "" {
				vagas[currentType] = currentVaga
			}

			currentType = cells[0]
			currentVaga = VagasInfo{
				Vagas:        toInt(cells[1]),
				Inscritos:    toInt(cells[2]),
				Pendentes:    toInt(cells[3]),
				Matriculados: toInt(cells[4]),
				Grupos:       make(map[string]VagasInfo),
			}
		} else if len(cells) == 6 {
			// Group details
			grupo := cells[1]
			groupVaga := VagasInfo{
				Vagas:        toInt(cells[2]),
				Inscritos:    toInt(cells[3]),
				Pendentes:    toInt(cells[4]),
				Matriculados: toInt(cells[5]),
			}
			currentVaga.Grupos[grupo] = groupVaga
		}
	})

	if currentType != "" {
		vagas[currentType] = currentVaga
	}

	return vagas
}

func parseDisciplineInfo(doc *goquery.Document) (*DisciplineInfo, error) {
	info := &DisciplineInfo{}

	doc.Find("table").Each(func(i int, table *goquery.Selection) {
		// Skip nested tables
		if table.Parents().FilterFunction(func(i int, s *goquery.Selection) bool {
			return s.Is("table")
		}).Length() > 0 {
			return
		}

		tableText := table.Text()

		// Parse header information
		if strings.Contains(tableText, "Disciplina:") {
			// Extract from table rows with specific structure
			rows := table.Find("tr")
			var rowTexts []string
			rows.Each(func(j int, tr *goquery.Selection) {
				td := tr.Find("td").First()
				text := strings.TrimSpace(td.Text())
				if text != "" && len(text) < 200 { // Avoid getting huge text blocks
					rowTexts = append(rowTexts, text)
				}
			})

			// Look for unit, department, and discipline in row texts
			for idx, text := range rowTexts {
				if strings.HasPrefix(text, "Disciplina:") {
					// Found discipline row, previous rows should be department and unit
					if idx >= 2 {
						info.Unidade = rowTexts[idx-2]
						info.Departamento = rowTexts[idx-1]
					} else if idx >= 1 {
						info.Unidade = rowTexts[idx-1]
					}

					// Parse discipline name and code
					disciplineRe := regexp.MustCompile(`Disciplina:\s+([A-Z0-9\s]{7})\s*-\s*(.+)`)
					if matches := disciplineRe.FindStringSubmatch(text); len(matches) >= 3 {
						info.Codigo = strings.TrimSpace(matches[1])
						info.Nome = matches[2]
					}
					break
				}
			}

			// Get campus from unit
			info.Campus = "São Paulo" // Default
		}

		// Parse objectives
		trs := table.Find("tr")
		if trs.Length() >= 2 {
			firstRow := strings.TrimSpace(trs.First().Text())
			if firstRow == "Objetivos" {
				info.Objetivos = strings.TrimSpace(trs.Eq(1).Text())
			} else if firstRow == "Programa Resumido" {
				info.ProgramaResumido = strings.TrimSpace(trs.Eq(1).Text())
			}
		}

		// Parse credits
		if strings.Contains(tableText, "Créditos Aula") {
			credits := parseCredits(table)
			info.CreditosAula = credits.Aula
			info.CreditosTrabalho = credits.Trabalho
		}
	})

	// Validate that we have basic info
	if info.Codigo == "" {
		return nil, nil
	}

	return info, nil
}

type creditsInfo struct {
	Aula     int
	Trabalho int
}

func parseCredits(table *goquery.Selection) creditsInfo {
	credits := creditsInfo{}

	table.Find("tr").Each(func(i int, tr *goquery.Selection) {
		tds := tr.Find("td")
		if tds.Length() < 2 {
			return
		}

		label := strings.TrimSpace(tds.First().Text())
		value := strings.TrimSpace(tds.Eq(1).Text())

		if strings.Contains(label, "Créditos Aula:") {
			credits.Aula = toInt(value)
		} else if strings.Contains(label, "Créditos Trabalho:") {
			credits.Trabalho = toInt(value)
		}
	})

	return credits
}

func toInt(s string) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return 0
}
