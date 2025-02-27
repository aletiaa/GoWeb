package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
)

var tmpl *template.Template

func init() {
	var err error
	tmpl, err = template.ParseFiles(filepath.Join("templates", "index.html"))
	if err != nil {
		log.Fatalf("Error loading template: %v", err)
	}
}

// Wrapper struct to safely handle both Damages and Reliability Results
type TemplateData struct {
	DamagesResult     *DamagesResultModel
	ReliabilityResult *ReliabilityResultModel
}

// Damages Models
type DamagesInputModel struct {
	FailureFrequency float64
	RestoreTime      float64
	Pm               float64
	Tm               float64
	Kp               float64
	Za               float64
	Zp               float64
}

type DamagesResultModel struct {
	MWa float64
	MWp float64
	Mz  float64
}

func calculateDamages(input DamagesInputModel) DamagesResultModel {
	MWa := input.FailureFrequency * input.RestoreTime * input.Pm * input.Tm
	MWp := input.Kp * input.Pm * input.Tm
	Mz := input.Za*MWa + input.Zp*MWp
	return DamagesResultModel{MWa, MWp, Mz}
}

// Reliability Models
type ReliabilityInputModel struct {
	ElectricGasSwitch          float64
	Pl110                      float64
	Transformer                float64
	InputSwitch                float64
	Connections                float64
	ElectricGasSwitchT         float64
	Pl110T                     float64
	TransformerT               float64
	InputSwitchT               float64
	ConnectionsT               float64
	Kppmax                     float64
	FailureFreqSectionSwitcher float64
}

type ReliabilityResultModel struct {
	FailureFrequency                    float64
	AverageRecoveryDuration             float64
	EmergencyCoeff                      float64
	PlanCoeff                           float64
	FailureFreqForTwoSys                float64
	FailureFrequencyWithSectionSwitcher float64
}

func calculateReliability(input ReliabilityInputModel) ReliabilityResultModel {
	// Step 1: Compute total failure frequency (ω_oc)
	failureFrequency := input.ElectricGasSwitch + input.Pl110 + input.Transformer + input.InputSwitch + input.Connections
	log.Printf("Failure Frequency (ω_oc): %.6f", failureFrequency)

	// Step 2: Debug each input value before calculating t_b.oc
	log.Printf("Electric Gas Switch: %.6f, Time: %.6f", input.ElectricGasSwitch, input.ElectricGasSwitchT)
	log.Printf("PL110: %.6f, Time: %.6f", input.Pl110, input.Pl110T)
	log.Printf("Transformer: %.6f, Time: %.6f", input.Transformer, input.TransformerT)
	log.Printf("Input Switch: %.6f, Time: %.6f", input.InputSwitch, input.InputSwitchT)
	log.Printf("Connections: %.6f, Time: %.6f", input.Connections, input.ConnectionsT)

	// Step 3: Compute numerator and denominator for t_b.oc
	numerator := (input.ElectricGasSwitch * input.ElectricGasSwitchT) +
		(input.Pl110 * input.Pl110T) +
		(input.Transformer * input.TransformerT) +
		(input.InputSwitch * input.InputSwitchT) +
		(input.Connections * input.ConnectionsT)

	denominator := failureFrequency

	log.Printf("Numerator for t_b.oc: %.6f", numerator)
	log.Printf("Denominator for t_b.oc: %.6f", denominator)

	// Step 4: Compute t_b.oc
	var averageRecoveryDuration float64
	if denominator > 0 {
		averageRecoveryDuration = numerator / denominator
	} else {
		averageRecoveryDuration = 0
	}

	log.Printf("Computed Average Recovery Duration (t_b.oc): %.6f", averageRecoveryDuration)

	// Step 5: Compute emergency coefficient (k_a.oc)
	emergencyCoeff := (failureFrequency * averageRecoveryDuration) / 8760.0
	log.Printf("Computed Emergency Coefficient (k_a.oc): %.6f", emergencyCoeff)

	// Step 6: Compute planned coefficient (k_n.oc)
	planCoeff := (1.2 * input.Kppmax) / 8760.0
	log.Printf("Computed Plan Coefficient (k_n.oc): %.6f", planCoeff)

	// Step 7: Compute outage frequency for two-system network (ω_uk)
	failureFreqForTwoSys := 2 * failureFrequency * (emergencyCoeff + planCoeff)
	log.Printf("Computed Failure Frequency for Two Systems (ω_uk): %.6f", failureFreqForTwoSys)

	// Step 8: Compute final outage frequency with section switcher (ω_dc)
	failureFrequencyWithSectionSwitcher := failureFreqForTwoSys + input.FailureFreqSectionSwitcher
	log.Printf("Computed Failure Frequency with Section Switcher (ω_dc): %.6f", failureFrequencyWithSectionSwitcher)

	// Return calculated values
	return ReliabilityResultModel{
		FailureFrequency:                    failureFrequency,
		AverageRecoveryDuration:             averageRecoveryDuration,
		EmergencyCoeff:                      emergencyCoeff,
		PlanCoeff:                           planCoeff,
		FailureFreqForTwoSys:                failureFreqForTwoSys,
		FailureFrequencyWithSectionSwitcher: failureFrequencyWithSectionSwitcher,
	}
}

