package main

import (
	"encoding/csv"
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

type RawPalmOilData struct {
	Year           int
	Region         string
	RegionID       string
	ParentRegion   string
	ParentRegionID string
	PlantedArea    float64
}

type ProvinceModel struct {
	Province             string
	TotalArea2003        float64
	TotalArea2022        float64
	GrowthRate20Years    float64
	AnnualGrowthRate     float64
	MarketShare2022      float64
	Rank2022             int
	Trend                string
	ProductionEfficiency float64
	Competitiveness      float64
	InvestmentPotential  string
	RiskLevel            string
	Recommendations      []string
	Projection2030       float64
	PeakYear             int
	PeakArea             float64
	StabilityIndex       float64
	YearlyData           map[int]float64
	GrowthPhases         []GrowthPhase
	DominantPeriod       string
}

type GrowthPhase struct {
	Period      string
	GrowthRate  float64
	Description string
}

type NationalTrend struct {
	Year            int
	TotalArea       float64
	GrowthRate      float64
	TopProvince     string
	TopProvinceArea float64
	NewProvinces    []string
	AnnualChange    float64
}

type DecadalAnalysis struct {
	Decade          string
	TotalGrowth     float64
	AverageAnnual   float64
	LeadingProvince string
	EmergingRegions []string
	KeyEvents       []string
}

func main() {
	fmt.Println("🌴 MODEL ANALISIS PROVINSI KELAPA SAWIT INDONESIA 2003-2022")
	fmt.Println("Memproses data 20 tahun...")

	rawData := readCSVData()
	provinceModels := buildProvinceModels(rawData)
	nationalTrends := analyzeNationalTrends(rawData)
	decadalAnalysis := analyzeDecadalTrends(rawData, provinceModels)

	createProvinceAnalysisExcel(provinceModels, nationalTrends, decadalAnalysis)
	createProvinceCharts(provinceModels, nationalTrends)
	createStrategicReport(provinceModels, nationalTrends, decadalAnalysis)

	fmt.Println("\n✅ PEMODELAN PROVINSI 2003-2022 SELESAI!")
	fmt.Println("📁 File Output:")
	fmt.Println("   - model_provinsi_2003_2022.xlsx (Analisis detail per provinsi)")
	fmt.Println("   - peta_heatmap_provinsi_20tahun.png")
	fmt.Println("   - trend_pertumbuhan_provinsi_20tahun.png")
	fmt.Println("   - proyeksi_2030.png")
	fmt.Println("   - matriks_investasi_provinsi_20tahun.png")
	fmt.Println("   - rekomendasi_strategis_provinsi_20tahun.md")
}

func readCSVData() []RawPalmOilData {
	file, err := os.Open("spatial-metrics-indonesia-palm-oil-oil_palm_ha_kabupaten.csv")
	if err != nil {
		log.Fatal("Error membuka file CSV:", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal("Error membaca CSV:", err)
	}

	var data []RawPalmOilData

	for i, record := range records {
		if i == 0 {
			continue
		}

		year, err := strconv.Atoi(record[0])
		if err != nil {
			continue
		}

		if year < 2003 || year > 2022 {
			continue
		}

		area, err := strconv.ParseFloat(record[7], 64)
		if err != nil {
			area = 0
		}

		rawData := RawPalmOilData{
			Year:           year,
			Region:         record[3],
			RegionID:       record[4],
			ParentRegion:   record[5],
			ParentRegionID: record[6],
			PlantedArea:    area,
		}

		data = append(data, rawData)
	}

	fmt.Printf("📊 Data berhasil dibaca: %d records (2003-2022)\n", len(data))
	return data
}

func buildProvinceModels(rawData []RawPalmOilData) []ProvinceModel {
	provinceMap := make(map[string]*ProvinceModel)
	yearlyProvinceData := make(map[string]map[int]float64)

	for _, data := range rawData {
		province := data.ParentRegion
		if province == "" {
			continue
		}

		if provinceMap[province] == nil {
			provinceMap[province] = &ProvinceModel{
				Province:   province,
				YearlyData: make(map[int]float64),
			}
		}

		if yearlyProvinceData[province] == nil {
			yearlyProvinceData[province] = make(map[int]float64)
		}

		yearlyProvinceData[province][data.Year] += data.PlantedArea
		provinceMap[province].YearlyData[data.Year] += data.PlantedArea
	}

	totalNational2022 := 0.0
	for province := range provinceMap {
		if yearlyProvinceData[province] != nil {
			totalNational2022 += yearlyProvinceData[province][2022]
		}
	}

	var models []ProvinceModel
	for province, yearlyData := range yearlyProvinceData {
		if yearlyData == nil {
			continue
		}

		model := ProvinceModel{
			Province:      province,
			TotalArea2003: yearlyData[2003],
			TotalArea2022: yearlyData[2022],
			YearlyData:    yearlyProvinceData[province],
		}

		if model.TotalArea2003 > 0 {
			model.GrowthRate20Years = ((model.TotalArea2022 - model.TotalArea2003) / model.TotalArea2003) * 100
			model.AnnualGrowthRate = model.GrowthRate20Years / 20
		}

		if totalNational2022 > 0 {
			model.MarketShare2022 = (model.TotalArea2022 / totalNational2022) * 100
		}

		model.Trend = analyzeProvinceTrend20Years(yearlyData)

		model.GrowthPhases = analyzeGrowthPhases(yearlyData)
		model.DominantPeriod = identifyDominantPeriod(model.GrowthPhases)

		model.PeakYear, model.PeakArea = findPeakYearAndArea(yearlyData)

		model.StabilityIndex = calculateStabilityIndex(yearlyData)

		model.ProductionEfficiency = calculateEfficiency(model.TotalArea2022, model.GrowthRate20Years, model.StabilityIndex)

		model.Competitiveness = calculateCompetitivenessScore(model)

		model.InvestmentPotential = assessInvestmentPotential(model)

		model.RiskLevel = assessRiskLevel(model)

		model.Recommendations = generateProvinceRecommendations(model)

		model.Projection2030 = calculateProjection2030(model)

		models = append(models, model)
	}

	sort.Slice(models, func(i, j int) bool {
		return models[i].TotalArea2022 > models[j].TotalArea2022
	})

	for i := range models {
		models[i].Rank2022 = i + 1
	}

	fmt.Printf("🏛️  Model provinsi dibangun: %d provinsi (2003-2022)\n", len(models))
	return models
}

func analyzeNationalTrends(rawData []RawPalmOilData) []NationalTrend {
	yearlyData := make(map[int]*NationalTrend)

	for year := 2003; year <= 2022; year++ {
		yearlyData[year] = &NationalTrend{
			Year:         year,
			TotalArea:    0,
			GrowthRate:   0,
			TopProvince:  "Unknown",
			AnnualChange: 0,
		}
	}

	for _, data := range rawData {
		if yearlyData[data.Year] != nil {
			yearlyData[data.Year].TotalArea += data.PlantedArea
		}
	}

	for year := 2003; year <= 2022; year++ {
		trend := yearlyData[year]
		if trend == nil {
			continue
		}

		provinceAreas := make(map[string]float64)
		for _, data := range rawData {
			if data.Year == year && data.ParentRegion != "" {
				provinceAreas[data.ParentRegion] += data.PlantedArea
			}
		}

		var topProvince string
		var topArea float64
		for province, area := range provinceAreas {
			if area > topArea {
				topArea = area
				topProvince = province
			}
		}

		if topProvince != "" {
			trend.TopProvince = topProvince
			trend.TopProvinceArea = topArea
		}

		if year > 2003 {
			prevYearData := yearlyData[year-1]
			if prevYearData != nil && prevYearData.TotalArea > 0 {
				growth := ((trend.TotalArea - prevYearData.TotalArea) / prevYearData.TotalArea) * 100
				trend.GrowthRate = growth
				trend.AnnualChange = trend.TotalArea - prevYearData.TotalArea
			}
		}
	}

	var trends []NationalTrend
	for year := 2003; year <= 2022; year++ {
		if yearlyData[year] != nil {
			trends = append(trends, *yearlyData[year])
		}
	}

	return trends
}

func analyzeDecadalTrends(rawData []RawPalmOilData, models []ProvinceModel) []DecadalAnalysis {
	decades := []struct {
		name  string
		start int
		end   int
	}{
		{"2003-2012", 2003, 2012},
		{"2013-2022", 2013, 2022},
	}

	var analysis []DecadalAnalysis

	for _, decade := range decades {
		decadeAnalysis := DecadalAnalysis{
			Decade: decade.name,
		}

		startArea := calculateNationalAreaForYear(rawData, decade.start)
		endArea := calculateNationalAreaForYear(rawData, decade.end)

		if startArea > 0 {
			decadeAnalysis.TotalGrowth = ((endArea - startArea) / startArea) * 100
			decadeAnalysis.AverageAnnual = decadeAnalysis.TotalGrowth / float64(decade.end-decade.start)
		}

		decadeAnalysis.LeadingProvince = findLeadingProvince(models, decade.start, decade.end)

		decadeAnalysis.EmergingRegions = findEmergingRegions(models, decade.start, decade.end)

		decadeAnalysis.KeyEvents = getKeyEventsForDecade(decade.name)

		analysis = append(analysis, decadeAnalysis)
	}

	return analysis
}

func createProvinceAnalysisExcel(models []ProvinceModel, trends []NationalTrend, decadalAnalysis []DecadalAnalysis) {
	f := excelize.NewFile()

	f.SetSheetName("Sheet1", "Dashboard_Provinsi_20Tahun")

	headers := []string{"Rank", "Provinsi", "Area 2022 (ha)", "Area 2003 (ha)",
		"Growth Rate 20 Tahun (%)", "Market Share 2022 (%)", "Trend",
		"Efisiensi Produksi", "Daya Saing", "Potensi Investasi", "Tingkat Risiko",
		"Proyeksi 2030 (ha)", "Tahun Puncak", "Area Puncak (ha)", "Indeks Stabilitas",
		"Periode Dominan", "Rekomendasi Utama"}

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue("Dashboard_Provinsi_20Tahun", cell, header)
		f.SetColWidth("Dashboard_Provinsi_20Tahun", cell, cell, 18)
	}

	for i, model := range models {
		row := i + 2
		f.SetCellValue("Dashboard_Provinsi_20Tahun", fmt.Sprintf("A%d", row), model.Rank2022)
		f.SetCellValue("Dashboard_Provinsi_20Tahun", fmt.Sprintf("B%d", row), model.Province)
		f.SetCellValue("Dashboard_Provinsi_20Tahun", fmt.Sprintf("C%d", row), formatNumber(model.TotalArea2022))
		f.SetCellValue("Dashboard_Provinsi_20Tahun", fmt.Sprintf("D%d", row), formatNumber(model.TotalArea2003))
		f.SetCellValue("Dashboard_Provinsi_20Tahun", fmt.Sprintf("E%d", row), fmt.Sprintf("%.1f%%", model.GrowthRate20Years))
		f.SetCellValue("Dashboard_Provinsi_20Tahun", fmt.Sprintf("F%d", row), fmt.Sprintf("%.1f%%", model.MarketShare2022))
		f.SetCellValue("Dashboard_Provinsi_20Tahun", fmt.Sprintf("G%d", row), model.Trend)
		f.SetCellValue("Dashboard_Provinsi_20Tahun", fmt.Sprintf("H%d", row), fmt.Sprintf("%.1f/10", model.ProductionEfficiency))
		f.SetCellValue("Dashboard_Provinsi_20Tahun", fmt.Sprintf("I%d", row), fmt.Sprintf("%.1f/10", model.Competitiveness))
		f.SetCellValue("Dashboard_Provinsi_20Tahun", fmt.Sprintf("J%d", row), model.InvestmentPotential)
		f.SetCellValue("Dashboard_Provinsi_20Tahun", fmt.Sprintf("K%d", row), model.RiskLevel)
		f.SetCellValue("Dashboard_Provinsi_20Tahun", fmt.Sprintf("L%d", row), formatNumber(model.Projection2030))
		f.SetCellValue("Dashboard_Provinsi_20Tahun", fmt.Sprintf("M%d", row), model.PeakYear)
		f.SetCellValue("Dashboard_Provinsi_20Tahun", fmt.Sprintf("N%d", row), formatNumber(model.PeakArea))
		f.SetCellValue("Dashboard_Provinsi_20Tahun", fmt.Sprintf("O%d", row), fmt.Sprintf("%.2f", model.StabilityIndex))
		f.SetCellValue("Dashboard_Provinsi_20Tahun", fmt.Sprintf("P%d", row), model.DominantPeriod)

		mainRec := "Tidak tersedia"
		if len(model.Recommendations) > 0 {
			mainRec = model.Recommendations[0]
		}
		f.SetCellValue("Dashboard_Provinsi_20Tahun", fmt.Sprintf("Q%d", row), mainRec)
	}

	f.NewSheet("Trend_Nasional_20Tahun")

	trendHeaders := []string{"Tahun", "Total Area (ha)", "Pertumbuhan (%)",
		"Provinsi Teratas", "Area Provinsi Teratas (ha)", "Perubahan Tahunan (ha)"}

	for i, header := range trendHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue("Trend_Nasional_20Tahun", cell, header)
		f.SetColWidth("Trend_Nasional_20Tahun", cell, cell, 20)
	}

	for i, trend := range trends {
		row := i + 2
		f.SetCellValue("Trend_Nasional_20Tahun", fmt.Sprintf("A%d", row), trend.Year)
		f.SetCellValue("Trend_Nasional_20Tahun", fmt.Sprintf("B%d", row), formatNumber(trend.TotalArea))
		f.SetCellValue("Trend_Nasional_20Tahun", fmt.Sprintf("C%d", row), fmt.Sprintf("%.1f%%", trend.GrowthRate))
		f.SetCellValue("Trend_Nasional_20Tahun", fmt.Sprintf("D%d", row), trend.TopProvince)
		f.SetCellValue("Trend_Nasional_20Tahun", fmt.Sprintf("E%d", row), formatNumber(trend.TopProvinceArea))
		f.SetCellValue("Trend_Nasional_20Tahun", fmt.Sprintf("F%d", row), formatNumber(trend.AnnualChange))
	}

	f.NewSheet("Analisis_Dekade")

	f.SetCellValue("Analisis_Dekade", "A1", "ANALISIS PER DEKADE 2003-2022")
	f.SetCellValue("Analisis_Dekade", "A2", "Dekade")
	f.SetCellValue("Analisis_Dekade", "B2", "Total Growth (%)")
	f.SetCellValue("Analisis_Dekade", "C2", "Rata2 Tahunan (%)")
	f.SetCellValue("Analisis_Dekade", "D2", "Provinsi Terdepan")
	f.SetCellValue("Analisis_Dekade", "E2", "Region Emerging")
	f.SetCellValue("Analisis_Dekade", "F2", "Event Penting")

	for i, analysis := range decadalAnalysis {
		row := i + 3
		f.SetCellValue("Analisis_Dekade", fmt.Sprintf("A%d", row), analysis.Decade)
		f.SetCellValue("Analisis_Dekade", fmt.Sprintf("B%d", row), fmt.Sprintf("%.1f%%", analysis.TotalGrowth))
		f.SetCellValue("Analisis_Dekade", fmt.Sprintf("C%d", row), fmt.Sprintf("%.1f%%", analysis.AverageAnnual))
		f.SetCellValue("Analisis_Dekade", fmt.Sprintf("D%d", row), analysis.LeadingProvince)
		f.SetCellValue("Analisis_Dekade", fmt.Sprintf("E%d", row), strings.Join(analysis.EmergingRegions, ", "))
		f.SetCellValue("Analisis_Dekade", fmt.Sprintf("F%d", row), strings.Join(analysis.KeyEvents, "; "))
	}

	f.NewSheet("Kelompok_Provinsi_20Tahun")

	primeProvinces := filterProvinces(models, "PRIME")
	growthProvinces := filterProvinces(models, "GROWTH")
	emergingProvinces := filterProvinces(models, "EMERGING")
	stableProvinces := filterProvinces(models, "STABLE")
	matureProvinces := filterProvinces(models, "MATURE")

	f.SetCellValue("Kelompok_Provinsi_20Tahun", "A1", "KELOMPOK PROVINSI BERDASARKAN POTENSI (2003-2022)")
	f.SetCellValue("Kelompok_Provinsi_20Tahun", "A2", "PRIME (Area > 1M ha, Growth > 100%)")
	writeProvinceGroup(f, primeProvinces, 3)

	startRow := len(primeProvinces) + 5
	f.SetCellValue("Kelompok_Provinsi_20Tahun", fmt.Sprintf("A%d", startRow), "GROWTH (Growth > 200%)")
	writeProvinceGroup(f, growthProvinces, startRow+1)

	startRow += len(growthProvinces) + 3
	f.SetCellValue("Kelompok_Provinsi_20Tahun", fmt.Sprintf("A%d", startRow), "EMERGING (Area < 500k, Growth > 300%)")
	writeProvinceGroup(f, emergingProvinces, startRow+1)

	startRow += len(emergingProvinces) + 3
	f.SetCellValue("Kelompok_Provinsi_20Tahun", fmt.Sprintf("A%d", startRow), "STABLE (Area > 500k, Growth 50-150%)")
	writeProvinceGroup(f, stableProvinces, startRow+1)

	startRow += len(stableProvinces) + 3
	f.SetCellValue("Kelompok_Provinsi_20Tahun", fmt.Sprintf("A%d", startRow), "MATURE (Area besar, Growth < 50%)")
	writeProvinceGroup(f, matureProvinces, startRow+1)

	f.NewSheet("Matriks_Strategi_20Tahun")

	strategicMatrix := [][]string{
		{"Kategori", "Strategi Inti", "Target Provinsi", "Timeline", "Expected Impact"},
		{"PRIME", "Leadership & Innovation", "Riau, Kalimantan Barat, Sumatra Utara", "2023-2025", "Productivity +20%"},
		{"GROWTH", "Sustainable Expansion", "Kalimantan Tengah, Kalimantan Timur", "2023-2027", "Market Share +15%"},
		{"EMERGING", "Strategic Development", "Papua, Sulawesi, Maluku", "2023-2030", "New Growth Centers"},
		{"STABLE", "Optimization & Tech Adoption", "Sumatra Selatan, Jambi, Aceh", "2023-2026", "Efficiency +25%"},
		{"MATURE", "Diversification & Value Add", "Lampung, Jawa Barat", "2023-2028", "Revenue Diversity +30%"},
		{"ALL", "Sustainability & Certification", "Semua Provinsi", "2023-2030", "100% Certified by 2030"},
	}

	for i, row := range strategicMatrix {
		for j, value := range row {
			cell, _ := excelize.CoordinatesToCellName(j+1, i+1)
			f.SetCellValue("Matriks_Strategi_20Tahun", cell, value)
			f.SetColWidth("Matriks_Strategi_20Tahun", cell, cell, 20)
		}
	}

	if err := f.SaveAs("model_provinsi_2003_2022.xlsx"); err != nil {
		log.Fatal("Error menyimpan Excel:", err)
	}

	fmt.Printf("📈 File Excel berhasil dibuat: model_provinsi_2003_2022.xlsx (%d provinsi)\n", len(models))
}

