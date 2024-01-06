package pkg

import (
	"encoding/json"
	"errors"
	"os"
	"os/signal"
	"sync"
)

type Sign struct {
	Lines     [4]string `json:"lines"`
	Timestamp string    `json:"timestamp"`
	Area      int       `json:"area"`
}

type LocationSigns struct {
	sync.RWMutex
	SignsMap map[string][]Sign `json:"signsMap"` // map a location to its signs
}

func NewLocationSigns() *LocationSigns {
	return &LocationSigns{
		SignsMap: make(map[string][]Sign),
	}
}

func (ls *LocationSigns) AddOrUpdateSign(location string, sign Sign) {
	ls.Lock()
	defer ls.Unlock()

	signs, exists := ls.SignsMap[location]
	if !exists {
		ls.SignsMap[location] = append(ls.SignsMap[location], sign)
		return
	}

	for i, existingSign := range signs {
		if equalLines(existingSign.Lines, sign.Lines) {
			if existingSign.Area < sign.Area {
				signs[i] = sign // Replace the existing sign with the new one
				return
			} else {
				return // Don't add or update if the area is not greater
			}
		}
	}

	ls.SignsMap[location] = append(signs, sign) // No existing sign with same lines found, so append
}

func equalLines(lines1, lines2 [4]string) bool {
	for i, line := range lines1 {
		if lines2[i] != line {
			return false
		}
	}
	return true
}

func (ls *LocationSigns) SaveToFile(filename string) error {
	data, err := json.Marshal(ls.SignsMap)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func (ls *LocationSigns) LoadFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	ls.Lock()
	defer ls.Unlock()
	return json.Unmarshal(data, &ls.SignsMap)
}

func AutoSavedLocationSigns(filename string) (*LocationSigns, error) {

	locationSigns := &LocationSigns{
		SignsMap: make(map[string][]Sign),
	}

	if err := locationSigns.LoadFromFile(filename); err != nil {
		return nil, err
	}

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)

	go func() {
		<-sigCh

		if err := locationSigns.SaveToFile(filename); err != nil {
			panic(err)
		}

		os.Exit(1)
	}()

	return locationSigns, nil
}

// func main() {
// 	const storageFile = "signs.json"
// 	locationSigns := NewLocationSigns()

// 	// Load existing signs from JSON file
// 	if err := locationSigns.LoadFromFile(storageFile); err != nil {
// 		panic(err)
// 	}

// 	// Example usage: add/update some signs
// 	sign1 := Sign{[4]string{"Line1", "Line2", "Line3", "Line4"}, "2023-01-01T12:00:00Z", 100}
// 	locationSigns.AddOrUpdateSign("LocationA", sign1)

// 	sign2 := Sign{[4]string{"Line1", "Line2", "Line3", "Line4"}, "2023-01-01T13:00:00Z", 50}
// 	locationSigns.AddOrUpdateSign("LocationA", sign2) // This won't replace the previous one since the area is lower

// 	// More operations can be performed here...

// 	// Save the current state to the JSON file before exiting
// 	if err := locationSigns.SaveToFile(storageFile); err != nil {
// 		panic(err)
// 	}
// }
