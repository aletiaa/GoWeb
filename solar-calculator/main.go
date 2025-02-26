package main

import (
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"strconv"
)

// Struct to hold calculation results
type CalculationResult struct {
	W1, W2, ProfitBefore, PenaltyBefore, FinalProfitBefore float64
	W3, W4, ProfitAfter, PenaltyAfter, FinalProfitAfter    float64
}

// Normal distribution function
func normalDistribution(p, Pc, stdDev float64) float64 {
	return (1 / (stdDev * math.Sqrt(2*math.Pi))) * math.Exp(-math.Pow(p-Pc, 2)/(2*math.Pow(stdDev, 2)))
}

// Numerical integration to approximate the probability
func integrateNormalDistribution(Pc, stdDev, P_lower, P_upper float64) float64 {
	n := 1000
	step := (P_upper - P_lower) / float64(n)
	area := 0.0
	for i := 0; i < n; i++ {
		x1 := P_lower + float64(i)*step
		x2 := P_lower + float64(i+1)*step
		y1 := normalDistribution(x1, Pc, stdDev)
		y2 := normalDistribution(x2, Pc, stdDev)
		area += 0.5 * (y1 + y2) * step
	}
	return area
}

// Serve the main page
func serveHome(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		log.Println("Template parsing error:", err)
		return
	}

	err = tmpl.Execute(w, nil) // No data needed on initial load
	if err != nil {
		log.Println("Error executing template:", err)
	}
}

// Handle form submission and perform calculations
func calculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse form values
	dailyPower, err := strconv.ParseFloat(r.FormValue("dailyPower"), 64)
	if err != nil {
		http.Error(w, "Invalid input for dailyPower", http.StatusBadRequest)
		return
	}

	currentStdDev, err := strconv.ParseFloat(r.FormValue("currentStdDev"), 64)
	if err != nil {
		http.Error(w, "Invalid input for currentStdDev", http.StatusBadRequest)
		return
	}

	futureStdDev, err := strconv.ParseFloat(r.FormValue("futureStdDev"), 64)
	if err != nil {
		http.Error(w, "Invalid input for futureStdDev", http.StatusBadRequest)
		return
	}

	energyCost, err := strconv.ParseFloat(r.FormValue("energyCost"), 64)
	if err != nil {
		http.Error(w, "Invalid input for energyCost", http.StatusBadRequest)
		return
	}

	// Perform calculations
	P_lower := dailyPower - futureStdDev
	P_upper := dailyPower + futureStdDev

	deltaW1 := integrateNormalDistribution(dailyPower, currentStdDev, P_lower, P_upper)
	W1 := dailyPower * 24 * deltaW1
	profitBefore := W1 * energyCost

	W2 := dailyPower * 24 * (1 - deltaW1)
	penaltyBefore := W2 * energyCost
	finalProfitBefore := profitBefore - penaltyBefore

	deltaW2 := integrateNormalDistribution(dailyPower, futureStdDev, P_lower, P_upper)
	W3 := dailyPower * 24 * deltaW2
	profitAfter := W3 * energyCost

	W4 := dailyPower * 24 * (1 - deltaW2)
	penaltyAfter := W4 * energyCost
	finalProfitAfter := profitAfter - penaltyAfter

	// Create result struct
	result := CalculationResult{
		W1, W2, profitBefore, penaltyBefore, finalProfitBefore,
		W3, W4, profitAfter, penaltyAfter, finalProfitAfter,
	}

	// Load and execute template with data
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		log.Println("Template parsing error:", err)
		return
	}

	err = tmpl.Execute(w, result)
	if err != nil {
		log.Println("Error executing template:", err)
	}
}

func main() {
	// Serve index.html when visiting "/"
	http.HandleFunc("/", serveHome)

	// Handle form submission
	http.HandleFunc("/calculate", calculate)

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