func createProvinceCharts(models []ProvinceModel, trends []NationalTrend) {
	createProvinceHeatmap20Years(models)
	createGrowthTrendChart20Years(models)
	createProjectionChart2030(models)
	createInvestmentScatterPlot20Years(models)
	createNationalTrendChart(trends)
}

func createProvinceHeatmap20Years(models []ProvinceModel) {
	p := plot.New()
	p.Title.Text = "PETA SEBARAN KELAPA SAWIT INDONESIA 2003-2022"
	p.Title.TextStyle.Font.Size = vg.Points(16)
	p.X.Label.Text = "Market Share 2022 (%)"
	p.Y.Label.Text = "Growth Rate 20 Tahun (%)"

	maxMarketShare := getMaxMarketShare(models)
	maxGrowthRate := getMaxGrowthRate(models)
	minGrowthRate := getMinGrowthRate(models)

	p.X.Max = maxMarketShare * 1.2
	p.Y.Max = math.Max(maxGrowthRate*1.2, 50)
	p.Y.Min = math.Min(minGrowthRate*1.2, -20)

	points := make(plotter.XYs, len(models))
	labels := make([]string, len(models))

	for i, province := range models {
		points[i].X = province.MarketShare2022
		points[i].Y = province.GrowthRate20Years
		labels[i] = getShortProvinceName(province.Province)
	}

	for i := range models {
		individualBubble, err := plotter.NewScatter(plotter.XYs{points[i]})
		if err != nil {
			log.Fatal(err)
		}

		province := models[i]
		if province.GrowthRate20Years > 200 {
			individualBubble.GlyphStyle.Color = color.RGBA{R: 0, G: 100, B: 0, A: 255}
		} else if province.GrowthRate20Years > 100 {
			individualBubble.GlyphStyle.Color = color.RGBA{R: 34, G: 139, B: 34, A: 255}
		} else if province.GrowthRate20Years > 50 {
			individualBubble.GlyphStyle.Color = color.RGBA{R: 173, G: 255, B: 47, A: 255}
		} else if province.GrowthRate20Years > 0 {
			individualBubble.GlyphStyle.Color = color.RGBA{R: 255, G: 255, B: 0, A: 255}
		} else {
			individualBubble.GlyphStyle.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255}
		}

		radius := vg.Points(4)
		if province.MarketShare2022 > 10 {
			radius = vg.Points(12)
		} else if province.MarketShare2022 > 5 {
			radius = vg.Points(9)
		} else if province.MarketShare2022 > 2 {
			radius = vg.Points(6)
		} else {
			radius = vg.Points(4)
		}
		individualBubble.GlyphStyle.Radius = radius
		individualBubble.GlyphStyle.Shape = draw.CircleGlyph{}

		p.Add(individualBubble)
	}

	labelPoints, err := plotter.NewLabels(plotter.XYLabels{
		XYs:    points,
		Labels: labels,
	})
	if err != nil {
		log.Fatal(err)
	}

	p.Add(labelPoints)

	p.Add(plotter.NewGrid())

	if err := p.Save(20*vg.Inch, 16*vg.Inch, "peta_heatmap_provinsi_20tahun.png"); err != nil {
		log.Fatal(err)
	}
}

