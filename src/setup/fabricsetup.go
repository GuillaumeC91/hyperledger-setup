package main

import (
	"fmt"
	"log"
	"docker"
	"configuration"
	"strconv"
	"strings"
	"errors"
	"github.com/mholt/archiver"
	"os"
	"net/http"
	"io"
)

func dockerFabricPull(){
	
	arch := configuration.GetImageArch()
	
	for i := 0; i < len(configuration.IMAGES); i++ {
		
		imageName := configuration.DOCKER_IMG_PREFIX + configuration.IMAGES[i] + ":" + arch + "-" + configuration.VERSION
		tagName := configuration.DOCKER_IMG_PREFIX + configuration.IMAGES[i]
		
		fmt.Println("Pulling Docker image: " + imageName)
		
		//Docker Pull images
		message, err := docker.ExecDockerCmd("pull", imageName)
		checkErr("Error while running docker pull", err) //Will stop here if an error is encountered
		fmt.Println("Docker Pull:", message)
		
		//Docker Tag
		message, err = docker.ExecDockerCmd("tag", imageName, tagName)
		checkErr("Error while running docker tag", err) //Will stop here if an error is encountered
		fmt.Println("Docker Tag:", message)
	}
	
	imageName := configuration.DOCKER_IMG_PREFIX + configuration.BASE_DOCKER_NAME + ":" + arch + "-" + configuration.BASE_DOCKER_TAG
	
	//Docker Pull baseos
	message, err := docker.ExecDockerCmd("pull", imageName)
	checkErr("Error while running docker pull", err) //Will stop here if an error is encountered
	fmt.Println("Docker Pull:", message)
}

func getBinaries(location string){

	for i := 0; i < len(configuration.DOWNLOADS); i++ {
		fmt.Println("Downloading file: " + configuration.DOWNLOADS[i])
	
		location, tarFile := downloadFromUrl(configuration.DOWNLOADS[i], location)
		
		dlLocation := location + tarFile
		fmt.Println("Downloaded file to: " + dlLocation)
		
		err := archiver.TarGz.Open(dlLocation, location)
		
		//Will stop here if an error is encountered
		checkErr("Error while extracting TAR file: " + tarFile, err)
	}	
}

func downloadFromUrl(url string, location string) (string, string) {
	tokens := strings.Split(url, "/")
	fileName := tokens[len(tokens)-1]
	
	if !strings.HasSuffix(location, "/"){
		location = location + "/"
	}
	
	fmt.Println("Downloading", url, "to", location + fileName)
	
	err := os.MkdirAll(location, os.ModeDir)
	checkErr("Error while creating directory: " + location, err)

	newFile := location + fileName
	output, err := os.Create(newFile)	
	defer output.Close()
	checkErr("Error while creating: " + fileName, err)

	response, err := http.Get(url)
	defer response.Body.Close()
	checkErr("Error while downloading from: " + url, err)

	n, err := io.Copy(output, response.Body)
	checkErr("Error while downloading from: " + url, err)

	fmt.Println(n, "bytes downloaded.")
	
	return location, fileName
}

func checkDockerReq() (bool, error){
	
	if docker.IsDockerInstalled() {
		
		version, err := docker.GetDockerVersion()
		
		checkErr("Error while getting Docker version", err)
			
		verInt, err := strconv.Atoi(strings.Replace(version, ".", "", -1))
		checkErr("Error while parsing Docker version", err)
		
		minVerInt, err := strconv.Atoi(strings.Replace(configuration.MIN_DOCKER_VER, ".", "", -1))
		checkErr("Error while parsing minimum Docker version", err)
		
		if verInt < minVerInt {
			//Minimum version not satisfied
			return false, errors.New("Minimum Docker version [" + configuration.MIN_DOCKER_VER + "] not satisfied")
		} else{
			return true, nil
		}
		
	} else{
		//Error, docker not installed
		return false, errors.New("Docker not installed")
	}
	
}

func checkErr(message string, err error) {
    if err != nil {
        log.Fatal("ERROR: ", message, ": ", err)
    }
}

func main(){
	fmt.Println("Setup starting")
	fmt.Println("Verifying Docker requirements")
	
	//Check Docker Requirements
	dockerReq, err := checkDockerReq()
	checkErr("Docker requirements unsatisfied", err)
	
	fmt.Println("Docker requirements satisfied: " + strconv.FormatBool(dockerReq))
	
	fmt.Println("Pulling Docker images")
	
	//Pull down images
	dockerFabricPull()
	
	fmt.Println("Downloading binaries")
	
	//Get binaries, current directory
	getBinaries(".")
}