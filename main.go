package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
)

type CocktailResponse struct {
	Query  string     `json:"search"`
	Drinks []Cocktail `json:"drinks"`
}

type Cocktail struct {
	ID           string `json:"idDrink"`
	Name         string `json:"strDrink"`
	Instructions string `json:"strInstructions"`
	Thumbnail    string `json:"strDrinkThumb"`
	Category     string `json:"strCategory"`
	Alcoholic    string `json:"strAlcoholic"`
	Glass        string `json:"strGlass"`
	Ingredients  map[string]string
}

func main() {
	tmpl := template.Must(template.ParseFiles("index.html"))
	tmpl1 := template.Must(template.ParseFiles("cocktails.html"))
	fs := http.FileServer(http.Dir("css"))
	http.Handle("/css/", http.StripPrefix("/css/", fs))

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.FormValue("cocktail")
		cocktailResp, err := searchCocktail(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = tmpl1.Execute(w, cocktailResp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/cocktails", func(w http.ResponseWriter, r *http.Request) {
		tmpl1.Execute(w, nil)
	})

	http.HandleFunc("/cocktail-details", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		cocktail, err := getCocktailDetails(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl2 := template.Must(template.ParseFiles("cocktail-details.html"))
		tmpl2.Execute(w, cocktail)
	})

	http.ListenAndServe(":8080", nil)
}

func searchCocktail(query string) (*CocktailResponse, error) {
	baseURL := "https://www.thecocktaildb.com/api/json/v1/1/search.php"
	url, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	params := url.Query()
	params.Add("s", query)
	url.RawQuery = params.Encode()
	resp, err := http.Get(url.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var cocktailResp struct {
		Drinks []Cocktail `json:"drinks"`
	}
	err = json.NewDecoder(resp.Body).Decode(&cocktailResp)
	if err != nil {
		return nil, err
	}
	cocktailRespObj := &CocktailResponse{
		Query:  query,
		Drinks: []Cocktail{},
	}
	for _, drink := range cocktailResp.Drinks {
		cocktailRespObj.Drinks = append(cocktailRespObj.Drinks, drink)
	}
	return cocktailRespObj, nil
}

func extractIngredientsAndMeasures(cocktailData map[string]interface{}) map[string]string {
	ingredients := make(map[string]string)
	for i := 1; i <= 15; i++ {
		ingredientKey := fmt.Sprintf("strIngredient%d", i)
		measureKey := fmt.Sprintf("strMeasure%d", i)

		ingredient, ingredientExists := cocktailData[ingredientKey]
		measure, measureExists := cocktailData[measureKey]

		if ingredientExists && ingredient != nil && ingredient != "" {
			if measureExists && measure != nil && measure != "" {
				ingredients[ingredient.(string)] = measure.(string)
			} else {
				ingredients[ingredient.(string)] = ""
			}
		}
	}
	return ingredients
}

func getCocktailDetails(id string) (*Cocktail, error) {
	baseURL := "https://www.thecocktaildb.com/api/json/v1/1/lookup.php"
	url, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	params := url.Query()
	params.Add("i", id)
	url.RawQuery = params.Encode()
	resp, err := http.Get(url.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var cocktailResp struct {
		Drinks []map[string]interface{} `json:"drinks"`
	}
	err = json.NewDecoder(resp.Body).Decode(&cocktailResp)
	if err != nil {
		return nil, err
	}
	if len(cocktailResp.Drinks) == 0 {
		return nil, errors.New("Cocktail not found")
	}

	cocktailData := cocktailResp.Drinks[0]
	ingredients := extractIngredientsAndMeasures(cocktailData)

	cocktail := &Cocktail{
		ID:           cocktailData["idDrink"].(string),
		Name:         cocktailData["strDrink"].(string),
		Instructions: cocktailData["strInstructions"].(string),
		Thumbnail:    cocktailData["strDrinkThumb"].(string),
		Ingredients:  ingredients,
	}

	return cocktail, nil
}
