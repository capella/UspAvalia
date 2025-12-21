package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
	"uspavalia/internal/config"
	"uspavalia/internal/database"
	"uspavalia/internal/models"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
)

var fetchCoursesCMD = &cobra.Command{
	Use:   "fetch-courses [output-directory]",
	Short: "Fetch USP course curricula from Jupiter Web",
	Long:  `Fetches course curriculum information from USP Jupiter Web. Use --store to save to JSON files.`,
	Args:  cobra.MaximumNArgs(1),
	Run:   runFetchCourses,
}

type CourseInfo struct {
	Codigo   string                       `json:"codigo"`
	Nome     string                       `json:"nome"`
	Unidade  string                       `json:"unidade"`
	Periodo  string                       `json:"periodo"`
	Periodos map[string][]DisciplinaCurso `json:"periodos"`
}

type DisciplinaCurso struct {
	Codigo      string   `json:"codigo"`
	Tipo        string   `json:"tipo"`
	ReqFraco    []string `json:"req_fraco"`
	ReqForte    []string `json:"req_forte"`
	IndConjunto []string `json:"ind_conjunto"`
}

var (
	fetchCoursesUnits       []string
	fetchCoursesTimeout     int
	fetchCoursesOut         string
	fetchCoursesConcurrency int
	fetchCoursesNoGzip      bool
	fetchCoursesStore       bool
)

func init() {
	rootCmd.AddCommand(fetchCoursesCMD)

	fetchCoursesCMD.Flags().
		StringSliceVarP(&fetchCoursesUnits, "units", "u", nil, "Fetch only these unit codes")
	fetchCoursesCMD.Flags().
		IntVarP(&fetchCoursesTimeout, "timeout", "t", 120, "HTTP request timeout in seconds")
	fetchCoursesCMD.Flags().
		StringVarP(&fetchCoursesOut, "output", "o", "cursos.json", "Output filename")
	fetchCoursesCMD.Flags().
		IntVarP(&fetchCoursesConcurrency, "concurrency", "c", 10, "Number of concurrent requests")
	fetchCoursesCMD.Flags().
		BoolVar(&fetchCoursesNoGzip, "no-gzip", false, "Don't create gzipped output")
	fetchCoursesCMD.Flags().
		BoolVar(&fetchCoursesStore, "store", false, "Store courses to JSON files in output directory")
}