func createGrowthTrendChart20Years(models []ProvinceModel) {
	p := plot.New()
	p.Title.Text = "TREND PERTUMBUHAN PROVINSI 2003-2022 (20 TAHUN)"
	p.Title.TextStyle.Font.Size = vg.Points(16)
	p.X.Label.Text = "Provinsi"
	p.Y.Label.Text = "Growth Rate 20 Tahun (%)"

	values := make(plotter.Values, len(models))
	labels := make([]string, len(models))

	for i, province := range models {
		values[i] = province.GrowthRate20Years
		labels[i] = getShortProvinceName(province.Province)
	}

	bars, err := plotter.NewBarChart(values, vg.Points(20))
	if err != nil {
		log.Fatal(err)
	}

	bars.Color = color.RGBA{R: 70, G: 130, B: 180, A: 255}
	bars.LineStyle.Width = vg.Length(0)

	p.Add(bars)

	p.NominalX(labels...)
	p.X.Tick.Label.Rotation = math.Pi / 3
	p.X.Tick.Label.YAlign = draw.YCenter
	p.X.Tick.Label.XAlign = draw.XCenter

	minGrowth := getMinGrowthRate(models)
	maxGrowth := getMaxGrowthRate(models)

	p.Y.Min = minGrowth - math.Abs(minGrowth)*0.15
	if p.Y.Min > 0 {
		p.Y.Min = 0
	}
	p.Y.Max = maxGrowth * 1.15

	for i, val := range values {
		if val > 0 {
			x := float64(i)
			y := val + (maxGrowth * 0.02)
			label, _ := plotter.NewLabels(plotter.XYLabels{
				XYs:    []plotter.XY{{X: x, Y: y}},
				Labels: []string{fmt.Sprintf("%.0f%%", val)},
			})
			p.Add(label)
		}
	}

	if err := p.Save(24*vg.Inch, 12*vg.Inch, "trend_pertumbuhan_provinsi_20tahun.png"); err != nil {
		log.Fatal(err)
	}
}

