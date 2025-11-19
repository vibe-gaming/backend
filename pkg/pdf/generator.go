package pdf

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/signintech/gopdf"
	"github.com/vibe-gaming/backend/internal/domain"
)

type Generator struct {
	pdf      *gopdf.GoPdf
	hasFont  bool
	fontName string
}

// NewGenerator создает новый генератор PDF
func NewGenerator() *Generator {
	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{
		PageSize: *gopdf.PageSizeA4,
		Unit:     gopdf.Unit_PT,
	})

	hasFont := false
	fontName := "dejavu"

	// Выводим текущую рабочую директорию для отладки
	if wd, err := os.Getwd(); err == nil {
		fmt.Printf("📂 PDF: Current working directory: %s\n", wd)
	}

	// Проверяем содержимое /app/fonts
	if entries, err := os.ReadDir("/app/fonts"); err == nil {
		fmt.Printf("📂 PDF: Contents of /app/fonts/:\n")
		for _, entry := range entries {
			fmt.Printf("  - %s\n", entry.Name())
		}
	} else {
		fmt.Printf("❌ PDF: Cannot read /app/fonts directory: %v\n", err)
	}

	// Пути для поиска шрифта (production и development)
	fontPaths := []string{
		"/app/fonts/DejaVuSans.ttf",      // Production path в Docker контейнере
		"./fonts/DejaVuSans.ttf",         // Development path (из корня проекта)
		"./backend/fonts/DejaVuSans.ttf", // Alternative development path
	}

	for _, path := range fontPaths {
		fmt.Printf("🔍 PDF: Checking font path: %s\n", path)
		if stat, err := os.Stat(path); err == nil {
			fmt.Printf("📁 PDF: File exists at %s (size: %d bytes)\n", path, stat.Size())
			if err := pdf.AddTTFFont(fontName, path); err == nil {
				hasFont = true
				fmt.Printf("✅ PDF: Font successfully loaded from %s\n", path)
				break
			} else {
				fmt.Printf("❌ PDF: Failed to load font from %s: %v\n", path, err)
			}
		} else {
			fmt.Printf("❌ PDF: File not found at %s: %v\n", path, err)
		}
	}

	if !hasFont {
		fmt.Printf("⚠️  PDF: Font not found in any of the paths: %v\n", fontPaths)
	}

	return &Generator{
		pdf:      pdf,
		hasFont:  hasFont,
		fontName: fontName,
	}
}

