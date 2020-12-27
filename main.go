package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

var operations [4]string = [4]string{"GET", "SETX", "INCX", "DECX"}

type Reactor struct {
	Storage map[uuid.UUID]float64
}

type Message struct {
	Operation string
	Key       uuid.UUID
	Value     float64
}

type Response struct {
	Code  int
	Value float64
	Error string
}

func (m *Message) logMessage() {
	fmt.Println("*---**---*")
	fmt.Println("NEW MESSAGE")
	fmt.Println("Operation: ", m.Operation)
	fmt.Println("Key: ", m.Key)
	fmt.Println("Value: ", m.Value)
	fmt.Println("//------//")
}

func logError(s string) {
	fmt.Println("*---**---*")
	fmt.Println("ERROR: ", s)
	fmt.Println("//------//")
}

func readOperation(s string) (string, error) {
	operation := strings.ToUpper(s)
	for _, op := range operations {
		if operation == op {
			return operation, nil
		}
	}
	err := errors.New("Unknown Operation")
	return operation, err
}

func (r Reactor) readMesage(req http.Request) (Message, error) {
	q, err := url.ParseQuery(req.URL.RawQuery)

	if err != nil {
		logError(err.Error())
		return Message{}, err
	}

	UUID, err := uuid.Parse(q["key"][0])

	if err != nil {
		logError(err.Error())
		return Message{}, err
	}

	value, err := strconv.ParseFloat(q["value"][0], 64)

	if err != nil {
		logError(err.Error())
		return Message{}, err
	}

	operation, err := readOperation(q["operation"][0])

	if err != nil {
		logError(err.Error())
		return Message{}, err
	}

	msg := Message{
		Operation: operation,
		Key:       UUID,
		Value:     value,
	}

	msg.logMessage()
	return msg, nil
}

func (r *Reactor) processMessage(m Message) (float64, error) {
	switch m.Operation {
	case "GET":
		return r.Storage[m.Key], nil
	case "SETX":
		_, ok := r.Storage[m.Key]
		if ok == true {
			return 0, errors.New("Value already exists")
		}
		r.Storage[m.Key] = m.Value
		return m.Value, nil
	case "INCX":
		value, ok := r.Storage[m.Key]
		if ok == false {
			return 0, errors.New("Specified key is nil")
		}
		m.Value = value + 1
		r.Storage[m.Key] = m.Value
		return m.Value, nil
	case "DECX":
		value, ok := r.Storage[m.Key]
		if ok == false {
			return 0, errors.New("Specified key is nil")
		}
		m.Value = value - 1
		r.Storage[m.Key] = m.Value
		return m.Value, nil
	default:
		return 0, errors.New("Missing action for operation")
	}
}

func (r *Reactor) processRequest(req *http.Request) Response {
	response := Response{}

	msg, err := r.readMesage(*req)
	if err != nil {
		response.Code = 422
		response.Error = err.Error()
		return response
	}

	value, err := r.processMessage(msg)
	if err != nil {
		response.Code = 422
		response.Error = err.Error()
		return response
	}

	response.Code = 200
	response.Value = value
	return response
}

func (r *Reactor) handler(w http.ResponseWriter, req *http.Request) {
	response := r.processRequest(req)
	json_response, _ := json.Marshal(response)
	fmt.Fprintf(w, string(json_response))
}

func main() {
	reactor := Reactor{
		Storage: make(map[uuid.UUID]float64),
	}
	http.HandleFunc("/", reactor.handler)
	http.ListenAndServe(":3000", nil)
}
