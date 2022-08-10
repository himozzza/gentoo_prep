package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
)

func main() {
	user, _ := user.Current()
	if user.Username != "root" {
		log.Fatalf("\nYou must be root, not %s!\n", user.Username)
	}
	arguments := os.Args
	targetDir := prepare()
	if arguments[1] == "--mount" || arguments[1] == "-m" {
		mounting(targetDir)
	} else if arguments[1] == "--help" || arguments[1] == "-h" {
		fmt.Printf("-m, --mount    Mounting and chroot without downloading stage3.\n-h, --help    Print this help page.\n\n")
	} else {
		selectDistr, DistRelease := selectDist(targetDir)
		release, pattern := parsingData(selectDistr, DistRelease)
		downloadData(release, pattern, targetDir)
		mounting(targetDir)
	}
}

func prepare() string {
	/*
		Выбор диска и монтирование раздела в /mnt/gentoo.
	*/
	var targetDir string = "/mnt/gentoo"
	var numberOfDrive int
	fmt.Printf("Welcome to Gentoo easy chrooting!\n\n")
	re := regexp.MustCompile("/dev/[a-z]{3}[0-9]|/dev/(.*?)p[0-9]")
	lsblk, _ := exec.Command("lsblk", "-lnpo", "KNAME").Output()
	drives := re.FindAllString(string(lsblk), -1)

	for n, i := range drives {
		n++
		fmt.Printf("%d) %s", n, string(i))
		fmt.Println(" ")
	}

	fmt.Println("\nSelect root drive: ")
	_, err := fmt.Scanf("%d", &numberOfDrive)
	if err != nil {
		log.Fatalf("Please, input digit.")
	} else if numberOfDrive > len(drives) {
		log.Fatal("Please, input number of range.")
	}
	numberOfDrive--
	fmt.Printf("\n----------\n\n")
	os.MkdirAll(targetDir, os.ModePerm)

	_, err = exec.Command("mount", drives[numberOfDrive], targetDir).Output()
	if err != nil {
		fmt.Printf("Mounting error.\n1. Try umount %s after run this script.\n2. Check mounting drive for valid.\n\n ", drives[numberOfDrive])
		os.Exit(0)
	}
	return targetDir
}

func selectDist(targetDir string) (string, string) {
	/*
		Выбор дистрибутива.
	*/
	homeURL := fmt.Sprintf("https://mirror.yandex.ru/gentoo-distfiles/releases/amd64/autobuilds/")
	resp, _ := http.Get(homeURL)

	url, _ := io.ReadAll(resp.Body)
	re := regexp.MustCompile("stage3-(.*?)[^txt]\"")
	prepre := re.FindAllString(string(url), -1)
	DistRelease := make(map[int]string)
	for n, i := range prepre {
		n++
		i = strings.Replace(i, "/\"", "", -1)
		DistRelease[n] = string(i)
	}

	var selectDistr int

	keys := make([]int, 0)

	for k := range DistRelease {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	for _, k := range keys {
		fmt.Printf("%d. %s\n", k, DistRelease[k])
	}

	fmt.Printf("\nSelect Distr: ")

	_, err := fmt.Scanf("%d", &selectDistr)
	fmt.Printf("\n----------\n")
	if err != nil {
		exec.Command("umount", "-R", targetDir).Run()
		log.Fatal("\nInput error! Select the gentoo redaction number.\n\n")
	} else if selectDistr > len(DistRelease) {
		exec.Command("umount", "-R", targetDir).Run()
		log.Fatal("\nInput error! Number out of range.\n\n")
	}
	sel := fmt.Sprintf("%scurrent-%s", homeURL, DistRelease[selectDistr])
	return sel, DistRelease[selectDistr]
}

func parsingData(selectDistr, DistRelease string) (string, string) {
	/*
		Получаем url до выбранного дистрибутива.
	*/
	pattern := fmt.Sprintf("%s-[0-9]{8}[A-Z][0-9]{6}[A-Z].tar.xz", DistRelease)
	re := regexp.MustCompile(pattern)

	prepareURL := fmt.Sprintf("%s/", selectDistr)
	resp, err := http.Get(prepareURL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	url, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("\nError Downloading.\n\n")
	}
	pattern = re.FindString(string(url))
	release := fmt.Sprintf("%s%s", prepareURL, pattern)

	return release, pattern
}

func downloadData(release, pattern, targetDir string) {
	/*
		Скачивание и распаковка дистрибутива.
	*/
	fmt.Printf("\nDownloading %s...\n", pattern)
	os.Chdir(targetDir)

	req, _ := http.NewRequest("GET", release, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	f, _ := os.OpenFile(pattern, os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"",
	)
	io.Copy(io.MultiWriter(f, bar), resp.Body)

	fmt.Printf("Unpacking tar.xz file...\n")
	exec.Command("tar", "xpvf", pattern, "--xattrs-include='*.*'", "--numeric-owner").Run()
	time.Sleep(1 * time.Second)
	os.Remove(pattern)
}

func mounting(targetDir string) {
	/*
		Монтирование.
	*/
	os.Chdir(targetDir)
	fmt.Printf("Mounting and Chrooting...\n")
	exec.Command("cp", "--dereference", "/etc/resolv.conf", "etc/").Run()
	exec.Command("mount", "--types", "proc", "/proc", "proc").Run()
	exec.Command("mount", "--rbind", "/sys", "sys").Run()
	exec.Command("mount", "--make-rslave", "sys").Run()
	exec.Command("mount", "--rbind", "/dev", "dev").Run()
	exec.Command("mount", "--make-rslave", "dev").Run()
	exec.Command("mount", "--bind", "/run", "run").Run()
	exec.Command("mount", "--make-slave", "run").Run()

	fmt.Printf("Complete!\n\n")

	fmt.Printf("ATTENTION!!!\nInput 'chroot %s /bin/bash' and 'source /etc/profile' for chrooting to your new Gentoo :)\n\n", targetDir)
}
