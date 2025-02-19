package main

import (
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"strconv"
)

// --- A simple struct to hold the result for the template ---
type ResultData struct {
	FuelType string
	KValue   float64
	EValue   float64
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/calculate", calculateHandler)

	log.Println("Server running on http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// indexHandler serves the form (index.html).
func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Помилка завантаження сторінки", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// calculateHandler reads form data, does the emission math, and returns a result page.
func calculateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1) Parse the form
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Невірні вхідні дані (ParseForm failed)", http.StatusBadRequest)
		return
	}

	// 2) Extract the fuel type and amount from the form
	fuelType := r.FormValue("fuelType") // "coal", "oilFuel", or "gas"
	amountStr := r.FormValue("fuelAmount")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		http.Error(w, "Невірне значення кількості палива", http.StatusBadRequest)
		return
	}

	// 3) We'll store default parameters for each fuel right here
	//    These mirror the logic from your previous JSON-based approach
	//    but are now handled directly in Go.
	var Qri, A, Ar, G, N, Ks, alpha float64

	switch fuelType {
	case "coal":
		Qri = 20.47
		A = 1.0
		Ar = 25.2
		G = 1.5
		N = 0.985
		Ks = 0
		alpha = 0.8
	case "oilFuel":
		Qri = 40.40
		A = 1.0
		Ar = 0.15
		G = 0
		N = 0.985
		Ks = 0
		alpha = 1.0
	case "gas":
		Qri = 33.08
		A = 0
		Ar = 0
		G = 0
		N = 0
		Ks = 0
	default:
		http.Error(w, "Невірно обрано паливо", http.StatusBadRequest)
		return
	}

	// 4) Calculate K and E using your formulas
	//    K = (10^6 / Qri) * A * (Ar/(100 - G)) * (1 - N) + Ks
	KValue := (math.Pow(10, 6)/Qri)*A*alpha*(Ar/(100-G))*(1-N) + Ks

	//    E = 10^(-6) * K * Qri * B
	EValue := math.Pow(10, -6) * KValue * Qri * amount

	// 5) Construct a simple HTML response OR use a separate template
	//    For simplicity, let's just build a small HTML right here
	//    Or you could parse a "templates/result.html" if you prefer
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html lang="uk">
<head>
  <meta charset="UTF-8"/>
  <title>Результат розрахунку</title>
</head>
<body style="font-family: sans-serif; margin: 2rem;">
  <h1>Результати</h1>
  <p><strong>Тип палива:</strong> %s</p>
  <p><strong>K:</strong> %.3f</p>
  <p><strong>E:</strong> %.3f</p>
  <hr>
  <a href="/">Повернутися назад</a>
</body>
</html>
`, fuelType, KValue, EValue)
}
