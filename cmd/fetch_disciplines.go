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

// Campus mapping - manually maintained as in original Python
var campusPorUnidade = map[string]string{
	// São Paulo
	"86": "São Paulo", "27": "São Paulo", "39": "São Paulo", "7": "São Paulo",
	"22": "São Paulo", "3": "São Paulo", "16": "São Paulo", "9": "São Paulo",
	"2": "São Paulo", "12": "São Paulo", "48": "São Paulo", "8": "São Paulo",
	"5": "São Paulo", "10": "São Paulo", "67": "São Paulo", "23": "São Paulo",
	"6": "São Paulo", "66": "São Paulo", "14": "São Paulo", "26": "São Paulo",
	"93": "São Paulo", "41": "São Paulo", "92": "São Paulo", "42": "São Paulo",
	"4": "São Paulo", "37": "São Paulo", "43": "São Paulo", "44": "São Paulo",
	"45": "São Paulo", "83": "São Paulo", "47": "São Paulo", "46": "São Paulo",
	"87": "São Paulo", "21": "São Paulo", "31": "São Paulo", "85": "São Paulo",
	"71": "São Paulo", "32": "São Paulo", "38": "São Paulo", "33": "São Paulo",
	// Ribeirão Preto
	"98": "Ribeirão Preto", "94": "Ribeirão Preto", "60": "Ribeirão Preto",
	"89": "Ribeirão Preto", "81": "Ribeirão Preto", "59": "Ribeirão Preto",
	"96": "Ribeirão Preto", "91": "Ribeirão Preto", "17": "Ribeirão Preto",
	"58": "Ribeirão Preto", "95": "Ribeirão Preto",
	// Other campuses
	"88": "Lorena",
	"18": "São Carlos", "97": "São Carlos", "99": "São Carlos",
	"55": "São Carlos", "76": "São Carlos", "75": "São Carlos", "90": "São Carlos",
	"11": "Piracicaba", "64": "Piracicaba",
	"25": "Bauru", "61": "Bauru",
	"74": "Pirassununga",
	"30": "São Sebastião",
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

	fmt.Println("- Obtaining list of all teaching units -")

	// Get teaching units (reuse from parse_courses.go)
	units, err := getTeachingUnits()
	if err != nil {
		fmt.Printf("Error getting teaching units: %v\n", err)
		os.Exit(1)
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
		fmt.Printf("Error processing disciplines: %v\n", err)
		os.Exit(1)
	}

	if fetchDisciplinesStore {
		// Store in database
		fmt.Println("\nStoring disciplines in database...")

		cfg := config.Load()
		db, err := database.Initialize(cfg)
		if err != nil {
			fmt.Printf("Error: Failed to initialize database: %v\n", err)
			os.Exit(1)
		}

		storedDisciplines := 0
		storedProfessors := 0
		storedAssociations := 0

		for _, disc := range disciplines {
			// Get or create unit by name
			var unit models.Unit
			result := db.Where("NOME = ?", disc.Unidade).First(&unit)
			if result.Error != nil {
				// Unit doesn't exist, create it
				unit = models.Unit{Name: disc.Unidade}
				if err := db.Create(&unit).Error; err != nil {
					fmt.Printf("Warning: Failed to create unit %s: %v\n", disc.Unidade, err)
					continue
				}
			}

			// Create or update discipline
			dbDiscipline := models.Discipline{
				Code:   disc.Codigo,
				Name:   disc.Nome,
				UnitID: unit.ID,
			}

			var existingDiscipline models.Discipline
			result = db.Where("code = ?", disc.Codigo).First(&existingDiscipline)
			if result.Error != nil {
				// Discipline doesn't exist, create it
				if err := db.Create(&dbDiscipline).Error; err != nil {
					fmt.Printf("Warning: Failed to store discipline %s: %v\n", disc.Codigo, err)
					continue
				}
			} else {
				// Discipline exists, update it and use existing ID
				db.Model(&existingDiscipline).Updates(dbDiscipline)
				dbDiscipline.ID = existingDiscipline.ID
			}
			storedDisciplines++

			// Update discipline with MatrUSP fields
			db.Model(&dbDiscipline).Updates(map[string]interface{}{
				"Department":   disc.Departamento,
				"Campus":       disc.Campus,
				"CreditsClass": disc.CreditosAula,
				"CreditsWork":  disc.CreditosTrabalho,
				"Objectives":   disc.Objetivos,
				"Summary":      disc.ProgramaResumido,
			})

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
					}
				} else {
					// Offering exists, update it
					db.Model(&existingOffering).Updates(offering)
				}
			}

			// Extract professors from turmas
			professorNames := make(map[string]bool)
			for _, turma := range disc.Turmas {
				if turma.Horario == nil {
					continue
				}
				for _, horario := range turma.Horario {
					for _, professorRaw := range horario.Professores {
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
			}

			// Store professors and create associations
			for professorName := range professorNames {
				// Create or get professor
				var professor models.Professor
				result = db.Where("name = ?", professorName).First(&professor)
				if result.Error != nil {
					// Professor doesn't exist, create it
					professor = models.Professor{
						Name:   professorName,
						UnitID: unit.ID,
					}
					if err := db.Create(&professor).Error; err != nil {
						fmt.Printf(
							"Warning: Failed to create professor %s: %v\n",
							professorName,
							err,
						)
						continue
					}
					storedProfessors++
				}

				// Create ClassProfessor association
				var classProfessor models.ClassProfessor
				result = db.Where("class_id = ? AND professor_id = ?", dbDiscipline.ID, professor.ID).
					First(&classProfessor)
				if result.Error != nil {
					// Association doesn't exist, create it
					classProfessor = models.ClassProfessor{
						ClassID:     dbDiscipline.ID,
						ProfessorID: professor.ID,
					}
					if err := db.Create(&classProfessor).Error; err != nil {
						fmt.Printf(
							"Warning: Failed to create association for %s-%s: %v\n",
							disc.Codigo,
							professorName,
							err,
						)
						continue
					}
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

	fmt.Println("- DONE! -")
	fmt.Printf("- Execution time: %v seconds -\n", time.Since(startTime).Seconds())
}

func createCampusMapping(units []UnitInfo) map[string][]string {
	campi := make(map[string][]string)

	for _, unit := range units {
		campus := campusPorUnidade[unit.Code]
		if campus == "" {
			campus = "Outro"
		}
		campi[campus] = append(campi[campus], unit.Name)
	}

	return campi
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
		fmt.Printf(
			"Warning: Discipline %s has no information registered. Skipping...\n",
			discipline.Code,
		)
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

	doc.Find("table").Each(func(i int, table *goquery.Selection) {
		tableText := strings.Join(strings.Fields(table.Text()), " ")

		// Check for class code table - must be small tables (< 1000 chars) to avoid giant wrapper tables
		if strings.Contains(tableText, "Código da Turma") && len(tableText) < 1000 {
			// Save previous turma if exists
			if currentTurma != nil {
				// Save turma with whatever data we have (even if incomplete)
				currentTurma.Horario = currentHorario
				currentTurma.Vagas = currentVagas
				turmas = append(turmas, *currentTurma)

				if currentHorario == nil {
					fmt.Printf(
						"Warning: Class %s has no schedule registered\n",
						currentTurma.Codigo,
					)
				}
				if currentVagas == nil {
					fmt.Printf(
						"Warning: Class %s has no enrollment data registered\n",
						currentTurma.Codigo,
					)
				}
			}

			// Parse new turma info
			currentTurma = parseTurmaInfo(table)
			currentHorario = nil
			currentVagas = nil
		}

		if strings.Contains(tableText, "Horário") {
			// Parse schedule
			currentHorario = parseHorario(table)
		}

		if strings.Contains(tableText, "Vagas") {
			// Parse enrollment data
			currentVagas = parseVagas(table)
		}
	})

	// Don't forget the last turma
	if currentTurma != nil {
		currentTurma.Horario = currentHorario
		currentTurma.Vagas = currentVagas
		turmas = append(turmas, *currentTurma)
	}

	return turmas, nil
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

func extractTableStrings(table *goquery.Selection) []string {
	var textStrings []string
	table.Find("*").Contents().Each(func(i int, s *goquery.Selection) {
		if s.Get(0).Type == 3 { // Text node
			text := strings.TrimSpace(s.Text())
			if text != "" {
				textStrings = append(textStrings, text)
			}
		}
	})
	return textStrings
}

func toInt(s string) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return 0
}
