package quiz
import (
	"fmt"
	"net/http"
	"encoding/json"
	"github.com/manish119/chor/jobiface"
	"math/rand"
	"bytes"
	"os"
	"strings"
	"strconv"
	"html"
)

type QuizQuestion interface {
	/*Get question (with options for multiple) and answer 
	in the form readable to quiz client*/
	AsQuestion() string
	Answer() string
	/*Check answer and tell whether it is valid and
	whether it is right*/
	/*First return value is whether answer is corrent
	and second is whether it is valid
	First value is irrelevant if second is false*/
	CheckAnswer(ans string) (bool,bool)
}
/*Multiple choice question*/
type MultipleQuestion struct {
	Question string //The question
	Options []string //Options
	AnsInd int //index of correct answer in options
}
/*Yes no question*/
type BoolQuestion struct {
	Question string
	Ans  bool
}
/*Describes a job of converting map[string]interface{}
of a question to QuizQuestion type*/
type BuildQuestionJob map[string]interface{}


/*implement DoJob method for BuildQuestionJob to use
jobiface multitasking method*/
/*Basically the wrapper for makeQuestion function*/

func (qmap BuildQuestionJob) DoJob() interface{}{
	var qmap1 map[string]interface{} =qmap
	return makeQuestion(qmap1)

}

//Name of the job is <QuestionJob>
func (qmap BuildQuestionJob) Name() string{
	return "<QuestionJob>"
}

/*Print multiple choice questions with options*/
func (question *MultipleQuestion) AsQuestion() string{
	var bytebuf bytes.Buffer
	bytebuf.WriteString(fmt.Sprintf("%s\n",question.Question))
	for i,opt:=range question.Options{
		bytebuf.WriteString(fmt.Sprintf("%c: %s\n",i+97,opt))
	}
	return bytebuf.String()

}
/*Print a yes/no question*/
func (question *BoolQuestion) AsQuestion() string{
	return fmt.Sprintf("%s\n True/False Question\n",question.Question)
}

/*Print answer of MultipleQuestion*/
func (question *MultipleQuestion) Answer() string{
	return fmt.Sprintf("%c",question.AnsInd+97)
}
/*Print true/false answer of yes/no question*/
func (question *BoolQuestion) Answer() string{
	return fmt.Sprintf("%v",question.Ans)
}
/*Check answer for a multiple choice question
It should be of format a/b/c/d..(case insensitive)  */
func (question *MultipleQuestion) CheckAnswer(ans string) (bool,bool) {
	//Invalidate any answer except a,b,c,d 
	//(case insensitive) and not higher than
	//Options
	if len(ans)!=1{
		return false,false
	}
	ans=strings.ToLower(ans)

	if int(ans[0])<97 || int(ans[0])>96+len(question.Options) {
		return false,false
	}
	right_ans:=fmt.Sprintf("%c",97+question.AnsInd)

	return right_ans==ans,true
}

/*Check answer for boolean question, It can be either
true/false or t/f (case insensitive)
*/


func (question *BoolQuestion) CheckAnswer(ans string) (bool,bool) {
	//Invalidate any answer except t/f or true/false
	//case insensitive
	ans=strings.ToLower(ans)
	if ans!="true" && ans!="false" && ans!="t" && ans!="f"{
		return false,false

	}
	right_ans:=strconv.FormatBool(question.Ans)
	if right_ans==ans{
		return true,true
	} else {
		return ans==right_ans[0:1],true
	}

}

var urlfmt string="https://opentdb.com/api.php?amount=%v&type=%s"
const (
	buflen=1024
)

