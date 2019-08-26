package main

import (
	"fmt"
	"github.com/hashicorp/terraform/config"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
)

func main() {

	fmt.Println("Test")
	//var config jettison.
	//
	//pwd, _ := os.Getwd()
	//data, err := ioutil.ReadFile(pwd + "/nsxt/config.yml")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//err2 := yaml.Unmarshal(data, &config)
	//if err2 != nil {
	//	log.Fatal(err)
	//}
	//
	//fmt.Println("vCenter hostname", config.Ecs.Vcenter.Hostname)
	//fmt.Println("vCenter username", config.Ecs.Vcenter.Username)
	//fmt.Println("----------------------------------------------")
	//fmt.Println("vCenter hostname", config.Ecs.Nsxt.Hostname)
	//fmt.Println("vCenter username", config.Ecs.Nsxt.Username)
	//fmt.Println("----------------------------------------------")
	//
	//fmt.Println(config.Ecs.Controllers[0])

}
