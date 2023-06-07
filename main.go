package main

import (
	"fmt"
	"music/routes"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()

	if os.Getenv("PORT") == "" {
		fmt.Println("No port provided, using default port 3000")
		os.Setenv("PORT", "3000")
	}
}

func GenerateAudioClip(inputFile string, time string) {
	os.Mkdir("songs/"+inputFile, 0777)
	app := "ffmpeg"
	arg0 := "-i"
	arg1 := "songs/" + inputFile + ".mp3"
	arg2 := "-ss"
	arg3 := "0"
	arg4 := "-t"
	arg5 := time
	arg6 := "songs/" + inputFile + "/" + time + ".mp3"
	arg7 := "-y"

	cmd := exec.Command(app, arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
	_, err := cmd.Output()
	if err != nil {
		fmt.Println("1")
		fmt.Println(err.Error())
		return
	}
}

func InitialiseSong(songName string) {
	app := "ffmpeg"
	arg0 := "-i"
	arg1 := "inputs/" + songName
	arg2 := "-ss"
	arg3 := "0"
	arg4 := "-t"
	arg5 := "30"
	arg6 := "-af"
	arg7 := "silenceremove=1:0:-50dB"
	arg8 := "songs/" + songName
	arg9 := "-y"

	cmd := exec.Command(app, arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9)
	_, err := cmd.Output()
	if err != nil {
		fmt.Println("2")
		fmt.Println(err.Error())
		return
	}

	os.Remove("inputs/" + songName)

	times := []string{"0.5", "1", "3", "5", "15", "30"}
	for _, time := range times {
		GenerateAudioClip(strings.Split(songName, ".mp3")[0], time)
	}
	os.Remove("songs/" + songName)
}

func main() {
	allSongs, _ := os.ReadDir("inputs")
	for _, song := range allSongs {
		InitialiseSong(song.Name())
	}

	fmt.Println("Server started on port " + os.Getenv("PORT"))

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.CleanPath)
	r.Use(middleware.Heartbeat("/ping"))
	r.Handle("/css/*", http.StripPrefix("/css/", http.FileServer(http.Dir("./public/assets/css"))))
	r.Handle("/img/*", http.StripPrefix("/img/", http.FileServer(http.Dir("./public/assets/img"))))
	r.Handle("/js/*", http.StripPrefix("/js/", http.FileServer(http.Dir("./public/assets/js"))))

	r.Get("/", routes.GetMusic)
	r.Get("/audio", routes.GetAudio)
	r.Get("/skip", routes.Skip)
	r.Post("/guess", routes.Guess)
	r.Post("/finish", routes.Finish)

	http.ListenAndServe(":"+os.Getenv("PORT"), r)
}