func createProjectionChart2030(models []ProvinceModel) {
	p := plot.New()
	p.Title.Text = "PROYEKSI AREA KELAPA SAWIT 2030 vs 2022"
	p.Title.TextStyle.Font.Size = vg.Points(16)
	p.X.Label.Text = "Area 2022 (juta ha)"
	p.Y.Label.Text = "Proyeksi 2030 (juta ha)"

	points := make(plotter.XYs, len(models))
	labels := make([]string, len(models))

	maxArea2022 := 0.0
	maxProjection := 0.0

	for i, province := range models {
		area2022 := province.TotalArea2022 / 1000000
		projection := province.Projection2030 / 1000000

		points[i].X = area2022
		points[i].Y = projection
		labels[i] = getShortProvinceName(province.Province)

		if area2022 > maxArea2022 {
			maxArea2022 = area2022
		}
		if projection > maxProjection {
			maxProjection = projection
		}
	}

	scatter, err := plotter.NewScatter(points)
	if err != nil {
		log.Fatal(err)
	}

	scatter.GlyphStyle.Color = color.RGBA{R: 139, G: 0, B: 0, A: 255}
	scatter.GlyphStyle.Radius = vg.Points(4)

	p.Add(scatter)
	p.Add(plotter.NewGrid())

	line := plotter.NewFunction(func(x float64) float64 { return x })
	line.Color = color.RGBA{R: 0, G: 0, B: 0, A: 255}
	line.Dashes = []vg.Length{vg.Points(5), vg.Points(5)}
	p.Add(line)

	maxValue := math.Max(maxArea2022, maxProjection) * 1.2
	p.X.Max = maxValue
	p.Y.Max = maxValue
	p.X.Min = 0
	p.Y.Min = 0

	labelPoints, err := plotter.NewLabels(plotter.XYLabels{
		XYs:    points,
		Labels: labels,
	})
	if err != nil {
		log.Fatal(err)
	}
	p.Add(labelPoints)

	if err := p.Save(20*vg.Inch, 16*vg.Inch, "proyeksi_2030.png"); err != nil {
		log.Fatal(err)
	}
}

