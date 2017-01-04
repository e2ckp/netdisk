package netlayer

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"time"
	"util"
)

var wgetChan = make(chan int)
var playChan = make(chan uint64)

func Get(url string) []byte {
	response, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	return body
}

func Post(url string, contentType string, body io.Reader) []byte {
	response, err := http.Post(url, contentType, body)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	bodyStr, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	return bodyStr
}

func WgetDownload(url string, saveas string, size uint64, hash string) {
	startTime := time.Now()
	fmt.Println(url)
	const thread uint64 = 5
	var last uint64 = thread - 1
	var thunkSize = size / thread
	var start uint64 = 0
	fmt.Println(size)
	fmt.Println(thunkSize)
	var i uint64
	filesList := make([]string, thread, thread)
	for i = 0; i < thread; i++ {
		partName := saveas + ".part" + strconv.FormatUint(i, 10)
		if i == last {
			go startChunkDownload(url, partName, start, size)
		} else {
			go startChunkDownload(url, partName, start, start+thunkSize-1)
		}
		start = start + thunkSize
		filesList[i] = partName
	}
	var j uint64
	for j = 0; j < thread; j++ {
		<-wgetChan
	}
	endTime := time.Since(startTime)
	speed := float64(size/1024) / endTime.Seconds()
	catFile(filesList, saveas)
	fmt.Printf("\n下载完毕,耗时%s,%.2fKB/s,校验MD5中...\n", endTime.String(), speed)
	util.PrintMd5(saveas)
}

func catFile(filesList []string, saveas string) {
	f, err := os.Create(saveas)
	if err != nil {
		panic(err)
	} else {
		f.Close()
	}
	file, err := os.OpenFile(saveas, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	for _, item := range filesList {
		part, err := os.Open(item)
		if err != nil {
			panic(err)
		}
		defer part.Close()
		defer os.Remove(item)
		io.Copy(file, part)
	}
	defer file.Close()
}

func startChunkDownload(url string, saveas string, start uint64, end uint64) {
	fmt.Println("\n")
	fmt.Printf("%s %d %d ", saveas, start, end)
	fmt.Println("\n")
	time.Sleep(100 * time.Millisecond)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	f, err := os.Create(saveas)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Range", fmt.Sprintf(" bytes=%d-%d", start, end))
	res, err := client.Do(req)
	defer res.Body.Close()

	counter := &WriteCounter{}
	counter.Size = end - start
	counter.Part = saveas
	src := io.TeeReader(res.Body, counter)
	count, err := io.Copy(f, src)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if count < 1 {
		os.Exit(1)
	}
	wgetChan <- 1
}

// WriteCounter counts the number of bytes written to it.
type WriteCounter struct {
	Total uint64 // Total # of bytes written
	Size  uint64
	Part  string
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	var per float64 = float64(wc.Total) / float64(wc.Size)
	var i int = int(per * 100)
	fmt.Printf("\r%s%d%% %s %s", util.Bar(i, 25), i, util.ByteFormat(wc.Total), wc.Part)
	return n, nil
}

func Download(url string, saveas string, size uint64, hash string) {
	start := time.Now()
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	f, err := os.Create(saveas)
	if err != nil {
		panic(err)
	}
	counter := &WriteCounter{}
	counter.Size = size
	src := io.TeeReader(res.Body, counter)
	count, err := io.Copy(f, src)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if count < 1 {
		os.Exit(1)
	}
	end := time.Since(start)
	speed := float64(size/1024) / end.Seconds()
	fmt.Printf("\n下载完毕,耗时%s,%.2fKB/s,校验MD5中...\n", end.String(), speed)
	util.PrintMd5(saveas)
}

func DownloadPlay(url string, saveas string, size uint64, hash string) {
	fmt.Println(url)
}

var playNo uint64 = 0
var waitNo uint64 = 1

func PlayStream(url string, saveas string, size uint64, hash string) {
	startTime := time.Now()
	var playerRun bool = false
	var thread uint64 = 4
	var thunkSize uint64 = 1048576
	var startContinue uint64 = 0
	var forceRange uint64 = 0
	if len(os.Args) >= 4 {
		var fl string
		if len(os.Args) > 4 {
			fl = os.Args[4]
		} else {
			fl = os.Args[3]
		}
		match, matchStart, matchEnd := getRange(fl, startContinue, size)
		if match {
			forceRange = matchStart
			size = matchEnd
			startContinue = matchStart
		} else if fl == "-c" {
			thread = thread * 2
		} else if fl == "-t" {
			thunkSize = thunkSize * 5
		} else if fl == "-f" {
			thread = thread * 2
			thunkSize = thunkSize * 5
		} else {

		}
	}

	if size < thunkSize*thread {
		thunkSize = thunkSize / 4
		if size < thunkSize*thread {
			thread = thread / 2
		}
	}

	fmt.Printf("下载中...线程%d,分块大小%dKB\n", thread, thunkSize/1024)

	if stat, err := os.Stat(saveas); os.IsNotExist(err) {
		os.Create(saveas)
		f, err := os.Create(saveas)
		if err != nil {
			panic(err)
		}
		f.Close()
	} else {
		if forceRange > 1 {
			startContinue = forceRange
		} else {
			startContinue = uint64(stat.Size())
		}

		i := int((float64(startContinue) / float64(size)) * 100)
		fmt.Printf("\r%s%d%% %s  %s ", util.Bar(i, 25), i, util.ByteFormat(startContinue), util.BoolString(i > 5, "★", "☆"))
		if !playerRun && (i > 5) {
			playerRun = true
			go callPlayer(saveas)
		}
		if startContinue >= size {
			fmt.Println("\n已下载完毕,校验MD5中...")
			util.PrintMd5(saveas)
			os.Exit(0)
		}
	}
	var start uint64 = startContinue
	var i uint64
	var chunEnd uint64
	for i = 1; i <= thread; i++ {
		playNo = playNo + 1
		chunEnd = start + thunkSize*playNo
		if chunEnd > size {
			chunEnd = size + 1
		}
		go startPlayChunkDownload(url, saveas, start, chunEnd-1, playNo)
		start = chunEnd
	}

	for {
		s := <-playChan
		endTime := time.Since(startTime).Seconds()
		speed := float64((s-startContinue)/1024) / endTime
		i := int((float64(s) / float64(size)) * 100)
		fmt.Printf("\r%s%d%% %s %.2fKB/s %.1fs %s ", util.Bar(i, 25), i, util.ByteFormat(s), speed, endTime, util.BoolString(i > 5, "★", "☆"))
		if !playerRun && (i > 5) {
			playerRun = true
			go callPlayer(saveas)
		}
		if s >= size {
			break
		} else {
			playNo = playNo + 1
			lastEnd := chunEnd
			if lastEnd < size {
				chunEnd = start + thunkSize*playNo
				if chunEnd > size {
					chunEnd = size
				}
				go startPlayChunkDownload(url, saveas, lastEnd, chunEnd-1, playNo)
			}
		}
	}

	endTime := time.Since(startTime)
	speed := float64((size-startContinue)/1024) / endTime.Seconds()
	fmt.Printf("\n下载完毕,耗时%s,%.2fKB/s,校验MD5中...\n", endTime.String(), speed)
	util.PrintMd5(saveas)

}

func callPlayer(file string) {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("PotPlayerMini.exe", file)
		cmd.Start()
	} else {
		cmd := exec.Command("mpv", file)
		cmd.Start()
	}
}

