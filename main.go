package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/obfio/phonemore-scraper/phonemore"
)

func main() {
	devices := []*phonemore.Model{}
	for i := 1; i < 5; i++ {
		models, err := phonemore.ScrapeModels(i)
		if err != nil {
			fmt.Println(err)
			break
		}
		for _, m := range models {
			fmt.Printf("%+v\n", m)
			err = m.FillData()
			if err != nil {
				panic(err)
			}
			devices = append(devices, m)
		}
	}
	b, err := json.MarshalIndent(devices, "", "	")
	if err != nil {
		panic(err)
	}
	os.WriteFile("devices.json", b, 0666)
}
