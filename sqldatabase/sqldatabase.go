package sqldatabase

import (
	"database/sql"
	"fmt"
	"indicartcoin/structs"
	"log"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

var db *sql.DB
var dbMutex sync.Mutex

func InitDatabase() error {
	var err error
	//=======>IMPORTANT
	// Initialize the database connection. Replace with your own credentials.
	db, err = sql.Open("mysql", "your_username:your_password@tcp(your_db_host:3306)/your_database_name")
	if err != nil {
		return fmt.Errorf("Failed to open database: %v", err)
	}
	// Check if the database is accessible
	if err = db.Ping(); err != nil {
		return fmt.Errorf("Could not connect to the database: %v", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Minute * 10)

	return nil
}

func CloseDatabase() {
	db.Close()
}

func LoadValidators() []structs.Validator {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	rows, err := db.Query("SELECT address, stake FROM validators")
	if err != nil {
		log.Println("Error loading validators:", err)
		return nil
	}
	defer rows.Close()

	var vals []structs.Validator
	for rows.Next() {
		var val structs.Validator
		if err := rows.Scan(&val.Address, &val.Stake); err != nil {
			log.Println("Error scanning validator row:", err)
			continue
		}
		vals = append(vals, val)
		//fmt.Println(vals)
	}

	return vals
}

// AddValidator adds a new validator to the SQL database.
func AddValidator(val structs.Validator) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	_, err := db.Exec("INSERT INTO validators (address, stake) VALUES (?, ?)",
		string(val.Address), float64(val.Stake))
	if err != nil {
		log.Println("Error adding validator:", err)
	}
}

func DeleteValidator(address string) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	_, err := db.Exec("DELETE FROM validators WHERE address = ?", address)
	if err != nil {
		log.Println("Error deleting validator:", err)
	}
}

// LoadTransactions fetches all pending transactions from the SQL database.
func LoadTransactions() []structs.Transaction {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	rows, err := db.Query("SELECT id, type, ArtID, FromAddress, ToAddress, Amount, Fee, Signature, Status, block_index FROM transactions LIMIT 5")
	if err != nil {
		log.Println("Error loading transactions:", err)
		return nil
	}
	defer rows.Close()

	var txs []structs.Transaction
	for rows.Next() {
		var tx structs.Transaction
		var bockIndex = ""
		var status = ""
		if err := rows.Scan(&tx.TransactionId, &tx.Type, &tx.ArtID, &tx.From, &tx.To, &tx.Amount, &tx.Fee, &tx.Signature, &status, &bockIndex); err != nil {
			log.Println("Error scanning transaction row:", err)
			continue
		}
		if status == "Pending" {
			tx.Status = structs.Pending
		} else {
			tx.Status = structs.Completed
		}
		txs = append(txs, tx)
	}

	return txs
}

// AddTransaction adds a new transaction to the SQL database.
func AddTransaction(tx structs.Transaction, block_index int) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	_, err := db.Exec("INSERT INTO transactions (id, type, ArtID, FromAddress, ToAddress, Amount, Fee, Signature, Status, block_index) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		tx.TransactionId, tx.Type, tx.ArtID, tx.From, tx.To, tx.Amount, tx.Fee, tx.Signature, tx.Status.String(), block_index)
	if err != nil {
		log.Println("Error adding transaction:", err)
	}
}

func LoadPendingTransactions() []structs.Transaction {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	rows, err := db.Query("SELECT id, type, ArtID, FromAddress, ToAddress, Amount, Fee,Signature,Status FROM pending_transactions WHERE Status = 'Pending' LIMIT 5")
	if err != nil {
		log.Println("Error loading transactions:", err)
		return nil
	}
	defer rows.Close()

	var txs []structs.Transaction
	for rows.Next() {
		var tx structs.Transaction
		var status = ""
		if err := rows.Scan(&tx.TransactionId, &tx.Type, &tx.ArtID, &tx.From, &tx.To, &tx.Amount, &tx.Fee, &tx.Signature, &status); err != nil {
			log.Println("Error scanning transaction row:", err)
			continue
		}
		if status == "Pending" {
			tx.Status = structs.Pending
		} else {
			tx.Status = structs.Completed
		}
		txs = append(txs, tx)
	}

	return txs
}

