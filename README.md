# Indicartcoin: A Simplified Blockchain for Digital Art Ownership

-----

Indicartcoin is a basic blockchain implementation designed to track and manage digital art ownership, transfers, and interactions (like liking art). It features a proof-of-stake-like validator selection, a secure user management system using RSA key pairs and AES encryption, and a robust SQL database backend for persistent storage.

## Table of Contents

  * [Features](https://www.google.com/search?q=%23features)
  * [How it Works](https://www.google.com/search?q=%23how-it-works)
      * [Blockchain & Transactions](https://www.google.com/search?q=%23blockchain--transactions)
      * [Validator & Consensus](https://www.google.com/search?q=%23validator--consensus)
      * [Art Ownership & Media](https://www.google.com/search?q=%23art-ownership--media)
      * [User Management & Security](https://www.google.com/search?q=%23user-management--security)
      * [Database Persistence](https://www.google.com/search?q=%23database-persistence)
  * [Project Structure](https://www.google.com/search?q=%23project-structure)
  * [API Endpoints](https://www.google.com/search?q=%23api-endpoints)
  * [Setup and Installation](https://www.google.com/search?q=%23setup-and-installation)
      * [Prerequisites](https://www.google.com/search?q=%23prerequisites)
      * [Database Setup](https://www.google.com/search?q=%23database-setup)
      * [Running the Application](https://www.google.com/search?q=%23running-the-application)
  * [Usage Examples](https://www.google.com/search?q=%23usage-examples)
      * [User Signup & Login](https://www.google.com/search?q=%23user-signup--login)
      * [Validator Signup](https://www.google.com/search?q=%23validator-signup)
      * [Getting Blockchain Data](https://www.google.com/search?q=%23getting-blockchain-data)
      * [Getting Art Summary](https://www.google.com/search?q=%23getting-art-summary)
      * [Art Ownership by Owner](https://www.google.com/search?q=%23art-ownership-by-owner)
      * [Liking Art](https://www.google.com/search?q=%23liking-art)
      * [Handling Transactions (Websockets)](https://www.google.com/search?q=%23handling-transactions-websockets)
  * [Troubleshooting](https://www.google.com/search?q=%23troubleshooting)
  * [Security Considerations](https://www.google.com/search?q=%23security-considerations)
  * [Contributing](https://www.google.com/search?q=%23contributing)
  * [License](https://www.google.com/search?q=%23license)

-----

## Features

  * **Custom Blockchain:** A simplified blockchain to record transactions.
  * **Proof-of-Stake (PoS) Validator Selection:** Validators are rewarded based on their stake and a randomized factor for block finalization.
  * **Digital Art Ownership Tracking:** Manages ownership, prices, descriptions, and media links for digital art.
  * **Art Liking System:** Users can "like" art pieces, incrementing a counter.
  * **User Management:** Secure user signup and login using RSA key pairs (2048-bit) and AES encryption for private keys.
  * **Transaction Types:** Supports `CoinTransfer`, `ArtUpload`, `ArtTransfer`, and `ArtUpdate` transactions.
  * **SQL Database Persistence:** Utilizes MySQL for persistent storage of blockchain data, user information, balances, art ownership, and more.
  * **RESTful API & WebSockets:** Provides HTTP endpoints for data retrieval and a WebSocket endpoint for submitting transactions.
  * **Periodic Data Fetching:** Automatically reloads critical application state (users, balances, validators, etc.) from the database at regular intervals.

-----

## How it Works

### Blockchain & Transactions

  * **Blocks:** Each block contains an `Index`, `Timestamp`, `Hash`, `PrevHash`, and a list of `Transactions`.
  * **Transaction Types:**
      * `CoinTransfer`: Standard transfer of Indicartcoin between users.
      * `ArtUpload`: Registers a new piece of art and its initial ownership on the blockchain.
      * `ArtTransfer`: Transfers ownership of an art piece from one user to another.
      * `ArtUpdate`: Allows updating details of an existing art piece.
  * **Transaction Processing:**
      * Transactions are initially added to a `PendingTransactions` pool.
      * When `MaxTransactionsPerBlock` (currently 5) pending transactions accumulate, a new block is created.
      * Blocks are added to the `Blockchain`, and transactions are "finalized" by applying their effects to the `AppState` (balances, art ownership) and moving them from pending to confirmed status in the SQL database.

### Validator & Consensus

  * **Validators:** Participants who stake Indicartcoin can become validators.
  * **Reward Distribution:** When a block is finalized, validators are rewarded based on their stake. The rewards are distributed using an exponential decay formula, favoring validators with higher stakes.
  * **Validator Removal:** Validators are removed from the active validator set after participating in block finalization (this might be a temporary or specific design choice for this simple implementation).

### Art Ownership & Media

  * **`ArtOwnership` Struct:** Stores details like `Id`, `ArtOwner`, `Price`, `Description`, `Format`, `Art` (media ID/URL), `RelatedImages`, `RelatedVideos`, `ArtName`, `ArtLikes`, `ForSale` status, and `Thumbnail`.
  * **Media Storage:** `Art` and `Thumbnail` fields likely store IDs that link to actual media data (bytes and media type) stored in a `media` table, accessible via the `/media/{mediaID}` endpoint.
  * **Liking Art:** Users can "like" art, which is recorded in the `art_likes` table and increments the `ArtLikes` counter in the `art_ownership` table.

### User Management & Security

  * **Key Pair Generation:** On signup, users generate a 2048-bit RSA private and public key pair.
  * **Private Key Encryption:** The private key is encrypted using AES (CBC mode with PKCS7 padding) with a user-provided passphrase (16, 24, or 32 bytes in length).
  * **Database Storage:** The encrypted private key and the public key are stored in the SQL database.
  * **Login:** Users can retrieve their encrypted private key and public key from the database by providing their username. The system attempts to decrypt the private key to verify the passphrase.
  * **Transaction Signature Verification:** Transactions are signed using the sender's private key and verified using their public key.

### Database Persistence

  * Uses `github.com/go-sql-driver/mysql` for connecting to a MySQL database.
  * **Tables:** The application interacts with tables like `users`, `balances`, `validators`, `pending_transactions`, `transactions`, `blocks`, `art_ownership`, `art_likes`, and `media`.
  * **Data Loading:** On startup and at regular intervals (1 second), the `fetchData()` function loads various application states from the SQL database into in-memory Go variables.

-----

## Project Structure

```
indicartcoin/
├── blockchain/        # Logic for blockchain operations (e.g., signature verification)
│   └── blockchain.go  # (Contains VerifySignature)
├── database/          # In-memory application state and core blockchain logic (e.g., AddTransaction, finalizeValidation)
│   └── database.go
├── network/           # HTTP handlers and WebSocket communication
│   └── network.go
├── sqldatabase/       # Database interaction logic (CRUD operations for all tables)
│   └── sqldatabase.go
├── state/             # Application state definition and transaction validation logic
│   └── state.go
├── structs/           # Go structs defining data models (Block, Transaction, ArtOwnership, etc.)
│   └── structs.go
├── usercreator/       # User signup, login, key generation, and encryption/decryption
│   └── usercreator.go
└── main.go            # Main entry point, HTTP server setup, and data fetching loop
```

-----

## API Endpoints

The server listens on `0.0.0.0:8080` (or a port specified by the `PORT` environment variable).

### HTTP Endpoints

  * **`/signup` (GET)**
      * **Description:** Registers a new user, generating an RSA key pair and encrypting the private key.
      * **Query Params:**
          * `username`: Desired username.
          * `passphrase`: A passphrase (16, 24, or 32 characters) for encrypting the private key.
      * **Response:** `{"message": "SignUp successful", "privateKey": "...", "publicKey": "..."}`
  * **`/login` (GET)**
      * **Description:** Authenticates a user and returns their encrypted private key and public key.
      * **Query Params:**
          * `username`: User's username.
          * `passphrase`: User's passphrase to verify.
      * **Response:** `{"message": "Login successful", "privateKey": "...", "publicKey": "..."}`
  * **`/get_blockchain` (GET)**
      * **Description:** Returns the entire blockchain.
      * **Response:** JSON array of `Block` objects.
  * **`/get_art_summary` (GET)**
      * **Description:** Fetches a summary of art ownership.
      * **Query Params:**
          * `start`: (int) Starting index for pagination.
          * `count`: (int) Number of records to fetch.
      * **Response:** JSON map of `ArtID` to `ArtOwnershipSummary`.
  * **`/get_app_state` (GET)**
      * **Description:** Returns the current in-memory application state (balances, art ownership).
      * **Response:** JSON representation of `database.AppState`.
  * **`/get_validators` (GET)**
      * **Description:** Returns the list of currently active validators.
      * **Response:** JSON array of `Validator` objects.
  * **`/updateArtOwnership` (POST)**
      * **Description:** Updates the details of an existing art ownership record.
      * **Request Body (JSON):**
        ```json
        {
            "artID": "string",
            "artOwnership": {
                "id": "string",
                "artOwner": "string",
                "price": 0.0,
                "description": "string",
                "format": "string",
                "art": "string",
                "relatedImages": [],
                "relatedVideos": [],
                "artName": "string",
                "artLikes": 0,
                "forSale": false,
                "thumbnail": "string",
                "status": 0
            }
        }
        ```
      * **Response:** `{"success": true, "message": "Art ownership updated successfully"}`
  * **`/art/like` (GET)**
      * **Description:** Allows a user to "like" a piece of art.
      * **Query Params:**
          * `art_id`: ID of the art to like.
          * `user_id`: Public key/address of the user liking the art.
      * **Response:** `{"success": true}` or `{"success": false}` if already liked or an error occurred.
  * **`/art/is_already_liked` (GET)**
      * **Description:** Checks if a user has already liked a specific art piece.
      * **Query Params:**
          * `art_id`: ID of the art.
          * `user_id`: Public key/address of the user.
      * **Response:** `{"success": true, "message": "Art is already liked by this user"}` or `{"success": false, "message": "Successfully checked like status"}`.
  * **`/media/{mediaID}` (GET)**
      * **Description:** Serves media data (e.g., images, videos) associated with art pieces.
      * **URL Path:** Replace `{mediaID}` with the actual ID of the media.
      * **Response:** Raw media data with appropriate `Content-Type` header.
  * **`/fetch_art_by_owner` (GET)**
      * **Description:** Fetches all art pieces owned by a specific user.
      * **Query Params:**
          * `artOwner`: The public key/address of the art owner.
      * **Response:** JSON array of `ArtOwnership` objects.
  * **`/artOwner_by_Id` (GET)**
      * **Description:** Fetches the `ArtOwnership` details for a specific `artId`.
      * **Query Params:**
          * `artId`: The ID of the art piece.
      * **Response:** JSON `ArtOwnership` object.
  * **`/validator/signup` (GET)**
      * **Description:** Allows a user to sign up as a validator.
      * **Query Params:**
          * `address`: The public key/address of the validator.
          * `stake`: (int) The amount of stake the validator pledges.
      * **Response:** `{"success": true, "message": "Validator signup successful"}`

### WebSocket Endpoint

  * **`/ws`**
      * **Description:** Used for submitting new transactions to the blockchain.
      * **Request (JSON):** A `Transaction` object.
      * **Response (JSON):** `{"Status": "success", "Message": "Transaction added"}` or `{"Status": "error", "Message": "..."}` if validation fails.

-----

## Setup and Installation

### Prerequisites

  * **Go (Golang)**: Version 1.18+ (for generics support).
  * **MySQL Server**: A running MySQL database instance.
  * **Git**: For cloning the repository.

### Database Setup

1.  **Create a MySQL Database:**
    You'll need a database named `anyname` (or adjust the connection string in `sqldatabase.go`).

2.  **Create Tables:**
    You'll need to create the following tables in your MySQL database. Replace `anyname` with your actual database name if different.

    **`users` table:**

    ```sql
    CREATE TABLE IF NOT EXISTS users (
        username VARCHAR(255) PRIMARY KEY,
        data TEXT NOT NULL
    );
    ```

    **`balances` table:**

    ```sql
    CREATE TABLE IF NOT EXISTS balances (
        address VARCHAR(255) PRIMARY KEY,
        balance DECIMAL(30, 10) NOT NULL
    );
    ```

    **`validators` table:**

    ```sql
    CREATE TABLE IF NOT EXISTS validators (
        address VARCHAR(255) PRIMARY KEY,
        stake DECIMAL(30, 10) NOT NULL
    );
    ```

    **`pending_transactions` table:**

    ```sql
    CREATE TABLE IF NOT EXISTS pending_transactions (
        id VARCHAR(255) PRIMARY KEY,
        type INT NOT NULL,
        ArtID VARCHAR(255),
        FromAddress VARCHAR(255) NOT NULL,
        ToAddress VARCHAR(255) NOT NULL,
        Amount DECIMAL(30, 10) NOT NULL,
        Fee DECIMAL(30, 10) NOT NULL,
        Signature TEXT NOT NULL,
        Status VARCHAR(50) NOT NULL,
        block_index INT
    );
    ```

    **`transactions` table:**

    ```sql
    CREATE TABLE IF NOT EXISTS transactions (
        id VARCHAR(255) PRIMARY KEY,
        type INT NOT NULL,
        ArtID VARCHAR(255),
        FromAddress VARCHAR(255) NOT NULL,
        ToAddress VARCHAR(255) NOT NULL,
        Amount DECIMAL(30, 10) NOT NULL,
        Fee DECIMAL(30, 10) NOT NULL,
        Signature TEXT NOT NULL,
        Status VARCHAR(50) NOT NULL,
        block_index INT
    );
    ```

    **`blocks` table:**

    ```sql
    CREATE TABLE IF NOT EXISTS blocks (
        block_index INT PRIMARY KEY,
        timestamp VARCHAR(255) NOT NULL,
        hash VARCHAR(255) NOT NULL,
        prev_hash VARCHAR(255)
    );
    ```

    **`art_ownership` table:**

    ```sql
    CREATE TABLE IF NOT EXISTS art_ownership (
        Id VARCHAR(255) PRIMARY KEY,
        ArtOwner VARCHAR(255) NOT NULL,
        Price DECIMAL(30, 10) NOT NULL,
        Description TEXT,
        Format VARCHAR(50),
        Art TEXT, -- Store media ID/URL
        RelatedImages TEXT, -- Store comma-separated IDs/URLs
        RelatedVideos TEXT, -- Store comma-separated IDs/URLs
        ArtName VARCHAR(255),
        ArtLikes INT DEFAULT 0,
        ForSale BOOLEAN,
        Thumbnail TEXT, -- Store media ID/URL
        Status VARCHAR(50) NOT NULL
    );
    ```

    **`art_likes` table:**

    ```sql
    CREATE TABLE IF NOT EXISTS art_likes (
        art_id VARCHAR(255) NOT NULL,
        user_id VARCHAR(255) NOT NULL,
        PRIMARY KEY (art_id, user_id)
    );
    ```

    **`media` table:** (If you store actual media data in the DB)

    ```sql
    CREATE TABLE IF NOT EXISTS media (
        id VARCHAR(255) PRIMARY KEY,
        data LONGBLOB NOT NULL,
        media_type VARCHAR(255) NOT NULL
    );
    ```

3.  **Update Database Credentials:**
    Edit `sqldatabase/sqldatabase.go` and replace the placeholder credentials in `InitDatabase()` with your MySQL username, password, and host.

    ```go
    // In sqldatabase/sqldatabase.go
    func InitDatabase() error {
        // Replace with your actual database credentials
        db, err = sql.Open("mysql", "your_username:your_password@tcp(your_db_host:3306)/your_database_name")
        // ...
    }
    ```

### Running the Application

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/sidsrbh/ArtCoin.git # Replace with your actual repo URL
    cd indicartcoin
    ```

2.  **Download Go Modules:**

    ```bash
    go mod tidy
    ```

3.  **Run the server:**

    ```bash
    go run main.go sqldatabase/sqldatabase.go blockchain/blockchain.go database/database.go network/network.go state/state.go structs/structs.go usercreator/usercreator.go validator/validator.go
    ```

    Alternatively, build and run the executable:

    ```bash
    go build -o indicartcoin .
    ./indicartcoin
    ```

    The server will start listening on port `8080`.

-----

## Usage Examples

Here are some examples of how to interact with the API endpoints. You can use tools like `curl`, Postman, or write a simple Go/Python client.

**Note:** Replace `localhost:8080` with your server's actual address if running remotely.

### User Signup & Login

**Signup:**

```bash
curl "http://localhost:8080/signup?username=alice&passphrase=averysecret12345"
```

**Login:**

```bash
curl "http://localhost:8080/login?username=alice&passphrase=averysecret12345"
```

### Validator Signup

```bash
# Use a public key obtained from a signup, or a generated one
curl "http://localhost:8080/validator/signup?address=YOUR_PUBLIC_KEY_HERE&stake=100"
```

### Getting Blockchain Data

```bash
curl http://localhost:8080/get_blockchain
```

### Getting Art Summary

```bash
curl "http://localhost:8080/get_art_summary?start=0&count=10"
```

### Art Ownership by Owner

```bash
# Use a public key as the artOwner
curl "http://localhost:8080/fetch_art_by_owner?artOwner=ART_OWNER_PUBLIC_KEY_HERE"
```

### Liking Art

```bash
# Use actual ArtID and UserID (public key)
curl "http://localhost:8080/art/like?art_id=SOME_ART_ID&user_id=USER_PUBLIC_KEY_HERE"
```

### Handling Transactions (Websockets)

You'll need a WebSocket client for this. Here's a basic Python example:

```python
import websocket
import json
import time

def on_message(ws, message):
    print(f"Received: {message}")

def on_error(ws, error):
    print(f"Error: {error}")

def on_close(ws, close_status_code, close_msg):
    print(f"Closed: {close_status_code}, {close_msg}")

def on_open(ws):
    print("Opened connection")
    # Example: Send an ArtUpload transaction
    tx = {
        "TransactionId": "tx-123456",
        "Type": 1, # ArtUpload
        "ArtID": "art-001",
        "From": "SENDER_PUBLIC_KEY", # Replace with actual public key
        "To": "SENDER_PUBLIC_KEY",   # For ArtUpload, From and To are usually the same
        "Amount": 0.0,
        "Fee": 0.0,
        "Signature": "YOUR_TX_SIGNATURE_HERE", # Replace with actual signature
        "ArtOwnership": {
            "id": "art-001",
            "artOwner": "SENDER_PUBLIC_KEY",
            "price": 10.5,
            "description": "A beautiful digital painting.",
            "format": "PNG",
            "art": "media-id-of-main-art",
            "relatedImages": ["media-id-related-img-1"],
            "relatedVideos": [],
            "artName": "Sunset Glory",
            "artLikes": 0,
            "forSale": True,
            "thumbnail": "media-id-of-thumbnail",
            "status": 0 # Pending
        },
        "Status": 0 # Pending
    }
    ws.send(json.dumps(tx))
    print("Sent transaction")

if __name__ == "__main__":
    websocket.enableTrace(True)
    ws = websocket.WebSocketApp("ws://localhost:8080/ws",
                                on_open=on_open,
                                on_message=on_message,
                                on_error=on_error,
                                on_close=on_close)
    ws.run_forever()
```

-----

## Troubleshooting

  * **`Failed to initialize database: ...`**:
      * Ensure your MySQL server is running and accessible from where you're running the Go application.
      * Double-check the database credentials (username, password, host, port, database name) in `sqldatabase/sqldatabase.go`.
      * Verify that all necessary tables are created in your MySQL database as per the [Database Setup](https://www.google.com/search?q=%23database-setup) section.
  * **`crypto/rsa: verification error`**: This typically means the signature is invalid, the transaction string used for signing doesn't match, or the public key is incorrect. Ensure the transaction is serialized identically for signing and verification, and the correct key pair is used.
  * **"invalid passphrase length"**: Your AES passphrase must be exactly 16, 24, or 32 bytes long.
  * **"Error decoding hex string"**: Occurs if the encrypted private key stored or provided is not a valid hexadecimal string.
  * **`websocket.Error: ...`**: Check your network connection and ensure the server is running on the correct port and accessible.

-----

## Security Considerations

  * **Basic Implementation:** This project is a simplified blockchain for learning purposes. It lacks many advanced security features of production-grade blockchains (e.g., robust peer-to-peer networking, complex consensus mechanisms, advanced cryptoeconomics, replay attack protection).
  * **SQL Injections:** While standard Go database/sql practices generally mitigate basic SQL injection, review all queries, especially those constructed with user input, for potential vulnerabilities.
  * **DoS Attacks:** The current WebSocket handler processes every incoming message immediately. In a real application, you'd need rate limiting, transaction mempools, and more sophisticated validation to prevent denial-of-service attacks.
  * **Private Key Handling:** The private key is encrypted but sent over the network (even if encrypted). In a production system, private keys should ideally never leave the client's device or be handled by a secure key management system.
  * **Validator Selection:** The current validator selection is very basic. A real PoS system would have more sophisticated mechanisms for stake weighting, randomness, validator penalties, and reward distribution.

-----

## Contributing

Contributions are highly welcome\! If you have suggestions, bug reports, or want to contribute code, please feel free to:

1.  Fork the repository.
2.  Create a new branch (`git checkout -b feature/your-feature`).
3.  Make your changes.
4.  Commit your changes (`git commit -m 'feat: Add new feature'`).
5.  Push to the branch (`git push origin feature/your-feature`).
6.  Open a Pull Request.

-----

## License

This project is open-sourced under the [MIT License](https://www.google.com/search?q=LICENSE).

-----