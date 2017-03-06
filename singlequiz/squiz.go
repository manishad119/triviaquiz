package  main

import (
	"fmt"
	"flag"
	"time"
	"github.com/manish119/triviaquiz/quiz"
	"os"
)
	
func main() {
	nqptr:=flag.Int("nq",10,"Number of questions")
	qtptr:=flag.String("qt","","Type of question boolean/multiple. Omit if need both")
	tcptr:=flag.Int("tc",3,"Number of goroutines to use")

	flag.Parse()
	qt:=*qtptr
	if (qt!="" && qt!="boolean" && qt !="multiple"){
		fmt.Fprintf(os.Stderr,"Invalid type %s\n",qt)
		return
	}
	fmt.Printf("Downloading %v questions from opentdb.com...\n",*nqptr)
	time.Sleep(time.Second*2)
	
	jsonresult:=quiz.GetQuestionJSON(*nqptr,*qtptr)
	questions:=quiz.BuildQuestions(jsonresult,*tcptr)
	nq:=len(questions)
	fmt.Printf("\n%v of %v questions successfully downloaded and parsed\n",nq,*nqptr)
	fmt.Printf("Starting the quiz...\n")
	time.Sleep(time.Second*4)
	correctcount:=0
	for i,question:=range questions{
		fmt.Printf("Next question in ts..")
		for cdown:=5;cdown>0;cdown--{
			fmt.Printf("\b\b\b\b%vs..",cdown)
			time.Sleep(time.Second*1)
		}
		fmt.Printf("\b\b\b\b0s..\nQuestion %v of %v\n",i+1,nq)
		fmt.Printf("%s",question.AsQuestion())
		//question.PrintAnswer()
		var ans string
		fmt.Printf("\nYour answer >")
		fmt.Scanf("%s",&ans)
		correct,valid:=question.CheckAnswer(ans)
		for !valid{
			fmt.Printf("Invalid answer. Please answer again\nYour answer >")
			fmt.Scanf("%s",&ans)
			correct,valid=question.CheckAnswer(ans)
		}
		if correct {
			fmt.Printf("CONGRATS!!!  Correct answer\n\t\t#------------------------------------------#\n\n")
			correctcount++
		} else {
			fmt.Printf("SORRY Wrong answer\n")
			fmt.Printf("Correct answer:%s\n",question.Answer())
			fmt.Printf("\t\t#------------------------------------------#\n\n")

		}

		

	}

	fmt.Printf("Your score is %v/%v\n",correctcount,nq)

	//fmt.Printf("%s\n",jsonresult)
	
}