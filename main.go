package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

type Cocktail struct {
	Name        string       `json:"name"`
	Glass       string       `json:"glass"`
	Category    string       `json:"category"`
	Garnish     string       `json:"garnish"`
	Preparation string       `json:"preparation"`
	Ingredients []Ingredient `json:"ingredients"`
}

type Ingredient struct {
	Unit    string  `json:"unit"`
	Amount  float64 `json:"amount"`
	Label   string  `json:"label"`
	Name    string  `json:"ingredient"`
	Special string  `json:"special"`
}

func (c Cocktail) String() string {
	out := strings.Builder{}

	out.WriteString(fmt.Sprintf("%s:\n", c.Name))

	for _, i := range c.Ingredients {
		out.WriteString(fmt.Sprintf("  - %s\n", i.String()))
	}

	return out.String()
}

func (i Ingredient) PrintedName() string {
	if i.Special != "" {
		return i.Special
	}

	if i.Label != "" {
		return i.Label
	}

	return i.Name
}

func (i Ingredient) String() string {
	if i.Special != "" {
		return i.Special
	}

	name := i.PrintedName()
	amount := i.Amount
	unit := i.Unit

	if unit == "cl" {
		amount = amount * 10
		unit = "ml"
	}

	return fmt.Sprintf("%g %s %s", amount, unit, name)
}

var recipes []Cocktail
var associations map[string][]Ingredient

const recipeFn = "recipes.json"

func init() {
	// Read the file
	f, err := os.Open(recipeFn)
	if err != nil {
		panic(err)
	}

	// Load the raw data
	d := json.NewDecoder(f)
	d.DisallowUnknownFields()
	err = d.Decode(&recipes)
	if err != nil {
		panic(err)
	}

	// Build the associations
	associations = make(map[string][]Ingredient)
	for _, recipe := range recipes {
		for i, src := range recipe.Ingredients {
			if src.Name == "" {
				continue
			}

			if _, ok := associations[src.Name]; !ok {
				associations[src.Name] = make([]Ingredient, 0)
			}

			for j, dst := range recipe.Ingredients {
				if i != j {
					associations[src.Name] = append(associations[src.Name], dst)
				}
			}
		}
	}

	// Randomise seed
	rand.Seed(time.Now().Unix())
}

func Random() Cocktail {
	// All ingredients
	bar := make([]string, 0)
	for name, _ := range associations {
		bar = append(bar, name)
	}

	// Starting ingredient
	parts := make([]string, 0)
	parts = append(parts, bar[rand.Intn(len(bar))])

	// Begin cocktail
	count := 2 + rand.Intn(4)
	out := Cocktail{
		Ingredients: make([]Ingredient, 0),
	}

	for i := 0; i < count; i++ {
		// Gather the list of associated ingredients
		choices := make([]Ingredient, 0)
		for _, part := range parts {
			for _, choice := range associations[part] {
				choices = append(choices, choice)
			}
		}

		var choice Ingredient
	dedupe:
		for i := 0; i < len(choices); i++ { // Avoid repeated ingredients
			choice = choices[rand.Intn(len(choices))]

			if choice.Name == "" {
				continue
			}

			for _, i := range out.Ingredients {
				if choice.PrintedName() == i.PrintedName() {
					continue dedupe
				}
			}

			out.Ingredients = append(out.Ingredients, choice)
			parts = append(parts, choice.Name)
			break
		}
	}

	// Now let's find a name for it
	nameParts := make([]string, 0)
	for i := 0; i < 2+rand.Intn(len(out.Ingredients)-1); i++ {
		ingredient := out.Ingredients[rand.Intn(len(out.Ingredients))]

	pick:
		for i := 0; i < len(recipes); i++ {
			c := recipes[rand.Intn(len(recipes))]
			for _, j := range c.Ingredients {
				if ingredient.PrintedName() == j.PrintedName() {
					parts := strings.Split(c.Name, " ")
					nameParts = append(nameParts, parts[rand.Intn(len(parts))])
					break pick
				}
			}
		}
	}
	out.Name = strings.Join(nameParts, " ")

	return out
}

func Handler() (Cocktail, error) {
	return Random(), nil
}

func main() {
	lambda.Start(Handler)
}
