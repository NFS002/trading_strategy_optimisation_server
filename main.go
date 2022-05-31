package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func SHA256(text string) string {
	data := []byte(text)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

var (
	registeredVoters         = make(map[string]string)
	loginAttempts            = make(map[string][]int)
	sessionStore, sessionErr = NewRediStore(10, "tcp", ":6379", "", []byte("secret-key"))
)

type requestBody1x1 struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
}

type errorResponseBody1x1 struct {
	Error string `json:"error"`
}

func writeErrorResponse1x1(res http.ResponseWriter, errorName string, statusCode int) {
	errorResponse := errorResponseBody1x1{Error: errorName}
	res.WriteHeader(statusCode)
	jsonErr := json.NewEncoder(res).Encode(errorResponse)
	if jsonErr != nil {
		log.Warnf("%s caused %s when writing reponse", errorName, jsonErr)
	}
}

func writeErrorResponse1x2(res http.ResponseWriter, err error, statusCode int) {
	errorResponse := errorResponseBody1x1{Error: error.Error(err)}
	res.WriteHeader(statusCode)
	jsonErr := json.NewEncoder(res).Encode(errorResponse)
	if jsonErr != nil {
		log.Warnf("Failed login caused '%s' when writing reponse", jsonErr)
	}
}

func writeResponse1x1(res http.ResponseWriter, statusCode int) {
	res.WriteHeader(statusCode)
}

func writeResponse1x2(res http.ResponseWriter, statusCode int) {
	res.WriteHeader(statusCode)
}

func registerVoter(reqBody requestBody1x1) {
	fullName := reqBody.FirstName + " " + reqBody.LastName
	registeredVoters[fullName] = SHA256(reqBody.Password)
	loginAttempts[fullName] = []int{0}
}

func register(res http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	reqBody := requestBody1x1{}
	if jsonDecodeErr := decoder.Decode(&reqBody); jsonDecodeErr != nil {
		writeErrorResponse1x1(res, "Invalid JSON", 400)
	} else if userExists(reqBody) {
		writeErrorResponse1x1(res, "User already registered", 409)
	} else if !Verify(reqBody.Password) {
		writeErrorResponse1x1(res, "Invalid Password", 400)
	} else {
		registerVoter(reqBody)
		writeResponse1x2(res, 204)
	}
}

func tryLogin(req *http.Request, reqBody requestBody1x1) error {
	fullName := reqBody.FirstName + " " + reqBody.LastName
	savedPwdHash, exists := registeredVoters[fullName]
	receivedPwHash := SHA256(reqBody.Password)
	if !exists || savedPwdHash != receivedPwHash {
		return errors.New("user does not exist or password is incorrect")
	}
	session, err := sessionStore.Get(req, "my-cookie-store")
	if err != nil {
		return err
	}
	session.Values["authenticated"] = true
	return nil
}

func userExists(reqBody requestBody1x1) bool {
	fullName := reqBody.FirstName + " " + reqBody.LastName
	_, exists := registeredVoters[fullName]
	return exists
}

func login(res http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	reqBody := requestBody1x1{}
	if jsonDecodeErr := decoder.Decode(&reqBody); jsonDecodeErr != nil {
		writeErrorResponse1x1(res, "Invalid JSON", 400)
	} else {
		if err := tryLogin(req, reqBody); err != nil {
			writeErrorResponse1x2(res, err, 401)
		} else {
			writeResponse1x2(res, 204)
		}
	}
}

func home(res http.ResponseWriter, req *http.Request) {
	session, err := sessionStore.Get(req, "my-cookie-store")
	if err != nil {
		http.Error(res, "Forbidden", http.StatusForbidden)
	} else if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(res, "Forbidden", http.StatusForbidden)
	} else {
		_, _ = res.Write([]byte("Welcome home honey"))
	}
}

func main() {
	if sessionErr != nil {
		log.Fatalf("Could not connect to session store")
	} else {
		defer func(sessionStore *RediStore) {
			err := sessionStore.Close()
			if err != nil {
				log.Fatalf("Could not close session store")
			}
			router := mux.NewRouter()

			router.Path("/register").Methods("POST").HandlerFunc(register)

			router.Path("/login").Methods("POST").HandlerFunc(login)

			router.Path("/home").Methods("POST").HandlerFunc(home)

			port := ":8090"
			log.Info("Server listening on port ", port)
			if err := http.ListenAndServe(port, router); err != nil {
				log.Fatalf("Server failed to start: %v", err)
			}

		}(sessionStore)
	}
}
