package main

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"io/ioutil"
	"strings"
	"testing"
)

func coppycol(filename string,sheetname string){
	filebyte ,err := ioutil.ReadFile(filename)  //./interface_col.xlsx
	if err != nil {
		fmt.Println(err)
	}
	xlsx, err := excelize.OpenReader(strings.NewReader(string(filebyte)))
	if err != nil {
		fmt.Println(err)
		return
	}
	rows := xlsx.GetRows(sheetname) //interface_col

	var prentidSlice []string
	idNewidMap := make(map[string]string)

	for i, row := range rows {
		if i==0{
			continue
		}
		prentidSlice = append(prentidSlice, row[13])
		_id := row[0]
		newid := row[19]
		idNewidMap[_id] = newid
	}

	fmt.Println("prentidSlice:", prentidSlice)
	fmt.Println("idNewidMap:", idNewidMap)

	for j,parentid := range prentidSlice{
		for k,v := range idNewidMap{
			if parentid == k{
				prentidSlice[j]=v
			}
		}
	}
	fmt.Println("prentidSlice:", prentidSlice)
}

//读取exl放到map中
func creatIdpath(filename string,sheetname string)map[string]string{
	idpathMap := make(map[string]string)
	filebyte ,err := ioutil.ReadFile(filename)  //./interface_col.xlsx
	if err != nil {
		fmt.Println(err)
	}
	xlsx, err := excelize.OpenReader(strings.NewReader(string(filebyte)))
	if err != nil {
		fmt.Println(err)
		return idpathMap
	}
	rows := xlsx.GetRows(sheetname) //interface_col

	for i, row := range rows {
		if i==0{
			continue
		}
		_id := row[0]
		path := row[1]
		idpathMap[_id] = path
	}
	return idpathMap
}
type interfaceyapi struct {
	localfilename string
	localsheetname string
	yapifilename string
	yapisheetname string
}
//匹配相同ID
func Matchthesameid(interfaceyapi *interfaceyapi) []string {
	localmap := creatIdpath(interfaceyapi.localfilename,interfaceyapi.localsheetname) //"./interface.xlsx","interface"
	servermap := creatIdpath(interfaceyapi.yapifilename,interfaceyapi.yapisheetname)  //"./interfaceyapi.xlsx","interface"

	fmt.Println("loaclmap:", localmap)
	fmt.Println("servermap:", servermap)
	//找到匹配的interfaceid
	var importid []string
	for _,path := range servermap{
		for k,v := range localmap{
			if path == v{
				importid = append(importid, k)
			}
		}
	}
	fmt.Println("importid:", importid)
	return importid
}
func TestFile(t *testing.T)  {
	interfaceyapi := &interfaceyapi{
		localfilename : "./interface.xlsx",
		localsheetname : "interface",
		yapifilename : "./interfaceyapi.xlsx",
		yapisheetname : "interface",
	}
	id := Matchthesameid(interfaceyapi)
	fmt.Println(id)
}
