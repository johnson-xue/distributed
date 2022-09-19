package grades

import (
	"fmt"
	"sync"
)

type Student struct {
	ID        int
	FirstName string
	LastName  string
	Grades    []Grande
}

func (s Student) Average() float32 {
	var result float32
	for _, grade := range s.Grades {
		result += grade.Score
	}
	return result / float32(len(s.Grades))
}

type Students []Student

var (
	students     Students
	studentMutex sync.Mutex
)

func (ss Students) GetById(id int) (*Student, error) {
	for i := range ss {
		if ss[i].ID == id {
			return &ss[i], nil
		}
	}
	return nil, fmt.Errorf("Student with Id %d not found", id)
}

type GrandeType string

const (
	GrandeQuiz = GrandeType("Quiz")
	GrandeTest = GrandeType("Test")
	GrandeExam = GrandeType("Exam")
)

type Grande struct {
	Title string
	Type  GrandeType
	Score float32
}