func createNationalTrendChart(trends []NationalTrend) {
	p := plot.New()
	p.Title.Text = "TREND NASIONAL KELAPA SAWIT INDONESIA 2003-2022"
	p.Title.TextStyle.Font.Size = vg.Points(16)
	p.X.Label.Text = "Tahun"
	p.Y.Label.Text = "Total Area (juta ha)"

	points := make(plotter.XYs, len(trends))
	for i, trend := range trends {
		points[i].X = float64(trend.Year)
		points[i].Y = trend.TotalArea / 1000000
	}

	line, err := plotter.NewLine(points)
	if err != nil {
		log.Fatal(err)
	}
	line.Color = color.RGBA{R: 0, G: 100, B: 0, A: 255}
	line.Width = vg.Points(2)

	p.Add(line)
	p.Add(plotter.NewGrid())

	yearLabels := make([]string, len(trends))
	for i, trend := range trends {
		yearLabels[i] = fmt.Sprintf("%d", trend.Year)
	}
	p.NominalX(yearLabels...)

	if err := p.Save(16*vg.Inch, 8*vg.Inch, "trend_nasional_20tahun.png"); err != nil {
		log.Fatal(err)
	}
}

func createInvestmentScatterPlot20Years(models []ProvinceModel) {
	p := plot.New()
	p.Title.Text = "MATRIKS POTENSI INVESTASI PROVINSI 2003-2022"
	p.Title.TextStyle.Font.Size = vg.Points(16)
	p.X.Label.Text = "Daya Saing (0-10)"
	p.Y.Label.Text = "Growth Rate 20 Tahun (%)"

	points := make(plotter.XYs, len(models))
	labels := make([]string, len(models))

	for i, province := range models {
		points[i].X = province.Competitiveness
		points[i].Y = province.GrowthRate20Years
		labels[i] = getShortProvinceName(province.Province)
	}

	for i := range models {
		individualPoint, err := plotter.NewScatter(plotter.XYs{points[i]})
		if err != nil {
			log.Fatal(err)
		}

		province := models[i]

		switch province.InvestmentPotential {
		case "VERY HIGH":
			individualPoint.GlyphStyle.Color = color.RGBA{R: 0, G: 100, B: 0, A: 255}
		case "HIGH":
			individualPoint.GlyphStyle.Color = color.RGBA{R: 34, G: 139, B: 34, A: 255}
		case "MEDIUM":
			individualPoint.GlyphStyle.Color = color.RGBA{R: 255, G: 165, B: 0, A: 255}
		case "LOW":
			individualPoint.GlyphStyle.Color = color.RGBA{R: 255, G: 69, B: 0, A: 255}
		case "VERY LOW":
			individualPoint.GlyphStyle.Color = color.RGBA{R: 220, G: 20, B: 60, A: 255}
		default:
			individualPoint.GlyphStyle.Color = color.RGBA{R: 65, G: 105, B: 225, A: 255}
		}

		radius := vg.Points(6)
		if province.MarketShare2022 > 10 {
			radius = vg.Points(12)
		} else if province.MarketShare2022 > 5 {
			radius = vg.Points(9)
		} else if province.MarketShare2022 > 2 {
			radius = vg.Points(7)
		}
		individualPoint.GlyphStyle.Radius = radius

		p.Add(individualPoint)
	}

	labelPoints, err := plotter.NewLabels(plotter.XYLabels{
		XYs:    points,
		Labels: labels,
	})
	if err != nil {
		log.Fatal(err)
	}
	p.Add(labelPoints)

	p.Add(plotter.NewGrid())

	p.X.Min = 0
	p.X.Max = 10.5
	p.Y.Min = getMinGrowthRate(models) * 0.9
	p.Y.Max = getMaxGrowthRate(models) * 1.1

	if err := p.Save(20*vg.Inch, 16*vg.Inch, "matriks_investasi_provinsi_20tahun.png"); err != nil {
		log.Fatal(err)
	}
}

