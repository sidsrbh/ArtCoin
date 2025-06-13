package main

import (
	"encoding/json"
	"fmt"
	"indicartcoin/database"
	"indicartcoin/network"
	"indicartcoin/sqldatabase"
	"indicartcoin/structs"
	"indicartcoin/usercreator"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Define a struct for the response message
type ResponseMessage struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Define a struct for the incoming request body
type UpdateArtOwnershipRequest struct {
	ArtID        string               `json:"artID"`
	ArtOwnership structs.ArtOwnership `json:"artOwnership"`
}

func updateArtOwnershipHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var req UpdateArtOwnershipRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Call the UpdateArtOwnership function from the sqldatabase package
	sqldatabase.UpdateArtOwnership(req.ArtID, req.ArtOwnership)

	// Create a success response
	resp := ResponseMessage{
		Success: true,
		Message: "Art ownership updated successfully",
	}

	// Send the response back
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
func fetchData() {
	//Fetch users
	//fmt.Println("fetching user..")
	if users := sqldatabase.LoadUsers(); users != nil {
		usercreator.Database = users
		database.UserDatabase = users
	}
	//fmt.Println("fetched data..")

	//fmt.Println("fetchin balances..")
	// Fetch balances
	balances := sqldatabase.LoadBalances()
	if balances != nil {
		database.AppState.Balances = balances
	}
	//fmt.Println("balances fetched..")

	//fmt.Println("fetching art ownership..")
	// Fetch art ownership
	//artOwnership := sqldatabase.LoadArtOwnership()
	//if artOwnership != nil {
	//	database.AppState.ArtOwnership = artOwnership
	//}
	//fmt.Println("ownership fetched..")
	//

	//fmt.Println("fetching art summary..")
	artSummary := sqldatabase.LoadArtOwnershipSummary(0, 20)
	if artSummary != nil {
		database.ArtSummary = artSummary
	}

	//fmt.Println("summary fetched..")
	// Fetch validators

	//fmt.Println("fetching Validators..")
	validators := sqldatabase.LoadValidators()
	if validators != nil {
		database.Validators = validators
	}
	//fmt.Println("validators fetched..")
	// Fetch pending transactions
	//fmt.Println("fetching pending transactions..")
	pendingTransactions := sqldatabase.LoadPendingTransactions()
	if pendingTransactions != nil {
		database.PendingTransactions = pendingTransactions
	}
	//fmt.Println("Pending Transactions fetched..")
	Transactions := sqldatabase.LoadTransactions()
	if pendingTransactions != nil {
		database.Transactions = Transactions
	}
	//fmt.Println("Transaction fetched..")
	for _, Tx := range Transactions {
		Tx.ArtOwnership = database.AppState.ArtOwnership[Tx.ArtID]
	}
	// Fetch blocks
	Blocks, err := sqldatabase.LoadBlocks(nil)
	if err == nil && Blocks != nil {
		database.Blockchain.Blocks = Blocks
	}

}

func main() {
	err := sqldatabase.InitDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %s", err.Error())
		return
	}
	defer sqldatabase.CloseDatabase()
	fmt.Println("fetching data..")
	fetchData()
	fmt.Println("data fetche data..")
	// Fetch data at regular intervals
	timeInterval := 1 * time.Second // 10 seconds
	ticker := time.NewTicker(timeInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				fetchData()
			}
		}
	}()

	// Rest of your code
	// ...

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		network.HandleConnections(w, r, database.AppState, database.Validators, &database.Blockchain)
	})
	http.HandleFunc("/signup", usercreator.SignupHandler)
	http.HandleFunc("/login", usercreator.LoginHandler)
	http.HandleFunc("/get_blockchain", network.GetBlockchainHandler)
	http.HandleFunc("/get_art_summary", network.GetArtSummaryHandler)

	// New HTTP handler to get the current app state
	http.HandleFunc("/get_app_state", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(database.AppState)
	})

	// New HTTP handler to get the list of available validators
	http.HandleFunc("/get_validators", func(w http.ResponseWriter, r *http.Request) {
		validators := sqldatabase.LoadValidators()
		if validators == nil {
			json.NewEncoder(w).Encode([]structs.Validator{})
		} else {
			json.NewEncoder(w).Encode(validators)
		}
		//fmt.Println(validators)
	})

	http.HandleFunc("/updateArtOwnership", updateArtOwnershipHandler)

	http.HandleFunc("/art/like", network.LikeArtHandler)

	http.HandleFunc("/art/is_already_liked", network.HasUserLikedHandler)

	http.HandleFunc("/media/", func(w http.ResponseWriter, r *http.Request) {
		mediaID := r.URL.Path[len("/media/"):]

		mediaData, mediaType, err := sqldatabase.GetMediaData(mediaID)
		if err != nil {
			http.Error(w, "Media not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", mediaType)
		w.Write(mediaData)
	})

	http.HandleFunc("/fetch_art_by_owner", func(w http.ResponseWriter, r *http.Request) {
		artOwner := r.URL.Query().Get("artOwner")
		if artOwner == "" {
			http.Error(w, "ArtOwner parameter is required", http.StatusBadRequest)
			return
		}
		fmt.Println(artOwner)
		artOwnershipList, err := sqldatabase.FetchArtOwnershipByOwner(artOwner)
		if err != nil {
			http.Error(w, "Failed to fetch art ownership", http.StatusInternalServerError)
			return
		}
		fmt.Println("Art Count: ", len(artOwnershipList))
		// Convert the artOwnershipList to JSON
		jsonData, err := json.Marshal(artOwnershipList)
		if err != nil {
			http.Error(w, "Failed to convert art ownership to JSON", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	})

	type ValidatorSignupResponse struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	http.HandleFunc("/artOwner_by_Id", func(w http.ResponseWriter, r *http.Request) {
		// Get the artId from the query parameters
		artId := r.URL.Query().Get("artId")
		if artId == "" {
			http.Error(w, "artId parameter is required", http.StatusBadRequest)
			return
		}

		// Call the FetchArtOwnershipByArtID method to get ArtOwnership
		artOwnership, err := sqldatabase.FetchArtOwnershipByArtID(artId)
		if err != nil {
			http.Error(w, "Failed to fetch art ownership", http.StatusInternalServerError)
			log.Println("Error fetching art ownership by artId:", err)
			return
		}

		// Check if artOwnership is nil, which means no record was found
		if artOwnership == nil {
			http.Error(w, "No art ownership found for the given artId", http.StatusNotFound)
			return
		}

		// Convert the artOwnership to JSON
		jsonData, err := json.Marshal(artOwnership)
		if err != nil {
			http.Error(w, "Failed to convert art ownership to JSON", http.StatusInternalServerError)
			log.Println("Error converting art ownership to JSON:", err)
			return
		}

		// Set the Content-Type header and write the JSON response
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	})

	http.HandleFunc("/validator/signup", func(w http.ResponseWriter, r *http.Request) {
		var response ValidatorSignupResponse

		address := r.URL.Query().Get("address")
		stakeStr := r.URL.Query().Get("stake")
		stake, err := strconv.Atoi(stakeStr)
		if err != nil {
			response = ValidatorSignupResponse{
				Success: false,
				Message: "Invalid stake",
			}
			jsonResponse, _ := json.Marshal(response)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write(jsonResponse)
			return
		}

		newValidator := structs.Validator{
			Address: address,
			Stake:   float64(stake),
		}
		database.Validators = append(database.Validators, newValidator)
		sqldatabase.AddValidator(newValidator)

		response = ValidatorSignupResponse{
			Success: true,
			Message: "Validator signup successful",
		}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)
	})

	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatalf("Failed to start server: %s", err.Error())
		}
	}()

	fmt.Println("Server started at :8080")

	select {}
}
