package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"syscall"
)

func main() {
	targetDir := prepare()
	selectDistr := selectDist(targetDir)
	release, pattern := parsingData(selectDistr)
	downloadData(release, pattern, targetDir)
	mounting(targetDir)
}

func prepare() string {
	var targetDir string = "/mnt/gentoo"
	var n int
	fmt.Printf("Welcome to Gentoo easy chrooting!\n\n")
	re := regexp.MustCompile("/dev/[a-z]{3}[0-9]|/dev/(.*?)p[0-9]")
	a, _ := exec.Command("lsblk", "-lnpo", "KNAME").Output()
	drives := re.FindAllString(string(a), -1)
	// z := strings.SplitN(string(b), "\n", -1)

	for n, i := range drives {
		n++
		fmt.Printf("%d) %s", n, string(i))
		fmt.Println(" ")
	}

	fmt.Println("\nSelect root drive: ")
	_, err := fmt.Scanf("%d", &n)
	if err != nil {
		log.Fatalf("Please, input digit.")
	} else if n > len(drives) {
		log.Fatal("Please, input number of range.")
	}

	fmt.Printf("\n----------\n\n")
	os.MkdirAll(targetDir, os.ModePerm)

	_, err = exec.Command("mount", drives[n], targetDir).Output()
	if err != nil {
		fmt.Printf("Mounting error.\n1. Try umount %s after run this script.\n2. Check mounting drive for valid.\n\n ", drives[n])
		os.Exit(0)
	}
	return targetDir
}

func selectDist(targetDir string) string {
	var selectDistr int
	DistRelease := map[int]string{
		1: "stage3-amd64-desktop-openrc",
		2: "stage3-amd64-desktop-systemd",
		3: "stage3-amd64-nomultilib-openrc",
		4: "stage3-amd64-nomultilib-systemd",
	}

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
	fmt.Printf("\n----------\n\n")
	if err != nil {
		exec.Command("umount", "-R", targetDir).Run()
		log.Fatal("\nInput error! Select the gentoo redaction number.\n\n")
	} else if selectDistr > len(DistRelease) {
		exec.Command("umount", "-R", targetDir).Run()
		log.Fatal("\nInput error! Number out of range.\n\n")
	}

	return DistRelease[selectDistr]
}

func parsingData(selectDistr string) (string, string) {
	pattern := fmt.Sprintf("%s-[0-9]{8}[A-Z][0-9]{6}[A-Z].tar.xz", selectDistr)
	re := regexp.MustCompile(pattern)
	prepareURL := fmt.Sprintf("https://mirror.yandex.ru/gentoo-distfiles/releases/amd64/autobuilds/current-%s/", selectDistr)
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
	fmt.Printf("\nDownloadind %s...\n", pattern)
	os.Chdir(targetDir)
	download, err := http.Get(release)
	if err != nil {
		log.Println(err)
	}
	defer download.Body.Close()

	out, err := os.Create(pattern)
	if err != nil {
		log.Println(err)
	}
	defer out.Close()

	_, err = io.Copy(out, download.Body)

	fmt.Printf("Unpacking tar.xz file...\n")
	exec.Command("tar", "xpvf", pattern, "--xattrs-include='*.*'", "--numeric-owner").Run()
}

func mounting(targetDir string) {
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
	syscall.Chroot("/mnt/gentoo")
	os.Chdir("/")
	exec.Command("source", "/etc/profile").Run()

	exec.Command("emerge-webrsync").Run()
	fmt.Printf("Complete!\n\n")

	fmt.Printf("ATTENTION!!!\nInput 'chroot %s /bin/bash' and 'source /etc/profile' for chrooting to your new Gentoo :)\n\n", targetDir)
}