func runFetchCourses(cmd *cobra.Command, args []string) {
	var outputDir string
	if len(args) > 0 {
		outputDir = args[0]
	} else if fetchCoursesStore {
		outputDir = "./matrusp"
	}

	// Validate output directory if storing
	if fetchCoursesStore {
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			fmt.Printf("Error: Output directory '%s' does not exist\n", outputDir)
			os.Exit(1)
		}
	}

	startTime := time.Now()

	fmt.Println("- Obtaining list of all teaching units -")

	// Get teaching units
	units, err := getTeachingUnits()
	if err != nil {
		fmt.Printf("Error getting teaching units: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("- %d teaching units found -\n", len(units))

	// Filter units if specified
	var targetUnits []string
	if len(fetchCoursesUnits) > 0 {
		targetUnits = fetchCoursesUnits
	} else {
		for _, unit := range units {
			targetUnits = append(targetUnits, unit.Code)
		}
	}

	fmt.Println("- Starting course processing -")

	// Process courses concurrently
	courses, err := processCoursesConcurrently(targetUnits, units)
	if err != nil {
		fmt.Printf("Error processing courses: %v\n", err)
		os.Exit(1)
	}

	if fetchCoursesStore {
		// Store in database
		fmt.Println("\nStoring courses in database...")

		cfg := config.Load()
		db, err := database.Initialize(cfg)
		if err != nil {
			fmt.Printf("Error: Failed to initialize database: %v\n", err)
			os.Exit(1)
		}

		stored := 0
		for _, course := range courses {
			// Get or create unit by name
			var unit models.Unit
			result := db.Where("NOME = ?", course.Unidade).First(&unit)
			if result.Error != nil {
				// Unit doesn't exist, create it
				unit = models.Unit{Name: course.Unidade}
				if err := db.Create(&unit).Error; err != nil {
					fmt.Printf("Warning: Failed to create unit %s: %v\n", course.Unidade, err)
					continue
				}
			}

			// Marshal periods to JSON
			periodsJSON, err := json.Marshal(course.Periodos)
			if err != nil {
				fmt.Printf("Warning: Failed to marshal periods for %s: %v\n", course.Codigo, err)
				continue
			}

			// Create or update course
			dbCourse := models.Course{
				Code:    course.Codigo,
				Name:    course.Nome,
				UnitID:  unit.ID,
				Period:  course.Periodo,
				Periods: string(periodsJSON),
			}

			// Use FirstOrCreate to avoid duplicates based on code
			var existingCourse models.Course
			result = db.Where("code = ?", course.Codigo).First(&existingCourse)
			if result.Error != nil {
				// Course doesn't exist, create it
				if err := db.Create(&dbCourse).Error; err != nil {
					fmt.Printf("Warning: Failed to store course %s: %v\n", course.Codigo, err)
					continue
				}
			} else {
				// Course exists, update it
				db.Model(&existingCourse).Updates(dbCourse)
			}

			stored++
		}

		fmt.Printf("✓ Successfully stored %d courses in database\n", stored)
	} else {
		// Just display the data
		fmt.Printf("\n- Found %d courses -\n\n", len(courses))
		for i, course := range courses {
			if i >= 10 {
				fmt.Printf("... and %d more courses\n", len(courses)-10)
				break
			}
			fmt.Printf("%-10s | %-50s | %s\n", course.Codigo, course.Nome, course.Unidade)
		}
	}

	fmt.Println("- DONE! -")
	fmt.Printf("- Execution time: %v seconds -\n", time.Since(startTime).Seconds())
}

func processCoursesConcurrently(targetUnits []string, allUnits []UnitInfo) ([]CourseInfo, error) {
	// Create unit name lookup
	unitNames := make(map[string]string)
	for _, unit := range allUnits {
		unitNames[unit.Code] = unit.Name
	}

	// Channel for collecting results
	resultsChan := make(chan []CourseInfo, len(targetUnits))
	errorsChan := make(chan error, len(targetUnits))

	// Semaphore to limit concurrency
	sem := make(chan struct{}, fetchCoursesConcurrency)
	var wg sync.WaitGroup

	// Process each unit
	for _, unitCode := range targetUnits {
		wg.Add(1)
		go func(code string) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			courses, err := processUnit(code, unitNames[code])
			if err != nil {
				errorsChan <- err
				return
			}
			resultsChan <- courses
		}(unitCode)
	}

	// Close channels when done
	go func() {
		wg.Wait()
		close(resultsChan)
		close(errorsChan)
	}()

	// Collect results
	var allCourses []CourseInfo
	var errors []error

	for {
		select {
		case courses, ok := <-resultsChan:
			if !ok {
				resultsChan = nil
			} else {
				allCourses = append(allCourses, courses...)
			}
		case err, ok := <-errorsChan:
			if !ok {
				errorsChan = nil
			} else if err != nil {
				errors = append(errors, err)
			}
		}

		if resultsChan == nil && errorsChan == nil {
			break
		}
	}

	if len(errors) > 0 {
		return allCourses, fmt.Errorf("encountered %d errors during processing", len(errors))
	}

	return allCourses, nil
}

func processUnit(unitCode, unitName string) ([]CourseInfo, error) {
	url := fmt.Sprintf(
		"https://uspdigital.usp.br/jupiterweb/jupCursoLista?tipo=N&codcg=%s",
		unitCode,
	)

	doc, err := httpGetWithCharset(url, 120*time.Second)
	if err != nil {
		return nil, err
	}

	var courses []CourseInfo

	// Find course links
	doc.Find("a[href*='listarGradeCurricular']").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")

		// Get period from next table cell
		period := ""
		if tr := s.Parent().Parent(); tr != nil {
			if td := tr.Find("td").Last(); td != nil {
				period = strings.TrimSpace(td.Text())
			}
		}

		fmt.Printf("- Processing course from %s -\n", href)

		course, err := parseCourse(href, period, unitName)
		if err != nil {
			fmt.Printf("Error parsing course %s: %v\n", href, err)
			return
		}

		if course != nil {
			courses = append(courses, *course)
		}
	})

	return courses, nil
}

