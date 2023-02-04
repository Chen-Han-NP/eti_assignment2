package main

import (
	//"encoding/json"

	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"google.golang.org/api/option"
)

// ==== STRUCTs ========
// struct for user login credentials
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Struct for User object
type User struct {
	UserID         string                       `json:"user_id" firestore:"UserID"`
	UserType       string                       `json:"user_type" firestore:"UserType"`
	Name           string                       `json:"name" firestore:"Name"`
	Email          string                       `json:"email" firestore:"Email"`
	Password       string                       `json:"password" firestore:"Password"`
	AreaOfInterest map[string][]string          `json:"area_of_interest" firestore:"AreaOfInterest"`
	School         string                       `json:"school,omitempty" firestore:"School,omitempty"`
	HourlyRate     int                          `json:"hourly_rate,omitempty" firestore:"HourlyRate,omitempty"`
	Availability   map[string]map[string]string `json:"availability,omitempty" firestore:"Availability,omitempty"`
	CertOfEvidence []string                     `json:"cert_of_evidence,omitempty" firestore:"CertOfEvidence,omitempty"`
}

// Create a struct that can be encoded into a JWT
// Please ONlY Store insensitive data
type Claims struct {
	EmailAddress string `json:"email_address"`
	UserType     string `json:"user_type"`
	UserID       string `json:"user_id"`
	jwt.RegisteredClaims
}

// ====== GLOBAL VARIABLES ========
var jwtKey = []byte("lhdrDMjhveyEVcvYFCgh1dBR2t7GM0YK") // A secure JWT Token for decoding, DO NOT SHARE

func verifyJWT(w http.ResponseWriter, r *http.Request) (Claims, error) {
	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			return Claims{}, err
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		return Claims{}, err
	}

	// Get the JWT string from the cookie
	tknStr := c.Value
	// Initialize a new instance of `Claims`
	claims := &Claims{}
	// Parse the JWT string and store the result in `claims`.
	// Note that we are passing the key in this method as well. This method will return an error
	// if the token is invalid (if it has expired according to the expiry time we set on sign in),
	// or if the signature does not match
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return *claims, err
		}
		w.WriteHeader(http.StatusBadRequest)
		return *claims, err
	}

	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return *claims, err
	}
	// Token is valid

	return *claims, nil
}

// ========= HANDLER FUNCTIONS ==========

