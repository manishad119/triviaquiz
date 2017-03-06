package main

import (
	"net"
	"fmt"
	"os"
	"time"
	"flag"
	"github.com/manish119/triviaquiz/onlinequiz/common"
)
/*Answer to question*/
func sendAnswer(conn net.Conn,ans string){
	dataMap:=make(map[string]interface{})
	dataMap["type"]="answer"
	dataMap["answer"]=ans
	common.SendJSONData(conn,dataMap)
}
/*Request to repeat the question*/
func repeatRequest(conn net.Conn){
	dataMap:=make(map[string]interface{})
	dataMap["type"]="repeat"
	common.SendJSONData(conn,dataMap)

}

func sendNewAnswer(conn net.Conn) {
	fmt.Printf("\n Your answer>")
	var ans string
	fmt.Scanf("%s",&ans)
	sendAnswer(conn,ans)
}
/*Process next message and return whether to terminate*/
func processMessage(conn net.Conn) bool{
	data:=common.ReadMessage(conn)
	//fmt.Printf("%+v\n",data)
	/*Read the initial request data*/
	if (data==nil){
		fmt.Fprintf(os.Stderr,"Error:cannot parse response. terminating..\n",)
		return true
	}

	reqtyperaw,ok:=data["type"]
	if !ok {
		fmt.Fprintf(os.Stderr,"Error: invalid answer '%+v'\n",data)
		return false
	}
	reqtype,ok:=reqtyperaw.(string)
	if(!ok){
		fmt.Fprintf(os.Stderr,"Error: invalid type '%v'\n",reqtyperaw)
		return false
	}
	//Do for each kind of response
	if reqtype=="error"{
		//Error message
		fmt.Fprintf(os.Stderr,"Incoming error message %s: %s\n",data["name"],data["msg"])
		return false

	}

	if reqtype=="question" {
		quesraw,ok:=data["question"]
		if !ok {
			//If cannot find question, ask for repeat
			fmt.Fprintf(os.Stderr,"Error: no question. Asking for repeat\n")
			repeatRequest(conn)
			return false
		}
		question,ok:=quesraw.(string)
		//Question not string. Ask for repeat
		if !ok {
			fmt.Fprintf(os.Stderr,"Error: Invalid question '%v'. Asking for repeat\n",quesraw)
			repeatRequest(conn)
			return false
		}

		fmt.Printf("---------NEW QUESTION-----------\n%s\n",question)
		sendNewAnswer(conn)
		return false

	} else if reqtype=="result"{
		//Check whether data is valid
		validraw,ok:=data["valid"]
		if !ok {
			fmt.Fprintf(os.Stderr,"Missing valid? boolean\n")
			sendNewAnswer(conn)
			return false


			
		}
		valid,ok:=validraw.(bool)
		if !ok {
			fmt.Fprintf(os.Stderr,"Error: valid must be a bool but found %v\n",validraw)
			sendNewAnswer(conn)
			return false
		}
		if !valid {
			fmt.Fprintf(os.Stderr,"Server Response: invalid answer. Please give a/b.. for multiple question and t/f or true/false for boolean\n")
			sendNewAnswer(conn)
			return false

		} else {
			correctraw,ok:=data["correct"]
			if !ok {
				fmt.Fprintf(os.Stderr,"Error: Missing correct? don't know result\n")
				return false
			}
			correct,ok:=correctraw.(bool)
			if !ok {
				fmt.Fprintf(os.Stderr,"Error: invalid correct %v need bool",correctraw)
				return false
			}
			if correct {
				fmt.Printf("CONGRATS!!  Correct Answer!!!\n")
			} else {
				fmt.Printf("SORRY!! Wrong answer\n")
				cans,ok:=data["ans"]
				if ok {
					fmt.Printf("Correct answer: %v\n",cans)
				}

			}
			return false
		}



	}else if reqtype=="score"{
		//If score is sent, terminate the program
		rawscore,ok:=data["score"]
		if !ok {
			fmt.Fprintf(os.Stderr,"Error: Missing score\n")
			return true
		}
		score,ok:=rawscore.(float64)
		if !ok {
			fmt.Fprintf(os.Stderr,"Error: Invalid score %v\n",rawscore)
			return true
		}
		fmt.Printf("Your score is %v/%v\n",score,data["total"])

		return true
		//Score 
	} else {
		fmt.Fprintf(os.Stderr,"Invalid response type %v\n",reqtype)
		return false
	}





	return false


}

func sendInitialRequest(conn net.Conn,name string, nq int, qt string) {
	dataMap:=make(map[string]interface{})
	dataMap["type"]="request"
	dataMap["nq"]=nq
	if len(qt)>0{
		dataMap["qt"]=qt
	}
	if len(name)>0{
		dataMap["name"]=name
	}
	common.SendJSONData(conn,dataMap)
	
}

func main(){
	/*Dummy check*/

	serverptr:=flag.String("server","localhost","Quiz server")
	portptr:=flag.Int("port",8049,"server port")
	nqptr:=flag.Int("nq",10,"Number of questions")
	qtptr:=flag.String("qt","","Question type boolean/multiple")
	nameptr:=flag.String("name","","Your name")
	flag.Parse()

	
	conn,err:=net.Dial("tcp",fmt.Sprintf("%s:%v",*serverptr,*portptr))
	if err!=nil {
		fmt.Fprintf(os.Stderr,"Fatal error: %s\n",err.Error())
		return
	}	
	defer conn.Close()

	time.Sleep(time.Second)
	sendInitialRequest(conn,*nameptr,*nqptr,*qtptr)
	term:=processMessage(conn)
	for !term {
		term=processMessage(conn)

	}
	

}