// AddTransaction adds a new transaction to the SQL database.
func AddPendingTransaction(tx structs.Transaction) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	_, err := db.Exec("INSERT INTO pending_transactions (id, type, ArtID, FromAddress, ToAddress, Amount, Fee, Signature, Status, block_index) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		tx.TransactionId, tx.Type, tx.ArtID, tx.From, tx.To, tx.Amount, tx.Fee, tx.Signature, tx.Status.String(), nil)
	if err != nil {
		log.Println("Error adding transaction:", err)
	}
}

// DeletePendingTransaction deletes a pending transaction from the SQL database based on the transaction ID.
func DeletePendingTransaction(transactionId string) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	_, err := db.Exec("DELETE FROM pending_transactions WHERE id = ?", transactionId)
	if err != nil {
		log.Println("Error deleting pending transaction:", err)
	}
}

// UpdateBalance updates the balance for a given address in the SQL database.
func UpdateBalance(address string, balance float64) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	_, err := db.Exec("UPDATE balances SET balance=? WHERE address=?", balance, address)
	if err != nil {
		log.Println("Error updating balance:", err)
	}
}

// AddBalances inserts a new balance record for a given address in the SQL database.
func AddBalances(address string, balance float64) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	_, err := db.Exec("INSERT INTO balances (address, balance) VALUES (?, ?)", address, balance)
	if err != nil {
		log.Println("Error adding balance:", err)
	}
}

// LoadBalances fetches all balances from the SQL database and returns them as a map.
func LoadBalances() map[string]float64 {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	rows, err := db.Query("SELECT address, balance FROM balances")
	if err != nil {
		log.Println("Error loading balances:", err)
		return nil
	}
	defer rows.Close()

	balances := make(map[string]float64)
	for rows.Next() {
		var address string
		var balance float64
		if err := rows.Scan(&address, &balance); err != nil {
			log.Println("Error scanning balance row:", err)
			continue
		}
		balances[address] = balance
	}

	return balances
}

func UpdateArtOwnership(artID string, artOwnership structs.ArtOwnership) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	fmt.Println(artOwnership.Status)
	_, err := db.Exec("UPDATE art_ownership SET ArtOwner=?, Price=?, Description=?, Format=?, Art=?, RelatedImages=?, RelatedVideos=?, ArtName=?, ArtLikes=?, ForSale=?, Thumbnail=?, Status=? WHERE Id=?",
		artOwnership.ArtOwner, artOwnership.Price, artOwnership.Description, artOwnership.Format, artOwnership.Art, artOwnership.RelatedImages, artOwnership.RelatedVideos, artOwnership.ArtName, artOwnership.ArtLikes, artOwnership.ForSale, artOwnership.Thumbnail, artOwnership.Status.String(), artID)
	if err != nil {
		log.Println("Error updating art ownership:", err)
		return err
	}
	return nil
}

func FetchArtOwnershipByOwner(artOwner string) ([]structs.ArtOwnership, error) {
	dbMutex.Lock()
	fmt.Println(artOwner)
	defer dbMutex.Unlock()

	rows, err := db.Query("SELECT Id, ArtOwner, Price, Description, Format, Art, RelatedImages, RelatedVideos, ArtName, ArtLikes, ForSale, Thumbnail FROM art_ownership WHERE ArtOwner=?", artOwner)
	if err != nil {
		log.Println("Error fetching art ownership by owner:", err)
		return nil, err
	}
	defer rows.Close()

	var artOwnershipList []structs.ArtOwnership
	for rows.Next() {
		var artOwnership structs.ArtOwnership
		err := rows.Scan(&artOwnership.Id, &artOwnership.ArtOwner, &artOwnership.Price, &artOwnership.Description, &artOwnership.Format, &artOwnership.Art, &artOwnership.RelatedImages, &artOwnership.RelatedVideos, &artOwnership.ArtName, &artOwnership.ArtLikes, &artOwnership.ForSale, &artOwnership.Thumbnail)
		if err != nil {
			log.Println("Error scanning art ownership row:", err)
			continue
		}
		artOwnershipList = append(artOwnershipList, artOwnership)
		fmt.Println("Art ID: ", artOwnership.Id)
		fmt.Println("art owner: ", artOwner)
	}
	print("length:", len(artOwnershipList))
	return artOwnershipList, nil
}

