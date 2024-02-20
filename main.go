package main

import (
	"BOOKING-APP/helper"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	conferenceTickets int    = 50
	conferenceName    string = "Go Conference"
	tableName                = "bookings"
)

var (
	remainingTickets uint = 50
	bookings         []UserData
	wg               sync.WaitGroup
	db               *sql.DB
	dbMutex          sync.Mutex
)

type UserData struct {
	firstName       string
	lastName        string
	email           string
	numberOfTickets uint
}

func main() {
	for {

		initializeDB()
		err := createTable(tableName) // create the table before using it
		if err != nil {
			log.Fatal("Error creating table:", err)
			return
		}

		//updateRemainingTickets()
		greetUsers()

		firstName, lastName, email, userTickets := getUserInput()
		isValidName, isValidEmail, isValidTicketNumber := helper.ValidateUserInput(firstName, lastName, email, userTickets, remainingTickets)

		if isValidName && isValidEmail && isValidTicketNumber {
			bookTicket(userTickets, firstName, lastName, email)
			wg.Add(1)
			go sendTicket(userTickets, firstName, lastName, email)

			//firstNames := getFirstNames()
			//fmt.Printf("The first names of bookings are %v\n", firstNames)
			// prompt user to continue after sendTicket function is executed

			if remainingTickets != 0 {
				wg.Wait()
				fmt.Printf("%v tickets remaining for %v \n", remainingTickets, conferenceName)
				fmt.Print("Do you want to continue? (yes/no): ")
				var choice string
				fmt.Scan(&choice)
				if choice != "yes" {
					wg.Wait()
					break
				}
			}
			if remainingTickets == 0 {
				wg.Wait()
				fmt.Printf("Since %v tickets remaining for %v \n", remainingTickets, conferenceName)
				fmt.Println("Our conference is booked out. Come back next year.")
				break
			}
			//wg.Wait()

		} else {
			printInputErrors(isValidName, isValidEmail, isValidTicketNumber)
		}
	}
	wg.Wait()

	// provided a command to delete all records in the table only for unit testing
	// to clean up huge chunk of data stored in table
	fmt.Print("Do you want to delete all records in the table? (yes/no): ")
	var deleteChoice string
	fmt.Scan(&deleteChoice)
	if deleteChoice == "yes" {
		deleteAllRecords()
	}

}

func initializeDB() {
	var err error
	db, err = sql.Open("sqlite3", "conference.db")
	if err != nil {
		log.Fatal("Error opening database:", err)
		return
	}
}

func createTable(tableName string) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			firstName TEXT,
			lastName TEXT,
			email TEXT,
			numberOfTickets INTEGER
		)`, tableName)

	_, err := db.Exec(query)
	return err
}

func greetUsers() {
	fmt.Printf("Hello, Welcome to %v booking application\n", conferenceName)
	fmt.Printf("We have a total of %v tickets and %v tickets are still available\n", conferenceTickets, remainingTickets)
	fmt.Println("Get your tickets here to attend")
}

func getUserInput() (string, string, string, uint) {
	var firstName, lastName, email string
	var userTickets uint

	fmt.Print("Enter your first name: ")
	fmt.Scan(&firstName)
	fmt.Print("Enter your last name: ")
	fmt.Scan(&lastName)
	fmt.Print("Enter your email address: ")
	fmt.Scan(&email)
	fmt.Print("Enter the number of tickets: ")
	_, err := fmt.Scan(&userTickets)
	if err != nil {
		fmt.Println("Error reading number of tickets:", err)
		return "", "", "", 0
	}

	return firstName, lastName, email, userTickets
}

func bookTicket(userTickets uint, firstName, lastName, email string) {
	dbMutex.Lock()

	remainingTickets = remainingTickets - userTickets
	userData := UserData{
		firstName:       firstName,
		lastName:        lastName,
		email:           email,
		numberOfTickets: userTickets,
	}

	_, err := db.Exec(`
		INSERT INTO bookings (firstName, lastName, email, numberOfTickets)
		VALUES (?, ?, ?, ?)
	`, firstName, lastName, email, userTickets)
	if err != nil {
		fmt.Println("Error inserting into database:", err)
		return
	}

	bookings = append(bookings, userData)
	fmt.Printf("List of bookings is %v\n", bookings)
	fmt.Printf("Thank you %v %v for booking %v tickets. You will receive a confirmation email at %v \n",
		firstName, lastName, userTickets, email)

	defer dbMutex.Unlock()

}

func sendTicket(userTickets uint, firstName, lastName, email string) {
	time.Sleep(5 * time.Second)
	ticket := fmt.Sprintf("%v tickets for %v %v", userTickets, firstName, lastName)
	fmt.Println("********************************************")
	fmt.Printf("Sending Ticket:\n%v \nto email address %v\n", ticket, email)
	fmt.Println("********************************************")
	wg.Done()
}

func printInputErrors(isValidName, isValidEmail, isValidTicketNumber bool) {
	if !isValidName {
		fmt.Println("First name or last name entered is too short")
	}
	if !isValidEmail {
		fmt.Println("Email address you entered doesn't contain @ sign ")
	}
	if !isValidTicketNumber {
		fmt.Println("Number of tickets you entered is invalid")
	}
}

func deleteAllRecords() {
	// delete all records from the table
	_, err := db.Exec("DELETE FROM " + tableName)
	if err != nil {
		fmt.Println("Error deleting records:", err)
	} else {
		fmt.Println("All records deleted successfully.")
	}
}
