package pdf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
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

	// Получаем текущую рабочую директорию
	wd, _ := os.Getwd()

	// Пробуем добавить TTF шрифт для кириллицы
	// Используем несколько путей для поиска шрифта
	fontPaths := []string{
		filepath.Join(wd, "fonts", "DejaVuSans.ttf"),
		filepath.Join(wd, "backend", "fonts", "DejaVuSans.ttf"),
		"./fonts/DejaVuSans.ttf",
		"./backend/fonts/DejaVuSans.ttf",
		"fonts/DejaVuSans.ttf",
		"backend/fonts/DejaVuSans.ttf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/System/Library/Fonts/Supplemental/Arial Unicode.ttf",
		"/Library/Fonts/Arial Unicode.ttf",
	}

	hasFont := false
	fontName := "dejavu"
	loadedPath := ""

	for _, path := range fontPaths {
		// Проверяем, существует ли файл
		if _, err := os.Stat(path); err == nil {
			err := pdf.AddTTFFont(fontName, path)
			if err == nil {
				hasFont = true
				loadedPath = path
				break
			}
		}
	}

	// Логируем результат для отладки
	if hasFont {
		fmt.Printf("✅ PDF: Font loaded from %s\n", loadedPath)
	} else {
		fmt.Printf("⚠️  PDF: Font not found. Searched in: %v\nCurrent working directory: %s\n", fontPaths, wd)
	}

	// Если не удалось загрузить TTF, используем встроенный шрифт
	if !hasFont {
		// gopdf поддерживает встроенные шрифты, но они не поддерживают кириллицу
		// В этом случае будем использовать транслитерацию
		fontName = ""
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
		return nil, fmt.Errorf("TTF font not loaded. Please ensure DejaVuSans.ttf is in ./fonts/ directory")
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
