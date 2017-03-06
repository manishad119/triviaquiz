package main 

import (
	"net"
	"os"
	"fmt"
	"time"
	"flag"
	"github.com/manish119/triviaquiz/quiz"
	"github.com/manish119/triviaquiz/onlinequiz/common"
)

/*Read data from the buffer until '\x00' occurs
and unmarshall its JSON encoding*/



/*Data about the client we are talking to
including the questions. Acts as a wrapper for Conn*/

type QuizClient struct {
	Name string
	Conn net.Conn
	Questions []quiz.QuizQuestion
	NQues int //Requested number of question not actually built
	Score int
	QType string //Requested question type


}



/*Send score to client as JSON*/

func sendScore(client *QuizClient){
	dataMap:=make(map[string]interface{})
	dataMap["type"]="score"
	dataMap["score"]=client.Score
	dataMap["total"]=len(client.Questions)
	common.SendJSONData(client.Conn,dataMap)
}






/*Send question to client as JSON*/

func sendQuestion(client *QuizClient, qindex int){
	dataMap:=make(map[string]interface{})
	dataMap["type"]="question"
	dataMap["question"]=client.Questions[qindex].AsQuestion()
	common.SendJSONData(client.Conn,dataMap)

}

/*Check answer and send result (correct?,valid?) to client
Return whether or not we have valid answer to tell the
caller whether we move to next question*/

func checkAnswerResult(client *QuizClient, qindex int,ans string) bool{
	dataMap:=make(map[string]interface{})
	dataMap["type"]="result"
	correct,valid:=client.Questions[qindex].CheckAnswer(ans)
	
	dataMap["valid"]=valid
	dataMap["correct"]=correct
	if correct {
		client.Score++
	} 
	if valid {
		dataMap["ans"]=client.Questions[qindex].Answer()
	} 

	
	common.SendJSONData(client.Conn,dataMap)
	return valid

}

func processInitialRequest(conn net.Conn) *QuizClient {
	data:=common.ReadMessage(conn)
	//fmt.Printf("%+v\n",data)
	/*Read the initial request data*/
	if (data==nil){
		common.SendError(conn,"DataError","Error occured while reading request data")
		return nil
	}

	reqtyperaw,ok:=data["type"]
	if !ok {
		common.SendError(conn,"InvalidRequest","Request not proper")
		return nil
	}
	reqtype,ok:=reqtyperaw.(string)
	if(!ok){
		common.SendError(conn,"InvalidType","Request not proper")
		return nil
	}
	if(reqtype!="request"){
		common.SendError(conn,"InvalidType",fmt.Sprintf("Unknown type '%s'\n",reqtype))
		return nil
	}

	
	nqraw,ok:=data["nq"]
	if (!ok){
		common.SendError(conn,"MissingArgument","number of questions is missing")
		return nil
	}
	nq1,ok:=nqraw.(float64)
	if (!ok){
		common.SendError(conn,"InvalidNumber",fmt.Sprintf("invalid number of questions %v %T",nqraw,nqraw))
		return nil
	}
	nq:=int(nq1)

	if nq<0 {
		common.SendError(conn,"InvalidNumber",fmt.Sprintf("invalid number of questions %v",nq))
		return nil
	}

	qt:=""
	qtraw,ok:=data["qt"]
	if ok {
		qt,ok=qtraw.(string)
		if(!ok){
			common.SendError(conn,"InvalidQType",fmt.Sprintf("Question type '%v' is invalid",qtraw))
			return nil
		}
		if qt!="multiple" && qt!="boolean"  {
			common.SendError(conn,"InvalidQType",fmt.Sprintf("Invalid question type '%s', must be boolean/multiple",qt))
			return nil

		}
	}


	clientname:="<anonymous>"
	if vnam,ok:=data["name"];ok {
		if clientname,ok=vnam.(string);!ok {
			common.SendError(conn,"Name error",fmt.Sprintf("Invalid name'%s'",vnam))
			return nil
		}


	}

	return &QuizClient{Name:clientname,Conn:conn,NQues:nq,Score:0,QType:qt}



}
/*Processes client's answer to the current question and returns whether to continue
with next question and whether to terminate*/