type PageData struct {
	DamagesResult     *DamagesResultModel
	ReliabilityResult *ReliabilityResultModel
}

// Handle Damage Calculation Request
func handleDamagesRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		damagesInput := DamagesInputModel{
			FailureFrequency: parseFloat(r.FormValue("failureFrequency"), 0.01),
			RestoreTime:      parseFloat(r.FormValue("restoreTime"), 45.0),
			Pm:               parseFloat(r.FormValue("Pm"), 5.12),
			Tm:               parseFloat(r.FormValue("Tm"), 6451.0),
			Kp:               parseFloat(r.FormValue("Kp"), 4.0),
			Za:               parseFloat(r.FormValue("Za"), 23.6),
			Zp:               parseFloat(r.FormValue("Zp"), 17.6),
		}

		damagesResult := calculateDamages(damagesInput)

		data := PageData{
			DamagesResult:     &damagesResult,
			ReliabilityResult: nil, // No reliability data in this request
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			log.Println("Template execution error:", err)
		}
	} else {
		data := PageData{}
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			log.Println("Template execution error:", err)
		}
	}
}

// Handle Reliability Calculation Request
func handleReliabilityRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		reliabilityInput := ReliabilityInputModel{
			ElectricGasSwitch:          parseFloat(r.FormValue("electricGasSwitch"), 0.01),
			Pl110:                      parseFloat(r.FormValue("pl110"), 0.07),
			Transformer:                parseFloat(r.FormValue("transformer"), 0.015),
			InputSwitch:                parseFloat(r.FormValue("inputSwitch"), 0.02),
			Connections:                parseFloat(r.FormValue("connections"), 0.18),
			ElectricGasSwitchT:         parseFloat(r.FormValue("electricGasSwitchT"), 30.0),
			Pl110T:                     parseFloat(r.FormValue("pl110T"), 10.0),
			TransformerT:               parseFloat(r.FormValue("transformerT"), 100.0),
			InputSwitchT:               parseFloat(r.FormValue("inputSwitchT"), 15.0),
			ConnectionsT:               parseFloat(r.FormValue("connectionsT"), 2.0),
			Kppmax:                     parseFloat(r.FormValue("kppmax"), 43.0),
			FailureFreqSectionSwitcher: parseFloat(r.FormValue("failureFreqSectionSwitcher"), 0.02),
		}

		reliabilityResult := calculateReliability(reliabilityInput)

		data := PageData{
			DamagesResult:     nil, // No damages data in this request
			ReliabilityResult: &reliabilityResult,
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			log.Println("Template execution error:", err)
		}
	} else {
		data := PageData{}
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			log.Println("Template execution error:", err)
		}
	}
}

func parseFloat(str string, defaultValue float64) float64 {
	if value, err := strconv.ParseFloat(str, 64); err == nil {
		return value
	}
	return defaultValue
}

func main() {
	fs := http.FileServer(http.Dir("./static"))

	fmt.Println("Server is running on http://localhost:8080...")

	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/damages", handleDamagesRequest)
	http.HandleFunc("/reliability", handleReliabilityRequest)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := tmpl.Execute(w, TemplateData{}); err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			log.Println("Template execution error:", err)
			return
		}
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
