package commands

import (
	"config"
	"fmt"
	"layers/fslayer"
	"os"
	"path"
	"path/filepath"
	"strings"
	"util"
)

func Use() {
	fmt.Println("select default remote connect")
	fmt.Println(config.Cfg)
}

func Ls() {
	var path string
	if len(os.Args) == 3 {
		path = absNoRoot(os.Args[2])
	} else {
		path = config.Cfg.Path
	}
	fslayer.ListDir(path)
}

func Cd() {
	if len(os.Args) == 3 {
		config.Cfg.Path = absNoRoot(os.Args[2])
		ret := fslayer.ListDir(config.Cfg.Path)
		if ret {
			config.SaveConfig()
		}
	} else {
		fmt.Println("change dir ")
	}
}

func Pwd() {
	fmt.Println(util.DiskName(config.Cfg.Disk) + config.Cfg.Root + "  ➜  " + config.Cfg.Path)
}

func Mv() {
	if len(os.Args) == 4 {
		var source string = absPath(os.Args[2])
		var target string = absPath(os.Args[3])
		ok, _, _ := fslayer.GetFileInfo(source, false)
		if ok {
			fslayer.MoveFile(source, target)
		}
	} else {

	}

}

func Mkdir() {

	if len(os.Args) == 3 {
		var path string = absPath(os.Args[2])
		fslayer.Mkdir(path)
	} else {
		fmt.Println("mkdir")
	}

}

func Rm() {
	if len(os.Args) == 3 {
		var path string = absPath(os.Args[2])
		ok, _, _ := fslayer.GetFileInfo(path, false)
		if ok {
			fslayer.DeleteFile(path)
		}
	} else {

		fmt.Println("rm files")
	}
}

func Get() {
	if len(os.Args) >= 3 {
		var dist string = ""
		if len(os.Args) >= 4 {
			dist = filepath.Clean(os.Args[3])
		}
		var path = absPath(os.Args[2])
		ok, size, hash := fslayer.GetFileInfo(path, false)
		if ok {
			fslayer.Get(path, dist, size, hash)
		}
	} else {
		fmt.Println("Usage get filepath saveas")
	}
}

func Put() {
	if len(os.Args) >= 3 {
		var path string = absLocalPath(os.Args[2])
		var savePath string = absPath(os.Args[2])
		if util.FileOk(path) {
			var ondup string = util.BoolString(len(os.Args) >= 4, "overwrite", "newcopy")
			fslayer.PutFile(path, savePath, ondup)
		} else {
			fmt.Println(path + "不存在或不可读")
		}
	} else {
		fmt.Println("Usage put filepath ")
	}
}

func Wget() {
	if len(os.Args) >= 3 {
		var dist string = ""
		if len(os.Args) >= 4 {
			dist = filepath.Clean(os.Args[3])
		}
		var path = absPath(os.Args[2])
		ok, size, hash := fslayer.GetFileInfo(path, false)
		if ok {
			fslayer.Wget(path, dist, size, hash)
		}
	} else {
		fmt.Println("Usage get filepath saveas")
	}
}

func Sync() {

}

func Info() {
	if len(os.Args) == 2 {
		fslayer.GetInfo()
		fmt.Println("配置文件:" + config.ConfigPath)
	} else {
		fslayer.GetFileInfo(absPath(os.Args[2]), false)
	}
}

func Hash() {
	if len(os.Args) == 3 {
		var filepath string = absLocalPath(os.Args[2])
		util.PrintMd5(filepath)
	} else {
		fmt.Println("hash file")
	}
}

func Play() {
	if len(os.Args) >= 3 {
		var dist string = ""
		if len(os.Args) >= 4 && (!strings.Contains(os.Args[3], "-")) {
			dist = filepath.Clean(os.Args[3])
		}
		var path = absPath(os.Args[2])
		ok, size, hash := fslayer.GetFileInfo(path, false)
		if ok {
			fslayer.GetPlayStream(path, dist, size, hash)
		}
	} else {
		fmt.Println("Usage get filepath saveas")
	}
}

func Help() {

}

func Config() {
	if (len(os.Args) == 2) || (os.Args[2] == "list") {
		config.ConfigList()
	} else if len(os.Args) == 4 && os.Args[2] == "get" {
		config.ConfigGet(os.Args[3])
	} else if len(os.Args) == 5 && os.Args[2] == "set" {
		config.ConfigSet(os.Args[3], os.Args[4])
	} else {
		config.Error()
	}
}

func Task() {
	if (len(os.Args) == 2) || (os.Args[2] == "list") {
		fslayer.GetTaskList()
	} else if len(os.Args) == 5 && os.Args[2] == "add" {
		fslayer.AddTask(absPath(os.Args[3]), os.Args[4])
	} else if len(os.Args) == 4 && os.Args[2] == "remove" {
		fslayer.RemoveTask(os.Args[3])
	} else if len(os.Args) == 4 && os.Args[2] == "info" {
		fslayer.GetTaskInfo(os.Args[3])
	} else {
		config.Error()
	}
}

func Usage() {

	fmt.Println("ls cd rm get put hash ")

}

func Daemon() {
	fmt.Println("daemon start")
}

func absPath(filePath string) string {
	filePath = path.Clean(filePath)
	if path.IsAbs(filePath) {
		if !strings.HasPrefix(filePath, config.Cfg.Root) {
			filePath = fmt.Sprintf("%s/%s", config.Cfg.Root, "."+filePath)
		}
	} else {
		filePath = fmt.Sprintf("%s/%s/%s", config.Cfg.Root, config.Cfg.Path, filePath)
	}
	return path.Clean(filePath)
}

func absNoRoot(filePath string) string {
	filePath = path.Clean(filePath)
	if path.IsAbs(filePath) {
		filePath = filePath
	} else {
		filePath = fmt.Sprintf("%s/%s", config.Cfg.Path, filePath)
	}
	return path.Clean(filePath)
}

func absLocalPath(filePath string) string {
	filePath = filepath.Clean(filePath)
	if filepath.IsAbs(filePath) {
		filePath = filePath
	} else {
		dir, _ := os.Getwd()
		filePath = filepath.Join(dir, filepath.Base(filePath))
	}
	return filepath.Clean(filePath)
}