func processClientAnswer(client *QuizClient,qindex int) (bool,bool){
	data:=common.ReadMessage(client.Conn)
	//fmt.Printf("%+v\n",data)
	/*Read the initial request data*/
	if (data==nil){
		common.SendError(client.Conn,"DataError","Error occured while reading request data")
		return false,true
	}

	reqtyperaw,ok:=data["type"]
	if !ok {
		common.SendError(client.Conn,"InvalidRequest","Request not proper")
		return false,false
	}
	reqtype,ok:=reqtyperaw.(string)
	if(!ok){
		common.SendError(client.Conn,"InvalidType","Request not proper")
		return false,false
	}
	if(reqtype!="answer" && reqtype!="repeat"){
		common.SendError(client.Conn,"",fmt.Sprintf("InvalidType '%s'",reqtype))
		return false,false
	}
	//If it has asked to repeat question send question again and don't continue

	if reqtype=="repeat"{
		sendQuestion(client,qindex)
		return false,false

	}

	
	ansraw,ok:=data["answer"]
	if (!ok){
		common.SendError(client.Conn,"MissingArgument","Answer is missing")
		return false,false
	}
	ans,ok:=ansraw.(string)
	if (!ok){
		common.SendError(client.Conn,"InvalidAnswer",fmt.Sprintf("invalid answer %v",ansraw))
		return false,false
	}

	return checkAnswerResult(client,qindex,ans),false

	

}

/*Main function which */

func handleConnection(conn net.Conn,tc int){
	defer conn.Close()
	
	client:=processInitialRequest(conn)
	if (client==nil){
		fmt.Fprintf(os.Stderr,"Error processing initial request.. terminating connection\n")
		return
	}
	fmt.Printf("Downloading %v questions from opentdb.com...\n",client.NQues)
	time.Sleep(time.Second*2)
	
	jsonresult:=quiz.GetQuestionJSON(client.NQues,client.QType)
	client.Questions=quiz.BuildQuestions(jsonresult,tc)
	nq:=len(client.Questions)
	fmt.Printf("\n%v of %v questions successfully downloaded and parsed\n",nq,client.NQues)
	fmt.Printf("Starting the quiz...\n")
	
	for i,_:=range client.Questions{
		time.Sleep(time.Second)
		
		sendQuestion(client,i)
		time.Sleep(time.Second)
		valid,terminate:=processClientAnswer(client,i)
		//question.PrintAnswer()
		
		for !valid && !terminate{
			time.Sleep(time.Second)

			valid,terminate=processClientAnswer(client,i)
			//fmt.Printf("%v\n",valid)
		}
		/*If we are to terminate return from here which will close the connection*/

		if terminate {
			return
		}


	}
	time.Sleep(time.Second)

	sendScore(client)

}

func main(){
	portptr:=flag.Int("port",8049,"Port for the server")
	tcptr:=flag.Int("tc",3,"Number of goroutines to process each client question")
	flag.Parse()
	fmt.Printf("Starting server on port %v\n",*portptr)
	lis,err:=net.Listen("tcp",fmt.Sprintf(":%v",*portptr))
	if err!=nil {
		fmt.Fprintf(os.Stderr,"Fatal error: %s ..\nShutting down server\n",err.Error())
		return
	}
	defer lis.Close()
	//Wait for connections from quiz clients and handle them 
	for {
		conn,err:=lis.Accept()
		if err!=nil {
			fmt.Fprintf(os.Stderr,"Connection from client %s failed\n",conn.RemoteAddr().String())
			continue
		}
		fmt.Printf("Incoming connection from client %s accepted\n",conn.RemoteAddr().String())
		go handleConnection(conn,*tcptr)

	}

}