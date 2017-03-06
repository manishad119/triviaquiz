package common

import (
	"fmt"
	"os"
	"encoding/json"
	"net"
	"bufio"
)

/*Common functions
e.g. communication protocol shared by client
and server in online quiz*/



func ReadMessage(conn net.Conn) map[string]interface{}{
	bufreader:=bufio.NewReaderSize(conn,1024)
	

	jsondata,err:=bufreader.ReadBytes(byte(0))
	if err!=nil {
		fmt.Fprintf(os.Stderr,"Error reading JSON: %s\n",err.Error())
		return nil
	}
	if err!=nil {
		fmt.Fprintf(os.Stderr,"Error: %s\n",err.Error())
		return nil
	}

	jsondata=jsondata[:len(jsondata)-1]
	dataMap:=make(map[string]interface{})
	err=json.Unmarshal(jsondata,&dataMap)
	if(err!=nil){
		fmt.Fprintf(os.Stderr,"Error: %s\n",err.Error())
		return nil
	}
	return dataMap



}
/*Send map encoded as JSON*/

func SendJSONData(conn net.Conn,dataMap map[string]interface{}){
	encodedjson,err:=json.Marshal(dataMap)
	if err!=nil {
		fmt.Fprintf(os.Stderr,"Error: %s\n",err.Error())
		return
	}
	encodedjson=append(encodedjson,byte(0))
	_,err=conn.Write(encodedjson)
	if(err!=nil){
		fmt.Fprintf(os.Stderr,"Error: %s\n",err.Error())
		return
	}


}

/*Send error message to client as JSON*/

func SendError(conn net.Conn,name string, msg string){

	dataMap:=make(map[string]interface{})
	dataMap["type"]="error"
	dataMap["name"]=name
	dataMap["msg"]=msg
	SendJSONData(conn,dataMap)


} 