// RETURN 200 -> Registered
// RETURN 406 -> Duplicated account (email)
func SignUp(w http.ResponseWriter, r *http.Request) {

	// POST http://localhost:5050/api/auth/signup/student
	// {"name": "xyz", "email": "..", "password", "area of interest": {"olevel":"..."...}, "certificate":[]}
	// POST http://localhost:5050/api/auth/signup/tutor
	//{"name": "xyz", "email": "..", "password", "area of interest": {"olevel":"..."...}, "certificate":"..."}

	params := mux.Vars(r)
	user_type := params["user_type"]

	ctx := context.Background()
	sa := option.WithCredentialsFile("/app/eti-assignment-2-firebase-adminsdk-6r9lk-85fb98eda4.json")

	// ---Authentication--
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		fmt.Printf("error initializing app: %v\n", err)
	}

	// Access auth service from the default app
	client, err := app.Auth(ctx)
	if err != nil {
		fmt.Printf("error getting Auth client: %v\n", err)
	}

	// ----Firestore----
	app2, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		fmt.Println(err.Error())
	}

	client2, err := app2.Firestore(ctx)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer client2.Close()

	// new user
	var user User

	// Check req methods
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	} else if r.Method == "POST" {
		fmt.Println(r.Body)
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			panic(err.Error())
		}

		// Step 1: Check if student or tutor
		if user_type == "student" {
			user.UserType = "Student"
		} else if user_type == "tutor" {
			user.UserType = "Tutor"
		} else {
			w.WriteHeader(http.StatusNotFound) // 404
			return
		}

		// ---- Create a new Auth user ----
		params := (&auth.UserToCreate{}).
			Email(user.Email).
			Password(user.Password)

		newUser, err := client.CreateUser(ctx, params)
		if err != nil {
			w.WriteHeader(http.StatusNotAcceptable)
			fmt.Println(err.Error())
			json.NewEncoder(w).Encode(err.Error())
			return
		}
		fmt.Println("New user with id: ", newUser.UID)
		user.UserID = newUser.UID

		// ---- Create a new Firebase record
		// Set UserID as the document name
		success, err := client2.Collection("User").Doc(user.UserID).Set(ctx, user)
		if err != nil {
			// Handle any errors in an appropriate way, such as returning them.
			fmt.Printf("An error has occurred: %s", err)
		}
		fmt.Println(success)
		json.NewEncoder(w).Encode(user)

	} else { // Other req methods
		w.WriteHeader(http.StatusNotFound) //404
		return
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	var UserID string
	var user User

	ctx := context.Background()
	sa := option.WithCredentialsFile("/app/eti-assignment-2-firebase-adminsdk-6r9lk-85fb98eda4.json")

	// ---Authentication--
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		fmt.Printf("error initializing app: %v\n", err)
	}

	// Access auth service from the default app
	client, err := app.Auth(ctx)
	if err != nil {
		fmt.Printf("error getting Auth client: %v\n", err)
	}

	app2, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		fmt.Println(err.Error())
	}

	client2, err := app2.Firestore(ctx)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer client2.Close()

	// Check req methods
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK) // 200
		return
	} else if r.Method == "POST" {
		// Receive user login information in JSON
		// and decode into User
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			// If the structure of the body is wrong, return an HTTP error
			w.WriteHeader(http.StatusBadRequest) //400
			return
		}

		// verify user email and password

		u, err := client.GetUserByEmail(ctx, creds.Email)
		if err != nil {
			fmt.Printf("Error getting user: %v\n", err)
			return
		}

		UserID = u.UID

		// Get user password and verify (For now, do this way)
		dsnap, err := client2.Collection("User").Doc(UserID).Get(ctx)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		dsnap.DataTo(&user)

		if creds.Password != user.Password {
			w.WriteHeader(http.StatusNotAcceptable) //406
			json.NewEncoder(w).Encode("Password not matched!")
			return
		}

		// Now set the JWT token and the cookie
		// Declare the expiration time of the token to 1hr

		expirationTime := time.Now().Add(12 * time.Hour)

		claims := &Claims{
			EmailAddress: creds.Email,
			UserType:     user.UserType,
			UserID:       user.UserID,
			RegisteredClaims: jwt.RegisteredClaims{
				// In JWT, the expiry time is expressed as unix milliseconds
				ExpiresAt: jwt.NewNumericDate(expirationTime),
			},
		}

		// Declare the token with the algorithm used for signing, and the claims
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		// Create the JWT string
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			// If there is an error in creating the JWT return an internal server error
			w.WriteHeader(http.StatusInternalServerError) //500
			return
		}

		// Finally, we set the client cookie for "token" as the JWT we just generated
		// we also set an expiry time which is the same as the token itself
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    tokenString,
			Expires:  expirationTime,
			Path:     "/",
			HttpOnly: true,
		})

		// Remove the password
		user.Password = ""

		// Return user object
		json.NewEncoder(w).Encode(user)
		return

	} else {
		w.WriteHeader(http.StatusNotFound)
		return
	}

}

// TEST - Check Cookie JWT token and return something
func Welcome(w http.ResponseWriter, r *http.Request) {
	claims, err := verifyJWT(w, r)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable) //406
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	// Finally, return the welcome message to the user, along with their
	// email given in the token
	w.Write([]byte(fmt.Sprintf("Welcome %s! /n Your user type is: %s, your user id is : %s",
		claims.EmailAddress, claims.UserType, claims.UserID)))

}

// CLEAR THE TOKEN COOKIE AND JWT
func Logout(w http.ResponseWriter, r *http.Request) {
	// immediately clear the token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Expires:  time.Now(),
		Path:     "/",
		HttpOnly: true,
	})
	fmt.Println(r.Header)
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/api/auth/signup/{user_type}", SignUp).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/auth/login", Login).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/auth/welcome", Welcome).Methods("GET")
	//router.HandleFunc("/api/auth/refresh", Refresh)
	router.HandleFunc("/api/auth/logout", Logout)

	fmt.Println("Listening at port 5050")
	log.Fatal(http.ListenAndServe(":5050", router))
}