func parseCourse(relativeLink, period, unitName string) (*CourseInfo, error) {
	if relativeLink == "" {
		return nil, nil
	}

	link := "https://uspdigital.usp.br/jupiterweb/" + relativeLink

	doc, statusCode, err := httpGetWithCharsetAndStatus(
		link,
		time.Duration(fetchCoursesTimeout)*time.Second,
	)
	if err != nil {
		return nil, err
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", statusCode)
	}

	course := CourseInfo{
		Periodo: period,
		Unidade: unitName,
	}

	// Extract course code and name from URL
	codeRe := regexp.MustCompile(`codcur=(.+?)&codhab=(.+?)(&|$)`)
	matches := codeRe.FindStringSubmatch(link)
	if len(matches) >= 3 {
		course.Codigo = fmt.Sprintf("%s-%s", matches[1], matches[2])
	}

	// Extract course name
	courseText := doc.Text()
	nameRe := regexp.MustCompile(`Curso:\s*(.+)\s*`)
	nameMatches := nameRe.FindAllStringSubmatch(courseText, -1)
	var names []string
	for _, match := range nameMatches {
		if len(match) > 1 {
			names = append(names, strings.TrimSpace(match[1]))
		}
	}
	course.Nome = strings.Join(names, " - ")

	// Parse curriculum periods
	periods := parseCoursePeriods(doc)
	course.Periodos = periods

	return &course, nil
}

func parseCoursePeriods(doc *goquery.Document) map[string][]DisciplinaCurso {
	periods := make(map[string][]DisciplinaCurso)

	// Find the table with required disciplines
	doc.Find("table").Each(func(i int, table *goquery.Selection) {
		// Check if this table contains discipline information
		if table.Find("table").Length() > 0 {
			return // Skip tables with nested tables
		}

		hasRequiredDisciplines := false
		table.Find("tr").Each(func(j int, tr *goquery.Selection) {
			text := strings.TrimSpace(tr.Text())
			if strings.Contains(text, "Disciplinas Obrigatórias") {
				hasRequiredDisciplines = true
				return
			}
		})

		if !hasRequiredDisciplines {
			return
		}

		// Parse the periods from this table
		currentType := ""
		currentPeriod := ""

		table.Find("tr").Each(func(j int, tr *goquery.Selection) {
			text := strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(tr.Text(), " "))

			// Check for discipline type
			switch text {
			case "Disciplinas Obrigatórias":
				currentType = "obrigatoria"
				return
			case "Disciplinas Optativas Eletivas":
				currentType = "optativa_eletiva"
				return
			case "Disciplinas Optativas Livres":
				currentType = "optativa_livre"
				return
			}

			// Check for period
			periodRe := regexp.MustCompile(`([0-9]+)º Período Ideal`)
			if matches := periodRe.FindStringSubmatch(text); len(matches) > 1 {
				currentPeriod = matches[1]
				if periods[currentPeriod] == nil {
					periods[currentPeriod] = []DisciplinaCurso{}
				}
				return
			}

			// Parse discipline or requirement
			tds := tr.Find("td")
			if tds.Length() == 0 {
				return
			}

			firstCell := strings.TrimSpace(tds.First().Text())

			if len(firstCell) == 7 && currentPeriod != "" {
				// New discipline
				discipline := DisciplinaCurso{
					Codigo:      firstCell,
					Tipo:        currentType,
					ReqFraco:    []string{},
					ReqForte:    []string{},
					IndConjunto: []string{},
				}
				periods[currentPeriod] = append(periods[currentPeriod], discipline)
			} else if len(periods[currentPeriod]) > 0 && tds.Length() >= 2 {
				// Requirement for the last discipline
				secondCell := strings.TrimSpace(tds.Eq(1).Text())
				lastIdx := len(periods[currentPeriod]) - 1

				switch secondCell {
				case "Requisito fraco":
					if len(firstCell) >= 7 {
						periods[currentPeriod][lastIdx].ReqFraco = append(periods[currentPeriod][lastIdx].ReqFraco, firstCell[:7])
					}
				case "Requisito":
					if len(firstCell) >= 7 {
						periods[currentPeriod][lastIdx].ReqForte = append(periods[currentPeriod][lastIdx].ReqForte, firstCell[:7])
					}
				case "Indicação de Conjunto":
					if len(firstCell) >= 7 {
						periods[currentPeriod][lastIdx].IndConjunto = append(periods[currentPeriod][lastIdx].IndConjunto, firstCell[:7])
					}
				}
			}
		})
	})

	return periods
}