func LoadArtOwnership() map[string]structs.ArtOwnership {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	rows, err := db.Query("SELECT Id, ArtOwner, Price, Description, Format, Art, RelatedImages, RelatedVideos, ArtName, ArtLikes, ForSale, Thumbnail, Status FROM art_ownership")
	if err != nil {
		log.Println("Error loading art ownership:", err)
		return nil
	}
	defer rows.Close()

	artOwnershipMap := make(map[string]structs.ArtOwnership)
	for rows.Next() {
		var artOwnership structs.ArtOwnership
		var status = ""
		if err := rows.Scan(&artOwnership.Id, &artOwnership.ArtOwner, &artOwnership.Price, &artOwnership.Description, &artOwnership.Format, &artOwnership.Art, &artOwnership.RelatedImages, &artOwnership.RelatedVideos, &artOwnership.ArtName, &artOwnership.ArtLikes, &artOwnership.ForSale, &artOwnership.Thumbnail, &status); err != nil {
			log.Println("Error scanning art ownership row:", err)
			continue
		}
		if status == "Pending" || status == "" {
			artOwnership.Status = structs.Pending
		} else {
			artOwnership.Status = structs.Completed
		}
		artOwnershipMap[artOwnership.Id] = artOwnership
	}

	return artOwnershipMap
}

func AddArtOwnership(artOwnership structs.ArtOwnership) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	_, err := db.Exec("INSERT INTO art_ownership (Id, ArtOwner, Price, Description, Format, Art, RelatedImages, RelatedVideos, ArtName, ArtLikes, ForSale, Thumbnail,Status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		artOwnership.Id, artOwnership.ArtOwner, artOwnership.Price, artOwnership.Description, artOwnership.Format, artOwnership.Art, artOwnership.RelatedImages, artOwnership.RelatedVideos, artOwnership.ArtName, artOwnership.ArtLikes, artOwnership.ForSale, artOwnership.Thumbnail, artOwnership.Status.String())
	if err != nil {
		log.Println("Error adding art ownership:", err)
	}
}
func LoadArtOwnershipSummary(start int, count int) map[string]structs.ArtOwnershipSummary {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	rows, err := db.Query("SELECT Id, Thumbnail, ArtLikes, ForSale, Price, Status FROM art_ownership LIMIT ?, ?", start, count)
	if err != nil {
		log.Println("Error loading art ownership summary:", err)
		return nil
	}
	defer rows.Close()

	artOwnershipSummaryMap := make(map[string]structs.ArtOwnershipSummary)

	for rows.Next() {
		var artID string
		var summary structs.ArtOwnershipSummary

		var status = ""
		if err := rows.Scan(&artID, &summary.Thumbnail, &summary.ArtLikes, &summary.ForSale, &summary.Price, &status); err != nil {
			log.Println("Error scanning art ownership summary row:", err)
			continue
		}
		if status == "Pending" || status == "" {
			summary.Status = structs.Pending
		} else {
			summary.Status = structs.Completed
		}
		summary.Id = artID
		artOwnershipSummaryMap[artID] = summary
	}

	return artOwnershipSummaryMap
}

func AddBlock(block *structs.Block) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	tx, err := db.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		return
	}

	_, err = tx.Exec("INSERT INTO blocks (block_index, timestamp, hash, prev_hash) VALUES (?, ?, ?, ?)",
		block.Index, block.Timestamp, block.Hash, block.PrevHash)
	if err != nil {
		log.Println("Error adding block:", err)
		tx.Rollback()
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println("Error committing transaction:", err)
		tx.Rollback()
	}
}

// LoadBlocks fetches blocks and their transactions from the SQL database starting from the given index and returns them as a slice.
// It fetches a maximum of 100 blocks at a time.
func LoadBlocks(startBlockIndex *int) ([]*structs.Block, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	var rows *sql.Rows
	var err error

	if startBlockIndex != nil {
		rows, err = db.Query("SELECT block_index, timestamp, hash, prev_hash FROM blocks WHERE index > ? LIMIT 100", *startBlockIndex)
	} else {
		rows, err = db.Query("SELECT block_index, timestamp, hash, prev_hash FROM blocks LIMIT 100")
	}

	if err != nil {
		log.Println("Error loading blocks:", err)
		return nil, err
	}
	defer rows.Close()

	var blocks []*structs.Block
	for rows.Next() {
		var block structs.Block
		if err := rows.Scan(&block.Index, &block.Timestamp, &block.Hash, &block.PrevHash); err != nil {
			log.Println("Error scanning block row:", err)
			continue
		}

		// Load transactions for this block
		txRows, err := db.Query("SELECT id, type, ArtID, FromAddress, ToAddress, Amount, Fee, Signature FROM transactions WHERE block_index = ?", block.Index)
		if err != nil {
			log.Println("Error loading transactions for block:", err)
			continue
		}

		var transactions []structs.Transaction
		for txRows.Next() {
			var tx structs.Transaction
			if err := txRows.Scan(&tx.ArtID, &tx.Type, &tx.ArtID, &tx.From, &tx.To, &tx.Amount, &tx.Fee, &tx.Signature); err != nil {
				log.Println("Error scanning transaction row:", err)
				continue
			}
			transactions = append(transactions, tx)
		}
		txRows.Close()

		block.Transactions = transactions
		blocks = append(blocks, &block)
	}

	return blocks, nil
}