/*Retrieve a json object as string from 
opendb.com*/
func GetQuestionJSON(nquestion int,qt string) string{
	resp,err:=http.Get(fmt.Sprintf(urlfmt,nquestion,qt))
	if err!=nil {
		fmt.Fprintf(os.Stderr,"Error: %s\n",err.Error())
		return ""
	}
	defer resp.Body.Close()
	var bytebuf bytes.Buffer
	n,err:=bytebuf.ReadFrom(resp.Body)
	if n==0 {
		fmt.Fprintf(os.Stderr,"Error: cannot retrieve JSON data\n")
	}
	if (err!=nil){
		fmt.Fprintf(os.Stderr,"Error: %s\n",err.Error())
		return ""
	}

	return bytebuf.String()





}
/*Parse JSON into a slice of QuizQuestions (ptrs of
above two types of question) using thcnt goroutines*/
func BuildQuestions(jsonstr string,thcnt int) []QuizQuestion{
	jsonbytes:=[]byte(jsonstr)
	jsonobject:=make(map[string]interface{})
	err:=json.Unmarshal(jsonbytes,&jsonobject)
	if (err!=nil){
		fmt.Fprintf(os.Stderr,"Error: %s\n",err.Error())
		return nil
	}

	fmt.Printf("Response Code: %v\n",jsonobject["response_code"])
	questionar:=jsonobject["results"]
	questionarray,ok:=questionar.([]interface{})
	//Questions not an array
	if !ok {
		fmt.Fprintf(os.Stderr,"Error: cannot parse JSON questions\n")
		return nil
	}
	
	jobs:=make([]jobiface.DoableJob,0)
	for _,quest1:=range questionarray {
		//fmt.Printf("%+v\n",quest1)
		quest,ok:=quest1.(map[string]interface{})
		//Question not a map
		if !ok {
			fmt.Fprintf(os.Stderr,"Error: invalid question map format for '%+v'\n",quest1)
			continue
		}
		var questjob BuildQuestionJob =quest
		//Add &quest into an array of Jobs used to work
		jobs=append(jobs,&questjob)

		
	}
	/*Do all quesion building jobs from jobs
	using thcnt goroutines */

	jobstats:=jobiface.DoAllJobs(jobs,thcnt)
	questions:=make([]QuizQuestion,0)

	//Filter out the invalid questions or nil questions
	//Due to failure and add to questions
	for _,jobstat :=range jobstats{
		question1:=jobstat.Stat
		if question1==nil {
			continue
		}
		question,ok :=question1.(QuizQuestion)
		//question1 not a QuizQuestion
		//This should not happen but just in case
		if !ok {
			fmt.Fprintf(os.Stderr,"question at %v not valid QuizQuestion object\n",question1)
			continue

		}
		questions=append(questions,question)
	}


	//fmt.Printf("%+v\n",jsonobject)



 	return questions
}


//Build one QuizQuestion object from one map[string]interface{}

func makeQuestion(quest map[string]interface{}) QuizQuestion {
	
	quest_type,ok:=quest["type"].(string)

		//Question type not a string
	if !ok {
		fmt.Fprintf(os.Stderr,"Error: invalid type format '%v'\n",quest["type"])
		return nil
	}
	question,ok:=quest["question"].(string)
		//Question not a string
	if !ok {
		fmt.Fprintf(os.Stderr,"Error: invalid question format '%v'\n",quest["question"])
		return nil
	}
	question=html.UnescapeString(question)
		//Multiple choice question
	if quest_type=="multiple" {
			

		questionstruct:=MultipleQuestion{Question:question}
			//Make oprions
		options:=make([]string,0)
			
		wrongarray,ok:=quest["incorrect_answers"].([]interface{})
			//Wrong answers not an array
		if !ok {
			fmt.Fprintf(os.Stderr,"Error: Invalid wrong options format '%+v'\n",quest["incorrect_answers"])
				return nil
		}
		correct_ans,ok:=quest["correct_answer"].(string)
			//Correct answer not a string
		if !ok {
			fmt.Fprintf(os.Stderr,"Error: Invalid correct answer format '%v'\n",quest["correct_answer"])
				return nil
		}
		correct_ans=html.UnescapeString(correct_ans)
			//Choose a random position to insert the correct answer
		pos:=rand.Intn(len(wrongarray)+1)
		for i,option1:= range wrongarray {
				//If i equals chosen position pos enter answer there
			if i==pos {
				/*Actual position at which it is inserted is the
				current size of slice.
				Generally i and actualpos should be equal but
				If any previous wrong option is invalidated due 
				to parsing error, they may be different*/

				actualpos:=len(options)

				options=append(options,correct_ans)
				questionstruct.AnsInd=actualpos
			}
				//Append wrong answers to options
			option,ok:=option1.(string)
			//An option not a string
			if !ok {
				fmt.Fprintf(os.Stderr,"Error: Invalid wrong option format '%v'\n",option1)
				continue

			}
			option=html.UnescapeString(option)
			options=append(options,option)

		}
		//If chosen position is last
		if pos==len(wrongarray){
			actualpos:=len(options)
			options=append(options,correct_ans)
			questionstruct.AnsInd=actualpos

		}
		questionstruct.Options=options
		//(&questionstruct).PrintQuestion()
		//fmt.Printf("%v\n",questionstruct.AnsInd)
		return &questionstruct






	} else if quest_type=="boolean" {
		questionstruct:=BoolQuestion{Question:question}
		//fmt.Printf("%v\n",quest["correct_answer"])
		correct_ans,ok:=quest["correct_answer"].(string)
		//Correct answer not a string
		if !ok {
			fmt.Fprintf(os.Stderr,"Error: invalid correct_ans format '%v'\n",quest["correct_ans"])
			return nil

		}
		ans,err:=strconv.ParseBool(strings.ToLower(correct_ans))
		if err!=nil {
		fmt.Fprintf(os.Stderr,"Error: %s\n",err.Error())
			return nil
		}
		questionstruct.Ans=ans
			//(&questionstruct).PrintQuestion()
			//fmt.Printf("%v\n",ans)
		return &questionstruct




	} 
	fmt.Fprintf(os.Stderr,"Error: invalid question type '%v'\n",quest_type)
	return nil

	
}





