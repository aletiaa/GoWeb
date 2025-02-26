package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
)

type CableData struct {
	Unom float64 `json:"unom"`
	Ik   float64 `json:"ik"`
	Tf   float64 `json:"tf"`
	Sm   float64 `json:"sm"`
	Jek  float64 `json:"jek"`
	Ct   float64 `json:"ct"`
}

type CurrentData struct {
	Ukmax float64 `json:"ukmax"`
	Uvn   float64 `json:"uvn"`
	Unn   float64 `json:"unn"`
	Snomt float64 `json:"snomt"`
	Xch   float64 `json:"xch"`
	Xcmin float64 `json:"xcmin"`
	Rch   float64 `json:"rch"`
	Rcmin float64 `json:"rcmin"`
	Ll    float64 `json:"ll"`
	R0    float64 `json:"r0"`
	X0    float64 `json:"x0"`
}

type CurrentOnTenData struct {
	Sk    float64 `json:"sk"`
	Uch   float64 `json:"uch"`
	Snomt float64 `json:"snomt"`
	Uk    float64 `json:"uk"`
}

type CableResult struct {
	Im   float64 `json:"im"`
	Impa float64 `json:"impa"`
	Sek  float64 `json:"sek"`
	S    float64 `json:"s"`
}

type CurrentOnTenResult struct {
	Xc  float64 `json:"xc"`
	Xt  float64 `json:"xt"`
	X   float64 `json:"x"`
	Ip0 float64 `json:"ip0"`
}

type CurrentResult struct {
	Rsh      float64 `json:"Rsh"`
	Xsh      float64 `json:"Xsh"`
	Zsh      float64 `json:"Zsh"`
	Rshmin   float64 `json:"Rshmin"`
	Xshmin   float64 `json:"Xshmin"`
	Zshmin   float64 `json:"Zshmin"`
	I3sh     float64 `json:"I3sh"`
	I2sh     float64 `json:"I2sh"`
	I3shmin  float64 `json:"I3shmin"`
	I2shmin  float64 `json:"I2shmin"`
	Kpr      float64 `json:"kpr"`
	Rshn     float64 `json:"Rshn"`
	Xshn     float64 `json:"Xshn"`
	Zshn     float64 `json:"Zshn"`
	Rshnmin  float64 `json:"Rshnmin"`
	Xshnmin  float64 `json:"Xshnmin"`
	Zshnmin  float64 `json:"Zshnmin"`
	I3shn    float64 `json:"I3shn"`
	I2shn    float64 `json:"I2shn"`
	I3shnmin float64 `json:"I3shnmin"`
	I2shnmin float64 `json:"I2shnmin"`
	Rl       float64 `json:"Rl"`
	Xl       float64 `json:"Xl"`
	Rcn      float64 `json:"Rcn"`
	Xcn      float64 `json:"Xcn"`
	Zcn      float64 `json:"Zcn"`
	Rcnmin   float64 `json:"Rcnmin"`
	Xcnmin   float64 `json:"Xcnmin"`
	Zcnmin   float64 `json:"Zcnmin"`
	I3ln     float64 `json:"I3ln"`
	I2ln     float64 `json:"I2ln"`
	I3lnmin  float64 `json:"I3lnmin"`
	I2lnmin  float64 `json:"I2lnmin"`
}

func calculateCable(w http.ResponseWriter, r *http.Request) {
	var data CableData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	im := (data.Sm / 2) / (math.Sqrt(3.0) * data.Unom)
	impa := im * 2
	sek := im / data.Jek
	s := (data.Ik * math.Sqrt(data.Tf)) / data.Ct

	result := CableResult{Im: im, Impa: impa, Sek: sek, S: s}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func calculateCurrentOnTen(w http.ResponseWriter, r *http.Request) {
	var data CurrentOnTenData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println(data)
	xc := math.Pow(data.Uch, 2) / data.Sk
	xt := (data.Uk / 100) * (math.Pow(data.Uch, 2) / data.Snomt)
	x := xc + xt
	ip0 := data.Uch / (math.Sqrt(3) * x)

	result := CurrentOnTenResult{Xc: xc, Xt: xt, X: x, Ip0: ip0}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func calculateCurrent(w http.ResponseWriter, r *http.Request) {
	var inputData CurrentData
	err := json.NewDecoder(r.Body).Decode(&inputData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	Xt := (inputData.Ukmax * math.Pow(inputData.Uvn, 2)) / (100 * inputData.Snomt)

	Xsh := inputData.Xch + Xt
	Zsh := math.Sqrt(math.Pow(inputData.Rch, 2) + math.Pow(Xsh, 2))
	Rshmin := inputData.Rcmin
	Xshmin := inputData.Xcmin + Xt
	Zshmin := math.Sqrt(math.Pow(Rshmin, 2) + math.Pow(Xshmin, 2))

	I3sh := (inputData.Uvn * 1000) / (math.Sqrt(3) * Zsh)
	I2sh := I3sh * (math.Sqrt(3) / 2)
	I3shmin := (inputData.Uvn * 1000) / (math.Sqrt(3) * Zshmin)
	I2shmin := I3shmin * (math.Sqrt(3) / 2)

	kpr := math.Pow(inputData.Unn, 2) / math.Pow(inputData.Uvn, 2)

	Rshn := inputData.Rch * kpr
	Xshn := Xsh * kpr
	Zshn := math.Sqrt(math.Pow(Rshn, 2) + math.Pow(Xshn, 2))
	Rshnmin := Rshmin * kpr
	Xshnmin := Xshmin * kpr
	Zshnmin := math.Sqrt(math.Pow(Rshnmin, 2) + math.Pow(Xshnmin, 2))

	I3shn := (inputData.Unn * 1000) / (math.Sqrt(3) * Zshn)
	I2shn := I3shn * (math.Sqrt(3) / 2)
	I3shnmin := (inputData.Unn * 1000) / (math.Sqrt(3) * Zshnmin)
	I2shnmin := I3shnmin * (math.Sqrt(3) / 2)

	Rl := inputData.Ll * inputData.R0
	Xl := inputData.Ll * inputData.X0

	Rcn := Rl + Rshn
	Xcn := Xl + Xshn
	Zcn := math.Sqrt(math.Pow(Rcn, 2) + math.Pow(Xcn, 2))
	Rcnmin := Rl + Rshnmin
	Xcnmin := Xl + Xshnmin
	Zcnmin := math.Sqrt(math.Pow(Rcnmin, 2) + math.Pow(Xcnmin, 2))

	I3ln := (inputData.Unn * 1000) / (math.Sqrt(3) * Zcn)
	I2ln := I3ln * (math.Sqrt(3) / 2)
	I3lnmin := (inputData.Unn * 1000) / (math.Sqrt(3) * Zcnmin)
	I2lnmin := I3lnmin * (math.Sqrt(3) / 2)

	result := CurrentResult{
		inputData.Rch, Xsh, Zsh, Rshmin, Xshmin, Zshmin,
		I3sh, I2sh, I3shmin, I2shmin, kpr, Rshn,
		Xshn, Zshn, Rshnmin, Xshnmin, Zshnmin,
		I3shn, I2shn, I3shnmin, I2shnmin, Rl,
		Xl, Rcn, Xcn, Zcn, Rcnmin, Xcnmin,
		Zcnmin, I3ln, I2ln, I3lnmin, I2lnmin,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/api/calculate-cable", calculateCable)
	http.HandleFunc("/api/calculate-current-on-ten", calculateCurrentOnTen)
	http.HandleFunc("/api/calculate-current", calculateCurrent)

	fmt.Println("Server is running on http://localhost:8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