// AddUser adds a new user to the SQL database.
func AddUser(username string, data []string) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	// Serialize the data array to a string, for example, by joining the elements with a delimiter
	serializedData := strings.Join(data, "SEPARATE")

	_, err := db.Exec("INSERT INTO users (username, data) VALUES (?, ?)", username, serializedData)
	if err != nil {
		log.Println("Error adding user:", err)
	}
}

// UpdateUser updates the data for a given username in the SQL database.
func UpdateUser(username string, data []string) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	// Serialize the data array to a string
	serializedData := strings.Join(data, ",")

	_, err := db.Exec("UPDATE users SET data=? WHERE username=?", serializedData, username)
	if err != nil {
		log.Println("Error updating user:", err)
	}
}

// LoadUsers fetches all user records from the SQL database and returns them as a map.
func LoadUsers() map[string][]string {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	rows, err := db.Query("SELECT username, data FROM users")
	if err != nil {
		log.Println("Error loading users:", err)
		return nil
	}
	defer rows.Close()

	userDatabase := make(map[string][]string)
	for rows.Next() {
		var username, serializedData string
		if err := rows.Scan(&username, &serializedData); err != nil {
			log.Println("Error scanning user row:", err)
			continue
		}

		// Deserialize the data string to an array
		data := strings.Split(serializedData, "SEPARATE")

		userDatabase[username] = data
	}

	return userDatabase
}

func AlreadyLiked(artID string, userID string) (bool, error) {
	var exists bool

	query := `SELECT EXISTS(SELECT 1 FROM art_likes WHERE art_id=? AND user_id=?)`
	err := db.QueryRow(query, artID, userID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func AddLike(artID string, userID string) error {
	// First, check if the user has already liked this art piece
	liked, err := AlreadyLiked(artID, userID)
	if err != nil {
		return err
	}

	// If not liked, then add a like
	if !liked {
		_, err := db.Exec("INSERT INTO art_likes (art_id, user_id) VALUES (?, ?)", artID, userID)
		if err != nil {
			return err
		}
		fmt.Println("Came Here, Done that..")
		if err != nil {
			return err
		}
	} else {
		// Optionally, you can return an error or a message saying "Already Liked"
		return fmt.Errorf("User has already liked this art")
	}

	return nil
}

func IncrementArtLike(artID string) error {
	query := "UPDATE art_ownership SET ArtLikes = ArtLikes + 1 WHERE Id = ?"
	_, err := db.Exec(query, artID)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func FetchArtOwnershipByArtID(artID string) (*structs.ArtOwnership, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	// Prepare the SQL query
	query := `SELECT Id, ArtOwner, Price, Description, Format, Art, RelatedImages, RelatedVideos, ArtName, ArtLikes, ForSale, Thumbnail, Status FROM art_ownership WHERE Id = ?`

	// Execute the query
	row := db.QueryRow(query, artID)

	// Scan the result into an ArtOwnership struct
	var artOwnership structs.ArtOwnership
	var status string
	err := row.Scan(&artOwnership.Id, &artOwnership.ArtOwner, &artOwnership.Price, &artOwnership.Description, &artOwnership.Format, &artOwnership.Art, &artOwnership.RelatedImages, &artOwnership.RelatedVideos, &artOwnership.ArtName, &artOwnership.ArtLikes, &artOwnership.ForSale, &artOwnership.Thumbnail, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			// No result found
			return nil, nil
		}
		log.Println("Error scanning art ownership row:", err)
		return nil, err
	}

	// Convert the status string to the corresponding enum value
	if status == "Pending" {
		artOwnership.Status = structs.Pending
	} else {
		artOwnership.Status = structs.Completed
	}

	return &artOwnership, nil
}

func GetMediaData(mediaID string) ([]byte, string, error) {
	var (
		data      []byte
		mediaType string
	)

	query := "SELECT data, media_type FROM media WHERE id = ?"
	row := db.QueryRow(query, mediaID)
	err := row.Scan(&data, &mediaType)

	return data, mediaType, err
}