func createStrategicReport(models []ProvinceModel, trends []NationalTrend, decadalAnalysis []DecadalAnalysis) {
	file, err := os.Create("rekomendasi_strategis_provinsi_20tahun.md")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	report := `# LAPORAN STRATEGIS KELAPA SAWIT INDONESIA
## Analisis Berbasis Provinsi 2003-2022

### 📊 EXECUTIVE SUMMARY 20 TAHUN

`

	totalProvinces := len(models)
	highGrowthProvinces := countProvincesByGrowth(models, 100)
	primeProvinces := len(filterProvinces(models, "PRIME"))

	report += fmt.Sprintf("- **Total Provinsi Analyzed**: %d\n", totalProvinces)
	report += fmt.Sprintf("- **Provinsi Growth >100%% (20 tahun)**: %d\n", highGrowthProvinces)
	report += fmt.Sprintf("- **Prime Provinces**: %d\n", primeProvinces)

	if len(trends) > 0 {
		firstYear := trends[0]
		lastYear := trends[len(trends)-1]
		totalGrowth := ((lastYear.TotalArea - firstYear.TotalArea) / firstYear.TotalArea) * 100

		report += fmt.Sprintf("- **Total Area 2003**: %s ha\n", formatNumber(firstYear.TotalArea))
		report += fmt.Sprintf("- **Total Area 2022**: %s ha\n", formatNumber(lastYear.TotalArea))
		report += fmt.Sprintf("- **Total Growth 20 Tahun**: %.1f%%\n", totalGrowth)
	}
	report += fmt.Sprintf("- **Average Growth Rate 20 Tahun**: %.1f%%\n", calculateAverageGrowth(models))

	if len(decadalAnalysis) >= 2 {
		report += "\n### 📈 ANALISIS PER DEKADE\n\n"
		report += "| Dekade | Total Growth | Rata2 Tahunan | Provinsi Terdepan |\n"
		report += "|--------|--------------|---------------|-------------------|\n"
		for _, analysis := range decadalAnalysis {
			report += fmt.Sprintf("| %s | %.1f%% | %.1f%% | %s |\n",
				analysis.Decade, analysis.TotalGrowth, analysis.AverageAnnual, analysis.LeadingProvince)
		}
	}

	report += "\n### 📋 DATA SEMUA PROVINSI (2003-2022)\n\n"
	report += "| Rank | Provinsi | Area 2022 (ha) | Growth 20 Tahun | Market Share | Potensi Investasi | Periode Dominan |\n"
	report += "|------|----------|----------------|-----------------|--------------|-------------------|-----------------|\n"

	for _, model := range models {
		report += fmt.Sprintf("| %d | %s | %s | %.0f%% | %.1f%% | %s | %s |\n",
			model.Rank2022,
			model.Province,
			formatNumber(model.TotalArea2022),
			model.GrowthRate20Years,
			model.MarketShare2022,
			model.InvestmentPotential,
			model.DominantPeriod)
	}

	report += "\n### 🎯 KELOMPOK PROVINSI BERDASARKAN KINERJA 20 TAHUN\n"

	categories := []struct {
		name     string
		category string
	}{
		{"PRIME", "PRIME"},
		{"GROWTH", "GROWTH"},
		{"EMERGING", "EMERGING"},
		{"STABLE", "STABLE"},
		{"MATURE", "MATURE"},
	}

	for _, cat := range categories {
		provinces := filterProvinces(models, cat.category)
		if len(provinces) > 0 {
			report += fmt.Sprintf("\n#### %s (%d provinsi)\n", cat.name, len(provinces))
			for _, province := range provinces {
				report += fmt.Sprintf("- **%s**: Area %s ha, Growth %.0f%%, %s\n",
					province.Province, formatNumber(province.TotalArea2022),
					province.GrowthRate20Years, strings.Join(province.Recommendations[:1], ", "))
			}
		}
	}

	report += `
### 🚀 REKOMENDASI STRATEGIS 2023-2030

#### 1. OPTIMISASI PROVINSI PRIME
- **Fokus**: Provinsi dengan area >1 juta ha dan growth >100%
- **Strategi**: Technology leadership, precision agriculture
- **Target**: Productivity improvement 20-30%

#### 2. AKSELERASI PROVINSI GROWTH  
- **Fokus**: Provinsi dengan growth >200% dalam 20 tahun
- **Strategi**: Sustainable expansion dengan circular economy
- **Target**: Market share increase 15-25%

#### 3. PENGEMBANGAN PROVINSI EMERGING
- **Fokus**: Region baru dengan potensi tinggi
- **Strategi**: Integrated plantation development
- **Target**: Establish new sustainable growth centers

#### 4. TRANSFORMASI PROVINSI MATURE
- **Fokus**: Provinsi dengan growth rendah tapi area besar
- **Strategi**: Diversification dan value-added products
- **Target**: Revenue diversification 30-40%

#### 5. SUSTAINABILITY ROADMAP 2030
- **Scope**: Semua provinsi
- **Strategi**: ISPO/RSPO certification, NDPE compliance
- **Target**: 100% sustainable certification by 2030

### 📅 ROADMAP IMPLEMENTASI 2023-2030

**2023-2025**: 
- Digital transformation di provinsi prime
- Penyusunan masterplan sustainability
- Pilot projects di provinsi emerging

**2026-2028**:
- Scale up sustainable practices
- Technology adoption massal
- Market diversification

**2029-2030**:
- Full certification implementation
- Evaluation dan adjustment
- Preparation untuk fase berikutnya

---
*Generated by Palm Oil Analytics System - %s*
`

	report = fmt.Sprintf(report, time.Now().Format("2 January 2006"))

	file.WriteString(report)
	fmt.Println("📋 Laporan strategis 20 tahun berhasil dibuat: rekomendasi_strategis_provinsi_20tahun.md")
}

func analyzeProvinceTrend20Years(yearlyData map[int]float64) string {
	if len(yearlyData) < 4 {
		return "INSUFFICIENT_DATA"
	}

	startArea := yearlyData[2003]
	endArea := yearlyData[2022]

	if startArea == 0 || endArea == 0 {
		return "INCOMPLETE_DATA"
	}

	totalGrowth := ((endArea - startArea) / startArea) * 100

	growthRates := calculateYearlyGrowthRates(yearlyData)
	volatility := calculateVolatility(growthRates)

	if totalGrowth > 500 {
		return "EXPLOSIVE_GROWTH"
	} else if totalGrowth > 200 {
		return "HIGH_GROWTH"
	} else if totalGrowth > 100 {
		return "MODERATE_GROWTH"
	} else if totalGrowth > 50 {
		return "STABLE_GROWTH"
	} else if totalGrowth < 0 {
		return "DECLINING"
	} else if volatility > 30 {
		return "VOLATILE"
	}
	return "MATURE"
}

func analyzeGrowthPhases(yearlyData map[int]float64) []GrowthPhase {
	var phases []GrowthPhase

	phase1 := calculatePhaseGrowth(yearlyData, 2003, 2007)
	phases = append(phases, GrowthPhase{
		Period:      "2003-2007",
		GrowthRate:  phase1,
		Description: getPhaseDescription(phase1, "Early"),
	})

	phase2 := calculatePhaseGrowth(yearlyData, 2008, 2012)
	phases = append(phases, GrowthPhase{
		Period:      "2008-2012",
		GrowthRate:  phase2,
		Description: getPhaseDescription(phase2, "Mid"),
	})

	phase3 := calculatePhaseGrowth(yearlyData, 2013, 2017)
	phases = append(phases, GrowthPhase{
		Period:      "2013-2017",
		GrowthRate:  phase3,
		Description: getPhaseDescription(phase3, "Recent"),
	})

	phase4 := calculatePhaseGrowth(yearlyData, 2018, 2022)
	phases = append(phases, GrowthPhase{
		Period:      "2018-2022",
		GrowthRate:  phase4,
		Description: getPhaseDescription(phase4, "Current"),
	})

	return phases
}

