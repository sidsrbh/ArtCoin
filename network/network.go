package network

import (
	"encoding/json"
	"fmt"
	"indicartcoin/database"
	"indicartcoin/sqldatabase"
	"indicartcoin/state"
	"indicartcoin/structs"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleConnections(w http.ResponseWriter, r *http.Request, state *state.State, vals []structs.Validator, IndicBlockchain *structs.Blockchain) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	for {
		var tx structs.Transaction
		err := ws.ReadJSON(&tx)
		if err != nil {
			log.Printf("Error: %v", err)
			_ = ws.WriteJSON(structs.ResponseMessage{Status: "error", Message: "Invalid JSON"})
			break
		}
		fmt.Println(tx.ArtOwnership.Status.String())
		valid, err := state.IsValidTransaction(tx)
		// Validate the transaction
		if !valid {
			log.Printf("Invalid transaction: %v", tx)
			log.Printf("error: %v", err.Error())
			_ = ws.WriteJSON(structs.ResponseMessage{Status: "error", Message: err.Error()})
			continue
		}
		//Add transaction to Database
		database.AddTransaction(tx, IndicBlockchain, vals)

		// Send success message
		_ = ws.WriteJSON(structs.ResponseMessage{Status: "success", Message: "Transaction added"})
	}
}

func GetBlockchainHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	blockchainData, err := json.Marshal(database.Blockchain.Blocks)
	if err != nil {
		http.Error(w, "Failed to serialize blockchain", http.StatusInternalServerError)
		return
	}

	w.Write(blockchainData)
}

func GetArtSummaryHandler(w http.ResponseWriter, r *http.Request) {
	start, _ := strconv.Atoi(r.URL.Query().Get("start"))
	count, _ := strconv.Atoi(r.URL.Query().Get("count"))
	w.Header().Set("Content-Type", "application/json")

	artSummary := sqldatabase.LoadArtOwnershipSummary(start, count)
	if artSummary == nil {
		http.Error(w, "Failed to load art summary", http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(artSummary)
	if err != nil {
		http.Error(w, "Failed to serialize art summary", http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}

type LikeResponse struct {
	Success bool `json:"success"`
}

func LikeArtHandler(w http.ResponseWriter, r *http.Request) {
	var response LikeResponse

	artID := r.URL.Query().Get("art_id")
	userID := r.URL.Query().Get("user_id") // Assume you get userID from session or JWT

	// First, check if the user has already liked this art piece
	liked, err := sqldatabase.AlreadyLiked(artID, userID)
	if err != nil {
		response.Success = false
		json.NewEncoder(w).Encode(response)
		return
	}

	// If not liked, then add a like and increment the ArtLikes counter
	if !liked {
		err := sqldatabase.AddLike(artID, userID)
		if err != nil {
			response.Success = false
			json.NewEncoder(w).Encode(response)
			return
		}

		err = sqldatabase.IncrementArtLike(artID)
		if err != nil {
			response.Success = false
			json.NewEncoder(w).Encode(response)
			return
		}

		response.Success = true
		json.NewEncoder(w).Encode(response)
	} else {
		response.Success = false
		json.NewEncoder(w).Encode(response)
	}
}

type HasUserLikedResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func HasUserLikedHandler(w http.ResponseWriter, r *http.Request) {
	artID := r.URL.Query().Get("art_id")
	userID := r.URL.Query().Get("user_id")

	var alreadyLiked bool

	// Assume db is your database connection
	alreadyLiked, _ = sqldatabase.AlreadyLiked(artID, userID)

	response := HasUserLikedResponse{
		Success: !alreadyLiked,
		Message: "Successfully checked like status",
	}

	if alreadyLiked {
		response.Message = "Art is already liked by this user"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
