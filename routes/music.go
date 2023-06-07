package routes

import (
	"encoding/json"
	"html/template"
	"math/rand"
	"music/controllers"
	"music/types"
	"net/http"
	"os"
	"time"
)

var store = controllers.Store

func Skip(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	time := session.Values["time"].(string)
	times := []string{"0.5", "1", "3", "5", "15", "30"}                       // Array of possible length of song snippets
	session.Values["guesses-left"] = session.Values["guesses-left"].(int) - 1 // Decrement guesses left
	session.Save(r, w)                                                        // Save session

	for i, t := range times { // Loop through times
		if t == time { // If time is equal to the current time
			if i+1 < len(times) { // If the next time is within the bounds of the array
				session.Values["time"] = times[i+1] // Set the time to the next time
				session.Save(r, w)                  // Save session
				break                               // Break out of loop
			}
		}
	}

	songName := session.Values["song"].(string) // Get song name from session
	time = session.Values["time"].(string)      // Get time from session
	songs, _ := os.ReadDir("songs/" + songName) // Read directory of song containing all time snippets

	for _, songLength := range songs { // Loop through songs
		if songLength.Name() == time+".mp3" { // If time snippet is equal to the current time
			file, err := os.Open("songs/" + songName + "/" + songLength.Name()) // Open file
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer file.Close() // Close file when finished

			fileInfo, _ := file.Stat()           // Get file info
			var fileSize int64 = fileInfo.Size() // Get file size
			buffer := make([]byte, fileSize)     // Create buffer with file size

			file.Read(buffer) // Read file into buffer

			w.Write(buffer) // Write buffer to response
		}
	}
}

func GetAudio(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")       // Get session
	songName := session.Values["song"].(string) // Get song name from session
	time := session.Values["time"].(string)     // Get time from session
	songs, _ := os.ReadDir("songs/" + songName) // Read directory of song containing all time snippets

	for _, songLength := range songs { // Loop through songs
		if songLength.Name() == time+".mp3" { // If time snippet is equal to the current time
			file, err := os.Open("songs/" + songName + "/" + songLength.Name()) // Open file
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer file.Close() // Close file when finished

			fileInfo, _ := file.Stat()           // Get file info
			var fileSize int64 = fileInfo.Size() // Get file size
			buffer := make([]byte, fileSize)     // Create buffer with file size

			file.Read(buffer) // Read file into buffer

			w.Write(buffer) // Write buffer to response
		}
	}
}

func GetMusic(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./public/templates/music.html") // Parse template

	songs, _ := os.ReadDir("songs") // Read directory of songs

	// Get all names of each song in songs and put into an array
	allSongNames := []string{}
	for _, song := range songs {
		allSongNames = append(allSongNames, song.Name())
	}

	session, _ := store.Get(r, "session")              // Get session
	randSeed := rand.NewSource(time.Now().UnixNano())  // Create random seed
	song := songs[rand.New(randSeed).Intn(len(songs))] // Get random song from available songs
	session.Values["song"] = song.Name()               // Set song in session
	session.Values["time"] = "0.5"                     // Set time in session to start value of 0.5s
	session.Values["guesses-left"] = 6                 // Set guesses left in session to 6
	session.Save(r, w)                                 // Save session

	if err != nil { // Check for errors
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pageData := types.MusicPageData{AllSongNames: allSongNames} // Create page data

	err = tmpl.Execute(w, pageData) // Execute template
	if err != nil {                 // Check for errors
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func Guess(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")       // Get session
	songName := session.Values["song"].(string) // Get song name from session

	w.Header().Set("Content-Type", "application/json") // Set content type to json

	if session.Values["guesses-left"].(int) <= 0 { // If no more guesses left
		w.Write([]byte(`{"finished": true}`)) // Write finished to response
		return
	}
	session.Values["guesses-left"] = session.Values["guesses-left"].(int) - 1 // Decrement guesses left
	session.Save(r, w)                                                        // Save session

	decoder := json.NewDecoder(r.Body) // Create json decoder
	type Guess struct {                // Create guess struct
		Guess string `json:"guess"`
	}
	var guess Guess
	err := decoder.Decode(&guess) // Decode json into guess struct
	if err != nil {
		w.Write([]byte(`{"correct": false}`))
		return
	}

	guessesLeft := session.Values["guesses-left"].(int) // Get guesses left from session
	if guess.Guess == songName {                        // If guess is correct
		w.Write([]byte(`{"correct": true}`))
	} else if guessesLeft <= 0 { // If no more guesses left and wrong
		w.Write([]byte(`{"finished": true}`))
	} else { // If guess is wrong
		w.Write([]byte(`{"correct": false}`))
	}

}

func Finish(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")              // Get session
	song := session.Values["song"].(string)            // Get song name from session
	session.Options.MaxAge = -1                        // Set session max age to -1 to delete session
	session.Save(r, w)                                 // Save session
	w.Header().Set("Content-Type", "application/json") // Set content type to json
	w.Write([]byte(`{"song": "` + song + `"}`))        // Write song name to response
}