// GenerateBenefitPDF генерирует PDF-документ для льготы
func (g *Generator) GenerateBenefitPDF(benefit *domain.Benefit) ([]byte, error) {
	// Проверяем, загружен ли шрифт
	if !g.hasFont {
		return nil, fmt.Errorf("TTF font not loaded. Font should be at /app/fonts/DejaVuSans.ttf (production) or ./fonts/DejaVuSans.ttf (development)")
	}

	// Добавляем страницу
	g.pdf.AddPage()

	// Устанавливаем шрифт
	g.pdf.SetFont(g.fontName, "", 14)

	// Заголовок документа
	g.addHeader()

	// Название льготы
	if g.hasFont {
		g.pdf.SetFont(g.fontName, "", 18)
		g.pdf.SetX(50)
		g.pdf.SetY(100)
		g.pdf.Cell(nil, "Название: "+benefit.Title)
	}

	// Тип льготы
	if g.hasFont {
		g.pdf.SetY(g.pdf.GetY() + 30)
		g.pdf.SetX(50)
		g.pdf.SetFont(g.fontName, "", 12)
		levelText := g.getLevelText(benefit.Type)
		g.pdf.Cell(nil, "Уровень: "+levelText)
	}

	// Описание
	g.pdf.SetY(g.pdf.GetY() + 25)
	g.addSection("Описание", benefit.Description)

	// Целевые группы
	if len(benefit.TargetGroupIDs) > 0 {
		groups := []string{}
		for _, tg := range benefit.TargetGroupIDs {
			groups = append(groups, g.getTargetGroupText(tg))
		}
		g.addSection("Целевые группы", strings.Join(groups, ", "))
	}

	// Категория
	if benefit.Category != nil {
		categoryText := g.getCategoryText(*benefit.Category)
		g.addSection("Категория", categoryText)
	}

	// Требования
	if benefit.Requirement != "" {
		g.addSection("Требования для получения", benefit.Requirement)
	}

	// Как получить
	if benefit.HowToUse != nil && *benefit.HowToUse != "" {
		g.addSection("Как получить", *benefit.HowToUse)
	}

	// Период действия
	validFrom := "Не указано"
	validTo := "Не указано"
	if benefit.ValidFrom != nil {
		validFrom = benefit.ValidFrom.Format("02.01.2006")
	}
	if benefit.ValidTo != nil {
		validTo = benefit.ValidTo.Format("02.01.2006")
	}
	g.addSection("Период действия", fmt.Sprintf("С %s по %s", validFrom, validTo))

	// Источник
	if benefit.SourceURL != "" {
		g.addSection("Источник информации", benefit.SourceURL)
	}

	// Футер
	g.addFooter()

	// Получаем bytes
	var buf bytes.Buffer
	_, err := g.pdf.WriteTo(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to output PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// addHeader добавляет заголовок документа
func (g *Generator) addHeader() {
	// Синий прямоугольник
	g.pdf.SetFillColor(59, 130, 246)
	g.pdf.RectFromUpperLeftWithStyle(0, 0, 595, 70, "F")

	// Текст заголовка
	if g.hasFont {
		g.pdf.SetTextColor(255, 255, 255)
		g.pdf.SetFont(g.fontName, "", 24)
		g.pdf.SetX(50)
		g.pdf.SetY(30)
		g.pdf.Cell(nil, "ЛЬГОТА")
		// Сбрасываем цвет текста
		g.pdf.SetTextColor(0, 0, 0)
	}
}

// addSection добавляет секцию с заголовком и текстом
func (g *Generator) addSection(title, content string) {
	if !g.hasFont {
		return // Если нет шрифта, пропускаем
	}

	currentY := g.pdf.GetY() + 20

	// Проверяем, не выходим ли за пределы страницы
	if currentY > 750 {
		g.pdf.AddPage()
		currentY = 50
	}

	g.pdf.SetY(currentY)
	g.pdf.SetX(50)

	// Заголовок секции
	g.pdf.SetFont(g.fontName, "", 14)
	g.pdf.SetTextColor(0, 0, 0)
	g.pdf.Cell(nil, title)

	// Контент секции
	g.pdf.SetY(g.pdf.GetY() + 18)
	g.pdf.SetX(50)
	g.pdf.SetFont(g.fontName, "", 11)
	g.pdf.SetTextColor(50, 50, 50)

	// Разбиваем длинный текст на строки
	rect := &gopdf.Rect{W: 500, H: 15}
	g.pdf.MultiCell(rect, content)
}

// addFooter добавляет футер
func (g *Generator) addFooter() {
	if !g.hasFont {
		return // Если нет шрифта, пропускаем
	}

	g.pdf.SetY(780)
	g.pdf.SetX(50)
	g.pdf.SetFont(g.fontName, "", 9)
	g.pdf.SetTextColor(150, 150, 150)
	dateStr := time.Now().Format("02.01.2006")
	g.pdf.Cell(nil, fmt.Sprintf("Документ создан %s", dateStr))
}

// getLevelText возвращает текстовое представление уровня льготы
func (g *Generator) getLevelText(level domain.BenefitLevel) string {
	switch level {
	case domain.Federal:
		return "Федеральный"
	case domain.Regional:
		return "Региональный"
	case domain.Commercial:
		return "Коммерческий"
	default:
		return string(level)
	}
}

// getTargetGroupText возвращает текстовое представление целевой группы
func (g *Generator) getTargetGroupText(tg domain.TargetGroup) string {
	switch tg {
	case domain.Pensioners:
		return "Пенсионеры"
	case domain.Disabled:
		return "Инвалиды"
	case domain.YoungFamilies:
		return "Молодые семьи"
	case domain.LowIncome:
		return "Малоимущие"
	case domain.Students:
		return "Студенты"
	case domain.LargeFamilies:
		return "Многодетные семьи"
	case domain.Children:
		return "Дети"
	case domain.Veterans:
		return "Ветераны"
	default:
		return string(tg)
	}
}

// getCategoryText возвращает текстовое представление категории
func (g *Generator) getCategoryText(cat domain.Category) string {
	switch cat {
	case domain.Medicine:
		return "Медицина"
	case domain.Transport:
		return "Транспорт"
	case domain.Food:
		return "Продукты питания"
	case domain.Clothing:
		return "Одежда"
	case domain.Education:
		return "Образование"
	case domain.Payments:
		return "Платежи"
	case domain.Other:
		return "Прочее"
	default:
		return string(cat)
	}
}

// GenerateUserCertificatePDF генерирует PDF-документ удостоверения для любой социальной группы
func (g *Generator) GenerateUserCertificatePDF(user *domain.User, groupType domain.GroupType) ([]byte, error) {
	// Проверяем, загружен ли шрифт
	if !g.hasFont {
		return nil, fmt.Errorf("TTF font not loaded. Font should be at /app/fonts/DejaVuSans.ttf (production) or ./fonts/DejaVuSans.ttf (development)")
	}

	// Получаем заголовок и дополнительную информацию в зависимости от типа группы
	title, additionalInfo := g.getCertificateTitleAndInfo(groupType)

	// Добавляем страницу
	g.pdf.AddPage()

	// Устанавливаем шрифт
	if err := g.pdf.SetFont(g.fontName, "", 14); err != nil {
		return nil, fmt.Errorf("failed to set font: %w", err)
	}

	// Заголовок
	g.pdf.SetX(50)
	g.pdf.SetY(50)
	if err := g.pdf.SetFont(g.fontName, "", 20); err != nil {
		return nil, err
	}
	g.pdf.Cell(nil, title)

	// Разделитель
	g.pdf.SetY(80)
	g.pdf.SetX(50)
	g.pdf.Line(50, 75, 550, 75)

	// Информация о пенсионере
	if err := g.pdf.SetFont(g.fontName, "", 14); err != nil {
		return nil, err
	}

	currentY := 100.0

	// Фамилия
	g.pdf.SetY(currentY)
	g.pdf.SetX(80)
	g.pdf.Cell(nil, "Фамилия:")
	g.pdf.SetX(200)
	if user.LastName.Valid {
		g.pdf.Cell(nil, user.LastName.String)
	}
	currentY += 30

	// Имя
	g.pdf.SetY(currentY)
	g.pdf.SetX(80)
	g.pdf.Cell(nil, "Имя:")
	g.pdf.SetX(200)
	if user.FirstName.Valid {
		g.pdf.Cell(nil, user.FirstName.String)
	}
	currentY += 30

	// Отчество
	g.pdf.SetY(currentY)
	g.pdf.SetX(80)
	g.pdf.Cell(nil, "Отчество:")
	g.pdf.SetX(200)
	if user.MiddleName.Valid {
		g.pdf.Cell(nil, user.MiddleName.String)
	}
	currentY += 30

	// СНИЛС
	g.pdf.SetY(currentY)
	g.pdf.SetX(80)
	g.pdf.Cell(nil, "СНИЛС:")
	g.pdf.SetX(200)
	if user.SNILS.Valid {
		g.pdf.Cell(nil, user.SNILS.String)
	}
	currentY += 30

	// Номер удостоверения (используем ID пользователя)
	g.pdf.SetY(currentY)
	g.pdf.SetX(80)
	g.pdf.Cell(nil, "Номер удостоверения:")
	g.pdf.SetX(280)
	g.pdf.Cell(nil, user.ID.String()[:8])
	currentY += 50

	// Дата выдачи
	g.pdf.SetY(currentY)
	g.pdf.SetX(80)
	g.pdf.Cell(nil, "Дата выдачи:")
	g.pdf.SetX(200)

	// Ищем подтвержденную группу пенсионеров
	var issueDate time.Time
	for _, group := range user.GroupType {
		if group.Type == domain.UserGroupPensioners && group.Status == domain.VerificationStatusVerified {
			if group.VerifiedAt != nil {
				issueDate = *group.VerifiedAt
			}
			break
		}
	}

	if issueDate.IsZero() {
		issueDate = time.Now()
	}

	g.pdf.Cell(nil, issueDate.Format("02.01.2006"))
	currentY += 50

	// Дополнительная информация (если есть)
	if additionalInfo != "" {
		g.pdf.SetY(currentY)
		g.pdf.SetX(80)
		if err := g.pdf.SetFont(g.fontName, "", 12); err != nil {
			return nil, err
		}
		rect := &gopdf.Rect{W: 450, H: 15}
		g.pdf.MultiCell(rect, additionalInfo)
		currentY = g.pdf.GetY() + 30
	}

	// Футер
	g.pdf.SetY(currentY + 20)
	g.pdf.SetX(50)
	g.pdf.Line(50, currentY+15, 550, currentY+15)

	g.pdf.SetY(currentY + 30)
	g.pdf.SetX(80)
	if err := g.pdf.SetFont(g.fontName, "", 10); err != nil {
		return nil, err
	}
	g.pdf.Cell(nil, "Документ действителен на территории Российской Федерации")

	// Возвращаем PDF в виде байтов
	var buf bytes.Buffer
	if _, err := g.pdf.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("failed to output PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// getCertificateTitleAndInfo возвращает заголовок и дополнительную информацию для сертификата в зависимости от типа группы
func (g *Generator) getCertificateTitleAndInfo(groupType domain.GroupType) (string, string) {
	switch groupType {
	case domain.UserGroupPensioners:
		return "УДОСТОВЕРЕНИЕ ПЕНСИОНЕРА",
			"Предоставляет право на получение льгот и социальных выплат, предусмотренных для пенсионеров."
	case domain.UserGroupDisabled:
		return "СПРАВКА ОБ ИНВАЛИДНОСТИ",
			"Подтверждает статус инвалида и право на получение социальных льгот и мер поддержки."
	case domain.UserGroupStudents:
		return "СТУДЕНЧЕСКИЙ БИЛЕТ",
			"Подтверждает статус студента и право на получение студенческих льгот и скидок."
	case domain.UserGroupYoungFamilies:
		return "УДОСТОВЕРЕНИЕ МОЛОДОЙ СЕМЬИ",
			"Подтверждает статус молодой семьи и право на получение мер социальной поддержки."
	case domain.UserGroupLargeFamilies:
		return "УДОСТОВЕРЕНИЕ МНОГОДЕТНОЙ СЕМЬИ",
			"Подтверждает статус многодетной семьи и право на получение льгот и мер поддержки."
	case domain.UserGroupLowIncome:
		return "СПРАВКА О ПРИЗНАНИИ МАЛОИМУЩИМ",
			"Подтверждает статус малоимущего гражданина и право на получение социальной помощи."
	case domain.UserGroupChildren:
		return "СВИДЕТЕЛЬСТВО О РОЖДЕНИИ",
			"Подтверждает право на получение детских пособий и мер поддержки семей с детьми."
	case domain.UserGroupVeterans:
		return "УДОСТОВЕРЕНИЕ ВЕТЕРАНА",
			"Подтверждает статус ветерана и право на получение льгот и мер социальной поддержки."
	default:
		return "УДОСТОВЕРЕНИЕ",
			"Подтверждает право на получение социальных льгот и мер поддержки."
	}
}

// GeneratePensionerCertificatePDF генерирует PDF-документ удостоверения пенсионера (для обратной совместимости)
func (g *Generator) GeneratePensionerCertificatePDF(user *domain.User) ([]byte, error) {
	return g.GenerateUserCertificatePDF(user, domain.UserGroupPensioners)
}
