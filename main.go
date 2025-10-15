package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/gabriel-vasile/mimetype"
	"github.com/schollz/progressbar/v3"
	"github.com/tidwall/gjson"
	"github.com/u2takey/ffmpeg-go"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	start := time.Now()

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Provide a directory :)")
		os.Exit(1)
	}
	dir := args[0]

	files := make(chan string)
	videos := make(chan string)

	find(files, dir)

	bar := progressbar.Default(-1)

	go func() {
		for filename := range files {
			if isVideo(filename) {
				videos <- filename
				bar.AddMax(1)
			}
		}
		close(videos)
	}()

	var total float64
	var wg sync.WaitGroup

	for i := range videos {
		video := i
		wg.Add(1)
		go func() {
			dur := videoDuration(video)
			total += dur
			wg.Done()
			bar.Add(1)
		}()
	}

	bar.Finish()

	wg.Wait()
	end := time.Now()

	fmt.Println(colorString("total duration:", color.FgHiYellow), time.Duration(total*math.Pow(10, 9)))
	fmt.Println(colorString("program runtime:", color.FgHiGreen), duration(start, end))
}

func find(c chan string, dir string) {
	go func() {
		entries, _ := os.ReadDir(dir)
		channels := make(map[string]chan string)
		for _, i := range entries {
			fullPath := path.Join(dir, i.Name())
			c <- fullPath
			if i.IsDir() {
				channels[fullPath] = make(chan string)
				find(channels[fullPath], fullPath)
			}
		}
		for _, v := range channels {
			for filename := range v {
				c <- filename
			}
		}
		close(c)
	}()
}

func videoDuration(filename string) float64 {
	info, _ := ffmpeg_go.Probe(filename)
	d, _ := strconv.ParseFloat(gjson.Get(info, "format.duration").Str, 32)
	return float64(d)
}

func isVideo(filename string) bool {
	mime, e := mimetype.DetectFile(filename)
	if e != nil {
		return false
	}
	mstring := mime.String()
	if strings.Contains(mstring, "video/") || strings.Contains(mstring, "audio/") {
		return true
	}
	return false
}

func duration(start, end time.Time) string {
	return end.Sub(start).String()
}

func colorString(str string, col color.Attribute) string {
	return color.New(col).Sprint(str)
}
