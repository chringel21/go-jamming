package main

import (
	"fmt"
   	"log"
	"net/http"
	"io/ioutil"
)

func mainz() {
	fmt.Println("Hello, playground")
	resp, err := http.Get("https://brainbaking.com/notes")
	if err != nil {
		log.Fatalln(err)
	}

   body, err2 := ioutil.ReadAll(resp.Body)
   if err2 != nil {
   		log.Fatalln(err)
   } 

   fmt.Printf("tis ditte")
	fmt.Printf("%s", body)
}