func identifyDominantPeriod(phases []GrowthPhase) string {
	if len(phases) == 0 {
		return "UNKNOWN"
	}

	maxGrowth := 0.0
	dominantPeriod := ""

	for _, phase := range phases {
		if phase.GrowthRate > maxGrowth {
			maxGrowth = phase.GrowthRate
			dominantPeriod = phase.Period
		}
	}

	return dominantPeriod
}

func findPeakYearAndArea(yearlyData map[int]float64) (int, float64) {
	peakYear := 2003
	peakArea := yearlyData[2003]

	for year, area := range yearlyData {
		if area > peakArea {
			peakArea = area
			peakYear = year
		}
	}

	return peakYear, peakArea
}

func calculateStabilityIndex(yearlyData map[int]float64) float64 {
	growthRates := calculateYearlyGrowthRates(yearlyData)
	if len(growthRates) == 0 {
		return 5.0
	}

	mean := 0.0
	for _, gr := range growthRates {
		mean += gr
	}
	mean /= float64(len(growthRates))

	variance := 0.0
	for _, gr := range growthRates {
		variance += math.Pow(gr-mean, 2)
	}
	variance /= float64(len(growthRates))
	stdDev := math.Sqrt(variance)

	stability := 10.0 - math.Min(stdDev/10, 5.0)
	return math.Max(1.0, math.Min(10.0, stability))
}

func calculateEfficiency(area, growth, stability float64) float64 {
	baseEfficiency := 5.0

	if area > 1000000 {
		baseEfficiency += 2.0
	} else if area > 500000 {
		baseEfficiency += 1.0
	}

	if growth > 200 {
		baseEfficiency += 1.5
	} else if growth > 100 {
		baseEfficiency += 1.0
	}

	baseEfficiency += (stability - 5.0) / 2

	return math.Min(10.0, baseEfficiency)
}

func calculateCompetitivenessScore(model ProvinceModel) float64 {
	score := 5.0

	score += model.MarketShare2022 / 10
	score += model.GrowthRate20Years / 100
	score += model.ProductionEfficiency / 2
	score += model.StabilityIndex / 2

	return math.Min(10.0, score)
}

func assessInvestmentPotential(model ProvinceModel) string {
	score := model.Competitiveness

	if score >= 8.0 {
		return "VERY HIGH"
	} else if score >= 6.5 {
		return "HIGH"
	} else if score >= 5.0 {
		return "MEDIUM"
	} else if score >= 3.0 {
		return "LOW"
	}
	return "VERY LOW"
}

func assessRiskLevel(model ProvinceModel) string {
	if model.GrowthRate20Years > 500 {
		return "HIGH"
	} else if model.StabilityIndex < 4 {
		return "HIGH"
	} else if model.GrowthRate20Years < 0 {
		return "MEDIUM-HIGH"
	} else if model.MarketShare2022 < 1 && model.GrowthRate20Years < 50 {
		return "MEDIUM"
	}
	return "LOW-MEDIUM"
}

func generateProvinceRecommendations(model ProvinceModel) []string {
	var recs []string

	if model.MarketShare2022 > 15 {
		recs = append(recs, "Maintain market leadership through innovation")
		recs = append(recs, "Focus on sustainable intensification")
	}

	if model.GrowthRate20Years > 200 {
		recs = append(recs, "Ensure sustainable expansion practices")
		recs = append(recs, "Invest in supply chain optimization")
	} else if model.GrowthRate20Years < 50 && model.TotalArea2022 > 500000 {
		recs = append(recs, "Diversify revenue streams")
		recs = append(recs, "Explore value-added products")
	}

	if model.StabilityIndex < 5 {
		recs = append(recs, "Improve operational consistency")
		recs = append(recs, "Risk management implementation")
	}

	if len(recs) == 0 {
		recs = append(recs, "Continuous improvement with sustainability focus")
	}

	return recs
}

func calculateProjection2030(model ProvinceModel) float64 {
	annualGrowth := model.AnnualGrowthRate
	if annualGrowth <= 0 {
		annualGrowth = 3.0
	} else if annualGrowth > 10 {
		annualGrowth = 8.0
	}

	projection := model.TotalArea2022 * math.Pow(1+annualGrowth/100, 8)
	return projection
}

func calculateYearlyGrowthRates(yearlyData map[int]float64) []float64 {
	var growthRates []float64
	years := getSortedYears(yearlyData)

	for i := 1; i < len(years); i++ {
		currentYear := years[i]
		previousYear := years[i-1]

		if yearlyData[previousYear] > 0 {
			growth := ((yearlyData[currentYear] - yearlyData[previousYear]) / yearlyData[previousYear]) * 100
			growthRates = append(growthRates, growth)
		}
	}

	return growthRates
}

func calculateVolatility(growthRates []float64) float64 {
	if len(growthRates) == 0 {
		return 0
	}

	mean := 0.0
	for _, gr := range growthRates {
		mean += gr
	}
	mean /= float64(len(growthRates))

	variance := 0.0
	for _, gr := range growthRates {
		variance += math.Pow(gr-mean, 2)
	}
	variance /= float64(len(growthRates))

	return math.Sqrt(variance)
}

func calculatePhaseGrowth(yearlyData map[int]float64, startYear, endYear int) float64 {
	startArea := yearlyData[startYear]
	endArea := yearlyData[endYear]

	if startArea > 0 && endArea > 0 {
		return ((endArea - startArea) / startArea) * 100
	}
	return 0
}

func getPhaseDescription(growthRate float64, period string) string {
	if growthRate > 100 {
		return period + " explosive growth"
	} else if growthRate > 50 {
		return period + " high growth"
	} else if growthRate > 20 {
		return period + " moderate growth"
	} else if growthRate > 0 {
		return period + " slow growth"
	}
	return period + " decline"
}

func getSortedYears(yearlyData map[int]float64) []int {
	years := make([]int, 0, len(yearlyData))
	for year := range yearlyData {
		years = append(years, year)
	}
	sort.Ints(years)
	return years
}

func calculateNationalAreaForYear(rawData []RawPalmOilData, year int) float64 {
	total := 0.0
	for _, data := range rawData {
		if data.Year == year {
			total += data.PlantedArea
		}
	}
	return total
}

