package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

// Глобальна змінна для зберігання посилання на завантажений HTML-шаблон
var tmpl *template.Template

func main() {
	// Завантажуємо HTML-шаблон з файлу "index.html"
	var err error
	tmpl, err = template.ParseFiles("index.html")
	if err != nil {
		// Якщо виникла помилка при читанні шаблону, виводимо її та завершуємо програму
		fmt.Println("Error loading template:", err)
		return
	}

	// Налаштовуємо маршрути (URL-шляхи):
	// 1) "/" - головна сторінка (просто відображає шаблон)
	// 2) "/calculate1" - обробник для Завдання 1
	// 3) "/calculate2" - обробник для Завдання 2

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Використовуємо шаблон "index.html" без жодних даних (nil),
		// щоб просто відкрити головну сторінку
		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/calculate1", calculateTask1)
	http.HandleFunc("/calculate2", calculateTask2)

	// Запуск веб-сервера на порту 8080
	fmt.Println("Server running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

// Допоміжна функція для перетворення рядка у float64
// Повертає число та помилку (якщо рядок не вдалося розпарсити)
func checkAndToDouble(input string) (float64, error) {
	return strconv.ParseFloat(input, 64)
}

// calculateTask1 - обробник для розрахунків "Завдання 1"
func calculateTask1(w http.ResponseWriter, r *http.Request) {
	// Зчитуємо поля з HTML-форми (методом FormValue)
	hp, _ := checkAndToDouble(r.FormValue("hp"))
	cp, _ := checkAndToDouble(r.FormValue("cp"))
	sp, _ := checkAndToDouble(r.FormValue("sp"))
	np, _ := checkAndToDouble(r.FormValue("np"))
	op, _ := checkAndToDouble(r.FormValue("op"))
	wp, _ := checkAndToDouble(r.FormValue("wp"))
	ap, _ := checkAndToDouble(r.FormValue("ap"))

	// kpc (коефіцієнт для сухої маси) = 100 / (100 - wᵖ)
	kpc := 100 / (100 - wp)

	// krg (коефіцієнт для горючої маси) = 100 / (100 - wᵖ - aᵖ)
	krg := 100 / (100 - wp - ap)

	// Розрахунок складу сухої маси
	hc := hp * kpc
	cc := cp * kpc
	sc := sp * kpc
	nc := np * kpc
	oc := op * kpc
	ac := ap * kpc

	// Розрахунок складу горючої маси
	hg := hp * krg
	cg := cp * krg
	sg := sp * krg
	ng := np * krg
	og := op * krg

	// Розрахунок нижчої теплоти згорання (робочої маси)
	qph := (339*cp + 1030*hp - 108.8*(op-sp) - 25*wp) / 1000

	// Для сухої маси
	qch := (qph + 0.025*wp) * 100 / (100 - wp)

	// Для горючої маси
	qgh := (qph + 0.025*wp) * 100 / (100 - wp - ap)

	// Формуємо текстовий результат
	result := fmt.Sprintf(`
Коеф. (роб. -> суха): %.3f
Коеф. (роб. -> горюча): %.3f

Склад сухої маси:
  Hc = %.3f %%
  Cc = %.3f %%
  Sc = %.3f %%
  Nc = %.3f %%
  Oc = %.3f %%
  Ac = %.3f %%

Склад горючої маси:
  Hg = %.3f %%
  Cg = %.3f %%
  Sg = %.3f %%
  Ng = %.3f %%
  Og = %.3f %%

Теплота (роб. маса): %.3f МДж/кг
Теплота (суха маса): %.3f МДж/кг
Теплота (горюча маса): %.3f МДж/кг
`,
		kpc, krg,
		hc, cc, sc, nc, oc, ac,
		hg, cg, sg, ng, og,
		qph, qch, qgh,
	)

	// Передаємо результат у шаблон. "Result" — ключ, який відповідає {{.Result}} в index.html
	tmpl.Execute(w, map[string]string{
		"Result": result,
	})
}

// --------------------- Завдання 2 ---------------------
// calculateTask2 - обробник для розрахунків "Завдання 2"
func calculateTask2(w http.ResponseWriter, r *http.Request) {
	// Зчитуємо поля з HTML-форми
	cg, _ := checkAndToDouble(r.FormValue("cg"))
	hg, _ := checkAndToDouble(r.FormValue("hg"))
	og, _ := checkAndToDouble(r.FormValue("og"))
	sg, _ := checkAndToDouble(r.FormValue("sg"))
	qi, _ := checkAndToDouble(r.FormValue("qi"))
	vg, _ := checkAndToDouble(r.FormValue("vg"))
	wg, _ := checkAndToDouble(r.FormValue("wg"))
	ag, _ := checkAndToDouble(r.FormValue("ag"))

	// Перерахунок складу на робочу масу
	// cp, hp, op, sp, ap, vp (робоча маса)
	cp := cg * (100 - wg - ag) / 100
	hp := hg * (100 - wg - ag) / 100
	op := og * (100 - wg - ag) / 100
	sp := sg * (100 - wg - ag) / 100
	ap := ag * (100 - wg) / 100
	vp := vg * (100 - wg) / 100

	// Нижча теплота згоряння (з поправкою на робочу масу)
	qri := qi*(100-wg-ap)/100 - 0.025*wg

	// Формуємо текстовий результат
	result := fmt.Sprintf(`
Перерахунок елементарного складу мазуту на робочу масу:
  Cp = %.3f %%
  Hp = %.3f %%
  Op = %.3f %%
  Sp = %.3f %%
  Ap = %.3f %%
  Vp = %.3f мг/кг

Нижча теплота згоряння (роб. маса): %.3f МДж/кг
`,
		cp, hp, op, sp, ap, vp, qri,
	)

	// Відображаємо результат у шаблоні
	tmpl.Execute(w, map[string]string{
		"Result": result,
	})
}