func startPlayChunkDownload(url string, saveas string, start uint64, end uint64, playno uint64) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Range", fmt.Sprintf(" bytes=%d-%d", start, end))
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		go startPlayChunkDownload(url, saveas, start, end, playno)
	} else {
		defer res.Body.Close()
		if playno == 1 {
			file, err := os.OpenFile(saveas, os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				panic(err)
			}
			defer file.Close()
			io.Copy(file, res.Body)
			playChan <- (end + 1)
			waitNo = playno + 1

		} else {
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				panic(err)
			}
			for {
				if waitNo == playno {
					break
				} else {
					time.Sleep(100 * time.Millisecond)
				}
			}
			file, err := os.OpenFile(saveas, os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				panic(err)
			}
			defer file.Close()
			file.Write(body)
			waitNo = playno + 1
			playChan <- (end + 1)
		}
	}

}

func getRange(str string, start uint64, end uint64) (bool, uint64, uint64) {
	var rangeXp = regexp.MustCompile(`^(\d+)-(\d+)$`)
	var rangeXpAll = regexp.MustCompile(`^(\d+)-$`)
	var matched bool = false
	var matchStart uint64
	var matchEnd uint64
	if rangeXp.MatchString(str) {
		match := rangeXp.FindStringSubmatch(str)
		strSt1, _ := strconv.Atoi(match[1])
		matchStart = uint64(strSt1)
		strSt2, _ := strconv.Atoi(match[2])
		matchEnd = uint64(strSt2)
		if matchStart < 8192 {
			matchStart = matchStart * 1048576
		}
		if matchEnd < 8192 {
			matchEnd = matchEnd * 1048576
		}
		if (matchEnd < matchStart) || (matchStart >= end) {
			fmt.Println("error range")
			os.Exit(1)
		} else if matchEnd > end {
			matchEnd = end
		}
		matched = true

	} else if rangeXpAll.MatchString(str) {
		match := rangeXpAll.FindStringSubmatch(str)
		strSt1, _ := strconv.Atoi(match[1])
		matchStart = uint64(strSt1)
		if matchStart < 8192 {
			matchStart = matchStart * 1048576
		}
		if matchStart > end {
			fmt.Println("error range")
			os.Exit(1)
		} else {
			matchEnd = end
		}
		matched = true
	}
	return matched, matchStart, matchEnd

}
