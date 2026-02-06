package auth

import (
	db "forum/internal/db"
	utils "forum/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"strconv"
)

func SubmitRegister(w http.ResponseWriter, r *http.Request) {

	var input struct {
    Username        string `json:"username"`
    Age             int    `json:"age"`
    Gender          string `json:"gender"`
    FirstName       string `json:"first_name"`      
    LastName        string `json:"last_name"`      
    Email           string `json:"email"`
    Password        string `json:"password"`
    ConfirmPassword string `json:"confirmpassword"`
	}

	// Parse JSON
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		// fallback for HTML form submit
		input.Username = r.FormValue("username")
		fmt.Println(input.Username)
		input.Age, _ = strconv.Atoi(r.FormValue("age"))
		fmt.Println(input.Age)
		input.Gender = r.FormValue("gender")
		fmt.Println(input.Gender)
		input.FirstName = r.FormValue("firstname")
		fmt.Println(input.FirstName)
		input.LastName = r.FormValue("lastname")
		input.Email = r.FormValue("email")
		input.Password = r.FormValue("password")
		input.ConfirmPassword = r.FormValue("confirmpassword")
	}	
	
	if !utils.ValidUsername(input.Username) || !utils.ValidUsername(input.FirstName) || !utils.ValidUsername(input.LastName) {
		http.Error(w, `{"status":"error","message":"Invalid nickname or Name or last name"}`, http.StatusBadRequest)
		return
	}
	if !utils.ValidEmail(input.Email) {
		http.Error(w, `{"status":"error","message":"Invalid email"}`, http.StatusBadRequest)
		return
	}
	if !utils.ValidPassword(input.Password) {
		http.Error(w, `{"status":"error","message":"Invalid password"}`, http.StatusBadRequest)
		return
	}
	if input.Password != input.ConfirmPassword {
		http.Error(w, `{"status":"error","message":"Passwords do not match"}`, http.StatusBadRequest)
		return
	}
	// Validate gender - must be one of the allowed values
	validGenders := map[string]bool{"male": true, "female": true, "other": true}
	if input.Gender == "" || input.Gender == "Select gender" {
		http.Error(w, `{"status":"error","message":"Gender is required"}`, http.StatusBadRequest)
		return
	}

	if input.Age<= 13 || input.Age >= 120  {
		http.Error(w, `{"status":"error","message":"Age is required"}`, http.StatusBadRequest)
		return
	}
	if !validGenders[strings.ToLower(input.Gender)] {
		http.Error(w, `{"status":"error","message":"Invalid gender. Must be male, female, or other"}`, http.StatusBadRequest)
		return
	}
	// Normalize gender to lowercase for consistency
	input.Gender = strings.ToLower(input.Gender)

	// Hash password
	hash, err := utils.HashPassword(input.Password)
	if err != nil {
		http.Error(w, `{"status":"error","message":"Server error"}`, http.StatusInternalServerError)
		return
	}

	// Insert Into DB
	err = db.AddNewUser(
		input.Username,
		input.Email,
		hash,
		input.FirstName,
		input.LastName,
		input.Gender,
		input.Age,		 
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Error(w, `{"status":"error","message":"username or email already used"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"status":"error","message":"Server error"}`, http.StatusInternalServerError)
		return
	}

	// Success Response ...
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"status":  "ok",
		"message": "Registration successful! Please login.",
	})
}