func findLeadingProvince(models []ProvinceModel, startYear, endYear int) string {
	if len(models) == 0 {
		return "Unknown"
	}
	return models[0].Province
}

func findEmergingRegions(models []ProvinceModel, startYear, endYear int) []string {
	var emerging []string
	for _, model := range models {
		if model.GrowthRate20Years > 300 && model.MarketShare2022 < 5 {
			emerging = append(emerging, model.Province)
		}
	}
	return emerging
}

func getKeyEventsForDecade(decade string) []string {
	events := map[string][]string{
		"2003-2012": {
			"Ekspansi cepat kelapa sawit",
			"Peningkatan permintaan global",
			"Pembukaan lahan baru",
		},
		"2013-2022": {
			"Fokus sustainability",
			"Sertifikasi ISPO/RSPO",
			"Tekanan lingkungan global",
			"Peningkatan produktivitas",
		},
	}

	if eventList, exists := events[decade]; exists {
		return eventList
	}
	return []string{"Perkembangan industri normal"}
}

func formatNumber(num float64) string {
	if num >= 1000000 {
		return fmt.Sprintf("%.2fM", num/1000000)
	} else if num >= 1000 {
		return fmt.Sprintf("%.1fK", num/1000)
	}
	return fmt.Sprintf("%.0f", num)
}

func getShortProvinceName(fullName string) string {
	shortNames := map[string]string{
		"RIAU":                      "RIAU",
		"SUMATERA UTARA":            "SUMUT",
		"KALIMANTAN BARAT":          "KALBAR",
		"JAMBI":                     "JAMBI",
		"SUMATERA SELATAN":          "SUMSEL",
		"KALIMANTAN TENGAH":         "KALTENG",
		"ACEH":                      "ACEH",
		"SUMATERA BARAT":            "SUMBAR",
		"KALIMANTAN TIMUR":          "KALTIM",
		"BENGKULU":                  "BENGKULU",
		"LAMPUNG":                   "LAMPUNG",
		"SULAWESI TENGAH":           "SULTENG",
		"SULAWESI SELATAN":          "SULSEL",
		"PAPUA":                     "PAPUA",
		"KALIMANTAN SELATAN":        "KALSEL",
		"SULAWESI UTARA":            "SULUT",
		"BANTEN":                    "BANTEN",
		"JAWA BARAT":                "JABAR",
		"JAWA TIMUR":                "JATIM",
		"JAWA TENGAH":               "JATENG",
		"DI YOGYAKARTA":             "YOGYA",
		"BALI":                      "BALI",
		"NUSA TENGGARA BARAT":       "NTB",
		"NUSA TENGGARA TIMUR":       "NTT",
		"MALUKU":                    "MALUKU",
		"SULAWESI TENGGARA":         "SULTRA",
		"GORONTALO":                 "GORONTALO",
		"KEPULAUAN RIAU":            "KEPRI",
		"PAPUA BARAT":               "PAPUA BARAT",
		"MALUKU UTARA":              "MALUT",
		"KEPULAUAN BANGKA BELITUNG": "BABEL",
	}

	if short, exists := shortNames[fullName]; exists {
		return short
	}

	if len(fullName) > 8 {
		return fullName[:8]
	}
	return fullName
}

func filterProvinces(models []ProvinceModel, category string) []ProvinceModel {
	var filtered []ProvinceModel

	for _, model := range models {
		switch category {
		case "PRIME":
			if model.TotalArea2022 > 1000000 && model.GrowthRate20Years > 100 {
				filtered = append(filtered, model)
			}
		case "GROWTH":
			if model.GrowthRate20Years > 200 {
				filtered = append(filtered, model)
			}
		case "EMERGING":
			if model.TotalArea2022 < 500000 && model.GrowthRate20Years > 300 {
				filtered = append(filtered, model)
			}
		case "STABLE":
			if model.TotalArea2022 > 500000 && model.GrowthRate20Years >= 50 && model.GrowthRate20Years <= 150 {
				filtered = append(filtered, model)
			}
		case "MATURE":
			if model.TotalArea2022 > 500000 && model.GrowthRate20Years < 50 {
				filtered = append(filtered, model)
			}
		}
	}

	return filtered
}

func writeProvinceGroup(f *excelize.File, provinces []ProvinceModel, startRow int) {
	for i, province := range provinces {
		row := startRow + i
		f.SetCellValue("Kelompok_Provinsi_20Tahun", fmt.Sprintf("A%d", row), province.Province)
		f.SetCellValue("Kelompok_Provinsi_20Tahun", fmt.Sprintf("B%d", row), formatNumber(province.TotalArea2022))
		f.SetCellValue("Kelompok_Provinsi_20Tahun", fmt.Sprintf("C%d", row), fmt.Sprintf("%.1f%%", province.GrowthRate20Years))
		f.SetCellValue("Kelompok_Provinsi_20Tahun", fmt.Sprintf("D%d", row), province.InvestmentPotential)
	}
}

func countProvincesByGrowth(models []ProvinceModel, minGrowth float64) int {
	count := 0
	for _, model := range models {
		if model.GrowthRate20Years >= minGrowth {
			count++
		}
	}
	return count
}

func calculateAverageGrowth(models []ProvinceModel) float64 {
	total := 0.0
	count := 0

	for _, model := range models {
		if model.GrowthRate20Years > 0 {
			total += model.GrowthRate20Years
			count++
		}
	}

	if count > 0 {
		return total / float64(count)
	}
	return 0
}

func getMinGrowthRate(models []ProvinceModel) float64 {
	if len(models) == 0 {
		return 0
	}
	min := models[0].GrowthRate20Years
	for _, model := range models {
		if model.GrowthRate20Years < min {
			min = model.GrowthRate20Years
		}
	}
	return min
}

func getMaxGrowthRate(models []ProvinceModel) float64 {
	if len(models) == 0 {
		return 0
	}
	max := models[0].GrowthRate20Years
	for _, model := range models {
		if model.GrowthRate20Years > max {
			max = model.GrowthRate20Years
		}
	}
	return max
}

func getMaxMarketShare(models []ProvinceModel) float64 {
	if len(models) == 0 {
		return 0
	}
	max := models[0].MarketShare2022
	for _, model := range models {
		if model.MarketShare2022 > max {
			max = model.MarketShare2022
		}
	}
	return max
}
